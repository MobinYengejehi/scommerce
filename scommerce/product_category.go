package scommerce

import (
	"context"
	"io"
	"sync"
)

var _ ProductCategory[any] = &BuiltinProductCategory[any]{}

type productCategoryDatabase[AccountID comparable] interface {
	DBProductCategory[AccountID]
	DBProduct[AccountID]
	DBProductItem[AccountID]
	DBUserReview[AccountID]
}

type ProductCategoryForm[AccountID comparable] struct {
	ID                    uint64                             `json:"id"`
	Name                  *string                            `json:"name,omitempty"`
	ParentProductCategory *BuiltinProductCategory[AccountID] `json:"parent_product_category,omitempty"`
	ProductCount          *uint64                            `json:"product_count,omitempty"`
}

type BuiltinProductCategory[AccountID comparable] struct {
	ProductCategoryForm[AccountID]
	DB productCategoryDatabase[AccountID] `json:"-"`
	FS FileStorage                        `json:"-"`
	MU sync.RWMutex                       `json:"-"`
}

func (category *BuiltinProductCategory[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (category *BuiltinProductCategory[AccountID]) GetID(ctx context.Context) (uint64, error) {
	category.MU.RLock()
	defer category.MU.RUnlock()
	return category.ID, nil
}

func (category *BuiltinProductCategory[AccountID]) GetName(ctx context.Context) (string, error) {
	category.MU.RLock()
	if category.Name != nil {
		defer category.MU.RUnlock()
		return *category.Name, nil
	}
	category.MU.RUnlock()
	id, err := category.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := category.DB.GetProductCategoryName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.Name = &name
	return name, nil
}

func (category *BuiltinProductCategory[AccountID]) GetParentProductCategory(ctx context.Context) (ProductCategory[AccountID], error) {
	category.MU.RLock()
	if category.ParentProductCategory != nil {
		defer category.MU.RUnlock()
		return category.ParentProductCategory, nil
	}
	category.MU.RUnlock()
	id, err := category.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	catForm := ProductCategoryForm[AccountID]{}
	cid, err := category.DB.GetProductCategoryParent(ctx, &form, id, &catForm, category.FS)
	if err != nil {
		return nil, err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	cat := &BuiltinProductCategory[AccountID]{
		DB: category.DB,
		ProductCategoryForm: ProductCategoryForm[AccountID]{
			ID: cid,
		},
	}
	if err := cat.Init(ctx); err != nil {
		return nil, err
	}
	if err := cat.ApplyFormObject(ctx, &catForm); err != nil {
		return nil, err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ParentProductCategory = cat
	return cat, nil
}

func (category *BuiltinProductCategory[AccountID]) GetProductCount(ctx context.Context) (uint64, error) {
	category.MU.RLock()
	if category.ProductCount != nil {
		defer category.MU.RUnlock()
		return *category.ProductCount, nil
	}
	category.MU.RUnlock()
	id, err := category.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := category.DB.GetProductCategoryProductCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ProductCount = &count
	return count, nil
}

func (category *BuiltinProductCategory[AccountID]) newProduct(ctx context.Context, pid uint64, db productDatabase[AccountID], form *ProductForm[AccountID]) (*BuiltinProduct[AccountID], error) {
	product := &BuiltinProduct[AccountID]{
		DB: db,
		FS: category.FS,
		ProductForm: ProductForm[AccountID]{
			ID:              pid,
			ProductCategory: category,
		},
	}
	if err := product.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := product.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return product, nil
}

func (category *BuiltinProductCategory[AccountID]) GetProducts(ctx context.Context, products []Product[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]Product[AccountID], error) {
	var err error = nil
	id, err := category.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	productForms := make([]*ProductForm[AccountID], 0, cap(ids))
	ids, productForms, err = category.DB.GetProductCategoryProducts(ctx, &form, id, ids, productForms, skip, limit, queueOrder, category.FS)
	if err != nil {
		return nil, err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	procs := products
	if procs == nil {
		procs = make([]Product[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		proc, err := category.newProduct(ctx, ids[i], category.DB, productForms[i])
		if err != nil {
			return nil, err
		}
		procs = append(procs, proc)
	}
	return procs, nil
}

func (category *BuiltinProductCategory[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (category *BuiltinProductCategory[AccountID]) NewProduct(ctx context.Context, name string, description string, images []FileReader) (Product[AccountID], error) {
	var errRes error = nil
	id, err := category.GetID(ctx)
	if err != nil {
		return nil, err
	}
	tokens := make([]string, 0, len(images))
	imgs := make([]FileReader, 0, len(images))
	for _, image := range images {
		token, err := image.GetToken(ctx)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		tokens = append(tokens, token)
		imgs = append(imgs, image)
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	productForm := ProductForm[AccountID]{}
	pid, err := category.DB.NewProductCategoryProduct(ctx, &form, id, name, description, tokens, &productForm, category.FS)
	if err != nil {
		errRes = joinErr(errRes, err)
		return nil, errRes
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	proc, err := category.newProduct(ctx, pid, category.DB, &productForm)
	if err != nil {
		errRes = joinErr(errRes, err)
		return nil, errRes
	}
	for i, token := range tokens {
		category.FS.Delete(ctx, token)
		file, err := category.FS.Create(ctx, token)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		if _, err := io.Copy(file, imgs[i]); err != nil {
			errRes = joinErr(errRes, err)
		}
		file.Close()
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ProductCount = nil
	return proc, errRes
}

func (category *BuiltinProductCategory[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (category *BuiltinProductCategory[AccountID]) RemoveAllProducts(ctx context.Context) error {
	id, err := category.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := category.DB.RemoveAllProducts(ctx, &form, id); err != nil {
		return err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ProductCount = nil
	return nil
}

func (category *BuiltinProductCategory[AccountID]) RemoveProduct(ctx context.Context, product Product[AccountID]) error {
	pid, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := category.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := category.DB.RemoveProduct(ctx, &form, id, pid); err != nil {
		return err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ProductCount = nil
	return nil
}

func (category *BuiltinProductCategory[AccountID]) SetName(ctx context.Context, name string) error {
	id, err := category.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := category.DB.SetProductCategoryName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.Name = &name
	return nil
}

func (category *BuiltinProductCategory[AccountID]) SetParentProductCategory(ctx context.Context, parent ProductCategory[AccountID]) error {
	var cid *uint64 = nil
	if parent != nil {
		tcid, err := parent.GetID(ctx)
		if err != nil {
			return err
		}
		cid = &tcid
	}
	id, err := category.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := category.ProductCategoryForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := category.DB.SetProductCategoryParent(ctx, &form, id, cid, category.FS); err != nil {
		return err
	}
	if err := category.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	pcat, err := parent.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	category.MU.Lock()
	defer category.MU.Unlock()
	category.ParentProductCategory = pcat
	return nil
}

func (category *BuiltinProductCategory[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProductCategory[AccountID], error) {
	return category, nil
}

func (category *BuiltinProductCategory[AccountID]) ToFormObject(ctx context.Context) (*ProductCategoryForm[AccountID], error) {
	category.MU.RLock()
	defer category.MU.RUnlock()
	return &category.ProductCategoryForm, nil
}

func (category *BuiltinProductCategory[AccountID]) ApplyFormObject(ctx context.Context, form *ProductCategoryForm[AccountID]) error {
	category.MU.Lock()
	defer category.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		category.ID = form.ID
	}
	if form.Name != nil {
		category.Name = form.Name
	}
	if form.ParentProductCategory != nil {
		category.ParentProductCategory = form.ParentProductCategory
	}
	if form.ProductCount != nil {
		category.ProductCount = form.ProductCount
	}
	return nil
}

func (form *ProductCategoryForm[AccountID]) Clone(ctx context.Context) (ProductCategoryForm[AccountID], error) {
	var cloned ProductCategoryForm[AccountID] = *form
	return cloned, nil
}
