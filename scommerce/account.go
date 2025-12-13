package scommerce

import (
	"context"
	"io"
	"sync"
	"time"
)

var _ UserAccount[any] = &BuiltinUserAccount[any]{}

type userAccountDatabase[AccountID comparable] interface {
	DBUserAccount[AccountID]
	userAddressDatabase[AccountID]
	userPaymentMethodDatabase[AccountID]
	userOrderDatabase[AccountID]
	userShoppingCartDatabase[AccountID]
	productItemSubscriptionDatabase[AccountID]
	userFactorDatabase[AccountID]
	DBUserRole
}

type UserAccountForm[AccountID comparable] struct {
	ID                       AccountID                            `json:"id"`
	TotalDepts               *float64                             `json:"total_depts,omitempty"`
	TotalDeptsWithoutPenalty *float64                             `json:"total_depts_without_penalty,omitempty"`
	AddressCount             *uint64                              `json:"address_count,omitempty"`
	Bio                      *string                              `json:"bio,omitempty"`
	DefaultAddress           *BuiltinUserAddress[AccountID]       `json:"default_address,omitempty"`
	DefaultPaymentMethod     *BuiltinUserPaymentMethod[AccountID] `json:"default_payment_method,omitempty"`
	FirstName                *string                              `json:"first_name,omitempty"`
	LastName                 *string                              `json:"last_name,omitempty"`
	LastUpdatedAt            *time.Time                           `json:"last_updated_at,omitempty"`
	OrderCount               *uint64                              `json:"order_count,omitempty"`
	Password                 *string                              `json:"password,omitempty"`
	PaymentMethodCount       *uint64                              `json:"payment_method_count,omitempty"`
	ProfileImages            *[]string                            `json:"profile_images,omitempty"`
	Role                     *BuiltinUserRole                     `json:"role,omitempty"`
	ShoppingCartCount        *uint64                              `json:"shopping_cart_count,omitempty"`
	Token                    *string                              `json:"token,omitempty"`
	UserLevel                *int64                               `json:"user_level,omitempty"`
	WalletCurrency           *float64                             `json:"wallet_currency,omitempty"`
	Penalty                  *float64                             `json:"penalty,omitempty"`
	IsActiveState            *bool                                `json:"is_active_state,omitempty"`
	IsBannedState            *string                              `json:"is_banned_state,omitempty"`
	IsSuperUserState         *bool                                `json:"is_super_user_state,omitempty"`
	IsTradingAllowedState    *bool                                `json:"is_trading_allowed,omitempty"`
}

type BuiltinUserAccount[AccountID comparable] struct {
	UserAccountForm[AccountID]
	DB                 userAccountDatabase[AccountID] `json:"-"`
	FS                 FileStorage                    `json:"-"`
	OrderStatusManager OrderStatusManager             `json:"-"`
	MU                 sync.RWMutex                   `json:"-"`
}

func (account *BuiltinUserAccount[AccountID]) AllowTrading(ctx context.Context, state bool) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.AllowUserAccountTrading(ctx, &form, id, state); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsTradingAllowedState = &state
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) Ban(ctx context.Context, till time.Duration, reason string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.BanUserAccount(ctx, &form, id, till, reason); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsBannedState = &reason
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) CalculateTotalDepts(ctx context.Context) (currency float64, err error) {
	account.MU.RLock()
	if account.TotalDepts != nil {
		defer account.MU.RUnlock()
		return *account.TotalDepts, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	depts, err := account.DB.CalculateUserAccountTotalDepts(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.TotalDepts = &depts
	return depts, nil
}

func (account *BuiltinUserAccount[AccountID]) CalculateTotalDeptsWithoutPenalty(ctx context.Context) (currency float64, err error) {
	account.MU.RLock()
	if account.TotalDeptsWithoutPenalty != nil {
		defer account.MU.RUnlock()
		return *account.TotalDeptsWithoutPenalty, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	depts, err := account.DB.CalculateUserAccountTotalDeptsWithoutPenalty(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.TotalDeptsWithoutPenalty = &depts
	return depts, nil
}

func (account *BuiltinUserAccount[AccountID]) ChargeWallet(ctx context.Context, currency float64) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.ChargeUserAccountWallet(ctx, &form, id, currency); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.WalletCurrency = nil
	account.TotalDepts = nil
	account.TotalDeptsWithoutPenalty = nil
	account.Penalty = nil
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (account *BuiltinUserAccount[AccountID]) Fine(ctx context.Context, amount float64) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.FineUserAccount(ctx, &form, id, amount); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.WalletCurrency = nil
	account.TotalDepts = nil
	account.TotalDeptsWithoutPenalty = nil
	account.Penalty = nil
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) GetAddressCount(ctx context.Context) (uint64, error) {
	account.MU.RLock()
	if account.AddressCount != nil {
		defer account.MU.RUnlock()
		return *account.AddressCount, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountAddressCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.AddressCount = &count
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) newUserAddress(ctx context.Context, id uint64, db userAddressDatabase[AccountID], form *UserAddressForm[AccountID]) (*BuiltinUserAddress[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	addr := &BuiltinUserAddress[AccountID]{
		UserAddressForm: UserAddressForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
		DB: db,
	}
	if err := addr.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := addr.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return addr, nil
}

func (account *BuiltinUserAccount[AccountID]) GetAddresses(ctx context.Context, addresses []UserAddress[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAddress[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	addressForms := make([]*UserAddressForm[AccountID], 0, cap(ids))
	ids, addressForms, err = account.DB.GetUserAccountAddresses(ctx, &form, id, ids, addressForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	adds := addresses
	if adds == nil {
		adds = make([]UserAddress[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		addr, err := account.newUserAddress(ctx, ids[i], account.DB, addressForms[i])
		if err != nil {
			return nil, err
		}
		adds = append(adds, addr)
	}
	return adds, nil
}

func (account *BuiltinUserAccount[AccountID]) GetBio(ctx context.Context) (string, error) {
	account.MU.RLock()
	if account.Bio != nil {
		defer account.MU.RUnlock()
		return *account.Bio, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	bio, err := account.DB.GetUserAccountBio(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Bio = &bio
	return bio, err
}

func (account *BuiltinUserAccount[AccountID]) GetDefaultAddress(ctx context.Context) (UserAddress[AccountID], error) {
	account.MU.RLock()
	if account.DefaultAddress != nil {
		defer account.MU.RUnlock()
		return account.DefaultAddress, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	addressForm := UserAddressForm[AccountID]{}
	aid, err := account.DB.GetUserAccountDefaultAddress(ctx, &form, id, &addressForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	addr, err := account.newUserAddress(ctx, aid, account.DB, &addressForm)
	if err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.DefaultAddress = addr
	return addr, nil
}

func (account *BuiltinUserAccount[AccountID]) newPaymentMethod(ctx context.Context, pid uint64, db userPaymentMethodDatabase[AccountID], form *UserPaymentMethodForm[AccountID]) (*BuiltinUserPaymentMethod[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
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
	if form != nil {
		if err := method.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return method, nil
}

func (account *BuiltinUserAccount[AccountID]) GetDefaultPaymentMethod(ctx context.Context) (UserPaymentMethod[AccountID], error) {
	account.MU.RLock()
	if account.DefaultPaymentMethod != nil {
		defer account.MU.RUnlock()
		return account.DefaultPaymentMethod, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	paymentMethodForm := UserPaymentMethodForm[AccountID]{}
	pid, err := account.DB.GetUserAccountDefaultPaymentMethod(ctx, &form, id, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	method, err := account.newPaymentMethod(ctx, pid, account.DB, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.DefaultPaymentMethod = method
	return method, nil
}

func (account *BuiltinUserAccount[AccountID]) GetFirstName(ctx context.Context) (string, error) {
	account.MU.RLock()
	if account.FirstName != nil {
		defer account.MU.RUnlock()
		return *account.FirstName, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := account.DB.GetUserAccountFirstName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.FirstName = &name
	return name, nil
}

func (account *BuiltinUserAccount[AccountID]) GetID(ctx context.Context) (AccountID, error) {
	account.MU.RLock()
	defer account.MU.RUnlock()
	return account.ID, nil
}

func (account *BuiltinUserAccount[AccountID]) GetLastName(ctx context.Context) (string, error) {
	account.MU.RLock()
	if account.LastName != nil {
		defer account.MU.RUnlock()
		return *account.LastName, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := account.DB.GetUserAccountLastName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.LastName = &name
	return name, nil
}

func (account *BuiltinUserAccount[AccountID]) GetLastUpdatedAt(ctx context.Context) (time.Time, error) {
	account.MU.RLock()
	if account.LastUpdatedAt != nil {
		defer account.MU.RUnlock()
		return *account.LastUpdatedAt, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return time.Time{}, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return time.Time{}, err
	}
	updatedAt, err := account.DB.GetUserAccountLastUpdatedAt(ctx, &form, id)
	if err != nil {
		return time.Time{}, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return time.Time{}, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.LastUpdatedAt = &updatedAt
	return updatedAt, nil
}

func (account *BuiltinUserAccount[AccountID]) GetOrderCount(ctx context.Context) (uint64, error) {
	account.MU.RLock()
	if account.OrderCount != nil {
		defer account.MU.RUnlock()
		return *account.OrderCount, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountOrderCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.OrderCount = &count
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) newOrder(ctx context.Context, oid uint64, db userOrderDatabase[AccountID], form *UserOrderForm[AccountID]) (*BuiltinUserOrder[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	order := &BuiltinUserOrder[AccountID]{
		UserOrderForm: UserOrderForm[AccountID]{
			ID:            oid,
			UserAccountID: aid,
		},
		DB:                 db,
		FS:                 account.FS,
		OrderStatusManager: account.OrderStatusManager,
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

func (account *BuiltinUserAccount[AccountID]) GetOrders(ctx context.Context, orders []UserOrder[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrder[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	orderForms := make([]*UserOrderForm[AccountID], 0, cap(ids))
	ids, orderForms, err = account.DB.GetUserAccountOrders(ctx, &form, id, ids, orderForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	ods := orders
	if ods == nil {
		ods = make([]UserOrder[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		order, err := account.newOrder(ctx, ids[i], account.DB, orderForms[i])
		if err != nil {
			return nil, err
		}
		ods = append(ods, order)
	}
	return ods, nil
}

func (account *BuiltinUserAccount[AccountID]) GetPassword(ctx context.Context) (string, error) {
	account.MU.RLock()
	if account.Password != nil {
		defer account.MU.RUnlock()
		return *account.Password, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	password, err := account.DB.GetUserAccountPassword(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Password = &password
	return password, nil
}

func (account *BuiltinUserAccount[AccountID]) GetPaymentMethodCount(ctx context.Context) (uint64, error) {
	account.MU.RLock()
	if account.PaymentMethodCount != nil {
		defer account.MU.RUnlock()
		return *account.PaymentMethodCount, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountPaymentMethodCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.PaymentMethodCount = &count
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) GetPaymentMethods(ctx context.Context, paymentMethods []UserPaymentMethod[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserPaymentMethod[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	paymentMethodForms := make([]*UserPaymentMethodForm[AccountID], 0, cap(ids))
	ids, paymentMethodForms, err = account.DB.GetUserAccountPaymentMethods(ctx, &form, id, ids, paymentMethodForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	methods := paymentMethods
	if methods == nil {
		methods = make([]UserPaymentMethod[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		method, err := account.newPaymentMethod(ctx, ids[i], account.DB, paymentMethodForms[i])
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	return methods, nil
}

func (account *BuiltinUserAccount[AccountID]) GetProfileImages(ctx context.Context) ([]FileReadCloser, error) {
	var imageTokens []string = nil
	account.MU.RLock()
	if account.ProfileImages != nil {
		imageTokens = *account.ProfileImages
		account.MU.RUnlock()
	} else {
		account.MU.RUnlock()
		var err error = nil
		id, err := account.GetID(ctx)
		if err != nil {
			return nil, err
		}
		form, err := account.UserAccountForm.Clone(ctx)
		if err != nil {
			return nil, err
		}
		imageTokens, err = account.DB.GetUserAccountProfileImages(ctx, &form, id)
		if err != nil {
			return nil, err
		}
		if err := account.ApplyFormObject(ctx, &form); err != nil {
			return nil, err
		}
		account.MU.Lock()
		account.ProfileImages = &imageTokens
		account.MU.Unlock()
	}

	files := make([]FileReadCloser, 0, len(imageTokens))
	for _, token := range imageTokens {
		file, err := account.FS.Open(ctx, token)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (account *BuiltinUserAccount[AccountID]) GetRole(ctx context.Context) (UserRole, error) {
	account.MU.RLock()
	if account.Role != nil {
		defer account.MU.RUnlock()
		return account.Role, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	roleForm := UserRoleForm{}
	rid, err := account.DB.GetUserAccountRole(ctx, &form, id, &roleForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	role := &BuiltinUserRole{
		DB: account.DB,
		UserRoleForm: UserRoleForm{
			ID: rid,
		},
	}
	if err := role.Init(ctx); err != nil {
		return nil, err
	}
	if err := role.ApplyFormObject(ctx, &roleForm); err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Role = role
	return role, nil
}

func (account *BuiltinUserAccount[AccountID]) GetShoppingCartCount(ctx context.Context) (uint64, error) {
	account.MU.RLock()
	if account.ShoppingCartCount != nil {
		defer account.MU.RUnlock()
		return *account.ShoppingCartCount, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountShoppingCartCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.ShoppingCartCount = &count
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) newShoppingCart(ctx context.Context, id uint64, db userShoppingCartDatabase[AccountID], form *UserShoppingCartForm[AccountID]) (*BuiltinUserShoppingCart[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	cart := &BuiltinUserShoppingCart[AccountID]{
		DB:                 db,
		FS:                 account.FS,
		OrderStatusManager: account.OrderStatusManager,
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

func (account *BuiltinUserAccount[AccountID]) GetShoppingCarts(ctx context.Context, outCarts []UserShoppingCart[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCart[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	cartForms := make([]*UserShoppingCartForm[AccountID], 0, cap(ids))
	ids, cartForms, err = account.DB.GetUserAccountShoppingCarts(ctx, &form, id, ids, cartForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	cts := outCarts
	if cts == nil {
		cts = make([]UserShoppingCart[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		cart, err := account.newShoppingCart(ctx, ids[i], account.DB, cartForms[i])
		if err != nil {
			return nil, err
		}
		cts = append(cts, cart)
	}
	return cts, nil
}

func (account *BuiltinUserAccount[AccountID]) GetToken(ctx context.Context) (string, error) {
	account.MU.RLock()
	if account.Token != nil {
		defer account.MU.RUnlock()
		return *account.Token, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	token, err := account.DB.GetUserAccountToken(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Token = &token
	return token, nil
}

func (account *BuiltinUserAccount[AccountID]) GetUserLevel(ctx context.Context) (int64, error) {
	account.MU.RLock()
	if account.UserLevel != nil {
		defer account.MU.RUnlock()
		return *account.UserLevel, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	level, err := account.DB.GetUserAccountLevel(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.UserLevel = &level
	return level, nil
}

func (account *BuiltinUserAccount[AccountID]) GetWalletCurrency(ctx context.Context) (float64, error) {
	account.MU.RLock()
	if account.WalletCurrency != nil {
		defer account.MU.RUnlock()
		return *account.WalletCurrency, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	currency, err := account.DB.GetUserAccountWalletCurrency(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.WalletCurrency = &currency
	return currency, nil
}

func (account *BuiltinUserAccount[AccountID]) HasPenalty(ctx context.Context) (bool, error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	result, err := account.DB.HasUserAccountPenalty(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	return result, nil
}

func (account *BuiltinUserAccount[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (account *BuiltinUserAccount[AccountID]) IsActive(ctx context.Context) (bool, error) {
	account.MU.RLock()
	if account.IsActiveState != nil {
		defer account.MU.RUnlock()
		return *account.IsActiveState, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isActive, err := account.DB.IsUserAccountActive(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsActiveState = &isActive
	return isActive, nil
}

func (account *BuiltinUserAccount[AccountID]) IsBanned(ctx context.Context) (reason string, err error) {
	account.MU.RLock()
	if account.IsBannedState != nil {
		defer account.MU.RUnlock()
		return *account.IsBannedState, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	isBanned, err := account.DB.IsUserAccountBanned(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsBannedState = &isBanned
	return isBanned, nil
}

func (account *BuiltinUserAccount[AccountID]) IsSuperUser(ctx context.Context) (bool, error) {
	account.MU.RLock()
	if account.IsSuperUserState != nil {
		defer account.MU.RUnlock()
		return *account.IsSuperUserState, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isSuperUser, err := account.DB.IsUserAccountSuperUser(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsSuperUserState = &isSuperUser
	return isSuperUser, nil
}

func (account *BuiltinUserAccount[AccountID]) IsTradingAllowed(ctx context.Context) (bool, error) {
	account.MU.RLock()
	if account.IsTradingAllowedState != nil {
		defer account.MU.RUnlock()
		return *account.IsTradingAllowedState, nil
	}
	account.MU.RUnlock()
	id, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isAllowed, err := account.DB.IsUserAccountTradingAllowed(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsTradingAllowedState = &isAllowed
	return isAllowed, nil
}

func (account *BuiltinUserAccount[AccountID]) NewAddress(ctx context.Context, unitNumber string, street_number string, addressLine1 string, addressLine2 string, city string, region string, postalCode string, country Country, isDefault bool) (UserAddress[AccountID], error) {
	var cid *uint64 = nil
	if country != nil {
		tcid, err := country.GetID(ctx)
		if err != nil {
			return nil, err
		}
		cid = &tcid
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	addressForm := UserAddressForm[AccountID]{}
	aid, err := account.DB.NewUserAccountAddress(ctx, &form, id, unitNumber, street_number, addressLine1, addressLine2, city, region, postalCode, cid, isDefault, &addressForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	addr, err := account.newUserAddress(ctx, aid, account.DB, &addressForm)
	if err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.AddressCount = nil
	return addr, nil
}

func (account *BuiltinUserAccount[AccountID]) NewPaymentMethod(ctx context.Context, paymentType PaymentType, provider string, accoutNumber string, expiryDate time.Time, isDefault bool) (UserPaymentMethod[AccountID], error) {
	pid, err := paymentType.GetID(ctx)
	if err != nil {
		return nil, err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	paymentMethodForm := UserPaymentMethodForm[AccountID]{}
	pmid, err := account.DB.NewUserAccountPaymentMethod(ctx, &form, id, pid, provider, accoutNumber, expiryDate, isDefault, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	pType, err := account.newPaymentMethod(ctx, pmid, account.DB, &paymentMethodForm)
	if err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.PaymentMethodCount = nil
	return pType, nil
}

func (account *BuiltinUserAccount[AccountID]) NewShoppingCart(ctx context.Context, sessionText string) (UserShoppingCart[AccountID], error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	cartForm := UserShoppingCartForm[AccountID]{}
	cid, err := account.DB.NewUserAccountShoppingCart(ctx, &form, id, sessionText, &cartForm)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	cart, err := account.newShoppingCart(ctx, cid, account.DB, &cartForm)
	if err != nil {
		return nil, err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.ShoppingCartCount = nil
	return cart, nil
}

func (account *BuiltinUserAccount[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAddress(ctx context.Context, address UserAddress[AccountID]) error {
	aid, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveUserAccountAddress(ctx, &form, id, aid); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.AddressCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAllAddresses(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveAllUserAccountAddresses(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.AddressCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAllOrders(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveAllUserAccountOrders(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.OrderCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAllPaymentMethods(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveAllUserAccountPaymentMethods(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.PaymentMethodCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAllShoppingCarts(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveAllUserAccountShoppingCarts(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.ShoppingCartCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveOrder(ctx context.Context, order UserOrder[AccountID]) error {
	oid, err := order.GetID(ctx)
	if err != nil {
		return nil
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveUserAccountOrder(ctx, &form, id, oid); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.OrderCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemovePaymentMethod(ctx context.Context, paymentMethod UserPaymentMethod[AccountID]) error {
	pid, err := paymentMethod.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveUserAccountPaymentMethod(ctx, &form, id, pid); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.PaymentMethodCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveShoppingCart(ctx context.Context, cart UserShoppingCart[AccountID]) error {
	cid, err := cart.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveUserAccountShoppingCart(ctx, &form, id, cid); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.ShoppingCartCount = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetActive(ctx context.Context, state bool) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountActive(ctx, &form, id, state); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsActiveState = &state
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetBio(ctx context.Context, bio string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountBio(ctx, &form, id, bio); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Bio = &bio
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetFirstName(ctx context.Context, name string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountFirstName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.FirstName = &name
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetLastName(ctx context.Context, name string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountLastName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.LastName = &name
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetLastUpdatedAt(ctx context.Context, lastUpdateAt time.Time) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountLastUpdatedAt(ctx, &form, id, lastUpdateAt); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.LastUpdatedAt = &lastUpdateAt
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetPassword(ctx context.Context, password string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountPassword(ctx, &form, id, password); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Password = &password
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetPenalty(ctx context.Context, penalty float64) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountPenalty(ctx, &form, id, penalty); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Penalty = &penalty
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetProfileImages(ctx context.Context, images []FileReader) error {
	var errRes error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	tokens := make([]string, 0, len(images))
	imgs := make([]FileReader, 0, len(images))
	for _, image := range images {
		token, err := image.GetToken(ctx)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		tokens = append(tokens, token)
		imgs = append(imgs, image)
	}

	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}

	err = account.DB.SetUserAccountProfileImages(ctx, &form, id, tokens)
	if err != nil {
		errRes = joinErr(errRes, err)
		return errRes
	}

	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}

	account.MU.Lock()

	account.ProfileImages = &tokens
	account.LastUpdatedAt = nil

	account.MU.Unlock()

	for i, token := range tokens {
		account.FS.Delete(ctx, token)
		file, err := account.FS.Create(ctx, token)
		if err != nil {
			errRes = joinErr(errRes, err)
			continue
		}
		if _, err := io.Copy(file, imgs[i]); err != nil {
			errRes = joinErr(errRes, err)
		}
		file.Close()
	}

	return errRes
}

func (account *BuiltinUserAccount[AccountID]) SetRole(ctx context.Context, role UserRole) error {
	rid, err := role.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountRole(ctx, &form, id, rid); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Role, err = role.ToBuiltinObject(ctx)
	account.LastUpdatedAt = nil
	return err
}

func (account *BuiltinUserAccount[AccountID]) SetSuperUser(ctx context.Context, state bool) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountSuperUser(ctx, &form, id, state); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsSuperUserState = &state
	account.UserLevel = nil
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetToken(ctx context.Context, token string) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountToken(ctx, &form, id, token); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.Token = &token
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetUserLevel(ctx context.Context, level int64) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountLevel(ctx, &form, id, level); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.UserLevel = &level
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) SetWalletCurrency(ctx context.Context, currency float64) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.SetUserAccountWalletCurrency(ctx, &form, id, currency); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.WalletCurrency = &currency
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) TransferCurrency(ctx context.Context, to UserAccount[AccountID], amount float64) error {
	aid, err := to.GetID(ctx)
	if err != nil {
		return err
	}
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.TransferUserAccountCurrency(ctx, &form, id, aid, amount); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.WalletCurrency = nil
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) Unban(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.UnbanUserAccount(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	account.MU.Lock()
	defer account.MU.Unlock()
	account.IsBannedState = nil
	account.LastUpdatedAt = nil
	return nil
}

func (account *BuiltinUserAccount[AccountID]) ValidatePassword(ctx context.Context, password string) (bool, error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	result, err := account.DB.ValidateUserAccountPassword(ctx, &form, id, password)
	if err != nil {
		return false, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	return result, nil
}

func (account *BuiltinUserAccount[AccountID]) newUserReview(ctx context.Context, id uint64, db userReviewDatabase[AccountID], form *UserReviewForm[AccountID]) (*BuiltinUserReview[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	review := &BuiltinUserReview[AccountID]{
		UserReviewForm: UserReviewForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
		DB: db,
		FS: account.FS,
	}
	if err := review.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := review.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return review, nil
}

func (account *BuiltinUserAccount[AccountID]) GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	reviewForms := make([]*UserReviewForm[AccountID], 0, cap(ids))
	ids, reviewForms, err = account.DB.GetUserAccountUserReviews(ctx, &form, id, ids, reviewForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	revs := reviews
	if revs == nil {
		revs = make([]UserReview[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		review, err := account.newUserReview(ctx, ids[i], account.DB, reviewForms[i])
		if err != nil {
			return nil, err
		}
		revs = append(revs, review)
	}
	return revs, nil
}

func (account *BuiltinUserAccount[AccountID]) GetUserReviewCount(ctx context.Context) (uint64, error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountUserReviewCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) GetSubscriptions(ctx context.Context, subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	subscriptionForms := make([]*ProductItemSubscriptionForm[AccountID], 0, cap(ids))
	ids, subscriptionForms, err = account.DB.GetUserAccountSubscriptions(ctx, &form, id, ids, subscriptionForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}

	subs := subscriptions
	if subs == nil {
		subs = make([]ProductItemSubscription[AccountID], 0, len(ids))
	}

	for i := range len(ids) {
		sub := &BuiltinProductItemSubscription[AccountID]{
			ProductItemSubscriptionForm: ProductItemSubscriptionForm[AccountID]{
				ID:            ids[i],
				UserAccountID: id,
			},
			DB: account.DB,
			FS: account.FS,
		}
		if err := sub.Init(ctx); err != nil {
			return nil, err
		}
		if subscriptionForms[i] != nil {
			if err := sub.ApplyFormObject(ctx, subscriptionForms[i]); err != nil {
				return nil, err
			}
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func (account *BuiltinUserAccount[AccountID]) GetSubscriptionCount(ctx context.Context) (uint64, error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountSubscriptionCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveSubscription(ctx context.Context, subscription ProductItemSubscription[AccountID]) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	subscriptionID, err := subscription.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveUserAccountSubscription(ctx, &form, id, subscriptionID); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	return nil
}

func (account *BuiltinUserAccount[AccountID]) RemoveAllSubscriptions(ctx context.Context) error {
	id, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := account.DB.RemoveAllUserAccountSubscriptions(ctx, &form, id); err != nil {
		return err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	return nil
}

func (account *BuiltinUserAccount[AccountID]) newUserFactor(ctx context.Context, fid uint64, db userFactorDatabase[AccountID], form *UserFactorForm[AccountID]) (*BuiltinUserFactor[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
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

func (account *BuiltinUserAccount[AccountID]) GetUserFactors(ctx context.Context, factors []UserFactor[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserFactor[AccountID], error) {
	var err error = nil
	id, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, GetSafeLimit(limit))
	factorForms := make([]*UserFactorForm[AccountID], 0, cap(ids))
	ids, factorForms, err = account.DB.GetUserAccountUserFactors(ctx, &form, id, ids, factorForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	ftrs := factors
	if ftrs == nil {
		ftrs = make([]UserFactor[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		factor, err := account.newUserFactor(ctx, ids[i], account.DB, factorForms[i])
		if err != nil {
			return nil, err
		}
		ftrs = append(ftrs, factor)
	}
	return ftrs, nil
}

func (account *BuiltinUserAccount[AccountID]) GetUserFactorCount(ctx context.Context) (uint64, error) {
	id, err := account.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := account.UserAccountForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	count, err := account.DB.GetUserAccountUserFactorCount(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := account.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	return count, nil
}

func (account *BuiltinUserAccount[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserAccount[AccountID], error) {
	return account, nil
}

func (account *BuiltinUserAccount[AccountID]) ToFormObject(ctx context.Context) (*UserAccountForm[AccountID], error) {
	account.MU.RLock()
	defer account.MU.RUnlock()
	return &account.UserAccountForm, nil
}

func (account *BuiltinUserAccount[AccountID]) ApplyFormObject(ctx context.Context, form *UserAccountForm[AccountID]) error {
	account.MU.Lock()
	defer account.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	// Check if ID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.ID != zeroAccountID {
		account.ID = form.ID
	}
	if form.TotalDepts != nil {
		account.TotalDepts = form.TotalDepts
	}
	if form.TotalDeptsWithoutPenalty != nil {
		account.TotalDeptsWithoutPenalty = form.TotalDeptsWithoutPenalty
	}
	if form.AddressCount != nil {
		account.AddressCount = form.AddressCount
	}
	if form.Bio != nil {
		account.Bio = form.Bio
	}
	if form.DefaultAddress != nil {
		account.DefaultAddress = form.DefaultAddress
	}
	if form.DefaultPaymentMethod != nil {
		account.DefaultPaymentMethod = form.DefaultPaymentMethod
	}
	if form.FirstName != nil {
		account.FirstName = form.FirstName
	}
	if form.LastName != nil {
		account.LastName = form.LastName
	}
	if form.LastUpdatedAt != nil {
		account.LastUpdatedAt = form.LastUpdatedAt
	}
	if form.OrderCount != nil {
		account.OrderCount = form.OrderCount
	}
	if form.Password != nil {
		account.Password = form.Password
	}
	if form.PaymentMethodCount != nil {
		account.PaymentMethodCount = form.PaymentMethodCount
	}
	if form.ProfileImages != nil {
		account.ProfileImages = form.ProfileImages
	}
	if form.Role != nil {
		account.Role = form.Role
	}
	if form.ShoppingCartCount != nil {
		account.ShoppingCartCount = form.ShoppingCartCount
	}
	if form.Token != nil {
		account.Token = form.Token
	}
	if form.UserLevel != nil {
		account.UserLevel = form.UserLevel
	}
	if form.WalletCurrency != nil {
		account.WalletCurrency = form.WalletCurrency
	}
	if form.Penalty != nil {
		account.Penalty = form.Penalty
	}
	if form.IsActiveState != nil {
		account.IsActiveState = form.IsActiveState
	}
	if form.IsBannedState != nil {
		account.IsBannedState = form.IsBannedState
	}
	if form.IsSuperUserState != nil {
		account.IsSuperUserState = form.IsSuperUserState
	}
	if form.IsTradingAllowedState != nil {
		account.IsTradingAllowedState = form.IsTradingAllowedState
	}
	return nil
}

func (form *UserAccountForm[AccountID]) Clone(ctx context.Context) (UserAccountForm[AccountID], error) {
	var cloned UserAccountForm[AccountID] = *form
	return cloned, nil
}
