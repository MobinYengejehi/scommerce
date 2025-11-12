package scommerce

import (
	"context"
	"sync"
	"time"
)

var _ UserOrderManager[any] = &BuiltinUserOrderManager[any]{}
var _ UserOrder[any] = &BuiltinUserOrder[any]{}

type userOrderDatabase[AccountID comparable] interface {
	DBUserOrder[AccountID]
	userAddressDatabase[AccountID]
	userPaymentMethodDatabase[AccountID]
	DBShippingMethod
	DBOrderStatus
	productItemDatabase[AccountID]
}

type userOrderManagerDatabase[AccountID comparable] interface {
	DBUserOrderManager[AccountID]
	userOrderDatabase[AccountID]
}

type BuiltinUserOrderManager[AccountID comparable] struct {
	DB                 userOrderManagerDatabase[AccountID]
	FS                 FileStorage
	OrderStatusManager OrderStatusManager
}

type UserOrderProductItem[AccountID comparable] struct {
	ProductItem *BuiltinProductItem[AccountID] `json:"product_item"`
	Quantity    uint64                         `json:"quantity"`
}

type UserOrderForm[AccountID comparable] struct {
	ID                uint64                               `json:"id"`
	UserAccountID     AccountID                            `json:"account_id"`
	TotalPrice        *float64                             `json:"total_price,omitempty"`
	DeliveryComment   *string                              `json:"delivery_comment,omitempty"`
	DeliveryDate      *time.Time                           `json:"delivery_date,omitempty"`
	Date              *time.Time                           `json:"date,omitempty"`
	Total             *float64                             `json:"total,omitempty"`
	PaymentMethod     *BuiltinUserPaymentMethod[AccountID] `json:"payment_method,omitempty"`
	ProductItemCount  *uint64                              `json:"product_item_count,omitempty"`
	ShippingAddress   *BuiltinUserAddress[AccountID]       `json:"shipping_address,omitempty"`
	ShippingMethod    *BuiltinShippingMethod               `json:"shipping_method,omitempty"`
	Status            *BuiltinOrderStatus                  `json:"status,omitempty"`
	UserComment       *string                              `json:"user_comment,omitempty"`
	IsDeliveriedState *bool                                `json:"is_deliveried,omitempty"`
}

type BuiltinUserOrder[AccountID comparable] struct {
	UserOrderForm[AccountID]
	OrderStatusManager OrderStatusManager           `json:"-"`
	DB                 userOrderDatabase[AccountID] `json:"-"`
	FS                 FileStorage                  `json:"-"`
	MU                 sync.RWMutex                 `json:"-"`
}

func NewBuiltinUserOrderManager[AccountID comparable](db userOrderManagerDatabase[AccountID], osm OrderStatusManager, fs FileStorage) *BuiltinUserOrderManager[AccountID] {
	return &BuiltinUserOrderManager[AccountID]{
		DB:                 db,
		FS:                 fs,
		OrderStatusManager: osm,
	}
}

func (orderManager *BuiltinUserOrderManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (orderManager *BuiltinUserOrderManager[AccountID]) GetUserOrderCount(ctx context.Context) (uint64, error) {
	return orderManager.DB.GetUserOrderCount(ctx)
}

func (orderManager *BuiltinUserOrderManager[AccountID]) newUserOrder(ctx context.Context, oid uint64, aid AccountID, db userOrderDatabase[AccountID], form *UserOrderForm[AccountID]) (*BuiltinUserOrder[AccountID], error) {
	order := &BuiltinUserOrder[AccountID]{
		UserOrderForm: UserOrderForm[AccountID]{
			ID:            oid,
			UserAccountID: aid,
		},
		DB:                 db,
		FS:                 orderManager.FS,
		OrderStatusManager: orderManager.OrderStatusManager,
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

func (orderManager *BuiltinUserOrderManager[AccountID]) GetUserOrders(ctx context.Context, orders []UserOrder[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrder[AccountID], error) {
	var err error = nil
	ids := make([]DBUserOrderResult[AccountID], 0, GetSafeLimit(limit))
	orderForms := make([]*UserOrderForm[AccountID], 0, cap(ids))
	ids, orderForms, err = orderManager.DB.GetUserOrders(ctx, ids, orderForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	ods := orders
	if ods == nil {
		ods = make([]UserOrder[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		res := ids[i]
		order, err := orderManager.newUserOrder(ctx, res.ID, res.AID, orderManager.DB, orderForms[i])
		if err != nil {
			return nil, err
		}
		ods = append(ods, order)
	}
	return ods, nil
}

func (orderManager *BuiltinUserOrderManager[AccountID]) Init(ctx context.Context) error {
	return orderManager.DB.InitUserOrderManager(ctx)
}

func (orderManager *BuiltinUserOrderManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (orderManager *BuiltinUserOrderManager[AccountID]) RemoveAllUserOrders(ctx context.Context) error {
	return orderManager.RemoveAllUserOrders(ctx)
}

func (orderManager *BuiltinUserOrderManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserOrderManager[AccountID], error) {
	return orderManager, nil
}

func (order *BuiltinUserOrder[AccountID]) CalculateTotalPrice(ctx context.Context) (float64, error) {
	order.MU.RLock()
	if order.TotalPrice != nil {
		defer order.MU.RUnlock()
		return *order.TotalPrice, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	price, err := order.DB.CalculateUserOrderTotalPrice(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.TotalPrice = &price
	return price, nil
}

func (order *BuiltinUserOrder[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (order *BuiltinUserOrder[AccountID]) Deliver(ctx context.Context, date time.Time, comment string) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	delivered, err := order.OrderStatusManager.GetDeliveriedOrderStatus(ctx)
	if err != nil {
		return err
	}
	did, err := delivered.GetID(ctx)
	if err != nil {
		return err
	}
	dstat := &BuiltinOrderStatus{
		DB: order.DB,
		OrderStatusForm: OrderStatusForm{
			ID: did,
		},
	}
	if err := dstat.Init(ctx); err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.DeliverUserOrder(ctx, &form, id, did, date, comment); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	state := true
	order.IsDeliveriedState = &state
	order.Status = dstat
	return nil
}

func (order *BuiltinUserOrder[AccountID]) GetDeliveryComment(ctx context.Context) (string, error) {
	order.MU.RLock()
	if order.DeliveryComment != nil {
		defer order.MU.RUnlock()
		return *order.DeliveryComment, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	comment, err := order.DB.GetUserOrderDeliveryComment(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.DeliveryComment = &comment
	return comment, nil
}

func (order *BuiltinUserOrder[AccountID]) GetDeliveryDate(ctx context.Context) (time.Time, error) {
	order.MU.RLock()
	if order.DeliveryDate != nil {
		defer order.MU.RUnlock()
		return *order.DeliveryDate, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return time.Time{}, err
	}
	date, err := order.DB.GetUserOrderDeliveryDate(ctx, &form, id)
	if err != nil {
		return time.Time{}, nil
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.DeliveryDate = &date
	return date, nil
}

func (order *BuiltinUserOrder[AccountID]) GetID(ctx context.Context) (uint64, error) {
	order.MU.RLock()
	order.MU.RUnlock()
	return order.ID, nil
}

func (order *BuiltinUserOrder[AccountID]) GetOrderDate(ctx context.Context) (time.Time, error) {
	order.MU.RLock()
	if order.Date != nil {
		defer order.MU.RUnlock()
		return *order.Date, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return time.Time{}, err
	}
	date, err := order.DB.GetUserOrderDate(ctx, &form, id)
	if err != nil {
		return time.Time{}, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Date = &date
	return date, nil
}

func (order *BuiltinUserOrder[AccountID]) GetOrderTotal(ctx context.Context) (float64, error) {
	order.MU.RLock()
	if order.Total != nil {
		defer order.MU.RUnlock()
		return *order.Total, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	total, err := order.DB.GetUserOrderTotal(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Total = &total
	return total, nil
}

func (order *BuiltinUserOrder[AccountID]) GetPaymentMethod(ctx context.Context) (UserPaymentMethod[AccountID], error) {
	order.MU.RLock()
	if order.PaymentMethod != nil {
		defer order.MU.RUnlock()
		return order.PaymentMethod, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := order.GetUserAccountID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	paymentMethodForm := UserPaymentMethodForm[AccountID]{}
	pid, err := order.DB.GetUserOrderPaymentMethod(ctx, &form, id, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	method := &BuiltinUserPaymentMethod[AccountID]{
		UserPaymentMethodForm: UserPaymentMethodForm[AccountID]{
			ID:            pid,
			UserAccountID: uid,
		},
		DB: order.DB,
	}
	if err := method.Init(ctx); err != nil {
		return nil, err
	}
	if err := method.ApplyFormObject(ctx, &paymentMethodForm); err != nil {
		return nil, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.PaymentMethod = method
	return method, nil
}

func (order *BuiltinUserOrder[AccountID]) GetProductItemCount(ctx context.Context) (uint64, error) {
	order.MU.RLock()
	if order.ProductItemCount != nil {
		defer order.MU.RUnlock()
		return *order.ProductItemCount, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := order.DB.GetUserOrderProductItemCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	order.ProductItemCount = &count
	return count, nil
}

func (order *BuiltinUserOrder[AccountID]) newProductItem(ctx context.Context, id uint64, db productItemDatabase[AccountID], form *ProductItemForm[AccountID]) (*BuiltinProductItem[AccountID], error) {
	item := &BuiltinProductItem[AccountID]{
		ProductItemForm: ProductItemForm[AccountID]{
			ID: id,
		},
		DB: db,
		FS: order.FS,
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

func (order *BuiltinUserOrder[AccountID]) GetProductItems(ctx context.Context, items []UserOrderProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrderProductItem[AccountID], error) {
	var err error = nil
	id, err := order.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	dbOrderItems := make([]DBUserOrderProductItem, 0, GetSafeLimit(limit))
	dbOrderItems, err = order.DB.GetUserOrderProductItems(ctx, &form, id, dbOrderItems, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	itms := items
	if itms == nil {
		itms = make([]UserOrderProductItem[AccountID], 0, len(dbOrderItems))
	}
	for _, dbItem := range dbOrderItems {
		productItem, err := order.newProductItem(ctx, dbItem.ProductItemID, order.DB, &ProductItemForm[AccountID]{
			ID: dbItem.ProductItemID,
		})
		if err != nil {
			return nil, err
		}
		itms = append(itms, UserOrderProductItem[AccountID]{
			ProductItem: productItem,
			Quantity:    dbItem.Quantity,
		})
	}
	return itms, nil
}

func (order *BuiltinUserOrder[AccountID]) GetShippingAddress(ctx context.Context) (UserAddress[AccountID], error) {
	order.MU.RLock()
	if order.ShippingAddress != nil {
		defer order.MU.RUnlock()
		return order.ShippingAddress, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := order.GetUserAccountID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	addressForm := UserAddressForm[AccountID]{}
	aid, err := order.DB.GetUserOrderShippingAddress(ctx, &form, id, &addressForm)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	addr := &BuiltinUserAddress[AccountID]{
		UserAddressForm: UserAddressForm[AccountID]{
			ID:            aid,
			UserAccountID: uid,
		},
		DB: order.DB,
	}
	if err := addr.Init(ctx); err != nil {
		return nil, err
	}
	if err := addr.ApplyFormObject(ctx, &addressForm); err != nil {
		return nil, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.ShippingAddress = addr
	return addr, nil
}

func (order *BuiltinUserOrder[AccountID]) GetShippingMethod(ctx context.Context) (ShippingMethod, error) {
	order.MU.RLock()
	if order.ShippingMethod != nil {
		defer order.MU.RUnlock()
		return order.ShippingMethod, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	shippingMethodForm := ShippingMethodForm{}
	mid, err := order.DB.GetUserOrderShippingMethod(ctx, &form, id, &shippingMethodForm)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	method := &BuiltinShippingMethod{
		DB: order.DB,
		ShippingMethodForm: ShippingMethodForm{
			ID: mid,
		},
	}
	if err := method.Init(ctx); err != nil {
		return nil, err
	}
	if err := method.ApplyFormObject(ctx, &shippingMethodForm); err != nil {
		return nil, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.ShippingMethod = method
	return method, nil
}

func (order *BuiltinUserOrder[AccountID]) GetStatus(ctx context.Context) (OrderStatus, error) {
	order.MU.RLock()
	if order.Status != nil {
		defer order.MU.RUnlock()
		return order.Status, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	statusForm := OrderStatusForm{}
	sid, err := order.DB.GetUserOrderStatus(ctx, &form, id, &statusForm)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	status := &BuiltinOrderStatus{
		DB: order.DB,
		OrderStatusForm: OrderStatusForm{
			ID: sid,
		},
	}
	if err := status.Init(ctx); err != nil {
		return nil, err
	}
	if err := status.ApplyFormObject(ctx, &statusForm); err != nil {
		return nil, err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Status = status
	return status, nil
}

func (order *BuiltinUserOrder[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	order.MU.RLock()
	defer order.MU.RUnlock()
	return order.UserAccountID, nil
}

func (order *BuiltinUserOrder[AccountID]) GetUserComment(ctx context.Context) (string, error) {
	order.MU.RLock()
	if order.UserComment != nil {
		defer order.MU.RUnlock()
		return *order.UserComment, nil
	}
	order.MU.RUnlock()
	id, err := order.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	comment, err := order.DB.GetUserOrderUserComment(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	order.UserComment = &comment
	return comment, nil
}

func (order *BuiltinUserOrder[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (order *BuiltinUserOrder[AccountID]) IsDeliveried(ctx context.Context) (bool, error) {
	order.MU.RLock()
	if order.IsDeliveriedState != nil {
		defer order.MU.RUnlock()
		return *order.IsDeliveriedState, nil
	}
	order.MU.RUnlock()
	delivered, err := order.OrderStatusManager.GetDeliveriedOrderStatus(ctx)
	if err != nil {
		return false, err
	}
	did, err := delivered.GetID(ctx)
	if err != nil {
		return false, err
	}
	mStatus, err := order.GetStatus(ctx)
	if err != nil {
		return false, err
	}
	msid, err := mStatus.GetID(ctx)
	if err != nil {
		return false, err
	}
	state := did == msid
	order.MU.Lock()
	defer order.MU.Unlock()
	order.IsDeliveriedState = &state
	return state, nil
}

func (order *BuiltinUserOrder[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetDeliveryComment(ctx context.Context, comment string) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderDeliveryComment(ctx, &form, id, comment); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.DeliveryComment = &comment
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetDeliveryDate(ctx context.Context, date time.Time) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderDeliveryDate(ctx, &form, id, date); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.DeliveryDate = &date
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetOrderDate(ctx context.Context, date time.Time) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderDate(ctx, &form, id, date); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Date = &date
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetOrderTotal(ctx context.Context, price float64) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderTotal(ctx, &form, id, price); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Total = &price
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetPaymentMethod(ctx context.Context, paymentMethod UserPaymentMethod[AccountID]) error {
	pid, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderPaymentMethod(ctx, &form, id, pid); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	meth, err := paymentMethod.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.PaymentMethod = meth
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetProductItems(ctx context.Context, items []UserOrderProductItem[AccountID]) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	// Convert UserOrderProductItem to DBUserOrderProductItem
	dbItems := make([]DBUserOrderProductItem, 0, len(items))
	for _, item := range items {
		pid, err := item.ProductItem.GetID(ctx)
		if err != nil {
			return err
		}
		dbItems = append(dbItems, DBUserOrderProductItem{
			ProductItemID: pid,
			Quantity:      item.Quantity,
		})
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderProductItems(ctx, &form, id, dbItems); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.ProductItemCount = nil
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetShippingAddress(ctx context.Context, address UserAddress[AccountID]) error {
	aid, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderShippingAddress(ctx, &form, id, aid); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	addr, err := address.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.ShippingAddress = addr
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetShippingMethod(ctx context.Context, shippingMethod ShippingMethod) error {
	aid, err := shippingMethod.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderShippingMethod(ctx, &form, id, aid); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	meth, err := shippingMethod.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.ShippingMethod = meth
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetStatus(ctx context.Context, status OrderStatus) error {
	sid, err := status.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderStatus(ctx, &form, id, sid); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	stat, err := status.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.Status = stat
	return nil
}

func (order *BuiltinUserOrder[AccountID]) SetUserComment(ctx context.Context, comment string) error {
	id, err := order.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := order.UserOrderForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := order.DB.SetUserOrderUserComment(ctx, &form, id, comment); err != nil {
		return err
	}
	if err := order.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	order.MU.Lock()
	defer order.MU.Unlock()
	order.UserComment = &comment
	return nil
}

func (order *BuiltinUserOrder[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserOrder[AccountID], error) {
	return order, nil
}

func (order *BuiltinUserOrder[AccountID]) ToFormObject(ctx context.Context) (*UserOrderForm[AccountID], error) {
	order.MU.RLock()
	defer order.MU.RUnlock()
	return &order.UserOrderForm, nil
}

func (order *BuiltinUserOrder[AccountID]) ApplyFormObject(ctx context.Context, form *UserOrderForm[AccountID]) error {
	order.MU.Lock()
	defer order.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		order.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		order.UserAccountID = form.UserAccountID
	}
	if form.TotalPrice != nil {
		order.TotalPrice = form.TotalPrice
	}
	if form.DeliveryComment != nil {
		order.DeliveryComment = form.DeliveryComment
	}
	if form.DeliveryDate != nil {
		order.DeliveryDate = form.DeliveryDate
	}
	if form.Date != nil {
		order.Date = form.Date
	}
	if form.Total != nil {
		order.Total = form.Total
	}
	if form.PaymentMethod != nil {
		order.PaymentMethod = form.PaymentMethod
	}
	if form.ProductItemCount != nil {
		order.ProductItemCount = form.ProductItemCount
	}
	if form.ShippingAddress != nil {
		order.ShippingAddress = form.ShippingAddress
	}
	if form.ShippingMethod != nil {
		order.ShippingMethod = form.ShippingMethod
	}
	if form.Status != nil {
		order.Status = form.Status
	}
	if form.UserComment != nil {
		order.UserComment = form.UserComment
	}
	if form.IsDeliveriedState != nil {
		order.IsDeliveriedState = form.IsDeliveriedState
	}
	return nil
}

func (form *UserOrderForm[AccountID]) Clone(ctx context.Context) (UserOrderForm[AccountID], error) {
	var cloned UserOrderForm[AccountID] = *form
	return cloned, nil
}
