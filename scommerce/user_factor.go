package scommerce

import (
	"context"
	"encoding/json"
	"sync"
)

var _ UserFactorManager[any] = &BuiltinUserFactorManager[any]{}
var _ UserFactor[any] = &BuiltinUserFactor[any]{}

type userFactorDatabase[AccountID comparable] interface {
	DBUserFactor[AccountID]
}

type userFactorManagerDatabase[AccountID comparable] interface {
	DBUserFactorManager[AccountID]
	userFactorDatabase[AccountID]
}

type BuiltinUserFactorManager[AccountID comparable] struct {
	DB userFactorManagerDatabase[AccountID]
}

type UserFactorForm[AccountID comparable] struct {
	ID            uint64           `json:"id"`
	UserAccountID AccountID        `json:"account_id"`
	Products      *json.RawMessage `json:"products,omitempty"`
	Discount      *float64         `json:"discount,omitempty"`
	Tax           *float64         `json:"tax,omitempty"`
	AmountPaid    *float64         `json:"amount_paid,omitempty"`
}

type BuiltinUserFactor[AccountID comparable] struct {
	UserFactorForm[AccountID]
	DB userFactorDatabase[AccountID] `json:"-"`
	MU sync.RWMutex                  `json:"-"`
}

func NewBuiltinUserFactorManager[AccountID comparable](db userFactorManagerDatabase[AccountID]) *BuiltinUserFactorManager[AccountID] {
	return &BuiltinUserFactorManager[AccountID]{
		DB: db,
	}
}

func (factorManager *BuiltinUserFactorManager[AccountID]) newUserFactor(ctx context.Context, fid uint64, aid AccountID, db userFactorDatabase[AccountID], form *UserFactorForm[AccountID]) (*BuiltinUserFactor[AccountID], error) {
	factor := &BuiltinUserFactor[AccountID]{
		UserFactorForm: UserFactorForm[AccountID]{
			ID:            fid,
			UserAccountID: aid,
		},
		DB: db,
	}
	if err := factor.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := factor.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return factor, nil
}

func (factorManager *BuiltinUserFactorManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (factorManager *BuiltinUserFactorManager[AccountID]) GetUserFactorCount(ctx context.Context, account UserAccount[AccountID]) (uint64, error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	return factorManager.DB.GetUserFactorCount(ctx, aid)
}

func (factorManager *BuiltinUserFactorManager[AccountID]) GetFactorWithID(ctx context.Context, fid uint64, fill bool) (UserFactor[AccountID], error) {
	if !fill {
		var zeroAccountID AccountID
		return factorManager.newUserFactor(ctx, fid, zeroAccountID, factorManager.DB, nil)
	}
	factorForm := UserFactorForm[AccountID]{}
	err := factorManager.DB.FillUserFactorWithID(ctx, fid, &factorForm)
	if err != nil {
		return nil, err
	}
	return factorManager.newUserFactor(ctx, fid, factorForm.UserAccountID, factorManager.DB, &factorForm)
}

func (factorManager *BuiltinUserFactorManager[AccountID]) GetUserFactors(ctx context.Context, account UserAccount[AccountID], factors []UserFactor[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserFactor[AccountID], error) {
	var err error = nil
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	factorForms := make([]*UserFactorForm[AccountID], 0, cap(ids))
	ids, factorForms, err = factorManager.DB.GetUserFactors(ctx, aid, ids, factorForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	ftrs := factors
	if ftrs == nil {
		ftrs = make([]UserFactor[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		factor, err := factorManager.newUserFactor(ctx, ids[i], aid, factorManager.DB, factorForms[i])
		if err != nil {
			return nil, err
		}
		ftrs = append(ftrs, factor)
	}
	return ftrs, nil
}

func (factorManager *BuiltinUserFactorManager[AccountID]) Init(ctx context.Context) error {
	return factorManager.DB.InitUserFactorManager(ctx)
}

func (factorManager *BuiltinUserFactorManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (factorManager *BuiltinUserFactorManager[AccountID]) RemoveAllUserFactors(ctx context.Context) error {
	return factorManager.DB.RemoveAllUserFactors(ctx)
}

func (factorManager *BuiltinUserFactorManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserFactorManager[AccountID], error) {
	return factorManager, nil
}

func (factor *BuiltinUserFactor[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) GetID(ctx context.Context) (uint64, error) {
	factor.MU.RLock()
	defer factor.MU.RUnlock()
	return factor.ID, nil
}

func (factor *BuiltinUserFactor[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	factor.MU.RLock()
	defer factor.MU.RUnlock()
	return factor.UserAccountID, nil
}

func (factor *BuiltinUserFactor[AccountID]) GetProducts(ctx context.Context) (json.RawMessage, error) {
	factor.MU.RLock()
	if factor.Products != nil {
		defer factor.MU.RUnlock()
		return *factor.Products, nil
	}
	factor.MU.RUnlock()
	id, err := factor.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	products, err := factor.DB.GetUserFactorProducts(ctx, &form, id)
	if err != nil {
		return nil, err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Products = &products
	return products, nil
}

func (factor *BuiltinUserFactor[AccountID]) SetProducts(ctx context.Context, products json.RawMessage) error {
	id, err := factor.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := factor.DB.SetUserFactorProducts(ctx, &form, id, products); err != nil {
		return err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Products = &products
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) GetDiscount(ctx context.Context) (float64, error) {
	factor.MU.RLock()
	if factor.Discount != nil {
		defer factor.MU.RUnlock()
		return *factor.Discount, nil
	}
	factor.MU.RUnlock()
	id, err := factor.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	discount, err := factor.DB.GetUserFactorDiscount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Discount = &discount
	return discount, nil
}

func (factor *BuiltinUserFactor[AccountID]) SetDiscount(ctx context.Context, discount float64) error {
	id, err := factor.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := factor.DB.SetUserFactorDiscount(ctx, &form, id, discount); err != nil {
		return err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Discount = &discount
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) GetTax(ctx context.Context) (float64, error) {
	factor.MU.RLock()
	if factor.Tax != nil {
		defer factor.MU.RUnlock()
		return *factor.Tax, nil
	}
	factor.MU.RUnlock()
	id, err := factor.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	tax, err := factor.DB.GetUserFactorTax(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Tax = &tax
	return tax, nil
}

func (factor *BuiltinUserFactor[AccountID]) SetTax(ctx context.Context, tax float64) error {
	id, err := factor.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := factor.DB.SetUserFactorTax(ctx, &form, id, tax); err != nil {
		return err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.Tax = &tax
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) GetAmountPaid(ctx context.Context) (float64, error) {
	factor.MU.RLock()
	if factor.AmountPaid != nil {
		defer factor.MU.RUnlock()
		return *factor.AmountPaid, nil
	}
	factor.MU.RUnlock()
	id, err := factor.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	amountPaid, err := factor.DB.GetUserFactorAmountPaid(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.AmountPaid = &amountPaid
	return amountPaid, nil
}

func (factor *BuiltinUserFactor[AccountID]) SetAmountPaid(ctx context.Context, amountPaid float64) error {
	id, err := factor.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := factor.UserFactorForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := factor.DB.SetUserFactorAmountPaid(ctx, &form, id, amountPaid); err != nil {
		return err
	}
	if err := factor.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	factor.MU.Lock()
	defer factor.MU.Unlock()
	factor.AmountPaid = &amountPaid
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (factor *BuiltinUserFactor[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserFactor[AccountID], error) {
	return factor, nil
}

func (factor *BuiltinUserFactor[AccountID]) ToFormObject(ctx context.Context) (*UserFactorForm[AccountID], error) {
	factor.MU.RLock()
	defer factor.MU.RUnlock()
	return &factor.UserFactorForm, nil
}

func (factor *BuiltinUserFactor[AccountID]) ApplyFormObject(ctx context.Context, form *UserFactorForm[AccountID]) error {
	factor.MU.Lock()
	defer factor.MU.Unlock()
	if form.ID != 0 {
		factor.ID = form.ID
	}
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		factor.UserAccountID = form.UserAccountID
	}
	if form.Products != nil {
		factor.Products = form.Products
	}
	if form.Discount != nil {
		factor.Discount = form.Discount
	}
	if form.Tax != nil {
		factor.Tax = form.Tax
	}
	if form.AmountPaid != nil {
		factor.AmountPaid = form.AmountPaid
	}
	return nil
}

func (form *UserFactorForm[AccountID]) Clone(ctx context.Context) (UserFactorForm[AccountID], error) {
	var cloned UserFactorForm[AccountID] = *form
	return cloned, nil
}
