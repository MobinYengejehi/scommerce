package scommerce

import (
	"context"
	"encoding/json"
	"io"
	"sync"
)

var _ Product[any] = &BuiltinProduct[any]{}

type productDatabase[AccountID comparable] interface {
	DBProductCategory[AccountID]
	DBProduct[AccountID]
	DBProductItem[AccountID]
	DBUserReview[AccountID]
}

type ProductForm[AccountID comparable] struct {
	ID               uint64                             `json:"id,omitempty"`
	Description      *string                            `json:"description,omitempty"`
	Images           *[]string                          `json:"images,omitempty"`
	Name             *string                            `json:"name,omitempty"`
	ProductCategory  *BuiltinProductCategory[AccountID] `json:"product_category,omitempty"`
	ProductItemCount *uint64                            `json:"product_item_count,omitempty"`
}

type BuiltinProduct[AccountID comparable] struct {
	ProductForm[AccountID]
	DB productDatabase[AccountID] `json:"-"`
	FS FileStorage                `json:"-"`
	MU sync.RWMutex               `json:"-"`
}

func (product *BuiltinProduct[AccountID]) newProductItem(ctx context.Context, id uint64, db productItemDatabase[AccountID], form *ProductItemForm[AccountID]) (*BuiltinProductItem[AccountID], error) {
	item := &BuiltinProductItem[AccountID]{
		DB: db,
		FS: product.FS,
		ProductItemForm: ProductItemForm[AccountID]{
			ID:      id,
			Product: product,
		},
	}
	if err := item.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := item.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (product *BuiltinProduct[AccountID]) AddProductItem(ctx context.Context, sku string, name string, price float64, quantity uint64, images []FileReader, attrs json.RawMessage) (ProductItem[AccountID], error) {
	var errRes error = nil
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

	id, err := product.GetID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	itemForm := ProductItemForm[AccountID]{}
	pid, err := product.DB.AddProductProductItem(ctx, &form, id, sku, name, price, quantity, tokens, attrs, &itemForm, product.FS)
	if err != nil {
		return nil, err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	item, err := product.newProductItem(ctx, pid, product.DB, &itemForm)
	if err != nil {
		errRes = joinErr(errRes, err)
		return nil, errRes
	}

	for i, token := range tokens {
		product.FS.Delete(ctx, token)
		file, err := product.FS.Open(ctx, token)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		if _, err := io.Copy(file, imgs[i]); err != nil {
			errRes = joinErr(errRes, err)
		}
		file.Close()
	}

	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductItemCount = nil

	return item, nil
}

func (product *BuiltinProduct[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (product *BuiltinProduct[AccountID]) GetDescription(ctx context.Context) (string, error) {
	product.MU.RLock()
	if product.Description != nil {
		defer product.MU.RUnlock()
		return *product.Description, nil
	}
	product.MU.RUnlock()
	id, err := product.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	desc, err := product.DB.GetProductDescription(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.Description = &desc
	return desc, nil
}

func (product *BuiltinProduct[AccountID]) GetID(ctx context.Context) (uint64, error) {
	product.MU.RLock()
	defer product.MU.RUnlock()
	return product.ID, nil
}

func (product *BuiltinProduct[AccountID]) GetImages(ctx context.Context) ([]FileReadCloser, error) {
	var imageTokens []string = nil

	product.MU.RLock()
	if product.Images != nil {
		imageTokens = *product.Images
		product.MU.RUnlock()
	} else {
		product.MU.RUnlock()
		var err error = nil
		id, err := product.GetID(ctx)
		if err != nil {
			return nil, err
		}
		form, err := product.ProductForm.Clone(ctx)
		if err != nil {
			return nil, err
		}
		imageTokens, err = product.DB.GetProductImages(ctx, &form, id)
		if err != nil {
			return nil, err
		}
		if err := product.ApplyFormObject(ctx, &form); err != nil {
			return nil, err
		}
		product.MU.Lock()
		product.Images = &imageTokens
		product.MU.Unlock()
	}

	files := make([]FileReadCloser, 0, len(imageTokens))
	for _, token := range imageTokens {
		file, err := product.FS.Open(ctx, token)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (product *BuiltinProduct[AccountID]) GetName(ctx context.Context) (string, error) {
	product.MU.RLock()
	if product.Name != nil {
		defer product.MU.RUnlock()
		return *product.Name, nil
	}
	product.MU.RUnlock()
	id, err := product.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := product.DB.GetProductName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.Name = &name
	return name, nil
}

func (product *BuiltinProduct[AccountID]) GetProductCategory(ctx context.Context) (ProductCategory[AccountID], error) {
	product.MU.RLock()
	if product.ProductCategory != nil {
		defer product.MU.RUnlock()
		return product.ProductCategory, nil
	}
	product.MU.RUnlock()
	id, err := product.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	catForm := ProductCategoryForm[AccountID]{}
	cid, err := product.DB.GetProductCategory(ctx, &form, id, &catForm, product.FS)
	if err != nil {
		return nil, err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	cat := &BuiltinProductCategory[AccountID]{
		DB: product.DB,
		FS: product.FS,
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
	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductCategory = cat
	return cat, nil
}

func (product *BuiltinProduct[AccountID]) GetProductItemCount(ctx context.Context) (uint64, error) {
	product.MU.RLock()
	if product.ProductItemCount != nil {
		defer product.MU.RUnlock()
		return *product.ProductItemCount, nil
	}
	product.MU.RUnlock()
	id, err := product.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := product.DB.GetProductProductItemCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductItemCount = &count
	return count, nil
}

func (product *BuiltinProduct[AccountID]) GetProductItems(ctx context.Context, items []ProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItem[AccountID], error) {
	var err error = nil
	id, err := product.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	itemForms := make([]*ProductItemForm[AccountID], 0, cap(ids))
	ids, itemForms, err = product.DB.GetProductProductItems(ctx, &form, id, ids, itemForms, skip, limit, queueOrder, product.FS)
	if err != nil {
		return nil, err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	itms := items
	if itms == nil {
		itms = make([]ProductItem[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		item, err := product.newProductItem(ctx, ids[i], product.DB, itemForms[i])
		if err != nil {
			return nil, err
		}
		itms = append(itms, item)
	}
	return itms, nil
}

func (product *BuiltinProduct[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (product *BuiltinProduct[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (product *BuiltinProduct[AccountID]) RemoveAllProductItems(ctx context.Context) error {
	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := product.DB.RemoveAllProductProductItems(ctx, &form, id); err != nil {
		return err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductItemCount = nil
	return nil
}

func (product *BuiltinProduct[AccountID]) RemoveProductItem(ctx context.Context, item ProductItem[AccountID]) error {
	itid, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := product.DB.RemoveProductProductItem(ctx, &form, id, itid); err != nil {
		return err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductItemCount = nil
	return nil
}

func (product *BuiltinProduct[AccountID]) SetDescription(ctx context.Context, desc string) error {
	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := product.DB.SetProductDescription(ctx, &form, id, desc); err != nil {
		return err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.Description = &desc
	return nil
}

func (product *BuiltinProduct[AccountID]) SetImages(ctx context.Context, images []FileReader) error {
	var errRes error = nil

	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}

	tokens := make([]string, 0, len(images))
	imgs := make([]FileReader, 0, len(images))
	for _, image := range images {
		token, err := image.GetToken(ctx)
		if err != nil {
			errRes = joinErr(errRes, err)
		}
		tokens = append(tokens, token)
		imgs = append(imgs, image)
	}

	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}

	err = product.DB.SetProductImages(ctx, &form, id, tokens)
	if err != nil {
		errRes = joinErr(errRes, err)
		return errRes
	}

	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	product.MU.Lock()
	product.Images = &tokens
	product.MU.Unlock()

	for i, token := range tokens {
		product.FS.Delete(ctx, token)
		file, err := product.FS.Create(ctx, token)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		if _, err := io.Copy(file, imgs[i]); err != nil {
			errRes = joinErr(errRes, err)
		}
		file.Close()
	}

	return errRes
}

func (product *BuiltinProduct[AccountID]) SetName(ctx context.Context, name string) error {
	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := product.DB.SetProductName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.Name = &name
	return nil
}

func (product *BuiltinProduct[AccountID]) SetProductCategory(ctx context.Context, category ProductCategory[AccountID]) error {
	var cid *uint64 = nil
	if category != nil {
		tcid, err := category.GetID(ctx)
		if err != nil {
			return err
		}
		cid = &tcid
	}
	id, err := product.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := product.ProductForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := product.DB.SetProductCategory(ctx, &form, id, cid, product.FS); err != nil {
		return err
	}
	if err := product.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	product.MU.Lock()
	defer product.MU.Unlock()
	product.ProductCategory, err = category.ToBuiltinObject(ctx)
	return err
}

func (product *BuiltinProduct[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProduct[AccountID], error) {
	return product, nil
}

func (product *BuiltinProduct[AccountID]) ToFormObject(ctx context.Context) (*ProductForm[AccountID], error) {
	product.MU.RLock()
	defer product.MU.RUnlock()
	return &product.ProductForm, nil
}

func (product *BuiltinProduct[AccountID]) ApplyFormObject(ctx context.Context, form *ProductForm[AccountID]) error {
	product.MU.Lock()
	defer product.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		product.ID = form.ID
	}
	if form.Description != nil {
		product.Description = form.Description
	}
	if form.Images != nil {
		product.Images = form.Images
	}
	if form.Name != nil {
		product.Name = form.Name
	}
	if form.ProductCategory != nil {
		product.ProductCategory = form.ProductCategory
	}
	if form.ProductItemCount != nil {
		product.ProductItemCount = form.ProductItemCount
	}
	return nil
}

func (form *ProductForm[AccountID]) Clone(ctx context.Context) (ProductForm[AccountID], error) {
	var cloned ProductForm[AccountID] = *form
	return cloned, nil
}
