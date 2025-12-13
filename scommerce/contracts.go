package scommerce

import (
	"context"
	"encoding/json"
	"time"
)

type GeneralInitiative interface {
	Init(ctx context.Context) error
}

type GeneralClosable interface {
	Close(ctx context.Context) error
}

type GeneralPulsable interface {
	Pulse(ctx context.Context) error // use err.join or err group to handle and log multiple errors
}

type GeneralAppObject interface {
	GeneralInitiative
	GeneralClosable
	GeneralPulsable
}

type UserAccountManager[AccountID comparable] interface {
	GeneralAppObject

	GetAccount(ctx context.Context, token string) (UserAccount[AccountID], error)
	GetAccountWithID(ctx context.Context, aid AccountID) (UserAccount[AccountID], error)

	NewAccount(ctx context.Context, token string, password string, twoFactor string) (UserAccount[AccountID], error)
	RemoveAccount(ctx context.Context, account UserAccount[AccountID]) error
	RemoveAccountWithToken(ctx context.Context, token string, password string, twoFactor string) error
	RemoveAllAccounts(ctx context.Context) error
	GetAccounts(ctx context.Context, accounts []UserAccount[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAccount[AccountID], error)
	GetAccountCount(ctx context.Context) (uint64, error)

	Authenticate(ctx context.Context, token string, password string, twoFactor string) (UserAccount[AccountID], error)

	RequestTwoFactor(ctx context.Context, token string) (string, error)
	ValidateTwoFactor(ctx context.Context, token string, code string) (bool, error)
	CancelTwoFactor(ctx context.Context, token string) error

	ToBuiltinObject(ctx context.Context) (*BuiltinUserAccountManager[AccountID], error)
}

type UserAccount[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (AccountID, error)
	GetToken(ctx context.Context) (string, error)
	SetToken(ctx context.Context, token string) error

	GetPassword(ctx context.Context) (string, error)
	SetPassword(ctx context.Context, password string) error
	ValidatePassword(ctx context.Context, password string) (bool, error)

	GetFirstName(ctx context.Context) (string, error)
	SetFirstName(ctx context.Context, name string) error
	GetLastName(ctx context.Context) (string, error)
	SetLastName(ctx context.Context, name string) error

	GetLastUpdatedAt(ctx context.Context) (time.Time, error)
	SetLastUpdatedAt(ctx context.Context, lastUpdateAt time.Time) error

	GetRole(ctx context.Context) (UserRole, error)
	SetRole(ctx context.Context, role UserRole) error

	GetUserLevel(ctx context.Context) (int64, error) // admin level for example
	SetUserLevel(ctx context.Context, level int64) error

	IsSuperUser(ctx context.Context) (bool, error)
	SetSuperUser(ctx context.Context, state bool) error
	IsActive(ctx context.Context) (bool, error)
	SetActive(ctx context.Context, state bool) error

	GetProfileImages(ctx context.Context) ([]FileReadCloser, error)
	SetProfileImages(ctx context.Context, images []FileReader) error
	GetBio(ctx context.Context) (string, error)
	SetBio(ctx context.Context, bio string) error

	// Ban
	Ban(ctx context.Context, till time.Duration, reason string) error
	Unban(ctx context.Context) error
	IsBanned(ctx context.Context) (reason string, err error)

	AllowTrading(ctx context.Context, state bool) error
	IsTradingAllowed(ctx context.Context) (bool, error)

	// Fine
	Fine(ctx context.Context, amount float64) error
	SetPenalty(ctx context.Context, penalty float64) error

	// Wallet (Money, Currency)
	GetWalletCurrency(ctx context.Context) (float64, error)
	SetWalletCurrency(ctx context.Context, currency float64) error
	ChargeWallet(ctx context.Context, currency float64) error
	TransferCurrency(ctx context.Context, to UserAccount[AccountID], amount float64) error

	// Carts
	NewShoppingCart(ctx context.Context, sessionText string) (UserShoppingCart[AccountID], error)
	RemoveShoppingCart(ctx context.Context, cart UserShoppingCart[AccountID]) error
	RemoveAllShoppingCarts(ctx context.Context) error // for this user
	GetShoppingCartCount(ctx context.Context) (uint64, error)
	GetShoppingCarts(ctx context.Context, outCarts []UserShoppingCart[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCart[AccountID], error)

	// Addresses
	NewAddress(ctx context.Context, unitNumber, street_number, addressLine1, addressLine2, city, region, postalCode string, country Country, isDefault bool) (UserAddress[AccountID], error)
	RemoveAddress(ctx context.Context, address UserAddress[AccountID]) error
	RemoveAllAddresses(ctx context.Context) error // for this user
	GetAddresses(ctx context.Context, addresses []UserAddress[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAddress[AccountID], error)
	GetAddressCount(ctx context.Context) (uint64, error)
	GetDefaultAddress(ctx context.Context) (UserAddress[AccountID], error)

	// Payment methods
	NewPaymentMethod(ctx context.Context, paymentType PaymentType, provider, accoutNumber string, expiryDate time.Time, isDefault bool) (UserPaymentMethod[AccountID], error)
	RemovePaymentMethod(ctx context.Context, paymentMethod UserPaymentMethod[AccountID]) error
	RemoveAllPaymentMethods(ctx context.Context) error
	GetPaymentMethods(ctx context.Context, paymentMethods []UserPaymentMethod[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserPaymentMethod[AccountID], error)
	GetPaymentMethodCount(ctx context.Context) (uint64, error)
	GetDefaultPaymentMethod(ctx context.Context) (UserPaymentMethod[AccountID], error)

	// Orders
	RemoveOrder(ctx context.Context, order UserOrder[AccountID]) error
	RemoveAllOrders(ctx context.Context) error // orders for this user
	GetOrders(ctx context.Context, orders []UserOrder[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrder[AccountID], error)
	GetOrderCount(ctx context.Context) (uint64, error)

	// Dept
	HasPenalty(ctx context.Context) (bool, error)                                        // if wallet currency is negative it has penalty
	CalculateTotalDeptsWithoutPenalty(ctx context.Context) (currency float64, err error) // total shopping carts dept
	CalculateTotalDepts(ctx context.Context) (currency float64, err error)               // depts without penalty + penalty depts

	// Reviews
	GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error)
	GetUserReviewCount(ctx context.Context) (uint64, error)

	// Subscriptions
	GetSubscriptions(ctx context.Context, subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error)
	GetSubscriptionCount(ctx context.Context) (uint64, error)
	RemoveSubscription(ctx context.Context, subscription ProductItemSubscription[AccountID]) error
	RemoveAllSubscriptions(ctx context.Context) error

	// Tickets

	ToBuiltinObject(ctx context.Context) (*BuiltinUserAccount[AccountID], error)
	ToFormObject(ctx context.Context) (*UserAccountForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserAccountForm[AccountID]) error
}

type UserOrderManager[AccountID comparable] interface {
	GeneralAppObject

	RemoveAllUserOrders(ctx context.Context) error
	GetUserOrders(ctx context.Context, orders []UserOrder[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrder[AccountID], error)
	GetUserOrderCount(ctx context.Context) (uint64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserOrderManager[AccountID], error)
}

type UserOrder[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)

	GetOrderDate(ctx context.Context) (time.Time, error)
	SetOrderDate(ctx context.Context, date time.Time) error

	GetPaymentMethod(ctx context.Context) (UserPaymentMethod[AccountID], error)
	SetPaymentMethod(ctx context.Context, paymentMethod UserPaymentMethod[AccountID]) error
	GetShippingAddress(ctx context.Context) (UserAddress[AccountID], error)
	SetShippingAddress(ctx context.Context, address UserAddress[AccountID]) error
	GetShippingMethod(ctx context.Context) (ShippingMethod, error)
	SetShippingMethod(ctx context.Context, shippingMethod ShippingMethod) error

	GetOrderTotal(ctx context.Context) (float64, error)
	SetOrderTotal(ctx context.Context, price float64) error
	CalculateTotalPrice(ctx context.Context) (float64, error)

	GetStatus(ctx context.Context) (OrderStatus, error)
	SetStatus(ctx context.Context, status OrderStatus) error
	IsDeliveried(ctx context.Context) (bool, error)

	GetUserComment(ctx context.Context) (string, error)
	SetUserComment(ctx context.Context, comment string) error

	GetDeliveryDate(ctx context.Context) (time.Time, error)
	SetDeliveryDate(ctx context.Context, date time.Time) error
	GetDeliveryComment(ctx context.Context) (string, error)
	SetDeliveryComment(ctx context.Context, comment string) error
	Deliver(ctx context.Context, date time.Time, comment string) error

	GetProductItems(ctx context.Context, items []UserOrderProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserOrderProductItem[AccountID], error)
	GetProductItemCount(ctx context.Context) (uint64, error)
	SetProductItems(ctx context.Context, items []UserOrderProductItem[AccountID]) error

	ToBuiltinObject(ctx context.Context) (*BuiltinUserOrder[AccountID], error)
	ToFormObject(ctx context.Context) (*UserOrderForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserOrderForm[AccountID]) error
}

type UserPaymentMethodManager[AccountID comparable] interface {
	GeneralAppObject

	RemoveAllPaymentMethods(ctx context.Context) error
	GetPaymentMethods(ctx context.Context, paymentMethods []UserPaymentMethod[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserPaymentMethod[AccountID], error)
	GetPaymentMethodCount(ctx context.Context) (uint64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserPaymentMethodManager[AccountID], error)
}

type UserPaymentMethod[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)

	GetPaymentType(ctx context.Context) (PaymentType, error)
	SetPaymentType(ctx context.Context, paymentType PaymentType) error

	GetProvider(ctx context.Context) (string, error)
	SetProvider(ctx context.Context, provider string) error
	GetAccountNumber(ctx context.Context) (string, error)
	SetAccountNumber(ctx context.Context, accountNumber string) error

	GetExpiryDate(ctx context.Context) (time.Time, error)
	SetExpiryDate(ctx context.Context, date time.Time) error
	IsExpired(ctx context.Context) (bool, error)

	IsDefault(ctx context.Context) (bool, error)
	SetDefault(ctx context.Context, state bool) error

	ToBuiltinObject(ctx context.Context) (*BuiltinUserPaymentMethod[AccountID], error)
	ToFormObject(ctx context.Context) (*UserPaymentMethodForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserPaymentMethodForm[AccountID]) error
}

type UserAddressManager[AccountID comparable] interface {
	GeneralAppObject

	RemoveAllAddresses(ctx context.Context) error
	GetAddresses(ctx context.Context, addresses []UserAddress[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAddress[AccountID], error)
	GetAddressCount(ctx context.Context) (uint64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserAddressManager[AccountID], error)
}

type UserAddress[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)

	GetUnitNumber(ctx context.Context) (string, error)
	SetUnitNumber(ctx context.Context, unitNumber string) error
	GetStreetNumber(ctx context.Context) (string, error)
	SetStreetNumber(ctx context.Context, streetNumber string) error
	GetAddressLine1(ctx context.Context) (string, error)
	SetAddressLine1(ctx context.Context, address string) error
	GetAddressLine2(ctx context.Context) (string, error)
	SetAddressLine2(ctx context.Context, address string) error
	GetCity(ctx context.Context) (string, error)
	SetCity(ctx context.Context, city string) error
	GetRegion(ctx context.Context) (string, error)
	SetRegion(ctx context.Context, region string) error
	GetPostalCode(ctx context.Context) (string, error)
	SetPostalCode(ctx context.Context, code string) error

	GetCountry(ctx context.Context) (Country, error)
	SetCountry(ctx context.Context, country Country) error

	IsDefault(ctx context.Context) (bool, error)
	SetDefault(ctx context.Context) error

	ToBuiltinObject(ctx context.Context) (*BuiltinUserAddress[AccountID], error)
	ToFormObject(ctx context.Context) (*UserAddressForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserAddressForm[AccountID]) error
}

type UserRoleManager interface {
	GeneralAppObject

	NewUserRole(ctx context.Context, name string) (UserRole, error)
	RemoveUserRole(ctx context.Context, role UserRole) error
	RemoveAllUserRoles(ctx context.Context) error
	GetUserRoles(ctx context.Context, roles []UserRole, skip int64, limit int64, queueOrder QueueOrder) ([]UserRole, error)
	GetUserRoleCount(ctx context.Context) (uint64, error)
	GetUserRoleByName(ctx context.Context, name string) (UserRole, error)
	ExistsUserRole(ctx context.Context, name string) (bool, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserRoleManager, error)
}

type UserRole interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error

	ToBuiltinObject(ctx context.Context) (*BuiltinUserRole, error)
	ToFormObject(ctx context.Context) (*UserRoleForm, error)
	ApplyFormObject(ctx context.Context, form *UserRoleForm) error
}

type UserShoppingCartManager[AccountID comparable] interface {
	GeneralAppObject

	RemoveAllShoppingCarts(ctx context.Context) error
	GetShoppingCarts(ctx context.Context, carts []UserShoppingCart[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCart[AccountID], error) // for all users
	GetShoppingCartCount(ctx context.Context) (uint64, error)
	GetShoppingCartBySessionText(ctx context.Context, sessionText string) (UserShoppingCart[AccountID], error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCartManager[AccountID], error)
}

type UserShoppingCart[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)
	GetSessionText(ctx context.Context) (string, error)
	SetSessionText(ctx context.Context, sessionText string) error

	NewShoppingCartItem(ctx context.Context, item ProductItem[AccountID], count int64, attrs json.RawMessage) (UserShoppingCartItem[AccountID], error)
	RemoveShoppingCartItem(ctx context.Context, item UserShoppingCartItem[AccountID]) error
	RemoveAllShoppingCartItems(ctx context.Context) error
	GetShoppingCartItems(ctx context.Context, items []UserShoppingCartItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserShoppingCartItem[AccountID], error)
	GetShoppingCartItemCount(ctx context.Context) (uint64, error)

	CalculateDept(ctx context.Context, shippingMethod ShippingMethod) (float64, error)

	Order(ctx context.Context, paymentMethod UserPaymentMethod[AccountID], address UserAddress[AccountID], shippingMethod ShippingMethod, userComment string) (UserOrder[AccountID], error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCart[AccountID], error)
	ToFormObject(ctx context.Context) (*UserShoppingCartForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserShoppingCartForm[AccountID]) error
}

type UserShoppingCartItem[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)
	GetShoppingCart(ctx context.Context) (UserShoppingCart[AccountID], error)

	GetProductItem(ctx context.Context) (ProductItem[AccountID], error)

	GetQuantity(ctx context.Context) (int64, error)
	SetQuantity(ctx context.Context, quantity int64) error
	AddQuantity(ctx context.Context, delta int64) error

	GetAttributes(ctx context.Context) (json.RawMessage, error)
	SetAttributes(ctx context.Context, attrs json.RawMessage) error

	CalculateDept(ctx context.Context) (float64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserShoppingCartItem[AccountID], error)
	ToFormObject(ctx context.Context) (*UserShoppingCartItemForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserShoppingCartItemForm[AccountID]) error
}

type UserFactorManager[AccountID comparable] interface {
}

/*
this user factor structure must look like this:

	struct UserFactor[AccountID comparable] {
		UserAccountID AccountID
		ID  uint64
		Products json.RawMessage // this contains the list of items with 'id' and 'item display names' and its count and price
		Discount   float64
		Tax        float64
		AmountPaid float64
	}
*/
type UserFactor[AccountID comparable] interface {
}

type ProductManager[AccountID comparable] interface {
	GeneralAppObject

	NewProductCategory(ctx context.Context, name string, parentCategory ProductCategory[AccountID]) (ProductCategory[AccountID], error)
	RemoveProductCategory(ctx context.Context, category ProductCategory[AccountID]) error
	RemoveAllProductCategroies(ctx context.Context) error
	GetProductCategories(ctx context.Context, categories []ProductCategory[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductCategory[AccountID], error)
	GetProductCategoryCount(ctx context.Context) (uint64, error)
	SearchForProductCategories(ctx context.Context, searchText string, deepSearch bool, categories []ProductCategory[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductCategory[AccountID], error)
	SearchForProducts(ctx context.Context, searchText string, deepSearch bool, products []Product[AccountID], skip int64, limit int64, queueOrder QueueOrder, category ProductCategory[AccountID]) ([]Product[AccountID], error)
	SearchForProductItems(ctx context.Context, searchText string, deepSearch bool, items []ProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder, product Product[AccountID], category ProductCategory[AccountID]) ([]ProductItem[AccountID], error)

	ToBuiltinObject(ctx context.Context) (*BuiltinProductManager[AccountID], error)
}

type ProductCategory[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error
	GetParentProductCategory(ctx context.Context) (ProductCategory[AccountID], error)
	SetParentProductCategory(ctx context.Context, parent ProductCategory[AccountID]) error

	NewProduct(ctx context.Context, name string, description string, images []FileReader) (Product[AccountID], error)
	RemoveProduct(ctx context.Context, product Product[AccountID]) error
	RemoveAllProducts(ctx context.Context) error
	GetProducts(ctx context.Context, products []Product[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]Product[AccountID], error)
	GetProductCount(ctx context.Context) (uint64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinProductCategory[AccountID], error)
	ToFormObject(ctx context.Context) (*ProductCategoryForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *ProductCategoryForm[AccountID]) error
}

type Product[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error
	GetDescription(ctx context.Context) (string, error)
	SetDescription(ctx context.Context, desc string) error
	GetImages(ctx context.Context) ([]FileReadCloser, error)
	SetImages(ctx context.Context, images []FileReader) error
	GetProductCategory(ctx context.Context) (ProductCategory[AccountID], error)
	SetProductCategory(ctx context.Context, category ProductCategory[AccountID]) error

	AddProductItem(ctx context.Context, sku string, name string, price float64, quantity uint64, images []FileReader, attrs json.RawMessage) (ProductItem[AccountID], error)
	RemoveProductItem(ctx context.Context, item ProductItem[AccountID]) error
	RemoveAllProductItems(ctx context.Context) error
	GetProductItems(ctx context.Context, items []ProductItem[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItem[AccountID], error)
	GetProductItemCount(ctx context.Context) (uint64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinProduct[AccountID], error)
	ToFormObject(ctx context.Context) (*ProductForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *ProductForm[AccountID]) error
}

type ProductItem[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error
	GetSKU(ctx context.Context) (string, error) // slug
	SetSKU(ctx context.Context, sku string) error
	GetPrice(ctx context.Context) (float64, error)
	SetPrice(ctx context.Context, price float64) error

	GetQuantityInStock(ctx context.Context) (uint64, error)
	SetQuantityInStock(ctx context.Context, quantity uint64) error
	AddQuantityInStock(ctx context.Context, delta int64) error

	GetImages(ctx context.Context) ([]FileReadCloser, error)
	SetImages(ctx context.Context, images []FileReader) error

	GetAttributes(ctx context.Context) (json.RawMessage, error)
	SetAttributes(ctx context.Context, attrs json.RawMessage) error

	GetProduct(ctx context.Context) (Product[AccountID], error)
	SetProduct(ctx context.Context, product Product[AccountID]) error

	GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error)
	GetUserReviewCount(ctx context.Context) (uint64, error)
	CalculateAverageRating(ctx context.Context) (float64, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinProductItem[AccountID], error)
	ToFormObject(ctx context.Context) (*ProductItemForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *ProductItemForm[AccountID]) error
}

type RenewalHandlerFunc[AccountID comparable] func(ctx context.Context, subscription ProductItemSubscription[AccountID], account UserAccount[AccountID], productItem ProductItem[AccountID]) (success bool, amountCharged float64, err error)

type ProductItemSubscriptionManager[AccountID comparable] interface {
	GeneralAppObject

	NewSubscription(ctx context.Context, account UserAccount[AccountID], productItem ProductItem[AccountID], duration time.Duration, subscriptionType string, autoRenew bool) (ProductItemSubscription[AccountID], error)
	RemoveSubscription(ctx context.Context, subscription ProductItemSubscription[AccountID]) error
	RemoveAllSubscriptions(ctx context.Context) error
	GetSubscriptions(ctx context.Context, subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error)
	GetSubscriptionCount(ctx context.Context) (uint64, error)
	GetUserSubscriptions(ctx context.Context, account UserAccount[AccountID], subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error)
	GetUserSubscriptionCount(ctx context.Context, account UserAccount[AccountID]) (uint64, error)
	GetProductItemSubscriptions(ctx context.Context, productItem ProductItem[AccountID], subscriptions []ProductItemSubscription[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]ProductItemSubscription[AccountID], error)
	GetProductItemSubscriptionCount(ctx context.Context, productItem ProductItem[AccountID]) (uint64, error)

	SetRenewalHandler(ctx context.Context, handler RenewalHandlerFunc[AccountID]) error
	ProcessExpiredSubscriptions(ctx context.Context) error

	ToBuiltinObject(ctx context.Context) (*BuiltinProductItemSubscriptionManager[AccountID], error)
}

type ProductItemSubscription[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)

	GetProductItem(ctx context.Context) (ProductItem[AccountID], error)
	SetProductItem(ctx context.Context, productItem ProductItem[AccountID]) error

	GetSubscribedAt(ctx context.Context) (time.Time, error)
	SetSubscribedAt(ctx context.Context, subscribedAt time.Time) error

	GetExpiresAt(ctx context.Context) (time.Time, error)
	SetExpiresAt(ctx context.Context, expiresAt time.Time) error

	GetDuration(ctx context.Context) (time.Duration, error)
	SetDuration(ctx context.Context, duration time.Duration) error

	GetSubscriptionType(ctx context.Context) (string, error)
	SetSubscriptionType(ctx context.Context, subscriptionType string) error

	IsAutoRenew(ctx context.Context) (bool, error)
	SetAutoRenew(ctx context.Context, autoRenew bool) error

	IsActive(ctx context.Context) (bool, error)
	SetActive(ctx context.Context, isActive bool) error

	IsExpired(ctx context.Context) (bool, error)
	Cancel(ctx context.Context) error

	ToBuiltinObject(ctx context.Context) (*BuiltinProductItemSubscription[AccountID], error)
	ToFormObject(ctx context.Context) (*ProductItemSubscriptionForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *ProductItemSubscriptionForm[AccountID]) error
}

type CountryManager interface {
	GeneralAppObject

	NewCountry(ctx context.Context, name string) (Country, error)
	RemoveCountry(ctx context.Context, country Country) error
	RemoveAllCountries(ctx context.Context) error
	GetCountries(ctx context.Context, countries []Country, skip int64, limit int64, queueOrder QueueOrder) ([]Country, error)
	GetCountryCount(ctx context.Context) (uint64, error)
	GetCountryByName(ctx context.Context, name string) (Country, error)
	ExistsCountry(ctx context.Context, name string) (bool, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinCountryManager, error)
}

type Country interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error

	ToBuiltinObject(ctx context.Context) (*BuiltinCountry, error)
	ToFormObject(ctx context.Context) (*CountryForm, error)
	ApplyFormObject(ctx context.Context, form *CountryForm) error
}

type PaymentTypeManager interface {
	GeneralAppObject

	NewPaymentType(ctx context.Context, name string) (PaymentType, error)
	RemovePaymentType(ctx context.Context, paymentType PaymentType) error
	RemoveAllPaymentTypes(ctx context.Context) error
	GetPaymentTypes(ctx context.Context, paymentTypes []PaymentType, skip int64, limit int64, queueOrder QueueOrder) ([]PaymentType, error)
	GetPaymentTypeCount(ctx context.Context) (uint64, error)
	GetPaymentTypeByName(ctx context.Context, name string) (PaymentType, error)
	ExistsPaymentType(ctx context.Context, name string) (bool, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinPaymentTypeManager, error)
}

type PaymentType interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error

	ToBuiltinObject(ctx context.Context) (*BuiltinPaymentType, error)
	ToFormObject(ctx context.Context) (*PaymentTypeForm, error)
	ApplyFormObject(ctx context.Context, form *PaymentTypeForm) error
}

type ShippingMethodManager interface {
	GeneralAppObject

	NewShippingMethod(ctx context.Context, name string, price float64) (ShippingMethod, error)
	RemoveShippingMethod(ctx context.Context, shippingMethod ShippingMethod) error
	RemoveAllShippingMethods(ctx context.Context) error
	GetShippingMethods(ctx context.Context, shippingMethods []ShippingMethod, skip int64, limit int64, queueOrder QueueOrder) ([]ShippingMethod, error)
	GetShippingMethodCount(ctx context.Context) (uint64, error)
	GetShippingMethodByName(ctx context.Context, name string) (ShippingMethod, error)
	ExistsShippingMethod(ctx context.Context, name string) (bool, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinShippingMethodManager, error)
}

type ShippingMethod interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error
	GetPrice(ctx context.Context) (float64, error)
	SetPrice(ctx context.Context, price float64) error

	ToBuiltinObject(ctx context.Context) (*BuiltinShippingMethod, error)
	ToFormObject(ctx context.Context) (*ShippingMethodForm, error)
	ApplyFormObject(ctx context.Context, form *ShippingMethodForm) error
}

type OrderStatusManager interface {
	GeneralAppObject

	NewOrderStatus(ctx context.Context, name string) (OrderStatus, error)
	RemoveOrderStatus(ctx context.Context, status OrderStatus) error
	RemoveAllOrderStatuses(ctx context.Context) error
	GetOrderStatuses(ctx context.Context, orderStatuses []OrderStatus, skip int64, limit int64, queueOrder QueueOrder) ([]OrderStatus, error)
	GetOrderStatusCount(ctx context.Context) (uint64, error)
	GetOrderStatusByName(ctx context.Context, name string) (OrderStatus, error)
	ExistsOrderStatus(ctx context.Context, name string) (bool, error)

	GetDeliveriedOrderStatus(ctx context.Context) (OrderStatus, error)
	GetIdleOrderStatus(ctx context.Context) (OrderStatus, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinOrderStatusManager, error)
}

type OrderStatus interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetName(ctx context.Context) (string, error)
	SetName(ctx context.Context, name string) error

	IsDeliveried(ctx context.Context) (bool, error)

	ToBuiltinObject(ctx context.Context) (*BuiltinOrderStatus, error)
	ToFormObject(ctx context.Context) (*OrderStatusForm, error)
	ApplyFormObject(ctx context.Context, form *OrderStatusForm) error
}

type UserReviewManager[AccountID comparable] interface {
	GeneralAppObject

	NewUserReview(ctx context.Context, account UserAccount[AccountID], productItem ProductItem[AccountID], ratingValue int32, comment string) (UserReview[AccountID], error)
	RemoveUserReview(ctx context.Context, review UserReview[AccountID]) error
	RemoveAllUserReviews(ctx context.Context) error
	GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error)
	GetUserReviewCount(ctx context.Context) (uint64, error)
	GetUserReviewsForProductItem(ctx context.Context, productItem ProductItem[AccountID], reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error)
	GetUserReviewsForAccount(ctx context.Context, account UserAccount[AccountID], reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserReviewManager[AccountID], error)
}

type UserReview[AccountID comparable] interface {
	GeneralAppObject

	GetID(ctx context.Context) (uint64, error)
	GetUserAccountID(ctx context.Context) (AccountID, error)

	GetRatingValue(ctx context.Context) (int32, error)
	SetRatingValue(ctx context.Context, rating int32) error

	GetComment(ctx context.Context) (string, error)
	SetComment(ctx context.Context, comment string) error

	GetProductItem(ctx context.Context) (ProductItem[AccountID], error)

	ToBuiltinObject(ctx context.Context) (*BuiltinUserReview[AccountID], error)
	ToFormObject(ctx context.Context) (*UserReviewForm[AccountID], error)
	ApplyFormObject(ctx context.Context, form *UserReviewForm[AccountID]) error
}
