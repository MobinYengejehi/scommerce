package scommerce

import (
	"context"
	"errors"
	"time"
)

var _ GeneralAppObject = &App[any]{}

type App[AccountID comparable] struct {
	AccountManager        UserAccountManager[AccountID]
	OrderManager          UserOrderManager[AccountID]
	PaymentMethodManager  UserPaymentMethodManager[AccountID]
	AddressManager        UserAddressManager[AccountID]
	ShoppingCartManager   UserShoppingCartManager[AccountID]
	RoleManager           UserRoleManager
	ProductManager        ProductManager[AccountID]
	CountryManager        CountryManager
	PaymentTypeManager    PaymentTypeManager
	ShippingMethodManager ShippingMethodManager
	OrderStatusManager    OrderStatusManager
	UserReviewManager     UserReviewManager[AccountID]
	SubscriptionManager   ProductItemSubscriptionManager[AccountID]
	DiscountManager       UserDiscountManager[AccountID]
}

type AppConfig[AccountID comparable] struct {
	DB                         DBApplication[AccountID]
	FileStorage                FileStorage
	OTPCodeLength              int32
	OTPTokenLength             int32
	OTPTTL                     time.Duration
	SubscriptionRenewalHandler RenewalHandlerFunc[AccountID]
	DiscountCodeLength         int32
}

func NewBuiltinApplication[AccountID comparable](conf *AppConfig[AccountID]) (*App[AccountID], error) {
	orderStatusManager := NewBuiltinOrderStatusManager(conf.DB)
	shippingMethodManager := NewBuiltinShippingMethodManager(conf.DB)
	paymentTypeManager := NewBuiltinPaymentTypeManager(conf.DB)
	countryManager := NewBuiltinCountryManager(conf.DB)
	userRoleManager := NewBuiltinUserRoleManager(conf.DB)
	addressManager := NewBuiltinUserAddressManager(conf.DB)
	paymentMethodManager := NewBuiltinPaymentMethodManager(conf.DB)
	orderManager := NewBuiltinUserOrderManager(conf.DB, orderStatusManager, conf.FileStorage)
	productManager := NewBuiltinProductManager(conf.DB, conf.FileStorage)
	shoppingCartManager := NewBuiltinUserShoppingCartManager(conf.DB, conf.FileStorage, orderStatusManager)
	userReviewManager := NewBuiltinUserReviewManager(conf.DB, conf.FileStorage)
	subscriptionManager := NewBuiltinProductItemSubscriptionManager(conf.DB, conf.FileStorage, conf.SubscriptionRenewalHandler)
	
	discountCodeLength := conf.DiscountCodeLength
	if discountCodeLength == 0 {
		discountCodeLength = 8
	}
	discountManager := NewBuiltinUserDiscountManager(conf.DB, discountCodeLength)

	accountManager, err := NewBuiltinUserAccountManager(
		conf.DB,
		conf.FileStorage,
		conf.OTPCodeLength,
		conf.OTPTokenLength,
		conf.OTPTTL,
		orderStatusManager,
	)
	if err != nil {
		return nil, err
	}

	return &App[AccountID]{
		OrderStatusManager:    orderStatusManager,
		ShippingMethodManager: shippingMethodManager,
		PaymentTypeManager:    paymentTypeManager,
		CountryManager:        countryManager,
		RoleManager:           userRoleManager,
		AccountManager:        accountManager,
		AddressManager:        addressManager,
		PaymentMethodManager:  paymentMethodManager,
		OrderManager:          orderManager,
		ProductManager:        productManager,
		ShoppingCartManager:   shoppingCartManager,
		UserReviewManager:     userReviewManager,
		SubscriptionManager:   subscriptionManager,
		DiscountManager:       discountManager,
	}, nil
}

func (app *App[AccountID]) Close(ctx context.Context) error {
	var err error = nil

	err = joinErr(err, app.AccountManager.Close(ctx))
	err = joinErr(err, app.OrderManager.Close(ctx))
	err = joinErr(err, app.PaymentMethodManager.Close(ctx))
	err = joinErr(err, app.AddressManager.Close(ctx))
	err = joinErr(err, app.ShoppingCartManager.Close(ctx))
	err = joinErr(err, app.RoleManager.Close(ctx))
	err = joinErr(err, app.ProductManager.Close(ctx))
	err = joinErr(err, app.CountryManager.Close(ctx))
	err = joinErr(err, app.PaymentTypeManager.Close(ctx))
	err = joinErr(err, app.ShippingMethodManager.Close(ctx))
	err = joinErr(err, app.OrderStatusManager.Close(ctx))
	err = joinErr(err, app.SubscriptionManager.Close(ctx))
	err = joinErr(err, app.DiscountManager.Close(ctx))

	return err
}

func (app *App[AccountID]) Init(ctx context.Context) error {
	var err error = nil

	err = joinErr(err, app.RoleManager.Init(ctx))
	err = joinErr(err, app.OrderStatusManager.Init(ctx))
	err = joinErr(err, app.PaymentTypeManager.Init(ctx))
	err = joinErr(err, app.ShippingMethodManager.Init(ctx))
	err = joinErr(err, app.CountryManager.Init(ctx))
	err = joinErr(err, app.ProductManager.Init(ctx))
	err = joinErr(err, app.AccountManager.Init(ctx))
	err = joinErr(err, app.PaymentMethodManager.Init(ctx))
	err = joinErr(err, app.AddressManager.Init(ctx))
	err = joinErr(err, app.ShoppingCartManager.Init(ctx))
	err = joinErr(err, app.UserReviewManager.Init(ctx))
	err = joinErr(err, app.OrderManager.Init(ctx))
	err = joinErr(err, app.SubscriptionManager.Init(ctx))
	err = joinErr(err, app.DiscountManager.Init(ctx))

	return err
}

func (app *App[AccountID]) Pulse(ctx context.Context) error {
	var err error = nil

	err = joinErr(err, app.AccountManager.Pulse(ctx))
	err = joinErr(err, app.OrderManager.Pulse(ctx))
	err = joinErr(err, app.PaymentMethodManager.Pulse(ctx))
	err = joinErr(err, app.AddressManager.Pulse(ctx))
	err = joinErr(err, app.ShoppingCartManager.Pulse(ctx))
	err = joinErr(err, app.RoleManager.Pulse(ctx))
	err = joinErr(err, app.ProductManager.Pulse(ctx))
	err = joinErr(err, app.CountryManager.Pulse(ctx))
	err = joinErr(err, app.PaymentTypeManager.Pulse(ctx))
	err = joinErr(err, app.ShoppingCartManager.Pulse(ctx))
	err = joinErr(err, app.OrderStatusManager.Pulse(ctx))
	err = joinErr(err, app.UserReviewManager.Pulse(ctx))
	err = joinErr(err, app.SubscriptionManager.Pulse(ctx))
	err = joinErr(err, app.DiscountManager.Pulse(ctx))

	return err
}

func joinErr(err error, newErr error) error {
	if newErr != nil {
		if err == nil {
			err = errors.New("")
		}
		return errors.Join(err, newErr)
	}
	return err
}
