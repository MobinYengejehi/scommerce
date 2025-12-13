package scommerce

import (
	"context"
	"sync"
	"time"
)

type productItemSubscriptionDatabase[AccountID comparable] interface {
	DBProductItemSubscriptionManager[AccountID]
	DBProductItemSubscription[AccountID]
	DBUserAccount[AccountID]
	DBProductItem[AccountID]
	DBProductCategory[AccountID]
	DBProduct[AccountID]
	DBUserReview[AccountID]
	// Interfaces needed for userAccountDatabase compatibility
	userAddressDatabase[AccountID]
	userPaymentMethodDatabase[AccountID]
	userOrderDatabase[AccountID]
	userShoppingCartDatabase[AccountID]
	userFactorDatabase[AccountID]
	DBUserRole
}

var _ ProductItemSubscriptionManager[any] = &BuiltinProductItemSubscriptionManager[any]{}
var _ ProductItemSubscription[any] = &BuiltinProductItemSubscription[any]{}

type ProductItemSubscriptionForm[AccountID comparable] struct {
	ID               uint64                         `json:"id"`
	UserAccountID    AccountID                      `json:"user_account_id"`
	ProductItem      *BuiltinProductItem[AccountID] `json:"product_item,omitempty"`
	SubscribedAt     *time.Time                     `json:"subscribed_at,omitempty"`
	ExpiresAt        *time.Time                     `json:"expires_at,omitempty"`
	Duration         *time.Duration                 `json:"duration,omitempty"`
	SubscriptionType *string                        `json:"subscription_type,omitempty"`
	AutoRenew        *bool                          `json:"auto_renew,omitempty"`
	IsActive         *bool                          `json:"is_active,omitempty"`
}

type BuiltinProductItemSubscriptionManager[AccountID comparable] struct {
	DB             productItemSubscriptionDatabase[AccountID]
	FS             FileStorage
	RenewalHandler RenewalHandlerFunc[AccountID]
	MU             sync.RWMutex
}

type BuiltinProductItemSubscription[AccountID comparable] struct {
	ProductItemSubscriptionForm[AccountID]
	DB productItemSubscriptionDatabase[AccountID] `json:"-"`
	FS FileStorage                                `json:"-"`
	MU sync.RWMutex                               `json:"-"`
}

func NewBuiltinProductItemSubscriptionManager[AccountID comparable](db productItemSubscriptionDatabase[AccountID], fs FileStorage, renewalHandler RenewalHandlerFunc[AccountID]) *BuiltinProductItemSubscriptionManager[AccountID] {
	manager := &BuiltinProductItemSubscriptionManager[AccountID]{
		DB:             db,
		FS:             fs,
		RenewalHandler: renewalHandler,
	}
	if manager.RenewalHandler == nil {
		manager.RenewalHandler = manager.defaultRenewalHandler
	}
	return manager
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) defaultRenewalHandler(ctx context.Context, subscription ProductItemSubscription[AccountID], account UserAccount[AccountID], productItem ProductItem[AccountID]) (bool, float64, error) {
	price, err := productItem.GetPrice(ctx)
	if err != nil {
		return false, 0, err
	}

	walletBalance, err := account.GetWalletCurrency(ctx)
	if err != nil {
		return false, 0, err
	}

	if walletBalance < price {
		return false, 0, nil
	}

	expiresAt, err := subscription.GetExpiresAt(ctx)
	if err != nil {
		return false, 0, err
	}

	duration, err := subscription.GetDuration(ctx)
	if err != nil {
		return false, 0, err
	}

	newExpiresAt := expiresAt.Add(duration)
	if err := subscription.SetExpiresAt(ctx, newExpiresAt); err != nil {
		return false, 0, err
	}

	return true, price, nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) newBuiltinProductItemSubscription(ctx context.Context, id uint64, aid AccountID, db productItemSubscriptionDatabase[AccountID], form *ProductItemSubscriptionForm[AccountID]) (*BuiltinProductItemSubscription[AccountID], error) {
	subscription := &BuiltinProductItemSubscription[AccountID]{
		ProductItemSubscriptionForm: ProductItemSubscriptionForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
		DB: db,
		FS: manager.FS,
	}
	if err := subscription.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := subscription.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return subscription, nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetProductItemSubscriptionCount(ctx context.Context, productItem ProductItem[AccountID]) (uint64, error) {
	productItemID, err := productItem.GetID(ctx)
	if err != nil {
		return 0, err
	}
	return manager.DB.GetProductItemSubscriptionCountForProduct(ctx, productItemID)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetProductItemSubscriptions(ctx context.Context, productItem ProductItem[AccountID], subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error) {
	productItemID, err := productItem.GetID(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, GetSafeLimit(limit))
	forms := make([]*ProductItemSubscriptionForm[AccountID], 0, cap(ids))
	ids, forms, err = manager.DB.GetProductItemSubscriptionsForProduct(ctx, productItemID, ids, forms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}

	results := subscriptions
	if results == nil {
		results = make([]ProductItemSubscription[AccountID], 0, len(ids))
	}

	for i := range len(ids) {
		var aid AccountID
		if forms[i] != nil {
			aid = forms[i].UserAccountID
		}
		sub, err := manager.newBuiltinProductItemSubscription(ctx, ids[i], aid, manager.DB, forms[i])
		if err != nil {
			return nil, err
		}
		results = append(results, sub)
	}

	return results, nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetSubscriptionCount(ctx context.Context) (uint64, error) {
	return manager.DB.GetProductItemSubscriptionCount(ctx)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetSubscriptions(ctx context.Context, subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error) {
	ids := make([]uint64, 0, GetSafeLimit(limit))
	forms := make([]*ProductItemSubscriptionForm[AccountID], 0, cap(ids))
	ids, forms, err := manager.DB.GetProductItemSubscriptions(ctx, ids, forms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}

	results := subscriptions
	if results == nil {
		results = make([]ProductItemSubscription[AccountID], 0, len(ids))
	}

	for i := range len(ids) {
		var aid AccountID
		if forms[i] != nil {
			aid = forms[i].UserAccountID
		}
		sub, err := manager.newBuiltinProductItemSubscription(ctx, ids[i], aid, manager.DB, forms[i])
		if err != nil {
			return nil, err
		}
		results = append(results, sub)
	}

	return results, nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetUserSubscriptionCount(ctx context.Context, account UserAccount[AccountID]) (uint64, error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	return manager.DB.GetUserProductItemSubscriptionCount(ctx, aid)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) GetUserSubscriptions(ctx context.Context, account UserAccount[AccountID], subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, GetSafeLimit(limit))
	forms := make([]*ProductItemSubscriptionForm[AccountID], 0, cap(ids))
	ids, forms, err = manager.DB.GetUserProductItemSubscriptions(ctx, aid, ids, forms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}

	results := subscriptions
	if results == nil {
		results = make([]ProductItemSubscription[AccountID], 0, len(ids))
	}

	for i := range len(ids) {
		sub, err := manager.newBuiltinProductItemSubscription(ctx, ids[i], aid, manager.DB, forms[i])
		if err != nil {
			return nil, err
		}
		results = append(results, sub)
	}

	return results, nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) Init(ctx context.Context) error {
	return manager.DB.InitProductItemSubscriptionManager(ctx)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) NewSubscription(ctx context.Context, account UserAccount[AccountID], productItem ProductItem[AccountID], duration time.Duration, subscriptionType string, autoRenew bool) (ProductItemSubscription[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}

	productItemID, err := productItem.GetID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(duration)

	form := ProductItemSubscriptionForm[AccountID]{}
	id, err := manager.DB.NewProductItemSubscription(ctx, aid, productItemID, now, expiresAt, duration, subscriptionType, autoRenew, &form)
	if err != nil {
		return nil, err
	}

	return manager.newBuiltinProductItemSubscription(ctx, id, aid, manager.DB, &form)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) Pulse(ctx context.Context) error {
	return manager.ProcessExpiredSubscriptions(ctx)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) ProcessExpiredSubscriptions(ctx context.Context) error {
	now := time.Now()

	ids := make([]uint64, 0, 100)
	forms := make([]*ProductItemSubscriptionForm[AccountID], 0, 100)
	ids, forms, err := manager.DB.GetExpiredSubscriptionsForRenewal(ctx, now, ids, forms, 1000)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, 10)
	errChan := make(chan error, len(ids))
	var wg sync.WaitGroup

	for i := range len(ids) {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var aid AccountID
			if forms[idx] != nil {
				aid = forms[idx].UserAccountID
			}

			subscription, err := manager.newBuiltinProductItemSubscription(ctx, ids[idx], aid, manager.DB, forms[idx])
			if err != nil {
				errChan <- err
				return
			}

			accountForm := UserAccountForm[AccountID]{}
			accountID, err := manager.DB.GetProductItemSubscriptionUserAccountID(ctx, forms[idx], ids[idx])
			if err != nil {
				errChan <- err
				return
			}

			account := &BuiltinUserAccount[AccountID]{
				UserAccountForm: UserAccountForm[AccountID]{
					ID: accountID,
				},
				DB: manager.DB, // Now compatible - productItemSubscriptionDatabase implements all required interfaces
				FS: manager.FS,
			}
			if err := account.ApplyFormObject(ctx, &accountForm); err != nil {
				errChan <- err
				return
			}

			productItemID, err := subscription.GetProductItem(ctx)
			if err != nil {
				errChan <- err
				return
			}

			success, amountCharged, err := manager.RenewalHandler(ctx, subscription, account, productItemID)
			if err != nil {
				errChan <- err
				return
			}

			if !success {
				// Renewal failed but no error - insufficient funds or other business reason
				// Subscription remains in expired state for retry on next pulse
				return
			}

			// Deduct the charged amount from user's wallet if amount > 0
			if amountCharged > 0 {
				if err := account.ChargeWallet(ctx, -amountCharged); err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) RemoveAllSubscriptions(ctx context.Context) error {
	return manager.DB.RemoveAllProductItemSubscriptions(ctx)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) RemoveSubscription(ctx context.Context, subscription ProductItemSubscription[AccountID]) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}
	return manager.DB.RemoveProductItemSubscription(ctx, id)
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) SetRenewalHandler(ctx context.Context, handler RenewalHandlerFunc[AccountID]) error {
	manager.MU.Lock()
	defer manager.MU.Unlock()
	manager.RenewalHandler = handler
	return nil
}

func (manager *BuiltinProductItemSubscriptionManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProductItemSubscriptionManager[AccountID], error) {
	return manager, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) ApplyFormObject(ctx context.Context, form *ProductItemSubscriptionForm[AccountID]) error {
	subscription.MU.Lock()
	defer subscription.MU.Unlock()

	if form.ID != 0 {
		subscription.ID = form.ID
	}
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		subscription.UserAccountID = form.UserAccountID
	}
	if form.ProductItem != nil {
		subscription.ProductItem = form.ProductItem
	}
	if form.SubscribedAt != nil {
		subscription.SubscribedAt = form.SubscribedAt
	}
	if form.ExpiresAt != nil {
		subscription.ExpiresAt = form.ExpiresAt
	}
	if form.Duration != nil {
		subscription.Duration = form.Duration
	}
	if form.SubscriptionType != nil {
		subscription.SubscriptionType = form.SubscriptionType
	}
	if form.AutoRenew != nil {
		subscription.AutoRenew = form.AutoRenew
	}
	// Note: IsActive is a method, not a field - skip assignment

	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) Cancel(ctx context.Context) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.CancelProductItemSubscription(ctx, &form, id); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	falseVal := false
	subscription.AutoRenew = &falseVal
	// Note: IsActive should be set via SetActive method, not direct assignment

	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetDuration(ctx context.Context) (time.Duration, error) {
	subscription.MU.RLock()
	if subscription.Duration != nil {
		defer subscription.MU.RUnlock()
		return *subscription.Duration, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return 0, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return 0, err
	}

	duration, err := subscription.DB.GetProductItemSubscriptionDuration(ctx, &form, id)
	if err != nil {
		return 0, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.Duration = &duration
	return duration, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetExpiresAt(ctx context.Context) (time.Time, error) {
	subscription.MU.RLock()
	if subscription.ExpiresAt != nil {
		defer subscription.MU.RUnlock()
		return *subscription.ExpiresAt, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return time.Time{}, err
	}

	expiresAt, err := subscription.DB.GetProductItemSubscriptionExpiresAt(ctx, &form, id)
	if err != nil {
		return time.Time{}, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ExpiresAt = &expiresAt
	return expiresAt, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetID(ctx context.Context) (uint64, error) {
	subscription.MU.RLock()
	defer subscription.MU.RUnlock()
	return subscription.ID, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetProductItem(ctx context.Context) (ProductItem[AccountID], error) {
	subscription.MU.RLock()
	if subscription.ProductItem != nil {
		defer subscription.MU.RUnlock()
		return subscription.ProductItem, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return nil, err
	}

	productItemForm := ProductItemForm[AccountID]{}
	productItemID, err := subscription.DB.GetProductItemSubscriptionProductItem(ctx, &form, id, &productItemForm, subscription.FS)
	if err != nil {
		return nil, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}

	// Create product item with minimal DB interface - only use methods from DBProductItem
	productItem := &BuiltinProductItem[AccountID]{
		DB: subscription.DB, // subscription.DB implements DBProductItem[AccountID]
		FS: subscription.FS,
		ProductItemForm: ProductItemForm[AccountID]{
			ID: productItemID,
		},
	}
	if err := productItem.Init(ctx); err != nil {
		return nil, err
	}
	if err := productItem.ApplyFormObject(ctx, &productItemForm); err != nil {
		return nil, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ProductItem = productItem
	return productItem, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetSubscribedAt(ctx context.Context) (time.Time, error) {
	subscription.MU.RLock()
	if subscription.SubscribedAt != nil {
		defer subscription.MU.RUnlock()
		return *subscription.SubscribedAt, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return time.Time{}, err
	}

	subscribedAt, err := subscription.DB.GetProductItemSubscriptionSubscribedAt(ctx, &form, id)
	if err != nil {
		return time.Time{}, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.SubscribedAt = &subscribedAt
	return subscribedAt, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetSubscriptionType(ctx context.Context) (string, error) {
	subscription.MU.RLock()
	if subscription.SubscriptionType != nil {
		defer subscription.MU.RUnlock()
		return *subscription.SubscriptionType, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return "", err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return "", err
	}

	subscriptionType, err := subscription.DB.GetProductItemSubscriptionType(ctx, &form, id)
	if err != nil {
		return "", err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.SubscriptionType = &subscriptionType
	return subscriptionType, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	subscription.MU.RLock()
	defer subscription.MU.RUnlock()
	return subscription.UserAccountID, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) IsActive(ctx context.Context) (bool, error) {
	subscription.MU.RLock()
	if subscription.ProductItemSubscriptionForm.IsActive != nil {
		defer subscription.MU.RUnlock()
		return *subscription.ProductItemSubscriptionForm.IsActive, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return false, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return false, err
	}

	isActive, err := subscription.DB.IsProductItemSubscriptionActive(ctx, &form, id)
	if err != nil {
		return false, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ProductItemSubscriptionForm.IsActive = &isActive
	return isActive, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) IsAutoRenew(ctx context.Context) (bool, error) {
	subscription.MU.RLock()
	if subscription.AutoRenew != nil {
		defer subscription.MU.RUnlock()
		return *subscription.AutoRenew, nil
	}
	subscription.MU.RUnlock()

	id, err := subscription.GetID(ctx)
	if err != nil {
		return false, err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return false, err
	}

	autoRenew, err := subscription.DB.IsProductItemSubscriptionAutoRenew(ctx, &form, id)
	if err != nil {
		return false, err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.AutoRenew = &autoRenew
	return autoRenew, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) IsExpired(ctx context.Context) (bool, error) {
	expiresAt, err := subscription.GetExpiresAt(ctx)
	if err != nil {
		return false, err
	}
	return time.Now().After(expiresAt), nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetActive(ctx context.Context, isActive bool) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionActive(ctx, &form, id, isActive); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ProductItemSubscriptionForm.IsActive = &isActive
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetAutoRenew(ctx context.Context, autoRenew bool) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionAutoRenew(ctx, &form, id, autoRenew); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.AutoRenew = &autoRenew
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetDuration(ctx context.Context, duration time.Duration) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionDuration(ctx, &form, id, duration); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.Duration = &duration
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetExpiresAt(ctx context.Context, expiresAt time.Time) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionExpiresAt(ctx, &form, id, expiresAt); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ExpiresAt = &expiresAt
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetProductItem(ctx context.Context, productItem ProductItem[AccountID]) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	productItemID, err := productItem.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionProductItem(ctx, &form, id, productItemID, subscription.FS); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	builtinProductItem, err := productItem.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.ProductItem = builtinProductItem
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetSubscribedAt(ctx context.Context, subscribedAt time.Time) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionSubscribedAt(ctx, &form, id, subscribedAt); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.SubscribedAt = &subscribedAt
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) SetSubscriptionType(ctx context.Context, subscriptionType string) error {
	id, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}

	form, err := subscription.ProductItemSubscriptionForm.Clone(ctx)
	if err != nil {
		return err
	}

	if err := subscription.DB.SetProductItemSubscriptionType(ctx, &form, id, subscriptionType); err != nil {
		return err
	}

	if err := subscription.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	subscription.MU.Lock()
	defer subscription.MU.Unlock()
	subscription.SubscriptionType = &subscriptionType
	return nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinProductItemSubscription[AccountID], error) {
	return subscription, nil
}

func (subscription *BuiltinProductItemSubscription[AccountID]) ToFormObject(ctx context.Context) (*ProductItemSubscriptionForm[AccountID], error) {
	subscription.MU.RLock()
	defer subscription.MU.RUnlock()
	return &subscription.ProductItemSubscriptionForm, nil
}

func (form *ProductItemSubscriptionForm[AccountID]) Clone(ctx context.Context) (ProductItemSubscriptionForm[AccountID], error) {
	var cloned ProductItemSubscriptionForm[AccountID] = *form
	return cloned, nil
}
