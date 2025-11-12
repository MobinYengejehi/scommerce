package scommerce

import (
	"context"
	"sync"
)

type orderStatusDatabase interface {
	DBOrderStatusManager
	DBOrderStatus
}

var _ OrderStatusManager = &BuiltinOrderStatusManager{}
var _ OrderStatus = &BuiltinOrderStatus{}

type BuiltinOrderStatusManager struct {
	DB orderStatusDatabase
}

type OrderStatusForm struct {
	ID                uint64  `json:"id"`
	Name              *string `json:"name,omitempty"`
	IsDeliveriedState *bool   `json:"is_deliveried,omitempty"`
}

type BuiltinOrderStatus struct {
	OrderStatusForm
	DB DBOrderStatus `json:"-"`
	MU sync.RWMutex  `json:"-"`
}

func NewBuiltinOrderStatusManager(db orderStatusDatabase) *BuiltinOrderStatusManager {
	return &BuiltinOrderStatusManager{
		DB: db,
	}
}

func (orderStatusManager *BuiltinOrderStatusManager) newBuiltinOrderStatus(ctx context.Context, id uint64, db DBOrderStatus, form *OrderStatusForm) (*BuiltinOrderStatus, error) {
	status := &BuiltinOrderStatus{
		OrderStatusForm: OrderStatusForm{
			ID: id,
		},
		DB: db,
	}
	if err := status.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := status.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return status, nil
}

func (orderStatusManager *BuiltinOrderStatusManager) Close(ctx context.Context) error {
	return nil
}

func (orderStatusManager *BuiltinOrderStatusManager) ExistsOrderStatus(ctx context.Context, name string) (bool, error) {
	return orderStatusManager.DB.ExistsOrderStatus(ctx, name)
}

func (orderStatusManager *BuiltinOrderStatusManager) GetDeliveriedOrderStatus(ctx context.Context) (OrderStatus, error) {
	statusForm := OrderStatusForm{}
	id, err := orderStatusManager.DB.GetDeliveriedOrderStatus(ctx, &statusForm)
	if err != nil {
		return nil, err
	}
	return orderStatusManager.newBuiltinOrderStatus(ctx, id, orderStatusManager.DB, &statusForm)
}

func (orderStatusManager *BuiltinOrderStatusManager) GetIdleOrderStatus(ctx context.Context) (OrderStatus, error) {
	statusForm := OrderStatusForm{}
	id, err := orderStatusManager.DB.GetIdleOrderStatus(ctx, &statusForm)
	if err != nil {
		return nil, err
	}
	return orderStatusManager.newBuiltinOrderStatus(ctx, id, orderStatusManager.DB, &statusForm)
}

func (orderStatusManager *BuiltinOrderStatusManager) GetOrderStatusByName(ctx context.Context, name string) (OrderStatus, error) {
	statusForm := OrderStatusForm{}
	id, err := orderStatusManager.DB.GetOrderStatusByName(ctx, name, &statusForm)
	if err != nil {
		return nil, err
	}
	return orderStatusManager.newBuiltinOrderStatus(ctx, id, orderStatusManager.DB, &statusForm)
}

func (orderStatusManager *BuiltinOrderStatusManager) GetOrderStatusCount(ctx context.Context) (uint64, error) {
	return orderStatusManager.DB.GetOrderStatusCount(ctx)
}

func (orderStatusManager *BuiltinOrderStatusManager) GetOrderStatuses(ctx context.Context, orderStatuses []OrderStatus, skip int64, limit int64, queueOrder QueueOrder) ([]OrderStatus, error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	statusForms := make([]*OrderStatusForm, 0, cap(ids))
	ids, statusForms, err = orderStatusManager.DB.GetOrderStatuses(ctx, ids, statusForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	statuses := orderStatuses
	if statuses == nil {
		statuses = make([]OrderStatus, 0, len(ids))
	}
	for i := range len(ids) {
		status, err := orderStatusManager.newBuiltinOrderStatus(ctx, ids[i], orderStatusManager.DB, statusForms[i])
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func (orderStatusManager *BuiltinOrderStatusManager) Init(ctx context.Context) error {
	return orderStatusManager.DB.InitOrderStatusManager(ctx)
}

func (orderStatusManager *BuiltinOrderStatusManager) NewOrderStatus(ctx context.Context, name string) (OrderStatus, error) {
	statusForm := OrderStatusForm{}
	id, err := orderStatusManager.DB.NewOrderStatus(ctx, name, &statusForm)
	if err != nil {
		return nil, err
	}
	return orderStatusManager.newBuiltinOrderStatus(ctx, id, orderStatusManager.DB, &statusForm)
}

func (orderStatusManager *BuiltinOrderStatusManager) Pulse(ctx context.Context) error {
	return nil
}

func (orderStatusManager *BuiltinOrderStatusManager) RemoveAllOrderStatuses(ctx context.Context) error {
	return orderStatusManager.DB.RemoveAllOrderStatuses(ctx)
}

func (orderStatusManager *BuiltinOrderStatusManager) RemoveOrderStatus(ctx context.Context, status OrderStatus) error {
	id, err := status.GetID(ctx)
	if err != nil {
		return err
	}
	return orderStatusManager.DB.RemoveOrderStatus(ctx, id)
}

func (orderStatusManager *BuiltinOrderStatusManager) ToBuiltinObject(ctx context.Context) (*BuiltinOrderStatusManager, error) {
	return orderStatusManager, nil
}

func (orderStatus *BuiltinOrderStatus) Close(ctx context.Context) error {
	return nil
}

func (orderStatus *BuiltinOrderStatus) GetID(ctx context.Context) (uint64, error) {
	orderStatus.MU.RLock()
	defer orderStatus.MU.RUnlock()
	return orderStatus.ID, nil
}

func (orderStatus *BuiltinOrderStatus) GetName(ctx context.Context) (string, error) {
	orderStatus.MU.RLock()
	if orderStatus.Name != nil {
		defer orderStatus.MU.RUnlock()
		return *orderStatus.Name, nil
	}
	orderStatus.MU.RUnlock()
	id, err := orderStatus.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := orderStatus.OrderStatusForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := orderStatus.DB.GetOrderStatusName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := orderStatus.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	orderStatus.MU.Lock()
	defer orderStatus.MU.Unlock()
	orderStatus.Name = &name
	return name, nil
}

func (orderStatus *BuiltinOrderStatus) Init(ctx context.Context) error {
	return nil
}

func (orderStatus *BuiltinOrderStatus) IsDeliveried(ctx context.Context) (bool, error) {
	orderStatus.MU.RLock()
	if orderStatus.IsDeliveriedState != nil {
		defer orderStatus.MU.RUnlock()
		return *orderStatus.IsDeliveriedState, nil
	}
	orderStatus.MU.RUnlock()
	id, err := orderStatus.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := orderStatus.OrderStatusForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isDeliveried, err := orderStatus.DB.IsOrderStatusDeliveried(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := orderStatus.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	orderStatus.MU.Lock()
	defer orderStatus.MU.Unlock()
	orderStatus.IsDeliveriedState = &isDeliveried
	return isDeliveried, nil
}

func (orderStatus *BuiltinOrderStatus) Pulse(ctx context.Context) error {
	return nil
}

func (orderStatus *BuiltinOrderStatus) SetName(ctx context.Context, name string) error {
	id, err := orderStatus.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := orderStatus.OrderStatusForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := orderStatus.DB.SetOrderStatusName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := orderStatus.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	orderStatus.MU.Lock()
	defer orderStatus.MU.Unlock()
	orderStatus.Name = &name
	return nil
}

func (orderStatus *BuiltinOrderStatus) ToBuiltinObject(ctx context.Context) (*BuiltinOrderStatus, error) {
	return orderStatus, nil
}

func (orderStatus *BuiltinOrderStatus) ToFormObject(ctx context.Context) (*OrderStatusForm, error) {
	orderStatus.MU.RLock()
	defer orderStatus.MU.RUnlock()
	return &orderStatus.OrderStatusForm, nil
}

func (orderStatus *BuiltinOrderStatus) ApplyFormObject(ctx context.Context, form *OrderStatusForm) error {
	orderStatus.MU.Lock()
	defer orderStatus.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		orderStatus.ID = form.ID
	}
	if form.Name != nil {
		orderStatus.Name = form.Name
	}
	if form.IsDeliveriedState != nil {
		orderStatus.IsDeliveriedState = form.IsDeliveriedState
	}
	return nil
}

func (form *OrderStatusForm) Clone(ctx context.Context) (OrderStatusForm, error) {
	var cloned OrderStatusForm = *form
	return cloned, nil
}
