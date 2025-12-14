package scommerce

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce/otp"
)

var ErrExceededMaxRetries = errors.New("exceeded maximum retry attempts")

var _ UserDiscountManager[any] = &BuiltinUserDiscountManager[any]{}
var _ UserDiscount[any] = &BuiltinUserDiscount[any]{}

type userDiscountDatabase[AccountID comparable] interface {
	DBUserDiscount[AccountID]
}

type userDiscountManagerDatabase[AccountID comparable] interface {
	DBUserDiscountManager[AccountID]
	userDiscountDatabase[AccountID]
}

type BuiltinUserDiscountManager[AccountID comparable] struct {
	DB         userDiscountManagerDatabase[AccountID]
	CodeLength int32
	RandSrc    *rand.Rand
	MU         sync.Mutex
}

type UserDiscountForm[AccountID comparable] struct {
	ID            uint64       `json:"id"`
	UserAccountID AccountID    `json:"account_id"`
	Code          *string      `json:"code,omitempty"`
	Value         *float64     `json:"value,omitempty"`
	ValidCount    *int64       `json:"valid_count,omitempty"`
	UsedBy        *[]AccountID `json:"used_by,omitempty"`
}

type BuiltinUserDiscount[AccountID comparable] struct {
	UserDiscountForm[AccountID]
	DB userDiscountDatabase[AccountID] `json:"-"`
	MU sync.RWMutex                    `json:"-"`
}

func NewBuiltinUserDiscountManager[AccountID comparable](db userDiscountManagerDatabase[AccountID], codeLength int32) *BuiltinUserDiscountManager[AccountID] {
	pcg := uint64(time.Now().UnixNano())
	rng := uint64(time.Now().UnixNano() >> 32)
	return &BuiltinUserDiscountManager[AccountID]{
		DB:         db,
		CodeLength: codeLength,
		RandSrc:    rand.New(rand.NewPCG(pcg, rng)),
	}
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) generateUniqueCode(ctx context.Context) (string, error) {
	const maxTry = 1000

	discountManager.MU.Lock()
	defer discountManager.MU.Unlock()

	for try := 0; try < maxTry; try++ {
		code := string(otp.GenerateRandomBytes(discountManager.CodeLength, otp.TokenKeys, discountManager.RandSrc))
		exists, err := discountManager.DB.ExistsUserDiscountCode(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}

	return "", ErrExceededMaxRetries
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) newUserDiscount(ctx context.Context, did uint64, aid AccountID, db userDiscountDatabase[AccountID], form *UserDiscountForm[AccountID]) (*BuiltinUserDiscount[AccountID], error) {
	discount := &BuiltinUserDiscount[AccountID]{
		UserDiscountForm: UserDiscountForm[AccountID]{
			ID:            did,
			UserAccountID: aid,
		},
		DB: db,
	}
	if err := discount.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := discount.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return discount, nil
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) ExistsUserDiscountCode(ctx context.Context, code string) (bool, error) {
	return discountManager.DB.ExistsUserDiscountCode(ctx, code)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) GetUserDiscountByCode(ctx context.Context, code string) (UserDiscount[AccountID], error) {
	discountForm := UserDiscountForm[AccountID]{}
	result, err := discountManager.DB.GetUserDiscountByCode(ctx, code, &discountForm)
	if err != nil {
		return nil, err
	}
	return discountManager.newUserDiscount(ctx, result.ID, result.AID, discountManager.DB, &discountForm)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) GetUserDiscountCount(ctx context.Context) (uint64, error) {
	return discountManager.DB.GetUserDiscountCount(ctx)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) GetUserDiscounts(ctx context.Context, discounts []UserDiscount[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserDiscount[AccountID], error) {
	var err error = nil
	ids := make([]DBUserDiscountResult[AccountID], 0, GetSafeLimit(limit))
	discountForms := make([]*UserDiscountForm[AccountID], 0, cap(ids))
	ids, discountForms, err = discountManager.DB.GetUserDiscounts(ctx, ids, discountForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	discs := discounts
	if discs == nil {
		discs = make([]UserDiscount[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		discount, err := discountManager.newUserDiscount(ctx, ids[i].ID, ids[i].AID, discountManager.DB, discountForms[i])
		if err != nil {
			return nil, err
		}
		discs = append(discs, discount)
	}
	return discs, nil
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) Init(ctx context.Context) error {
	return discountManager.DB.InitUserDiscountManager(ctx)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) NewUserDiscount(ctx context.Context, ownerAccount UserAccount[AccountID], value float64, validCount int64) (UserDiscount[AccountID], error) {
	aid, err := ownerAccount.GetID(ctx)
	if err != nil {
		return nil, err
	}

	code, err := discountManager.generateUniqueCode(ctx)
	if err != nil {
		return nil, err
	}

	discountForm := UserDiscountForm[AccountID]{
		Code:       &code,
		Value:      &value,
		ValidCount: &validCount,
	}
	id, err := discountManager.DB.NewUserDiscount(ctx, aid, value, validCount, &discountForm)
	if err != nil {
		return nil, err
	}
	return discountManager.newUserDiscount(ctx, id, aid, discountManager.DB, &discountForm)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) RemoveAllUserDiscounts(ctx context.Context) error {
	return discountManager.DB.RemoveAllUserDiscounts(ctx)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) RemoveUserDiscount(ctx context.Context, discount UserDiscount[AccountID]) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	return discountManager.DB.RemoveUserDiscount(ctx, id)
}

func (discountManager *BuiltinUserDiscountManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserDiscountManager[AccountID], error) {
	return discountManager, nil
}

func (discount *BuiltinUserDiscount[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) DecrementValidCount(ctx context.Context) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := discount.DB.DecrementUserDiscountValidCount(ctx, &form, id); err != nil {
		return err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.ValidCount = nil
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetCode(ctx context.Context) (string, error) {
	discount.MU.RLock()
	if discount.Code != nil {
		defer discount.MU.RUnlock()
		return *discount.Code, nil
	}
	discount.MU.RUnlock()
	id, err := discount.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	code, err := discount.DB.GetUserDiscountCode(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.Code = &code
	return code, nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetID(ctx context.Context) (uint64, error) {
	discount.MU.RLock()
	defer discount.MU.RUnlock()
	return discount.ID, nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	discount.MU.RLock()
	defer discount.MU.RUnlock()
	return discount.UserAccountID, nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetUsedBy(ctx context.Context) ([]AccountID, error) {
	discount.MU.RLock()
	if discount.UsedBy != nil {
		defer discount.MU.RUnlock()
		return *discount.UsedBy, nil
	}
	discount.MU.RUnlock()
	id, err := discount.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	usedBy, err := discount.DB.GetUserDiscountUsedBy(ctx, &form, id)
	if err != nil {
		return nil, err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.UsedBy = &usedBy
	return usedBy, nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetValidCount(ctx context.Context) (int64, error) {
	discount.MU.RLock()
	if discount.ValidCount != nil {
		defer discount.MU.RUnlock()
		return *discount.ValidCount, nil
	}
	discount.MU.RUnlock()
	id, err := discount.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	validCount, err := discount.DB.GetUserDiscountValidCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.ValidCount = &validCount
	return validCount, nil
}

func (discount *BuiltinUserDiscount[AccountID]) GetValue(ctx context.Context) (float64, error) {
	discount.MU.RLock()
	if discount.Value != nil {
		defer discount.MU.RUnlock()
		return *discount.Value, nil
	}
	discount.MU.RUnlock()
	id, err := discount.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	value, err := discount.DB.GetUserDiscountValue(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.Value = &value
	return value, nil
}

func (discount *BuiltinUserDiscount[AccountID]) HasUserUsed(ctx context.Context, account UserAccount[AccountID]) (bool, error) {
	id, err := discount.GetID(ctx)
	if err != nil {
		return false, err
	}
	aid, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	hasUsed, err := discount.DB.HasUserUsedDiscount(ctx, &form, id, aid)
	if err != nil {
		return false, err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	return hasUsed, nil
}

func (discount *BuiltinUserDiscount[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) MarkAsUsedBy(ctx context.Context, account UserAccount[AccountID]) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	aid, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := discount.DB.AddUserDiscountUsedBy(ctx, &form, id, aid); err != nil {
		return err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.UsedBy = nil
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) SetCode(ctx context.Context, code string) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := discount.DB.SetUserDiscountCode(ctx, &form, id, code); err != nil {
		return err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.Code = &code
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) SetValidCount(ctx context.Context, validCount int64) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := discount.DB.SetUserDiscountValidCount(ctx, &form, id, validCount); err != nil {
		return err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.ValidCount = &validCount
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) SetValue(ctx context.Context, value float64) error {
	id, err := discount.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := discount.UserDiscountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := discount.DB.SetUserDiscountValue(ctx, &form, id, value); err != nil {
		return err
	}
	if err := discount.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	discount.MU.Lock()
	defer discount.MU.Unlock()
	discount.Value = &value
	return nil
}

func (discount *BuiltinUserDiscount[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserDiscount[AccountID], error) {
	return discount, nil
}

func (discount *BuiltinUserDiscount[AccountID]) ToFormObject(ctx context.Context) (*UserDiscountForm[AccountID], error) {
	discount.MU.RLock()
	defer discount.MU.RUnlock()
	return &discount.UserDiscountForm, nil
}

func (discount *BuiltinUserDiscount[AccountID]) ApplyFormObject(ctx context.Context, form *UserDiscountForm[AccountID]) error {
	discount.MU.Lock()
	defer discount.MU.Unlock()
	if form.ID != 0 {
		discount.ID = form.ID
	}
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		discount.UserAccountID = form.UserAccountID
	}
	if form.Code != nil {
		discount.Code = form.Code
	}
	if form.Value != nil {
		discount.Value = form.Value
	}
	if form.ValidCount != nil {
		discount.ValidCount = form.ValidCount
	}
	if form.UsedBy != nil {
		discount.UsedBy = form.UsedBy
	}
	return nil
}

func (form *UserDiscountForm[AccountID]) Clone(ctx context.Context) (UserDiscountForm[AccountID], error) {
	var cloned UserDiscountForm[AccountID] = *form
	return cloned, nil
}
