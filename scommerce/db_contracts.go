package scommerce

import (
	"context"
	"encoding/json"
	"time"
)

type DBApplication[AccountID comparable] interface {
	DBUserAccountManager[AccountID]
	DBUserAccount[AccountID]
	DBUserShoppingCartManager[AccountID]
	DBUserShoppingCart[AccountID]
	DBUserShoppingCartItem[AccountID]
	DBUserOrderManager[AccountID]
	DBUserOrder[AccountID]
	DBUserPaymentMethodManager[AccountID]
	DBUserPaymentMethod[AccountID]
	DBUserAddressManager[AccountID]
	DBUserAddress[AccountID]
	DBUserReviewManager[AccountID]
	DBUserReview[AccountID]
	DBUserRoleManager
	DBUserRole
	DBProductManager[AccountID]
	DBProductCategory[AccountID]
	DBProduct[AccountID]
	DBProductItem[AccountID]
	DBProductItemSubscriptionManager[AccountID]
	DBProductItemSubscription[AccountID]
	DBUserFactorManager[AccountID]
	DBUserFactor[AccountID]
	DBUserDiscountManager[AccountID]
	DBUserDiscount[AccountID]
	DBCountryManager
	DBCountry
	DBPaymentTypeManager
	DBPaymentType
	DBShippingMethodManager
	DBShippingMethod
	DBOrderStatusManager
	DBOrderStatus
}

type DBUserAccountManager[AccountID comparable] interface {
	RemoveAllUserAccounts(ctx context.Context) error
	RemoveUserAccountWithToken(ctx context.Context, token string, password string) error
	RemoveUserAccount(ctx context.Context, aid AccountID) error
	GetUserAccountCount(ctx context.Context) (uint64, error)
	GetUserAccount(ctx context.Context, token string, accountForm *UserAccountForm[AccountID]) (AccountID, error)
	GetUserAccounts(ctx context.Context, accounts []AccountID, accountForms []*UserAccountForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]AccountID, []*UserAccountForm[AccountID], error)
	NewUserAccount(ctx context.Context, token string, password string, accountForm *UserAccountForm[AccountID]) (AccountID, error)
	AuthenticateUserAccount(ctx context.Context, token string, password string, accountForm *UserAccountForm[AccountID]) (AccountID, error)
	InitUserAccountManager(ctx context.Context) error
	FillUserAccountWithID(ctx context.Context, aid AccountID, accountForm *UserAccountForm[AccountID]) error
}

type DBUserAccount[AccountID comparable] interface {
	AllowUserAccountTrading(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, state bool) error
	BanUserAccount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, till time.Duration, reason string) error
	CalculateUserAccountTotalDepts(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (float64, error)
	CalculateUserAccountTotalDeptsWithoutPenalty(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (float64, error)
	ChargeUserAccountWallet(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, currency float64) error
	FineUserAccount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, amount float64) error
	GetUserAccountAddressCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	GetUserAccountAddresses(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, addresses []uint64, addressForms []*UserAddressForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserAddressForm[AccountID], error)
	GetUserAccountBio(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	GetUserAccountDefaultAddress(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, addressForm *UserAddressForm[AccountID]) (uint64, error)
	GetUserAccountDefaultPaymentMethod(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, paymentMethodForm *UserPaymentMethodForm[AccountID]) (uint64, error)
	GetUserAccountFirstName(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	GetUserAccountLastName(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	GetUserAccountLastUpdatedAt(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (time.Time, error)
	GetUserAccountOrderCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	GetUserAccountOrders(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, orders []uint64, orderForms []*UserOrderForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserOrderForm[AccountID], error)
	GetUserAccountPassword(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	GetUserAccountPaymentMethodCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	GetUserAccountPaymentMethods(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, paymentMethods []uint64, paymentMethodForms []*UserPaymentMethodForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserPaymentMethodForm[AccountID], error)
	GetUserAccountProfileImages(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) ([]string, error)
	GetUserAccountRole(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, roleForm *UserRoleForm) (uint64, error)
	GetUserAccountShoppingCartCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	GetUserAccountShoppingCarts(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, carts []uint64, cartForms []*UserShoppingCartForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserShoppingCartForm[AccountID], error)
	GetUserAccountToken(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	GetUserAccountLevel(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (int64, error)
	GetUserAccountWalletCurrency(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (float64, error)
	GetUserAccountUserReviews(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, ids []uint64, reviewForms []*UserReviewForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserReviewForm[AccountID], error)
	GetUserAccountUserReviewCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	GetUserAccountSubscriptions(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, ids []uint64, subscriptionForms []*ProductItemSubscriptionForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*ProductItemSubscriptionForm[AccountID], error)
	GetUserAccountSubscriptionCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	RemoveUserAccountSubscription(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, subscriptionID uint64) error
	RemoveAllUserAccountSubscriptions(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	GetUserAccountUserFactors(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, ids []uint64, factorForms []*UserFactorForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserFactorForm[AccountID], error)
	GetUserAccountUserFactorCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	NewUserAccountDiscount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, value float64, validCount int64, discountForm *UserDiscountForm[AccountID]) (uint64, error)
	RemoveUserAccountDiscount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, discountID uint64) error
	RemoveAllUserAccountDiscounts(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	GetUserAccountDiscounts(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, ids []uint64, discountForms []*UserDiscountForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserDiscountForm[AccountID], error)
	GetUserAccountDiscountCount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (uint64, error)
	HasUserAccountPenalty(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (bool, error)
	IsUserAccountActive(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (bool, error)
	IsUserAccountBanned(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (string, error)
	IsUserAccountSuperUser(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (bool, error)
	IsUserAccountTradingAllowed(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) (bool, error)
	NewUserAccountAddress(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, unitNumber string, street_number string, addressLine1 string, addressLine2 string, city string, region string, postalCode string, country *uint64, isDefault bool, addressForm *UserAddressForm[AccountID]) (uint64, error)
	NewUserAccountPaymentMethod(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, paymentType uint64, provider string, accoutNumber string, expiryDate time.Time, isDefault bool, paymentMethodForm *UserPaymentMethodForm[AccountID]) (uint64, error)
	NewUserAccountShoppingCart(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, sessionText string, cartForm *UserShoppingCartForm[AccountID]) (uint64, error)
	RemoveUserAccountAddress(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, addrID uint64) error
	RemoveAllUserAccountAddresses(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	RemoveAllUserAccountOrders(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	RemoveAllUserAccountPaymentMethods(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	RemoveAllUserAccountShoppingCarts(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	RemoveUserAccountShoppingCart(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, cid uint64) error
	RemoveUserAccountPaymentMethod(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, pid uint64) error
	RemoveUserAccountOrder(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, oid uint64) error
	SetUserAccountActive(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, state bool) error
	SetUserAccountBio(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, bio string) error
	SetUserAccountFirstName(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, name string) error
	SetUserAccountLastName(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, name string) error
	SetUserAccountLastUpdatedAt(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, lastUpdatedAt time.Time) error
	SetUserAccountPassword(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, password string) error
	SetUserAccountPenalty(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, penalty float64) error
	SetUserAccountProfileImages(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, images []string) error
	SetUserAccountRole(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, role uint64) error
	SetUserAccountSuperUser(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, state bool) error
	SetUserAccountToken(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, token string) error
	SetUserAccountLevel(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, level int64) error
	SetUserAccountWalletCurrency(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, currency float64) error
	TransferUserAccountCurrency(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, to AccountID, currency float64) error
	UnbanUserAccount(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID) error
	ValidateUserAccountPassword(ctx context.Context, form *UserAccountForm[AccountID], aid AccountID, password string) (bool, error)
}

type DBUserUserShoppingCartResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}
type DBUserShoppingCartManager[AccountID comparable] interface {
	GetShoppingCartBySessionText(ctx context.Context, sessionText string, cartForm *UserShoppingCartForm[AccountID]) (uint64, error)
	GetShoppingCartCount(ctx context.Context) (uint64, error)
	GetShoppingCarts(ctx context.Context, ids []DBUserUserShoppingCartResult[AccountID], cartForms []*UserShoppingCartForm[AccountID], skip int64, limit int64, order QueueOrder) ([]DBUserUserShoppingCartResult[AccountID], []*UserShoppingCartForm[AccountID], error)
	RemoveAllShoppingCarts(ctx context.Context) error
	InitUserShoppingCartManager(ctx context.Context) error
	FillUserShoppingCartWithID(ctx context.Context, cid uint64, cartForm *UserShoppingCartForm[AccountID]) error
	FillUserShoppingCartItemWithID(ctx context.Context, iid uint64, cartForm *UserShoppingCartItemForm[AccountID]) error
}

type DBUserShoppingCart[AccountID comparable] interface {
	CalculateUserShoppingCartDept(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, shippingMethod uint64) (float64, error)
	GetUserShoppingCartSessionText(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64) (string, error)
	GetUserShoppingCartItemCount(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64) (uint64, error)
	GetUserShoppingCartItems(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, items []uint64, itemForms []*UserShoppingCartItemForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, fs FileStorage, osm OrderStatusManager) ([]uint64, []*UserShoppingCartItemForm[AccountID], error)
	NewUserShoppingCartShoppingCartItem(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, productItem uint64, count int64, attrs json.RawMessage, itemForm *UserShoppingCartItemForm[AccountID], fs FileStorage, osm OrderStatusManager) (uint64, error)
	OrderUserShoppingCart(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, paymentMethod uint64, address uint64, shippingMethod uint64, userComment string, discountCode string, orderForm *UserOrderForm[AccountID]) (uint64, error)
	RemoveUserShoppingCartAllShoppingCartItems(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64) error
	RemoveUserShoppingCartShoppingCartItem(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, itid uint64) error
	SetUserShoppingCartSessionText(ctx context.Context, form *UserShoppingCartForm[AccountID], sid uint64, text string) error
}

type DBUserShoppingCartItem[AccountID comparable] interface {
	AddUserShoppingCartItemQuantity(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64, delta int64) error
	CalculateUserShoppingCartItemDept(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64) (float64, error)
	GetUserShoppingCartItemProductItem(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64, pItemForm *ProductItemForm[AccountID], db FileStorage) (uint64, error)
	GetUserShoppingCartItemQuantity(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64) (int64, error)
	GetUserShoppingCartItemShoppingCart(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64, cartForm *UserShoppingCartForm[AccountID], db FileStorage, osm OrderStatusManager) (uint64, error)
	SetUserShoppingCartItemQuantity(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64, quantity int64) error
	GetUserShoppingCartItemAttributes(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64) (json.RawMessage, error)
	SetUserShoppingCartItemAttributes(ctx context.Context, form *UserShoppingCartItemForm[AccountID], itid uint64, attrs json.RawMessage) error
}

type DBUserOrderResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}

type DBUserOrderManager[AccountID comparable] interface {
	GetUserOrderCount(ctx context.Context) (uint64, error)
	GetUserOrders(ctx context.Context, orders []DBUserOrderResult[AccountID], orderForms []*UserOrderForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserOrderResult[AccountID], []*UserOrderForm[AccountID], error)
	RemoveAllUserOrders(ctx context.Context) error
	InitUserOrderManager(ctx context.Context) error
	FillUserOrderWithID(ctx context.Context, oid uint64, orderForm *UserOrderForm[AccountID]) error
}

type DBUserOrderProductItem struct {
	ProductItemID uint64          `json:"product_item_id"`
	Quantity      uint64          `json:"quantity"`
	Attributes    json.RawMessage `json:"attributes,omitempty"`
}
type DBUserOrder[AccountID comparable] interface {
	CalculateUserOrderTotalPrice(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (float64, error)
	DeliverUserOrder(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, sid uint64, date time.Time, comment string) error
	GetUserOrderDeliveryComment(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (string, error)
	GetUserOrderDeliveryDate(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (time.Time, error)
	GetUserOrderDate(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (time.Time, error)
	GetUserOrderTotal(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (float64, error)
	GetUserOrderPaymentMethod(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, paymentMethodForm *UserPaymentMethodForm[AccountID]) (uint64, error)
	GetUserOrderProductItemCount(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (uint64, error)
	GetUserOrderProductItems(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, items []DBUserOrderProductItem, skip int64, limit int64, queueOrder QueueOrder) ([]DBUserOrderProductItem, error)
	GetUserOrderShippingAddress(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, addressForm *UserAddressForm[AccountID]) (uint64, error)
	GetUserOrderShippingMethod(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, shippingMethodForm *ShippingMethodForm) (uint64, error)
	GetUserOrderStatus(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, statusForm *OrderStatusForm) (uint64, error)
	GetUserOrderUserComment(ctx context.Context, form *UserOrderForm[AccountID], oid uint64) (string, error)
	SetUserOrderDeliveryComment(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, comment string) error
	SetUserOrderDeliveryDate(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, date time.Time) error
	SetUserOrderDate(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, date time.Time) error
	SetUserOrderTotal(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, total float64) error
	SetUserOrderPaymentMethod(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, method uint64) error
	SetUserOrderProductItems(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, items []DBUserOrderProductItem) error
	SetUserOrderShippingAddress(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, address uint64) error
	SetUserOrderShippingMethod(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, method uint64) error
	SetUserOrderStatus(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, status uint64) error
	SetUserOrderUserComment(ctx context.Context, form *UserOrderForm[AccountID], oid uint64, comment string) error
}

type DBUserPaymentMethodResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}
type DBUserPaymentMethodManager[AccountID comparable] interface {
	GetUserPaymentMethodCount(ctx context.Context) (uint64, error)
	GetUserPaymentMethods(ctx context.Context, methods []DBUserPaymentMethodResult[AccountID], methodForms []*UserPaymentMethodForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserPaymentMethodResult[AccountID], []*UserPaymentMethodForm[AccountID], error)
	RemoveAllUserPaymentMethods(ctx context.Context) error
	InitUserPaymentMethodManager(ctx context.Context) error
	FillUserPaymentMethodWithID(ctx context.Context, mid uint64, paymentMethodForm *UserPaymentMethodForm[AccountID]) error
}

type DBUserPaymentMethod[AccountID comparable] interface {
	GetUserPaymentMethodAccountNumber(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64) (string, error)
	GetUserPaymentMethodExpiryDate(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64) (time.Time, error)
	GetUserPaymentMethodPaymentType(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, paymentTypeFrom *PaymentTypeForm) (uint64, error)
	GetUserPaymentMethodProvider(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64) (string, error)
	IsUserPaymentMethodDefault(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64) (bool, error)
	IsUserPaymentMethodExpired(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64) (bool, error)
	SetUserPaymentMethodAccountNumber(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, number string) error
	SetUserPaymentMethodDefault(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, state bool) error
	SetUserPaymentMethodExpiryDate(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, date time.Time) error
	SetUserPaymentMethodPaymentType(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, pType uint64) error
	SetUserPaymentMethodProvider(ctx context.Context, form *UserPaymentMethodForm[AccountID], mid uint64, provider string) error
}

type DBUserAddressResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}
type DBUserAddressManager[AccountID comparable] interface {
	GetUserAddressCount(ctx context.Context) (uint64, error)
	GetUserAddresses(ctx context.Context, addresses []DBUserAddressResult[AccountID], addressForms []*UserAddressForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserAddressResult[AccountID], []*UserAddressForm[AccountID], error)
	RemoveAllUserAddresses(ctx context.Context) error
	InitUserAddressManager(ctx context.Context) error
	FillUserAddressWithID(ctx context.Context, aid uint64, addressForm *UserAddressForm[AccountID]) error
}

type DBUserAddress[AccountID comparable] interface {
	GetUserAddressAddressLine1(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressAddressLine2(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressCity(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressCountry(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, countryForm *CountryForm) (uint64, error)
	GetUserAddressPostalCode(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressRegion(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressStreetNumber(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressUnitNumber(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (string, error)
	GetUserAddressAccountID(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (AccountID, error)
	IsUserAddressDefault(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) (bool, error)
	SetUserAddressDefault(ctx context.Context, form *UserAddressForm[AccountID], addr uint64) error
	SetUserAddressAddressLine1(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, line string) error
	SetUserAddressAddressLine2(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, line string) error
	SetUserAddressCity(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, city string) error
	SetUserAddressCountry(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, country *uint64) error
	SetUserAddressPostalCode(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, code string) error
	SetUserAddressRegion(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, region string) error
	SetUserAddressStreetNumber(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, number string) error
	SetUserAddressUnitNumber(ctx context.Context, form *UserAddressForm[AccountID], addr uint64, number string) error
}

type DBUserRoleManager interface {
	ExistsUserRole(ctx context.Context, name string) (bool, error)
	GetUserRoleByName(ctx context.Context, name string, roleForm *UserRoleForm) (uint64, error)
	GetUserRoleCount(ctx context.Context) (uint64, error)
	GetUserRoles(ctx context.Context, roles []uint64, roleForms []*UserRoleForm, skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserRoleForm, error)
	NewUserRole(ctx context.Context, name string, roleForm *UserRoleForm) (uint64, error)
	RemoveAllUserRoles(ctx context.Context) error
	RemoveUserRole(ctx context.Context, country uint64) error
	InitUserRoleManager(ctx context.Context) error
	FillUserRoleWithID(ctx context.Context, rid uint64, roleForm *UserRoleForm) error
}

type DBUserRole interface {
	GetUserRoleName(ctx context.Context, form *UserRoleForm, id uint64) (string, error)
	SetUserRoleName(ctx context.Context, form *UserRoleForm, id uint64, name string) error
}

type DBProductManager[AccountID comparable] interface {
	GetProductCategories(ctx context.Context, categories []uint64, catForms []*ProductCategoryForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, fs FileStorage) ([]uint64, []*ProductCategoryForm[AccountID], error)
	GetProductCategoryCount(ctx context.Context) (uint64, error)
	NewProductCategory(ctx context.Context, name string, parentCategory *uint64, catForm *ProductCategoryForm[AccountID], fs FileStorage) (uint64, error)
	RemoveAllProductCategories(ctx context.Context) error
	RemoveProductCategory(ctx context.Context, category uint64) error
	SearchForProductCategories(ctx context.Context, searchText string, deepSearch bool, categories []uint64, catForms []*ProductCategoryForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, fs FileStorage) ([]uint64, []*ProductCategoryForm[AccountID], error)
	SearchForProducts(ctx context.Context, searchText string, deepSearch bool, products []uint64, productForms []*ProductForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, category_id *uint64, fs FileStorage) ([]uint64, []*ProductForm[AccountID], error)
	SearchForProductItems(ctx context.Context, searchText string, deepSearch bool, items []uint64, itemForms []*ProductItemForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, product_id *uint64, category_id *uint64, fs FileStorage) ([]uint64, []*ProductItemForm[AccountID], error)
	InitProductManager(ctx context.Context) error
	FillProductCategoryWithID(ctx context.Context, cid uint64, catForm *ProductCategoryForm[AccountID], fs FileStorage) error
	FillProductWithID(ctx context.Context, pid uint64, productForm *ProductForm[AccountID], fs FileStorage) error
	FillProductItemWithID(ctx context.Context, iid uint64, itemForm *ProductItemForm[AccountID], fs FileStorage) error
}

type DBProductCategory[AccountID comparable] interface {
	GetProductCategoryName(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64) (string, error)
	GetProductCategoryParent(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, catForm *ProductCategoryForm[AccountID], fs FileStorage) (uint64, error)
	GetProductCategoryProductCount(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64) (uint64, error)
	GetProductCategoryProducts(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, products []uint64, productForms []*ProductForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, fs FileStorage) ([]uint64, []*ProductForm[AccountID], error)
	NewProductCategoryProduct(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, name string, description string, images []string, productForm *ProductForm[AccountID], fs FileStorage) (uint64, error)
	RemoveAllProducts(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64) error
	RemoveProduct(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, product uint64) error
	SetProductCategoryName(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, name string) error
	SetProductCategoryParent(ctx context.Context, form *ProductCategoryForm[AccountID], pid uint64, parent *uint64, fs FileStorage) error
}

type DBProduct[AccountID comparable] interface {
	AddProductProductItem(ctx context.Context, form *ProductForm[AccountID], pid uint64, sku string, name string, price float64, quantity uint64, images []string, attrs json.RawMessage, itemForm *ProductItemForm[AccountID], fs FileStorage) (uint64, error)
	GetProductDescription(ctx context.Context, form *ProductForm[AccountID], pid uint64) (string, error)
	GetProductName(ctx context.Context, form *ProductForm[AccountID], pid uint64) (string, error)
	GetProductImages(ctx context.Context, form *ProductForm[AccountID], pid uint64) ([]string, error)
	GetProductCategory(ctx context.Context, form *ProductForm[AccountID], pid uint64, catForm *ProductCategoryForm[AccountID], fs FileStorage) (uint64, error)
	GetProductProductItemCount(ctx context.Context, form *ProductForm[AccountID], pid uint64) (uint64, error)
	GetProductProductItems(ctx context.Context, form *ProductForm[AccountID], pid uint64, items []uint64, itemForms []*ProductItemForm[AccountID], skip int64, limit int64, queueOrder QueueOrder, fs FileStorage) ([]uint64, []*ProductItemForm[AccountID], error)
	RemoveAllProductProductItems(ctx context.Context, form *ProductForm[AccountID], pid uint64) error
	RemoveProductProductItem(ctx context.Context, form *ProductForm[AccountID], pid uint64, itid uint64) error
	SetProductDescription(ctx context.Context, form *ProductForm[AccountID], pid uint64, desc string) error
	SetProductName(ctx context.Context, form *ProductForm[AccountID], pid uint64, name string) error
	SetProductImages(ctx context.Context, form *ProductForm[AccountID], pid uint64, images []string) error
	SetProductCategory(ctx context.Context, form *ProductForm[AccountID], pid uint64, category *uint64, fs FileStorage) error
}

type DBProductItem[AccountID comparable] interface {
	AddProductItemQuantityInStock(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, delta int64) error
	GetProductItemAttributes(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (json.RawMessage, error)
	GetProductItemImages(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) ([]string, error)
	GetProductItemPrice(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (float64, error)
	GetProductItemProduct(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, productForm *ProductForm[AccountID], fs FileStorage) (uint64, error)
	GetProductItemQuantityInStock(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (uint64, error)
	GetProductItemName(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (string, error)
	GetProductItemSKU(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (string, error)
	SetProductItemAttributes(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, attrs json.RawMessage) error
	SetProductItemImages(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, images []string) error
	SetProductItemPrice(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, price float64) error
	SetProductItemProduct(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, product *uint64, fs FileStorage) error
	SetProductItemQuantityInStock(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, quantity uint64) error
	SetProductItemName(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, name string) error
	SetProductItemSKU(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, sku string) error
	GetProductItemUserReviews(ctx context.Context, form *ProductItemForm[AccountID], pid uint64, ids []uint64, reviewForms []*UserReviewForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserReviewForm[AccountID], error)
	GetProductItemUserReviewCount(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (uint64, error)
	CalculateProductItemAverageRating(ctx context.Context, form *ProductItemForm[AccountID], pid uint64) (float64, error)
}

type DBCountryManager interface {
	ExistsCountry(ctx context.Context, name string) (bool, error)
	GetCountryByName(ctx context.Context, name string, countForm *CountryForm) (uint64, error)
	GetCountryCount(ctx context.Context) (uint64, error)
	GetCountries(ctx context.Context, countries []uint64, countForms []*CountryForm, skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*CountryForm, error)
	NewCountry(ctx context.Context, name string, countForm *CountryForm) (uint64, error)
	RemoveAllCountries(ctx context.Context) error
	RemoveCountry(ctx context.Context, country uint64) error
	InitCountryManager(ctx context.Context) error
	FillCountryWithID(ctx context.Context, cid uint64, countryForm *CountryForm) error
}

type DBCountry interface {
	GetCountryName(ctx context.Context, form *CountryForm, id uint64) (string, error)
	SetCountryName(ctx context.Context, form *CountryForm, id uint64, name string) error
}

type DBPaymentTypeManager interface {
	ExistsPaymentType(ctx context.Context, name string) (bool, error)
	GetPaymentTypeByName(ctx context.Context, name string, typeForm *PaymentTypeForm) (uint64, error)
	GetPaymentTypeCount(ctx context.Context) (uint64, error)
	GetPaymentTypes(ctx context.Context, paymentTypes []uint64, typeForms []*PaymentTypeForm, skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*PaymentTypeForm, error)
	NewPaymentType(ctx context.Context, name string, typeForm *PaymentTypeForm) (uint64, error)
	RemoveAllPaymentTypes(ctx context.Context) error
	RemovePaymentType(ctx context.Context, paymentType uint64) error
	InitPaymentTypeManager(ctx context.Context) error
	FillPaymentTypeWithID(ctx context.Context, pid uint64, typeForm *PaymentTypeForm) error
}

type DBPaymentType interface {
	GetPaymentTypeName(ctx context.Context, form *PaymentTypeForm, id uint64) (string, error)
	SetPaymentTypeName(ctx context.Context, form *PaymentTypeForm, id uint64, name string) error
}

type DBShippingMethodManager interface {
	ExistsShippingMethod(ctx context.Context, name string) (bool, error)
	GetShippingMethodByName(ctx context.Context, name string, methodForm *ShippingMethodForm) (uint64, error)
	GetShippingMethodCount(ctx context.Context) (uint64, error)
	GetShippingMethods(ctx context.Context, shippingMethods []uint64, methodForms []*ShippingMethodForm, skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*ShippingMethodForm, error)
	NewShippingMethod(ctx context.Context, name string, price float64, methodForm *ShippingMethodForm) (uint64, error)
	RemoveAllShippingMethods(ctx context.Context) error
	RemoveShippingMethod(ctx context.Context, shippingMethod uint64) error
	InitShippingMethodManager(ctx context.Context) error
	FillShippingMethodWithID(ctx context.Context, sid uint64, methodForm *ShippingMethodForm) error
}

type DBShippingMethod interface {
	GetShippingMethodName(ctx context.Context, form *ShippingMethodForm, id uint64) (string, error)
	SetShippingMethodName(ctx context.Context, form *ShippingMethodForm, id uint64, name string) error
	GetShippingMethodPrice(ctx context.Context, form *ShippingMethodForm, id uint64) (float64, error)
	SetShippingMethodPrice(ctx context.Context, form *ShippingMethodForm, id uint64, price float64) error
}

type DBOrderStatusManager interface {
	ExistsOrderStatus(ctx context.Context, name string) (bool, error)
	GetDeliveriedOrderStatus(ctx context.Context, statusForm *OrderStatusForm) (uint64, error)
	GetIdleOrderStatus(ctx context.Context, statusForm *OrderStatusForm) (uint64, error)
	GetOrderStatusByName(ctx context.Context, name string, statusForm *OrderStatusForm) (uint64, error)
	GetOrderStatusCount(ctx context.Context) (uint64, error)
	GetOrderStatuses(ctx context.Context, status []uint64, statusForms []*OrderStatusForm, skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*OrderStatusForm, error)
	NewOrderStatus(ctx context.Context, name string, statusForm *OrderStatusForm) (uint64, error)
	RemoveAllOrderStatuses(ctx context.Context) error
	RemoveOrderStatus(ctx context.Context, status uint64) error
	InitOrderStatusManager(ctx context.Context) error
	FillOrderStatusWithID(ctx context.Context, sid uint64, statusForm *OrderStatusForm) error
}

type DBOrderStatus interface {
	GetOrderStatusName(ctx context.Context, form *OrderStatusForm, id uint64) (string, error)
	IsOrderStatusDeliveried(ctx context.Context, form *OrderStatusForm, id uint64) (bool, error)
	SetOrderStatusName(ctx context.Context, form *OrderStatusForm, id uint64, name string) error
}

type DBUserReviewResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}

type DBUserReviewManager[AccountID comparable] interface {
	InitUserReviewManager(ctx context.Context) error
	NewUserReview(ctx context.Context, userAccountID AccountID, productItemID uint64, ratingValue int32, comment string, reviewForm *UserReviewForm[AccountID]) (uint64, error)
	RemoveUserReview(ctx context.Context, reviewID uint64) error
	RemoveAllUserReviews(ctx context.Context) error
	GetUserReviewCount(ctx context.Context) (uint64, error)
	GetUserReviews(ctx context.Context, ids []DBUserReviewResult[AccountID], reviewForms []*UserReviewForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserReviewResult[AccountID], []*UserReviewForm[AccountID], error)
	GetUserReviewsForProductItem(ctx context.Context, productItemID uint64, ids []DBUserReviewResult[AccountID], reviewForms []*UserReviewForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserReviewResult[AccountID], []*UserReviewForm[AccountID], error)
	GetUserReviewsForAccount(ctx context.Context, accountID AccountID, ids []DBUserReviewResult[AccountID], reviewForms []*UserReviewForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserReviewResult[AccountID], []*UserReviewForm[AccountID], error)
	FillUserReviewWithID(ctx context.Context, rid uint64, reviewForm *UserReviewForm[AccountID]) error
}

type DBUserReview[AccountID comparable] interface {
	GetUserReviewRatingValue(ctx context.Context, form *UserReviewForm[AccountID], reviewID uint64) (int32, error)
	SetUserReviewRatingValue(ctx context.Context, form *UserReviewForm[AccountID], reviewID uint64, rating int32) error
	GetUserReviewComment(ctx context.Context, form *UserReviewForm[AccountID], reviewID uint64) (string, error)
	SetUserReviewComment(ctx context.Context, form *UserReviewForm[AccountID], reviewID uint64, comment string) error
	GetUserReviewProductItem(ctx context.Context, form *UserReviewForm[AccountID], reviewID uint64, productItemForm *ProductItemForm[AccountID], fs FileStorage) (uint64, error)
}

type DBProductItemSubscriptionManager[AccountID comparable] interface {
	InitProductItemSubscriptionManager(ctx context.Context) error
	NewProductItemSubscription(ctx context.Context, userAccountID AccountID, productItemID uint64, subscribedAt time.Time, expiresAt time.Time, duration time.Duration, subscriptionType string, autoRenew bool, form *ProductItemSubscriptionForm[AccountID]) (uint64, error)
	RemoveProductItemSubscription(ctx context.Context, subscriptionID uint64) error
	RemoveAllProductItemSubscriptions(ctx context.Context) error
	GetProductItemSubscriptionCount(ctx context.Context) (uint64, error)
	GetProductItemSubscriptions(ctx context.Context, ids []uint64, forms []*ProductItemSubscriptionForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*ProductItemSubscriptionForm[AccountID], error)
	GetUserProductItemSubscriptions(ctx context.Context, userAccountID AccountID, ids []uint64, forms []*ProductItemSubscriptionForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*ProductItemSubscriptionForm[AccountID], error)
	GetUserProductItemSubscriptionCount(ctx context.Context, userAccountID AccountID) (uint64, error)
	GetProductItemSubscriptionsForProduct(ctx context.Context, productItemID uint64, ids []uint64, forms []*ProductItemSubscriptionForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*ProductItemSubscriptionForm[AccountID], error)
	GetProductItemSubscriptionCountForProduct(ctx context.Context, productItemID uint64) (uint64, error)
	GetExpiredSubscriptionsForRenewal(ctx context.Context, now time.Time, ids []uint64, forms []*ProductItemSubscriptionForm[AccountID], limit int64) ([]uint64, []*ProductItemSubscriptionForm[AccountID], error)
	FillProductItemSubscriptionWithID(ctx context.Context, sid uint64, subscriptionForm *ProductItemSubscriptionForm[AccountID]) error
}

type DBProductItemSubscription[AccountID comparable] interface {
	GetProductItemSubscriptionUserAccountID(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (AccountID, error)
	GetProductItemSubscriptionProductItem(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, productItemForm *ProductItemForm[AccountID], fs FileStorage) (uint64, error)
	SetProductItemSubscriptionProductItem(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, productItemID uint64, fs FileStorage) error
	GetProductItemSubscriptionSubscribedAt(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (time.Time, error)
	SetProductItemSubscriptionSubscribedAt(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, subscribedAt time.Time) error
	GetProductItemSubscriptionExpiresAt(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (time.Time, error)
	SetProductItemSubscriptionExpiresAt(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, expiresAt time.Time) error
	GetProductItemSubscriptionDuration(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (time.Duration, error)
	SetProductItemSubscriptionDuration(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, duration time.Duration) error
	GetProductItemSubscriptionType(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (string, error)
	SetProductItemSubscriptionType(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, subscriptionType string) error
	IsProductItemSubscriptionAutoRenew(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (bool, error)
	SetProductItemSubscriptionAutoRenew(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, autoRenew bool) error
	IsProductItemSubscriptionActive(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) (bool, error)
	SetProductItemSubscriptionActive(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64, isActive bool) error
	CancelProductItemSubscription(ctx context.Context, form *ProductItemSubscriptionForm[AccountID], subscriptionID uint64) error
}

type DBUserFactorManager[AccountID comparable] interface {
	InitUserFactorManager(ctx context.Context) error
	GetUserFactorCount(ctx context.Context, aid AccountID) (uint64, error)
	GetUserFactors(ctx context.Context, aid AccountID, ids []uint64, factorForms []*UserFactorForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserFactorForm[AccountID], error)
	RemoveAllUserFactors(ctx context.Context) error
	RemoveUserAccountFactors(ctx context.Context, aid AccountID) error
	FillUserFactorWithID(ctx context.Context, fid uint64, factorForm *UserFactorForm[AccountID]) error
}

type DBUserFactor[AccountID comparable] interface {
	GetUserFactorProducts(ctx context.Context, form *UserFactorForm[AccountID], fid uint64) (json.RawMessage, error)
	SetUserFactorProducts(ctx context.Context, form *UserFactorForm[AccountID], fid uint64, products json.RawMessage) error
	GetUserFactorDiscount(ctx context.Context, form *UserFactorForm[AccountID], fid uint64) (float64, error)
	SetUserFactorDiscount(ctx context.Context, form *UserFactorForm[AccountID], fid uint64, discount float64) error
	GetUserFactorTax(ctx context.Context, form *UserFactorForm[AccountID], fid uint64) (float64, error)
	SetUserFactorTax(ctx context.Context, form *UserFactorForm[AccountID], fid uint64, tax float64) error
	GetUserFactorAmountPaid(ctx context.Context, form *UserFactorForm[AccountID], fid uint64) (float64, error)
	SetUserFactorAmountPaid(ctx context.Context, form *UserFactorForm[AccountID], fid uint64, amountPaid float64) error
}

type DBUserDiscountResult[AccountID comparable] struct {
	ID  uint64
	AID AccountID
}

type DBUserDiscountManager[AccountID comparable] interface {
	InitUserDiscountManager(ctx context.Context) error
	NewUserDiscount(ctx context.Context, ownerAccountID AccountID, value float64, validCount int64, discountForm *UserDiscountForm[AccountID]) (uint64, error)
	RemoveUserDiscount(ctx context.Context, discountID uint64) error
	RemoveAllUserDiscounts(ctx context.Context) error
	GetUserDiscountCount(ctx context.Context) (uint64, error)
	GetUserDiscounts(ctx context.Context, ids []DBUserDiscountResult[AccountID], discountForms []*UserDiscountForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]DBUserDiscountResult[AccountID], []*UserDiscountForm[AccountID], error)
	GetUserDiscountByCode(ctx context.Context, code string, discountForm *UserDiscountForm[AccountID]) (DBUserDiscountResult[AccountID], error)
	ExistsUserDiscountCode(ctx context.Context, code string) (bool, error)
	GetUserDiscountsForAccount(ctx context.Context, ownerAccountID AccountID, ids []uint64, discountForms []*UserDiscountForm[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]uint64, []*UserDiscountForm[AccountID], error)
	FillUserDiscountWithID(ctx context.Context, did uint64, discountForm *UserDiscountForm[AccountID]) error
}

type DBUserDiscount[AccountID comparable] interface {
	GetUserDiscountCode(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64) (string, error)
	SetUserDiscountCode(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64, code string) error
	GetUserDiscountValue(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64) (float64, error)
	SetUserDiscountValue(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64, value float64) error
	GetUserDiscountValidCount(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64) (int64, error)
	SetUserDiscountValidCount(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64, validCount int64) error
	DecrementUserDiscountValidCount(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64) error
	GetUserDiscountUsedBy(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64) ([]AccountID, error)
	AddUserDiscountUsedBy(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64, accountID AccountID) error
	HasUserUsedDiscount(ctx context.Context, form *UserDiscountForm[AccountID], discountID uint64, accountID AccountID) (bool, error)
}
