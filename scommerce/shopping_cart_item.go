package scommerce

import (
	"context"
	"sync"
)

var _ UserShoppingCartItem[any] = &BuiltinUserShoppingCartItem[any]{}

type userShoppingCartItemDatabase[AccountID comparable] interface {
	DBUserShoppingCartItem[AccountID]
	userShoppingCartDatabase[AccountID]
	productItemDatabase[AccountID]
}

type UserShoppingCartItemForm[AccountID comparable] struct {
	ID            uint64                              `json:"id"`
	UserAccountID AccountID                           `json:"account_id"`
	Dept          *float64                            `json:"dept,omitempty"`
	ProductItem   *BuiltinProductItem[AccountID]      `json:"product_item,omitempty"`
	Quantity      *int64                              `json:"quantity,omitempty"`
	ShoppingCart  *BuiltinUserShoppingCart[AccountID] `json:"shopping_cart,omitempty"`
}

type BuiltinUserShoppingCartItem[AccountID comparable] struct {
	UserShoppingCartItemForm[AccountID]
	DB                 userShoppingCartItemDatabase[AccountID] `json:"-"`
	FS                 FileStorage                             `json:"-"`
	OrderStatusManager OrderStatusManager                      `json:"-"`
	MU                 sync.RWMutex                            `json:"-"`
}

func (item *BuiltinUserShoppingCartItem[AccountID]) AddQuantity(ctx context.Context, delta int64) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.AddUserShoppingCartItemQuantity(ctx, &form, id, delta); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.RUnlock()
	item.Quantity = nil
	return nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) CalculateDept(ctx context.Context) (float64, error) {
	item.MU.RLock()
	if item.Dept != nil {
		defer item.MU.RUnlock()
		return *item.Dept, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	dept, err := item.DB.CalculateUserShoppingCartItemDept(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Dept = &dept
	return dept, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) GetID(ctx context.Context) (uint64, error) {
	item.MU.RLock()
	defer item.MU.RUnlock()
	return item.ID, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	item.MU.RLock()
	defer item.MU.RUnlock()
	return item.UserAccountID, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) newProductItem(ctx context.Context, id uint64, db productItemDatabase[AccountID], form *ProductItemForm[AccountID]) (*BuiltinProductItem[AccountID], error) {
	pItem := &BuiltinProductItem[AccountID]{
		DB: db,
		FS: item.FS,
		ProductItemForm: ProductItemForm[AccountID]{
			ID: id,
		},
	}
	if err := pItem.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := pItem.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return pItem, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) GetProductItem(ctx context.Context) (ProductItem[AccountID], error) {
	item.MU.RLock()
	if item.ProductItem != nil {
		defer item.MU.RUnlock()
		return item.ProductItem, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	pItemForm := ProductItemForm[AccountID]{}
	itid, err := item.DB.GetUserShoppingCartItemProductItem(ctx, &form, id, &pItemForm, item.FS)
	if err != nil {
		return nil, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	pItem, err := item.newProductItem(ctx, itid, item.DB, &pItemForm)
	if err != nil {
		return nil, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.ProductItem = pItem
	return pItem, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) GetQuantity(ctx context.Context) (int64, error) {
	item.MU.RLock()
	if item.Quantity != nil {
		defer item.MU.RUnlock()
		return *item.Quantity, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	quantity, err := item.DB.GetUserShoppingCartItemQuantity(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Quantity = &quantity
	return quantity, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) newShoppingCart(ctx context.Context, id uint64, db userShoppingCartDatabase[AccountID], form *UserShoppingCartForm[AccountID]) (*BuiltinUserShoppingCart[AccountID], error) {
	aid, err := item.GetUserAccountID(ctx)
	if err != nil {
		return nil, err
	}
	cart := &BuiltinUserShoppingCart[AccountID]{
		DB:                 db,
		FS:                 item.FS,
		OrderStatusManager: item.OrderStatusManager,
		UserShoppingCartForm: UserShoppingCartForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
	}
	if err := cart.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := cart.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return cart, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) GetShoppingCart(ctx context.Context) (UserShoppingCart[AccountID], error) {
	item.MU.RLock()
	if item.ShoppingCart != nil {
		defer item.MU.RUnlock()
		return item.ShoppingCart, nil
	}
	item.MU.RUnlock()
	id, err := item.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	cartForm := UserShoppingCartForm[AccountID]{}
	sid, err := item.DB.GetUserShoppingCartItemShoppingCart(ctx, &form, id, &cartForm, item.FS, item.OrderStatusManager)
	if err != nil {
		return nil, err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	cart, err := item.newShoppingCart(ctx, sid, item.DB, &cartForm)
	if err != nil {
		return nil, err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.ShoppingCart = cart
	return cart, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) SetQuantity(ctx context.Context, quantity int64) error {
	id, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := item.UserShoppingCartItemForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := item.DB.SetUserShoppingCartItemQuantity(ctx, &form, id, quantity); err != nil {
		return err
	}
	if err := item.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	item.MU.Lock()
	defer item.MU.Unlock()
	item.Quantity = &quantity
	return nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCartItem[AccountID], error) {
	return item, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) ToFormObject(ctx context.Context) (*UserShoppingCartItemForm[AccountID], error) {
	item.MU.RLock()
	defer item.MU.RUnlock()
	return &item.UserShoppingCartItemForm, nil
}

func (item *BuiltinUserShoppingCartItem[AccountID]) ApplyFormObject(ctx context.Context, form *UserShoppingCartItemForm[AccountID]) error {
	item.MU.Lock()
	defer item.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		item.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		item.UserAccountID = form.UserAccountID
	}
	if form.Dept != nil {
		item.Dept = form.Dept
	}
	if form.ProductItem != nil {
		item.ProductItem = form.ProductItem
	}
	if form.Quantity != nil {
		item.Quantity = form.Quantity
	}
	if form.ShoppingCart != nil {
		item.ShoppingCart = form.ShoppingCart
	}
	return nil
}

func (form *UserShoppingCartItemForm[AccountID]) Clone(ctx context.Context) (UserShoppingCartItemForm[AccountID], error) {
	var cloned UserShoppingCartItemForm[AccountID] = *form
	return cloned, nil
}
