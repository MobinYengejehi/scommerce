package scommerce

import "context"

var _ ProductManager[any] = &BuiltinProductManager[any]{}

type productManagerDatabase[AccountID comparable] interface {
	DBProductManager[AccountID]
	productCategoryDatabase[AccountID]
	productDatabase[AccountID]
	productItemDatabase[AccountID]
}

type BuiltinProductManager[AccountID comparable] struct {
	DB productManagerDatabase[AccountID]
	FS FileStorage
}

func NewBuiltinProductManager[AccountID comparable](db productManagerDatabase[AccountID], fs FileStorage) *BuiltinProductManager[AccountID] {
	return &BuiltinProductManager[AccountID]{
		DB: db,
		FS: fs,
	}
}

func (productManager *BuiltinProductManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (productManager *BuiltinProductManager[AccountID]) newProductCategory(ctx context.Context, id uint64, db productCategoryDatabase[AccountID], form *ProductCategoryForm[AccountID]) (*BuiltinProductCategory[AccountID], error) {
	cat := &BuiltinProductCategory[AccountID]{
		ProductCategoryForm: ProductCategoryForm[AccountID]{
			ID: id,
		},
		DB: db,
		FS: productManager.FS,
	}
	if err := cat.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := cat.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return cat, nil
}

func (productManager *BuiltinProductManager[AccountID]) GetProductCategories(ctx context.Context, categories []ProductCategory[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductCategory[AccountID], error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	catForms := make([]*ProductCategoryForm[AccountID], 0, cap(ids))
	ids, catForms, err = productManager.DB.GetProductCategories(ctx, ids, catForms, skip, limit, queueOrder, productManager.FS)
	if err != nil {
		return nil, err
	}
	cats := categories
	if cats == nil {
		cats = make([]ProductCategory[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		cat, err := productManager.newProductCategory(ctx, ids[i], productManager.DB, catForms[i])
		if err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

func (productManager *BuiltinProductManager[AccountID]) GetProductCategoryCount(ctx context.Context) (uint64, error) {
	return productManager.DB.GetProductCategoryCount(ctx)
}

func (productManager *BuiltinProductManager[AccountID]) Init(ctx context.Context) error {
	return productManager.DB.InitProductManager(ctx)
}

func (productManager *BuiltinProductManager[AccountID]) NewProductCategory(ctx context.Context, name string, parentCategory ProductCategory[AccountID]) (ProductCategory[AccountID], error) {
	var pid *uint64 = nil
	if parentCategory != nil {
		tpid, err := parentCategory.GetID(ctx)
		if err != nil {
			return nil, err
		}
		pid = &tpid
	}
	catForm := ProductCategoryForm[AccountID]{}
	cid, err := productManager.DB.NewProductCategory(ctx, name, pid, &catForm, productManager.FS)
	if err != nil {
		return nil, err
	}
	return productManager.newProductCategory(ctx, cid, productManager.DB, &catForm)
}

func (productManager *BuiltinProductManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (productManager *BuiltinProductManager[AccountID]) RemoveAllProductCategroies(ctx context.Context) error {
	return productManager.RemoveAllProductCategroies(ctx)
}

func (productManager *BuiltinProductManager[AccountID]) RemoveProductCategory(ctx context.Context, category ProductCategory[AccountID]) error {
	cid, err := category.GetID(ctx)
	if err != nil {
		return err
	}
	return productManager.DB.RemoveProductCategory(ctx, cid)
}

func (productManager *BuiltinProductManager[AccountID]) SearchForProductCategories(ctx context.Context, searchText string, deepSearch bool, categories []ProductCategory[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductCategory[AccountID], error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	catForms := make([]*ProductCategoryForm[AccountID], 0, cap(ids))
	ids, catForms, err = productManager.DB.SearchForProductCategories(ctx, searchText, deepSearch, ids, catForms, skip, limit, queueOrder, productManager.FS)
	if err != nil {
		return nil, err
	}
	cats := categories
	if cats == nil {
		cats = make([]ProductCategory[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		cat, err := productManager.newProductCategory(ctx, ids[i], productManager.DB, catForms[i])
		if err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

func (productManager *BuiltinProductManager[AccountID]) newProductItem(ctx context.Context, id uint64, db productItemDatabase[AccountID], form *ProductItemForm[AccountID]) (*BuiltinProductItem[AccountID], error) {
	item := &BuiltinProductItem[AccountID]{
		DB: db,
		ProductItemForm: ProductItemForm[AccountID]{
			ID: id,
		},
		FS: productManager.FS,
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

func (productManager *BuiltinProductManager[AccountID]) SearchForProductItems(ctx context.Context, searchText string, deepSearch bool, items []ProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder, product Product[AccountID], category ProductCategory[AccountID]) ([]ProductItem[AccountID], error) {
	var err error = nil
	var pid *uint64 = nil
	var cid *uint64 = nil
	if product != nil {
		tpid, err := product.GetID(ctx)
		if err != nil {
			return nil, err
		}
		pid = &tpid
	}
	if category != nil {
		tcid, err := category.GetID(ctx)
		if err != nil {
			return nil, err
		}
		cid = &tcid
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	itemForms := make([]*ProductItemForm[AccountID], 0, cap(ids))
	ids, itemForms, err = productManager.DB.SearchForProductItems(ctx, searchText, deepSearch, ids, itemForms, skip, limit, queueOrder, pid, cid, productManager.FS)
	if err != nil {
		return nil, err
	}
	itms := items
	if itms == nil {
		itms = make([]ProductItem[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		item, err := productManager.newProductItem(ctx, ids[i], productManager.DB, itemForms[i])
		if err != nil {
			return nil, err
		}
		itms = append(itms, item)
	}
	return itms, nil
}

func (productManager *BuiltinProductManager[AccountID]) newProduct(ctx context.Context, id uint64, db productDatabase[AccountID], form *ProductForm[AccountID]) (*BuiltinProduct[AccountID], error) {
	product := &BuiltinProduct[AccountID]{
		ProductForm: ProductForm[AccountID]{
			ID: id,
		},
		DB: db,
		FS: productManager.FS,
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

func (productManager *BuiltinProductManager[AccountID]) SearchForProducts(ctx context.Context, searchText string, deepSearch bool, products []Product[AccountID], skip int64, limit int64, queueOrder QueueOrder, category ProductCategory[AccountID]) ([]Product[AccountID], error) {
	var err error = nil
	var cid *uint64 = nil
	if category != nil {
		tcid, err := category.GetID(ctx)
		if err != nil {
			return nil, err
		}
		cid = &tcid
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	productForms := make([]*ProductForm[AccountID], 0, cap(ids))
	ids, productForms, err = productManager.DB.SearchForProducts(ctx, searchText, deepSearch, ids, productForms, skip, limit, queueOrder, cid, productManager.FS)
	if err != nil {
		return nil, err
	}
	procs := products
	if procs == nil {
		procs = make([]Product[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		proc, err := productManager.newProduct(ctx, ids[i], productManager.DB, productForms[i])
		if err != nil {
			return nil, err
		}
		procs = append(procs, proc)
	}
	return procs, nil
}

func (productManager *BuiltinProductManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProductManager[AccountID], error) {
	return productManager, nil
}
