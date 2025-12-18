package scommerce

import (
	"context"
	"sync"
	"time"
)

var _ UserPaymentMethodManager[any] = &BuiltinUserPaymentMethodManager[any]{}
var _ UserPaymentMethod[any] = &BuiltinUserPaymentMethod[any]{}

type userPaymentMethodDatabase[AccountID comparable] interface {
	DBUserPaymentMethod[AccountID]
	DBPaymentType
}

type userPaymentMethodManagerDatabase[AccountID comparable] interface {
	DBUserPaymentMethodManager[AccountID]
	userPaymentMethodDatabase[AccountID]
}

type BuiltinUserPaymentMethodManager[AccountID comparable] struct {
	DB userPaymentMethodManagerDatabase[AccountID]
}

type UserPaymentMethodForm[AccountID any] struct {
	ID             uint64              `json:"id"`
	UserAccountID  AccountID           `json:"account_id"`
	PaymentType    *BuiltinPaymentType `json:"payment_Type,omitempty"`
	Provider       *string             `json:"provider,omitempty"`
	AccountNumber  *string             `json:"account_number,omitempty"`
	ExpiryDate     *time.Time          `json:"expiry_date,omitempty"`
	IsExpiredState *bool               `json:"is_expired,omitempty"`
	IsDefaultState *bool               `json:"is_default,omitempty"`
}

type BuiltinUserPaymentMethod[AccountID comparable] struct {
	UserPaymentMethodForm[AccountID]
	DB userPaymentMethodDatabase[AccountID] `json:"-"`
	MU sync.RWMutex                         `json:"-"`
}

func NewBuiltinPaymentMethodManager[AccountID comparable](db userPaymentMethodManagerDatabase[AccountID]) *BuiltinUserPaymentMethodManager[AccountID] {
	return &BuiltinUserPaymentMethodManager[AccountID]{
		DB: db,
	}
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) GetPaymentMethodCount(ctx context.Context) (uint64, error) {
	return paymentMethodManager.DB.GetUserPaymentMethodCount(ctx)
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) GetPaymentMethodWithID(ctx context.Context, pid uint64, fill bool) (UserPaymentMethod[AccountID], error) {
	if !fill {
		var zeroAccountID AccountID
		return paymentMethodManager.newPaymentMethod(ctx, pid, zeroAccountID, paymentMethodManager.DB, nil)
	}
	paymentMethodForm := UserPaymentMethodForm[AccountID]{}
	err := paymentMethodManager.DB.FillUserPaymentMethodWithID(ctx, pid, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	return paymentMethodManager.newPaymentMethod(ctx, pid, paymentMethodForm.UserAccountID, paymentMethodManager.DB, &paymentMethodForm)
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) newPaymentMethod(ctx context.Context, pid uint64, aid AccountID, db userPaymentMethodDatabase[AccountID], form *UserPaymentMethodForm[AccountID]) (*BuiltinUserPaymentMethod[AccountID], error) {
	method := &BuiltinUserPaymentMethod[AccountID]{
		UserPaymentMethodForm: UserPaymentMethodForm[AccountID]{
			ID:            pid,
			UserAccountID: aid,
		},
		DB: db,
	}
	if err := method.Init(ctx); err != nil {
		return nil, err
	}
	if err := method.ApplyFormObject(ctx, form); err != nil {
		return nil, err
	}
	return method, nil
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) GetPaymentMethods(ctx context.Context, paymentMethods []UserPaymentMethod[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserPaymentMethod[AccountID], error) {
	var err error = nil
	ids := make([]DBUserPaymentMethodResult[AccountID], 0, GetSafeLimit(limit))
	methodForms := make([]*UserPaymentMethodForm[AccountID], 0, cap(ids))
	ids, methodForms, err = paymentMethodManager.DB.GetUserPaymentMethods(ctx, ids, methodForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	methods := paymentMethods
	if methods == nil {
		methods = make([]UserPaymentMethod[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		res := ids[i]
		method, err := paymentMethodManager.newPaymentMethod(ctx, res.ID, res.AID, paymentMethodManager.DB, methodForms[i])
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) Init(ctx context.Context) error {
	return paymentMethodManager.DB.InitUserPaymentMethodManager(ctx)
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) RemoveAllPaymentMethods(ctx context.Context) error {
	return paymentMethodManager.DB.RemoveAllUserPaymentMethods(ctx)
}

func (paymentMethodManager *BuiltinUserPaymentMethodManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserPaymentMethodManager[AccountID], error) {
	return paymentMethodManager, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetAccountNumber(ctx context.Context) (string, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.AccountNumber != nil {
		defer paymentMethod.MU.RUnlock()
		return *paymentMethod.AccountNumber, nil
	}
	paymentMethod.MU.RUnlock()
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	number, err := paymentMethod.DB.GetUserPaymentMethodAccountNumber(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.AccountNumber = &number
	return number, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetExpiryDate(ctx context.Context) (time.Time, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.ExpiryDate != nil {
		defer paymentMethod.MU.RUnlock()
		return *paymentMethod.ExpiryDate, nil
	}
	paymentMethod.MU.RUnlock()
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return time.Time{}, nil
	}
	date, err := paymentMethod.DB.GetUserPaymentMethodExpiryDate(ctx, &form, id)
	if err != nil {
		return time.Time{}, err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.ExpiryDate = &date
	return date, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetID(ctx context.Context) (uint64, error) {
	paymentMethod.MU.RLock()
	defer paymentMethod.MU.RUnlock()
	return paymentMethod.ID, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetPaymentType(ctx context.Context) (PaymentType, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.PaymentType != nil {
		defer paymentMethod.MU.RUnlock()
		return paymentMethod.PaymentType, nil
	}
	paymentMethod.MU.RUnlock()

	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return nil, err
	}

	typeForm := PaymentTypeForm{}
	pid, err := paymentMethod.DB.GetUserPaymentMethodPaymentType(ctx, &form, id, &typeForm)
	if err != nil {
		return nil, err
	}

	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}

	pType := &BuiltinPaymentType{
		DB: paymentMethod.DB,
		PaymentTypeForm: PaymentTypeForm{
			ID: pid,
		},
	}
	if err := pType.Init(ctx); err != nil {
		return nil, err
	}
	if err := pType.ApplyFormObject(ctx, &typeForm); err != nil {
		return nil, err
	}

	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()

	paymentMethod.PaymentType = pType

	return pType, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetProvider(ctx context.Context) (string, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.Provider != nil {
		defer paymentMethod.MU.RUnlock()
		return *paymentMethod.Provider, nil
	}
	paymentMethod.MU.RUnlock()
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	provider, err := paymentMethod.DB.GetUserPaymentMethodProvider(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.Provider = &provider
	return provider, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	paymentMethod.MU.RLock()
	defer paymentMethod.MU.RUnlock()
	return paymentMethod.UserAccountID, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) IsDefault(ctx context.Context) (bool, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.IsDefaultState != nil {
		defer paymentMethod.MU.RUnlock()
		return *paymentMethod.IsDefaultState, nil
	}
	paymentMethod.MU.RUnlock()
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isDefault, err := paymentMethod.DB.IsUserPaymentMethodDefault(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.IsDefaultState = &isDefault
	return isDefault, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) IsExpired(ctx context.Context) (bool, error) {
	paymentMethod.MU.RLock()
	if paymentMethod.IsExpiredState != nil {
		defer paymentMethod.MU.RUnlock()
		return *paymentMethod.IsExpiredState, nil
	}
	paymentMethod.MU.RUnlock()
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isExpired, err := paymentMethod.DB.IsUserPaymentMethodExpired(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.IsExpiredState = &isExpired
	return isExpired, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) SetAccountNumber(ctx context.Context, accountNumber string) error {
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentMethod.DB.SetUserPaymentMethodAccountNumber(ctx, &form, id, accountNumber); err != nil {
		return err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.AccountNumber = &accountNumber
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) SetDefault(ctx context.Context, state bool) error {
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentMethod.DB.SetUserPaymentMethodDefault(ctx, &form, id, state); err != nil {
		return err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.IsDefaultState = &state
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) SetExpiryDate(ctx context.Context, date time.Time) error {
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentMethod.DB.SetUserPaymentMethodExpiryDate(ctx, &form, id, date); err != nil {
		return err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.ExpiryDate = &date
	paymentMethod.IsExpiredState = nil
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) SetPaymentType(ctx context.Context, paymentType PaymentType) error {
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	pid, err := paymentType.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentMethod.DB.SetUserPaymentMethodPaymentType(ctx, &form, id, pid); err != nil {
		return err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.PaymentType, err = paymentType.ToBuiltinObject(ctx)
	return err
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) SetProvider(ctx context.Context, provider string) error {
	id, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := paymentMethod.UserPaymentMethodForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := paymentMethod.DB.SetUserPaymentMethodProvider(ctx, &form, id, provider); err != nil {
		return err
	}
	if err := paymentMethod.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	paymentMethod.Provider = &provider
	return nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserPaymentMethod[AccountID], error) {
	return paymentMethod, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) ToFormObject(ctx context.Context) (*UserPaymentMethodForm[AccountID], error) {
	paymentMethod.MU.RLock()
	defer paymentMethod.MU.RUnlock()
	return &paymentMethod.UserPaymentMethodForm, nil
}

func (paymentMethod *BuiltinUserPaymentMethod[AccountID]) ApplyFormObject(ctx context.Context, form *UserPaymentMethodForm[AccountID]) error {
	paymentMethod.MU.Lock()
	defer paymentMethod.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		paymentMethod.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		paymentMethod.UserAccountID = form.UserAccountID
	}
	if form.PaymentType != nil {
		paymentMethod.PaymentType = form.PaymentType
	}
	if form.Provider != nil {
		paymentMethod.Provider = form.Provider
	}
	if form.AccountNumber != nil {
		paymentMethod.AccountNumber = form.AccountNumber
	}
	if form.ExpiryDate != nil {
		paymentMethod.ExpiryDate = form.ExpiryDate
	}
	if form.IsExpiredState != nil {
		paymentMethod.IsExpiredState = form.IsExpiredState
	}
	if form.IsDefaultState != nil {
		paymentMethod.IsDefaultState = form.IsDefaultState
	}
	return nil
}

func (form *UserPaymentMethodForm[AccountID]) Clone(ctx context.Context) (UserPaymentMethodForm[AccountID], error) {
	var cloned UserPaymentMethodForm[AccountID] = *form
	return cloned, nil
}
