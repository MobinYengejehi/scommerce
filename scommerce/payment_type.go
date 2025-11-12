package scommerce

import (
	"context"
	"sync"
)

type paymentTypeDatabase interface {
	DBPaymentTypeManager
	DBPaymentType
}

var _ PaymentTypeManager = &BuiltinPaymentTypeManager{}
var _ PaymentType = &BuiltinPaymentType{}

type BuiltinPaymentTypeManager struct {
	DB paymentTypeDatabase
}

type PaymentTypeForm struct {
	ID   uint64  `json:"id"`
	Name *string `json:"name,omitempty"`
}

type BuiltinPaymentType struct {
	PaymentTypeForm
	DB DBPaymentType `json:"-"`
	MU sync.RWMutex  `json:"-"`
}

func NewBuiltinPaymentTypeManager(db paymentTypeDatabase) *BuiltinPaymentTypeManager {
	return &BuiltinPaymentTypeManager{
		DB: db,
	}
}

func (paymentTypeManager *BuiltinPaymentTypeManager) newBuiltinPaymentType(ctx context.Context, id uint64, db DBPaymentType, form *PaymentTypeForm) (*BuiltinPaymentType, error) {
	paymentType := &BuiltinPaymentType{
		PaymentTypeForm: PaymentTypeForm{
			ID: id,
		},
		DB: db,
	}
	if err := paymentType.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := paymentType.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return paymentType, nil
}

func (paymentTypeManager *BuiltinPaymentTypeManager) Close(ctx context.Context) error {
	return nil
}

func (paymentTypeManager *BuiltinPaymentTypeManager) ExistsPaymentType(ctx context.Context, name string) (bool, error) {
	return paymentTypeManager.DB.ExistsPaymentType(ctx, name)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) GetPaymentTypeByName(ctx context.Context, name string) (PaymentType, error) {
	typeForm := PaymentTypeForm{}
	id, err := paymentTypeManager.DB.GetPaymentTypeByName(ctx, name, &typeForm)
	if err != nil {
		return nil, err
	}
	return paymentTypeManager.newBuiltinPaymentType(ctx, id, paymentTypeManager.DB, &typeForm)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) GetPaymentTypeCount(ctx context.Context) (uint64, error) {
	return paymentTypeManager.DB.GetPaymentTypeCount(ctx)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) GetPaymentTypes(ctx context.Context, paymentTypes []PaymentType, skip int64, limit int64, queueOrder QueueOrder) ([]PaymentType, error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	typeForms := make([]*PaymentTypeForm, 0, cap(ids))
	ids, typeForms, err = paymentTypeManager.DB.GetPaymentTypes(ctx, ids, typeForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	types := paymentTypes
	if types == nil {
		types = make([]PaymentType, 0, len(ids))
	}
	for i := range len(ids) {
		pType, err := paymentTypeManager.newBuiltinPaymentType(ctx, ids[i], paymentTypeManager.DB, typeForms[i])
		if err != nil {
			return nil, err
		}
		types = append(types, pType)
	}
	return types, nil
}

func (paymentTypeManager *BuiltinPaymentTypeManager) Init(ctx context.Context) error {
	return paymentTypeManager.DB.InitPaymentTypeManager(ctx)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) NewPaymentType(ctx context.Context, name string) (PaymentType, error) {
	typeForm := PaymentTypeForm{}
	id, err := paymentTypeManager.DB.NewPaymentType(ctx, name, &typeForm)
	if err != nil {
		return nil, err
	}
	return paymentTypeManager.newBuiltinPaymentType(ctx, id, paymentTypeManager.DB, &typeForm)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) Pulse(ctx context.Context) error {
	return nil
}

func (paymentTypeManager *BuiltinPaymentTypeManager) RemoveAllPaymentTypes(ctx context.Context) error {
	return paymentTypeManager.DB.RemoveAllPaymentTypes(ctx)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) RemovePaymentType(ctx context.Context, status PaymentType) error {
	id, err := status.GetID(ctx)
	if err != nil {
		return err
	}
	return paymentTypeManager.DB.RemovePaymentType(ctx, id)
}

func (paymentTypeManager *BuiltinPaymentTypeManager) ToBuiltinObject(ctx context.Context) (*BuiltinPaymentTypeManager, error) {
	return paymentTypeManager, nil
}

func (paymentType *BuiltinPaymentType) Close(ctx context.Context) error {
	return nil
}

func (paymentType *BuiltinPaymentType) GetID(ctx context.Context) (uint64, error) {
	paymentType.MU.RLock()
	defer paymentType.MU.RUnlock()
	return paymentType.ID, nil
}

func (paymentType *BuiltinPaymentType) GetName(ctx context.Context) (string, error) {
	paymentType.MU.RLock()
	if paymentType.Name != nil {
		defer paymentType.MU.RUnlock()
		return *paymentType.Name, nil
	}
	paymentType.MU.RUnlock()
	id, err := paymentType.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := paymentType.PaymentTypeForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := paymentType.DB.GetPaymentTypeName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := paymentType.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	paymentType.MU.Lock()
	defer paymentType.MU.Unlock()
	paymentType.Name = &name
	return name, nil
}

func (paymentType *BuiltinPaymentType) Init(ctx context.Context) error {
	return nil
}

func (paymentType *BuiltinPaymentType) Pulse(ctx context.Context) error {
	return nil
}

func (paymentType *BuiltinPaymentType) SetName(ctx context.Context, name string) error {
	id, err := paymentType.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentType.PaymentTypeForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentType.DB.SetPaymentTypeName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := paymentType.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentType.MU.Lock()
	defer paymentType.MU.Unlock()
	paymentType.Name = &name
	return nil
}

func (paymentType *BuiltinPaymentType) ToBuiltinObject(ctx context.Context) (*BuiltinPaymentType, error) {
	return paymentType, nil
}

func (paymentType *BuiltinPaymentType) ToFormObject(ctx context.Context) (*PaymentTypeForm, error) {
	paymentType.MU.RLock()
	defer paymentType.MU.RUnlock()
	return &paymentType.PaymentTypeForm, nil
}

func (paymentType *BuiltinPaymentType) ApplyFormObject(ctx context.Context, form *PaymentTypeForm) error {
	paymentType.MU.Lock()
	defer paymentType.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		paymentType.ID = form.ID
	}
	if form.Name != nil {
		paymentType.Name = form.Name
	}
	return nil
}

func (form *PaymentTypeForm) Clone(ctx context.Context) (PaymentTypeForm, error) {
	var cloned PaymentTypeForm = *form
	return cloned, nil
}
