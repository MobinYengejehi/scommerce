package scommerce

import (
	"context"
	"sync"
)

var _ UserShoppingCartManager[any] = &BuiltinUserShoppingCartManager[any]{}
var _ UserShoppingCart[any] = &BuiltinUserShoppingCart[any]{}

type userShoppingCartManagerDatabase[AccountID comparable] interface {
	DBUserShoppingCartManager[AccountID]
	userShoppingCartDatabase[AccountID]
}

type userShoppingCartDatabase[AccountID comparable] interface {
	DBUserShoppingCart[AccountID]
	DBUserShoppingCartItem[AccountID]
	userOrderDatabase[AccountID]
	productItemDatabase[AccountID]
}

type BuiltinUserShoppingCartManager[AccountID comparable] struct {
	DB                 userShoppingCartManagerDatabase[AccountID]
	FS                 FileStorage
	OrderStatusManager OrderStatusManager
}

type UserShoppingCartForm[AccountID comparable] struct {
	ID                    uint64    `json:"id"`
	UserAccountID         AccountID `json:"account_id"`
	SessionText           *string   `json:"session_text,omitempty"`
	ShoppingCartItemCount *uint64   `json:"shopping_cart_item_count,omitempty"`
	Dept                  *float64  `json:"dept,omitempty"`
}

type BuiltinUserShoppingCart[AccountID comparable] struct {
	UserShoppingCartForm[AccountID]
	DB                 userShoppingCartDatabase[AccountID] `json:"-"`
	FS                 FileStorage                         `json:"-"`
	OrderStatusManager OrderStatusManager                  `json:"-"`
	MU                 sync.RWMutex                        `json:"-"`
}

func NewBuiltinUserShoppingCartManager[AccountID comparable](db userShoppingCartManagerDatabase[AccountID], fs FileStorage, osm OrderStatusManager) *BuiltinUserShoppingCartManager[AccountID] {
	return &BuiltinUserShoppingCartManager[AccountID]{
		DB:                 db,
		FS:                 fs,
		OrderStatusManager: osm,
	}
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) newShoppingCart(ctx context.Context, id uint64, aid *AccountID, db userShoppingCartDatabase[AccountID], form *UserShoppingCartForm[AccountID]) (*BuiltinUserShoppingCart[AccountID], error) {
	cart := &BuiltinUserShoppingCart[AccountID]{
		DB:                 db,
		FS:                 shoppingCartManager.FS,
		OrderStatusManager: shoppingCartManager.OrderStatusManager,
		UserShoppingCartForm: UserShoppingCartForm[AccountID]{
			ID: id,
		},
	}
	if aid != nil {
		cart.UserShoppingCartForm.UserAccountID = *aid
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

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) GetShoppingCartBySessionText(ctx context.Context, sessionText string) (UserShoppingCart[AccountID], error) {
	cartForm := UserShoppingCartForm[AccountID]{}
	cid, err := shoppingCartManager.DB.GetShoppingCartBySessionText(ctx, sessionText, &cartForm)
	if err != nil {
		return nil, err
	}
	cart, err := shoppingCartManager.newShoppingCart(ctx, cid, nil, shoppingCartManager.DB, &cartForm)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) GetShoppingCartCount(ctx context.Context) (uint64, error) {
	return shoppingCartManager.DB.GetShoppingCartCount(ctx)
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) GetShoppingCarts(ctx context.Context, carts []UserShoppingCart[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCart[AccountID], error) {
	var err error = nil
	ids := make([]DBUserUserShoppingCartResult[AccountID], 0, GetSafeLimit(limit))
	cartForms := make([]*UserShoppingCartForm[AccountID], 0, cap(ids))
	ids, cartForms, err = shoppingCartManager.DB.GetShoppingCarts(ctx, ids, cartForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	cts := carts
	if cts == nil {
		cts = make([]UserShoppingCart[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		cart, err := shoppingCartManager.newShoppingCart(ctx, ids[i].ID, &ids[i].AID, shoppingCartManager.DB, cartForms[i])
		if err != nil {
			return nil, err
		}
		cts = append(cts, cart)
	}
	return cts, nil
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) Init(ctx context.Context) error {
	return shoppingCartManager.DB.InitUserShoppingCartManager(ctx)
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) RemoveAllShoppingCarts(ctx context.Context) error {
	return shoppingCartManager.RemoveAllShoppingCarts(ctx)
}

func (shoppingCartManager *BuiltinUserShoppingCartManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCartManager[AccountID], error) {
	return shoppingCartManager, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) CalculateDept(ctx context.Context, shippingMethod ShippingMethod) (float64, error) {
	var sid uint64 = 0
	if shippingMethod != nil {
		tsid, err := shippingMethod.GetID(ctx)
		if err != nil {
			return 0, err
		}
		sid = tsid
	}
	shoppingCart.MU.RLock()
	if shoppingCart.Dept != nil {
		defer shoppingCart.MU.RUnlock()
		return *shoppingCart.Dept, nil
	}
	shoppingCart.MU.RUnlock()
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	dept, err := shoppingCart.DB.CalculateUserShoppingCartDept(ctx, &form, id, sid)
	if err != nil {
		return 0, err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.Dept = &dept
	return dept, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) GetID(ctx context.Context) (uint64, error) {
	shoppingCart.MU.RLock()
	defer shoppingCart.MU.RUnlock()
	return shoppingCart.ID, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) GetSessionText(ctx context.Context) (string, error) {
	shoppingCart.MU.RLock()
	if shoppingCart.SessionText != nil {
		defer shoppingCart.MU.RUnlock()
		return *shoppingCart.SessionText, nil
	}
	shoppingCart.MU.RUnlock()
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	text, err := shoppingCart.DB.GetUserShoppingCartSessionText(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.SessionText = &text
	return text, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) GetShoppingCartItemCount(ctx context.Context) (uint64, error) {
	shoppingCart.MU.RLock()
	if shoppingCart.ShoppingCartItemCount != nil {
		defer shoppingCart.MU.RUnlock()
		return *shoppingCart.ShoppingCartItemCount, nil
	}
	shoppingCart.MU.RUnlock()
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := shoppingCart.DB.GetUserShoppingCartItemCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.ShoppingCartItemCount = &count
	return count, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) newShoppingCartItem(ctx context.Context, id uint64, db userShoppingCartItemDatabase[AccountID], form *UserShoppingCartItemForm[AccountID]) (*BuiltinUserShoppingCartItem[AccountID], error) {
	aid, err := shoppingCart.GetUserAccountID(ctx)
	if err != nil {
		return nil, err
	}
	item := &BuiltinUserShoppingCartItem[AccountID]{
		DB:                 db,
		FS:                 shoppingCart.FS,
		OrderStatusManager: shoppingCart.OrderStatusManager,
		UserShoppingCartItemForm: UserShoppingCartItemForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
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

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) GetShoppingCartItems(ctx context.Context, items []UserShoppingCartItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCartItem[AccountID], error) {
	var err error = nil
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	itemForms := make([]*UserShoppingCartItemForm[AccountID], 0, cap(ids))
	ids, itemForms, err = shoppingCart.DB.GetUserShoppingCartItems(ctx, &form, id, ids, itemForms, skip, limit, queueOrder, shoppingCart.FS, shoppingCart.OrderStatusManager)
	if err != nil {
		return nil, err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	itms := items
	if itms == nil {
		itms = make([]UserShoppingCartItem[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		item, err := shoppingCart.newShoppingCartItem(ctx, ids[i], shoppingCart.DB, itemForms[i])
		if err != nil {
			return nil, err
		}
		itms = append(itms, item)
	}
	return itms, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	shoppingCart.MU.RLock()
	defer shoppingCart.MU.RUnlock()
	return shoppingCart.UserAccountID, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) NewShoppingCartItem(ctx context.Context, item ProductItem[AccountID], count int64) (UserShoppingCartItem[AccountID], error) {
	itid, err := item.GetID(ctx)
	if err != nil {
		return nil, err
	}
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	itemForm := UserShoppingCartItemForm[AccountID]{}
	sitid, err := shoppingCart.DB.NewUserShoppingCartShoppingCartItem(ctx, &form, id, itid, count, &itemForm, shoppingCart.FS, shoppingCart.OrderStatusManager)
	if err != nil {
		return nil, err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	sItem, err := shoppingCart.newShoppingCartItem(ctx, sitid, shoppingCart.DB, &itemForm)
	if err != nil {
		return nil, err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.ShoppingCartItemCount = nil
	return sItem, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) newOrder(ctx context.Context, id uint64, db userOrderDatabase[AccountID], form *UserOrderForm[AccountID]) (*BuiltinUserOrder[AccountID], error) {
	aid, err := shoppingCart.GetUserAccountID(ctx)
	if err != nil {
		return nil, err
	}
	order := &BuiltinUserOrder[AccountID]{
		DB:                 db,
		FS:                 shoppingCart.FS,
		OrderStatusManager: shoppingCart.OrderStatusManager,
		UserOrderForm: UserOrderForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
	}
	if err := order.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := order.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) Order(ctx context.Context, paymentMethod UserPaymentMethod[AccountID], address UserAddress[AccountID], shippingMethod ShippingMethod, userComment string) (UserOrder[AccountID], error) {
	pid, err := paymentMethod.GetID(ctx)
	if err != nil {
		return nil, err
	}
	aid, err := address.GetID(ctx)
	if err != nil {
		return nil, err
	}
	sid, err := shippingMethod.GetID(ctx)
	if err != nil {
		return nil, err
	}
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	orderForm := UserOrderForm[AccountID]{}
	oid, err := shoppingCart.DB.OrderUserShoppingCart(ctx, &form, id, pid, aid, sid, userComment, &orderForm)
	if err != nil {
		return nil, err
	}
	order, err := shoppingCart.newOrder(ctx, oid, shoppingCart.DB, &orderForm)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) RemoveAllShoppingCartItems(ctx context.Context) error {
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := shoppingCart.DB.RemoveUserShoppingCartAllShoppingCartItems(ctx, &form, id); err != nil {
		return err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.ShoppingCartItemCount = nil
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) RemoveShoppingCartItem(ctx context.Context, item UserShoppingCartItem[AccountID]) error {
	itid, err := item.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := shoppingCart.DB.RemoveUserShoppingCartShoppingCartItem(ctx, &form, id, itid); err != nil {
		return err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.ShoppingCartItemCount = nil
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) SetSessionText(ctx context.Context, sessionText string) error {
	id, err := shoppingCart.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := shoppingCart.UserShoppingCartForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := shoppingCart.DB.SetUserShoppingCartSessionText(ctx, &form, id, sessionText); err != nil {
		return err
	}
	if err := shoppingCart.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	shoppingCart.SessionText = &sessionText
	return nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCart[AccountID], error) {
	return shoppingCart, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) ToFormObject(ctx context.Context) (*UserShoppingCartForm[AccountID], error) {
	shoppingCart.MU.RLock()
	defer shoppingCart.MU.RUnlock()
	return &shoppingCart.UserShoppingCartForm, nil
}

func (shoppingCart *BuiltinUserShoppingCart[AccountID]) ApplyFormObject(ctx context.Context, form *UserShoppingCartForm[AccountID]) error {
	shoppingCart.MU.Lock()
	defer shoppingCart.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		shoppingCart.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		shoppingCart.UserAccountID = form.UserAccountID
	}
	if form.SessionText != nil {
		shoppingCart.SessionText = form.SessionText
	}
	if form.ShoppingCartItemCount != nil {
		shoppingCart.ShoppingCartItemCount = form.ShoppingCartItemCount
	}
	if form.Dept != nil {
		shoppingCart.Dept = form.Dept
	}
	return nil
}

func (form *UserShoppingCartForm[AccountID]) Clone(ctx context.Context) (UserShoppingCartForm[AccountID], error) {
	var cloned UserShoppingCartForm[AccountID] = *form
	return cloned, nil
}
