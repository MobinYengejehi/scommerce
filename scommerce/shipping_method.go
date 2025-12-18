package scommerce

import (
	"context"
	"sync"
)

type shippingMethodDatabase interface {
	DBShippingMethodManager
	DBShippingMethod
}

var _ ShippingMethodManager = &BuiltinShippingMethodManager{}
var _ ShippingMethod = &BuiltinShippingMethod{}

type BuiltinShippingMethodManager struct {
	DB shippingMethodDatabase
}

type ShippingMethodForm struct {
	ID    uint64   `json:"id"`
	Name  *string  `json:"name,omitempty"`
	Price *float64 `json:"price,omitempty"`
}

type BuiltinShippingMethod struct {
	ShippingMethodForm
	DB DBShippingMethod `json:"-"`
	MU sync.RWMutex     `json:"-"`
}

func NewBuiltinShippingMethodManager(db shippingMethodDatabase) *BuiltinShippingMethodManager {
	return &BuiltinShippingMethodManager{
		DB: db,
	}
}

func (shippingMethodManager *BuiltinShippingMethodManager) newBuiltinShippingMethod(ctx context.Context, id uint64, db DBShippingMethod, form *ShippingMethodForm) (*BuiltinShippingMethod, error) {
	method := &BuiltinShippingMethod{
		ShippingMethodForm: ShippingMethodForm{
			ID: id,
		},
		DB: db,
	}
	if err := method.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := method.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return method, nil
}

func (shippingMethodManager *BuiltinShippingMethodManager) Close(ctx context.Context) error {
	return nil
}

func (shippingMethodManager *BuiltinShippingMethodManager) ExistsShippingMethod(ctx context.Context, name string) (bool, error) {
	return shippingMethodManager.DB.ExistsShippingMethod(ctx, name)
}

func (shippingMethodManager *BuiltinShippingMethodManager) GetShippingMethodByName(ctx context.Context, name string) (ShippingMethod, error) {
	shippingForm := ShippingMethodForm{}
	id, err := shippingMethodManager.DB.GetShippingMethodByName(ctx, name, &shippingForm)
	if err != nil {
		return nil, err
	}
	return shippingMethodManager.newBuiltinShippingMethod(ctx, id, shippingMethodManager.DB, &shippingForm)
}

func (shippingMethodManager *BuiltinShippingMethodManager) GetShippingMethodCount(ctx context.Context) (uint64, error) {
	return shippingMethodManager.DB.GetShippingMethodCount(ctx)
}

func (shippingMethodManager *BuiltinShippingMethodManager) GetShippingMethods(ctx context.Context, shippingMethodes []ShippingMethod, skip int64, limit int64, queueOrder QueueOrder) ([]ShippingMethod, error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	shippingForms := make([]*ShippingMethodForm, 0, cap(ids))
	ids, shippingForms, err = shippingMethodManager.DB.GetShippingMethods(ctx, ids, shippingForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	methods := shippingMethodes
	if methods == nil {
		methods = make([]ShippingMethod, 0, len(ids))
	}
	for i := range len(ids) {
		method, err := shippingMethodManager.newBuiltinShippingMethod(ctx, ids[i], shippingMethodManager.DB, shippingForms[i])
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func (shippingMethodManager *BuiltinShippingMethodManager) GetShippingMethodWithID(ctx context.Context, sid uint64, fill bool) (ShippingMethod, error) {
	if !fill {
		return shippingMethodManager.newBuiltinShippingMethod(ctx, sid, shippingMethodManager.DB, nil)
	}
	methodForm := ShippingMethodForm{}
	err := shippingMethodManager.DB.FillShippingMethodWithID(ctx, sid, &methodForm)
	if err != nil {
		return nil, err
	}
	return shippingMethodManager.newBuiltinShippingMethod(ctx, sid, shippingMethodManager.DB, &methodForm)
}

func (shippingMethodManager *BuiltinShippingMethodManager) Init(ctx context.Context) error {
	return shippingMethodManager.DB.InitShippingMethodManager(ctx)
}

func (shippingMethodManager *BuiltinShippingMethodManager) NewShippingMethod(ctx context.Context, name string, price float64) (ShippingMethod, error) {
	shippingForm := ShippingMethodForm{}
	id, err := shippingMethodManager.DB.NewShippingMethod(ctx, name, price, &shippingForm)
	if err != nil {
		return nil, err
	}
	return shippingMethodManager.newBuiltinShippingMethod(ctx, id, shippingMethodManager.DB, &shippingForm)
}

func (shippingMethodManager *BuiltinShippingMethodManager) Pulse(ctx context.Context) error {
	return nil
}

func (shippingMethodManager *BuiltinShippingMethodManager) RemoveAllShippingMethods(ctx context.Context) error {
	return shippingMethodManager.DB.RemoveAllShippingMethods(ctx)
}

func (shippingMethodManager *BuiltinShippingMethodManager) RemoveShippingMethod(ctx context.Context, method ShippingMethod) error {
	id, err := method.GetID(ctx)
	if err != nil {
		return err
	}
	return shippingMethodManager.DB.RemoveShippingMethod(ctx, id)
}

func (shippingMethodManager *BuiltinShippingMethodManager) ToBuiltinObject(ctx context.Context) (*BuiltinShippingMethodManager, error) {
	return shippingMethodManager, nil
}

func (shippingMethod *BuiltinShippingMethod) Close(ctx context.Context) error {
	return nil
}

func (shippingMethod *BuiltinShippingMethod) GetID(ctx context.Context) (uint64, error) {
	shippingMethod.MU.RLock()
	defer shippingMethod.MU.RUnlock()
	return shippingMethod.ID, nil
}

func (shippingMethod *BuiltinShippingMethod) GetName(ctx context.Context) (string, error) {
	shippingMethod.MU.RLock()
	if shippingMethod.Name != nil {
		defer shippingMethod.MU.RUnlock()
		return *shippingMethod.Name, nil
	}
	shippingMethod.MU.RUnlock()
	id, err := shippingMethod.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := shippingMethod.ShippingMethodForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := shippingMethod.DB.GetShippingMethodName(ctx, &form, id)
	if err != nil {
		return "", nil
	}
	if err := shippingMethod.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	shippingMethod.MU.Lock()
	defer shippingMethod.MU.Unlock()
	shippingMethod.Name = &name
	return name, nil
}

func (shippingMethod *BuiltinShippingMethod) Init(ctx context.Context) error {
	return nil
}

func (shippingMethod *BuiltinShippingMethod) Pulse(ctx context.Context) error {
	return nil
}

func (shippingMethod *BuiltinShippingMethod) SetName(ctx context.Context, name string) error {
	id, err := shippingMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := shippingMethod.ShippingMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := shippingMethod.DB.SetShippingMethodName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := shippingMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	shippingMethod.MU.Lock()
	defer shippingMethod.MU.Unlock()
	shippingMethod.Name = &name
	return nil
}

func (shippingMethod *BuiltinShippingMethod) GetPrice(ctx context.Context) (float64, error) {
	shippingMethod.MU.RLock()
	if shippingMethod.Price != nil {
		defer shippingMethod.MU.RUnlock()
		return *shippingMethod.Price, nil
	}
	shippingMethod.MU.RUnlock()
	id, err := shippingMethod.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := shippingMethod.ShippingMethodForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	price, err := shippingMethod.DB.GetShippingMethodPrice(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := shippingMethod.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	shippingMethod.MU.Lock()
	defer shippingMethod.MU.Unlock()
	shippingMethod.Price = &price
	return price, nil
}

func (shippingMethod *BuiltinShippingMethod) SetPrice(ctx context.Context, price float64) error {
	id, err := shippingMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := shippingMethod.ShippingMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := shippingMethod.DB.SetShippingMethodPrice(ctx, &form, id, price); err != nil {
		return err
	}
	if err := shippingMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	shippingMethod.MU.Lock()
	defer shippingMethod.MU.Unlock()
	shippingMethod.Price = &price
	return nil
}

func (shippingMethod *BuiltinShippingMethod) ToBuiltinObject(ctx context.Context) (*BuiltinShippingMethod, error) {
	return shippingMethod, nil
}

func (shippingMethod *BuiltinShippingMethod) ToFormObject(ctx context.Context) (*ShippingMethodForm, error) {
	shippingMethod.MU.RLock()
	defer shippingMethod.MU.RUnlock()
	return &shippingMethod.ShippingMethodForm, nil
}

func (shippingMethod *BuiltinShippingMethod) ApplyFormObject(ctx context.Context, form *ShippingMethodForm) error {
	shippingMethod.MU.Lock()
	defer shippingMethod.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		shippingMethod.ID = form.ID
	}
	if form.Name != nil {
		shippingMethod.Name = form.Name
	}
	if form.Price != nil {
		shippingMethod.Price = form.Price
	}
	return nil
}

func (form *ShippingMethodForm) Clone(ctx context.Context) (ShippingMethodForm, error) {
	var cloned ShippingMethodForm = *form
	return cloned, nil
}
