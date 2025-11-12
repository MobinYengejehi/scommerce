package scommerce

import (
	"context"
	"encoding/json"
	"io"
	"sync"
)

var _ ProductItem[any] = &BuiltinProductItem[any]{}

type productItemDatabase[AccountID comparable] interface {
	DBProductCategory[AccountID]
	DBProductItem[AccountID]
	DBProduct[AccountID]
	DBUserReview[AccountID]
}

type ProductItemForm[AccountID comparable] struct {
	ID              uint64                     `json:"id"`
	Attributes      *json.RawMessage           `json:"attributes,omitempty"`
	Images          *[]string                  `json:"images,omitempty"`
	Price           *float64                   `json:"price,omitempty"`
	Product         *BuiltinProduct[AccountID] `json:"product,omitempty"`
	QuantityInStock *uint64                    `json:"quantity_in_stock,omitempty"`
	Name            *string                    `json:"name,omitempty"`
	SKU             *string                    `json:"sku,omitempty"`
}

type BuiltinProductItem[AccountID comparable] struct {
	ProductItemForm[AccountID]
	DB productItemDatabase[AccountID] `json:"-"`
	FS FileStorage                    `json:"-"`
	MU sync.RWMutex                   `json:"-"`
}

func (item *BuiltinProductItem[AccountID]) AddQuantityInStock(ctx context.Context, delta int64) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.AddProductItemQuantityInStock(ctx, &form, id, delta); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.QuantityInStock = nil
	return nil
}

func (item *BuiltinProductItem[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (item *BuiltinProductItem[AccountID]) GetAttributes(ctx context.Context) (json.RawMessage, error) {
	item.MU.RLock()
	if item.Attributes != nil {
		defer item.MU.RUnlock()
		return *item.Attributes, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	attrs, err := item.DB.GetProductItemAttributes(ctx, &form, id)
	if err != nil {
		return nil, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Attributes = &attrs
	return attrs, nil
}

func (item *BuiltinProductItem[AccountID]) GetID(ctx context.Context) (uint64, error) {
	item.MU.RLock()
	defer item.MU.RUnlock()
	return item.ID, nil
}

func (item *BuiltinProductItem[AccountID]) GetImages(ctx context.Context) ([]FileReadCloser, error) {
	var imageTokens []string = nil
	item.MU.RLock()
	if item.Images != nil {
		imageTokens = *item.Images
		item.MU.RUnlock()
	} else {
		item.MU.RUnlock()
		var err error = nil
		id, err := item.GetID(ctx)
		if err != nil {
			return nil, err
		}
		form, err := item.ProductItemForm.Clone(ctx)
		if err != nil {
			return nil, err
		}
		imageTokens, err = item.DB.GetProductItemImages(ctx, &form, id)
		if err != nil {
			return nil, err
		}
		if err := item.ApplyFormObject(ctx, &form); err != nil {
			return nil, err
		}
		item.MU.Lock()
		item.Images = &imageTokens
		item.MU.Unlock()
	}

	files := make([]FileReadCloser, 0, len(imageTokens))
	for _, token := range imageTokens {
		file, err := item.FS.Open(ctx, token)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (item *BuiltinProductItem[AccountID]) GetPrice(ctx context.Context) (float64, error) {
	item.MU.RLock()
	if item.Price != nil {
		defer item.MU.RUnlock()
		return *item.Price, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	price, err := item.DB.GetProductItemPrice(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Price = &price
	return price, nil
}

func (item *BuiltinProductItem[AccountID]) GetProduct(ctx context.Context) (Product[AccountID], error) {
	item.MU.RLock()
	if item.Product != nil {
		defer item.MU.RUnlock()
		return item.Product, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	productForm := ProductForm[AccountID]{}
	pid, err := item.DB.GetProductItemProduct(ctx, &form, id, &productForm, item.FS)
	if err != nil {
		return nil, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	product := &BuiltinProduct[AccountID]{
		DB: item.DB,
		ProductForm: ProductForm[AccountID]{
			ID: pid,
		},
	}
	if err := product.Init(ctx); err != nil {
		return nil, err
	}
	if err := product.ApplyFormObject(ctx, &productForm); err != nil {
		return nil, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Product = product
	return product, nil
}

func (item *BuiltinProductItem[AccountID]) GetQuantityInStock(ctx context.Context) (uint64, error) {
	item.MU.RLock()
	if item.QuantityInStock != nil {
		defer item.MU.RUnlock()
		return *item.QuantityInStock, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	quantity, err := item.DB.GetProductItemQuantityInStock(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.QuantityInStock = &quantity
	return quantity, nil
}

func (item *BuiltinProductItem[AccountID]) GetName(ctx context.Context) (string, error) {
	item.MU.RLock()
	if item.Name != nil {
		defer item.MU.RUnlock()
		return *item.Name, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := item.DB.GetProductItemName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Name = &name
	return name, nil
}

func (item *BuiltinProductItem[AccountID]) GetSKU(ctx context.Context) (string, error) {
	item.MU.RLock()
	if item.SKU != nil {
		defer item.MU.RUnlock()
		return *item.SKU, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	sku, err := item.DB.GetProductItemSKU(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.SKU = &sku
	return sku, nil
}

func (item *BuiltinProductItem[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (item *BuiltinProductItem[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetAttributes(ctx context.Context, attrs json.RawMessage) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemAttributes(ctx, &form, id, attrs); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Attributes = &attrs
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetImages(ctx context.Context, images []FileReader) error {
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

	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}

	err = item.DB.SetProductItemImages(ctx, &form, id, tokens)
	if err != nil {
		errRes = joinErr(errRes, err)
		return errRes
	}

	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	item.MU.Lock()
	item.Images = &tokens
	item.MU.Unlock()

	for i, token := range tokens {
		item.FS.Delete(ctx, token)
		file, err := item.FS.Create(ctx, token)
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

func (item *BuiltinProductItem[AccountID]) SetPrice(ctx context.Context, price float64) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemPrice(ctx, &form, id, price); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Price = &price
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetProduct(ctx context.Context, product Product[AccountID]) error {
	var pid *uint64 = nil
	if product != nil {
		tpid, err := product.GetID(ctx)
		if err != nil {
			return err
		}
		pid = &tpid
	}
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemProduct(ctx, &form, id, pid, item.FS); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	proc, err := product.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Product = proc
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetQuantityInStock(ctx context.Context, quantity uint64) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemQuantityInStock(ctx, &form, id, quantity); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.QuantityInStock = &quantity
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetName(ctx context.Context, name string) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Name = nil
	return nil
}

func (item *BuiltinProductItem[AccountID]) SetSKU(ctx context.Context, sku string) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetProductItemSKU(ctx, &form, id, sku); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.SKU = &sku
	return nil
}

func (item *BuiltinProductItem[AccountID]) newUserReview(ctx context.Context, id uint64, db userReviewDatabase[AccountID], form *UserReviewForm[AccountID]) (*BuiltinUserReview[AccountID], error) {
	review := &BuiltinUserReview[AccountID]{
		UserReviewForm: UserReviewForm[AccountID]{
			ID: id,
		},
		DB: db,
		FS: item.FS,
	}
	if err := review.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := review.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return review, nil
}

func (item *BuiltinProductItem[AccountID]) GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error) {
	var err error = nil
	id, err := item.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	reviewForms := make([]*UserReviewForm[AccountID], 0, cap(ids))
	ids, reviewForms, err = item.DB.GetProductItemUserReviews(ctx, &form, id, ids, reviewForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	revs := reviews
	if revs == nil {
		revs = make([]UserReview[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		review, err := item.newUserReview(ctx, ids[i], item.DB, reviewForms[i])
		if err != nil {
			return nil, err
		}
		revs = append(revs, review)
	}
	return revs, nil
}

func (item *BuiltinProductItem[AccountID]) GetUserReviewCount(ctx context.Context) (uint64, error) {
	id, err := item.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := item.DB.GetProductItemUserReviewCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	return count, nil
}

func (item *BuiltinProductItem[AccountID]) CalculateAverageRating(ctx context.Context) (float64, error) {
	id, err := item.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := item.ProductItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	average, err := item.DB.CalculateProductItemAverageRating(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	return average, nil
}

func (item *BuiltinProductItem[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProductItem[AccountID], error) {
	return item, nil
}

func (item *BuiltinProductItem[AccountID]) ToFormObject(ctx context.Context) (*ProductItemForm[AccountID], error) {
	item.MU.RLock()
	defer item.MU.RUnlock()
	return &item.ProductItemForm, nil
}

func (item *BuiltinProductItem[AccountID]) ApplyFormObject(ctx context.Context, form *ProductItemForm[AccountID]) error {
	item.MU.Lock()
	defer item.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		item.ID = form.ID
	}
	if form.Attributes != nil {
		item.Attributes = form.Attributes
	}
	if form.Images != nil {
		item.Images = form.Images
	}
	if form.Price != nil {
		item.Price = form.Price
	}
	if form.Product != nil {
		item.Product = form.Product
	}
	if form.QuantityInStock != nil {
		item.QuantityInStock = form.QuantityInStock
	}
	if form.Name != nil {
		item.Name = form.Name
	}
	if form.SKU != nil {
		item.SKU = form.SKU
	}
	return nil
}

func (form *ProductItemForm[AccountID]) Clone(ctx context.Context) (ProductItemForm[AccountID], error) {
	var cloned ProductItemForm[AccountID] = *form
	return cloned, nil
}
