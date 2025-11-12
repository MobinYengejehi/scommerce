# Contracts Reference

This document provides a complete reference for all interface contracts in S-Commerce. Contracts define the behavior of components without specifying implementation details.

## Table of Contents

- [Core Lifecycle Contracts](#core-lifecycle-contracts)
- [User Management Contracts](#user-management-contracts)
- [Product Management Contracts](#product-management-contracts)
- [Shopping Cart Contracts](#shopping-cart-contracts)
- [Order Management Contracts](#order-management-contracts)
- [Payment Contracts](#payment-contracts)
- [Address Contracts](#address-contracts)
- [Review Contracts](#review-contracts)
- [Reference Data Contracts](#reference-data-contracts)

##

 Core Lifecycle Contracts

These fundamental contracts define lifecycle behavior used throughout S-Commerce.

### GeneralInitiative

**Purpose:** Defines initialization behavior for components

**Method:**
- `Init(ctx context.Context) error` - Initialize the component, set up resources, create database schema

**Usage Context:** Called once during application startup after component creation

**When to Implement:** Any component requiring initialization (managers, entities if needed)

---

### GeneralClosable

**Purpose:** Defines cleanup behavior for components

**Method:**
- `Close(ctx context.Context) error` - Clean up resources, close connections, release memory

**Usage Context:** Called once during application shutdown

**When to Implement:** Components holding resources that need cleanup

---

### GeneralPulsable

**Purpose:** Defines periodic maintenance behavior

**Method:**
- `Pulse(ctx context.Context) error` - Perform periodic maintenance, cleanup expired data

**Usage Context:** Called regularly during application runtime (e.g., every minute)

**When to Implement:** Components requiring periodic tasks (OTP cleanup, cache expiration)

---

### GeneralAppObject

**Purpose:** Combines all three lifecycle contracts

**Composed Of:**
- GeneralInitiative
- GeneralClosable
- GeneralPulsable

**Usage Context:** Most managers and entities implement this interface

**Benefits:** Uniform lifecycle management across all components

---

## User Management Contracts

### UserAccountManager[AccountID]

**Purpose:** Manages collection of user accounts, handles authentication and OTP

**Lifecycle:** Singleton per application, lives for application lifetime

**Key Responsibilities:**
- Create and remove user accounts
- Authenticate users with password and two-factor
- Generate and validate OTP codes
- Retrieve accounts by token or ID
- List accounts with pagination

**Primary Methods:**

| Method | Purpose | Key Parameters | Returns |
|--------|---------|----------------|---------|
| NewAccount | Create new user account | token, password, twoFactor code | UserAccount instance |
| GetAccount | Retrieve account by token | token | UserAccount instance |
| GetAccountWithID | Retrieve account by ID | account ID | UserAccount instance |
| Authenticate | Login with credentials | token, password, twoFactor | UserAccount instance |
| RequestTwoFactor | Generate OTP code | token (username/email) | OTP code string |
| ValidateTwoFactor | Check OTP code | token, code | boolean (valid or not) |
| GetAccounts | List accounts with pagination | skip, limit, order | Array of UserAccount |
| RemoveAccount | Delete account | UserAccount instance | error |

**Related Contracts:**
- UserAccount: Individual account entity managed by this manager
- UserRole: Accounts have roles
- DBUserAccountManager: Database operations for accounts

**Usage Pattern:**
```
Flow: Request OTP → User receives code → Create account with code
```

---

### UserAccount[AccountID]

**Purpose:** Represents an individual user account with all properties and capabilities

**Lifecycle:** Created by UserAccountManager, lives until deleted or application shutdown

**Key Responsibilities:**
- Manage account properties (name, email, password, role)
- Handle wallet and currency operations
- Manage ban status and trading permissions
- Create and manage addresses
- Create and manage payment methods
- Create and manage shopping carts
- Access order history
- Profile image management

**Property Methods:**

| Property | Get Method | Set Method | Notes |
|----------|------------|------------|-------|
| ID | GetID | N/A | Immutable identifier |
| Token | GetToken | SetToken | Username/email |
| Password | GetPassword | SetPassword | Should be hashed |
| FirstName | GetFirstName | SetFirstName | User's first name |
| LastName | GetLastName | SetLastName | User's last name |
| Role | GetRole | SetRole | Returns UserRole instance |
| UserLevel | GetUserLevel | SetUserLevel | Admin level, custom use |
| Bio | GetBio | SetBio | Profile biography |
| WalletCurrency | GetWalletCurrency | SetWalletCurrency | Account balance |
| ProfileImages | GetProfileImages | SetProfileImages | Array of images |
| LastUpdatedAt | GetLastUpdatedAt | SetLastUpdatedAt | Last modification time |

**State Methods:**

| Method | Purpose | Parameters | Returns |
|--------|---------|------------|---------|
| IsActive | Check if account is active | none | boolean |
| SetActive | Enable/disable account | state boolean | error |
| IsSuperUser | Check super user status | none | boolean |
| SetSuperUser | Set super user status | state boolean | error |
| IsBanned | Check ban status | none | reason string (empty if not banned) |
| Ban | Ban account | till duration, reason | error |
| Unban | Remove ban | none | error |
| IsTradingAllowed | Check trading permission | none | boolean |
| AllowTrading | Set trading permission | state boolean | error |

**Financial Methods:**

| Method | Purpose | Parameters | Returns |
|--------|---------|------------|---------|
| ChargeWallet | Add currency to wallet | amount float64 | error |
| TransferCurrency | Send currency to another account | to account, amount | error |
| Fine | Deduct penalty from wallet | amount float64 | error |
| SetPenalty | Set penalty amount | penalty float64 | error |
| HasPenalty | Check if wallet is negative | none | boolean |
| CalculateTotalDepts | Calculate total debt including penalties | none | currency amount |
| CalculateTotalDeptsWithoutPenalty | Calculate debt from carts only | none | currency amount |

**Relationship Methods:**

| Method | Purpose | Returns |
|--------|---------|---------|
| NewShoppingCart | Create cart for this user | UserShoppingCart |
| GetShoppingCarts | List user's carts | Array of UserShoppingCart |
| NewAddress | Create address for this user | UserAddress |
| GetAddresses | List user's addresses | Array of UserAddress |
| GetDefaultAddress | Get default shipping address | UserAddress |
| NewPaymentMethod | Create payment method | UserPaymentMethod |
| GetPaymentMethods | List payment methods | Array of UserPaymentMethod |
| GetDefaultPaymentMethod | Get default payment | UserPaymentMethod |
| GetOrders | List user's orders | Array of UserOrder |
| GetUserReviews | List user's reviews | Array of UserReview |

**Special Behaviors:**
- **Ban System**: Temporary bans with duration and reason tracking
- **Trading Permissions**: Control ability to purchase
- **Wallet**: Can go negative (debt), tracked separately from shopping cart debt
- **Two-Factor**: Required for account creation and sensitive operations
- **Default Resources**: One default address and payment method per account

---

### UserRoleManager

**Purpose:** Manages user role definitions

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewUserRole | Create role with name |
| GetUserRoleByName | Find role by name |
| GetUserRoles | List all roles with pagination |
| RemoveUserRole | Delete role |
| ExistsUserRole | Check if role name exists |

**Usage:** Define roles like "Customer", "Admin", "Moderator" and assign to users

---

### UserRole

**Purpose:** Represents a user role definition

**Properties:**
- ID: Unique identifier
- Name: Role name (e.g., "Admin", "Customer")

**Usage:** Assigned to UserAccount via SetRole method

---

## Product Management Contracts

### ProductManager[AccountID]

**Purpose:** Manages product catalog including categories, products, and items

**Key Responsibilities:**
- Create and organize product categories
- Search products and items
- Coordinate product-related operations

**Primary Methods:**

| Method | Purpose | Key Parameters |
|--------|---------|----------------|
| NewProductCategory | Create category | name, parent category |
| GetProductCategories | List categories | skip, limit, order |
| SearchForProductCategories | Find categories | search text, deep search flag |
| SearchForProducts | Find products | search text, category filter |
| SearchForProductItems | Find items | search text, product/category filters |

**Search Capabilities:**
- **Text Search**: Find by name, description
- **Deep Search**: Search recursively through category hierarchy
- **Filtering**: Filter by category, product

---

### ProductCategory[AccountID]

**Purpose:** Represents a product category in hierarchical structure

**Key Responsibilities:**
- Organize products into groups
- Support category hierarchy (parent/child)
- Manage products within category

**Properties:**
- ID: Unique identifier
- Name: Category name
- ParentProductCategory: Parent in hierarchy (nil for root)

**Product Management:**

| Method | Purpose |
|--------|---------|
| NewProduct | Create product in this category |
| GetProducts | List products in category |
| RemoveProduct | Delete product from category |
| GetProductCount | Count products in category |

**Hierarchy:** Categories can be nested (e.g., Electronics → Smartphones → Apple)

---

### Product[AccountID]

**Purpose:** Represents a product with variants (product items)

**Properties:**
- ID: Unique identifier
- Name: Product name
- Description: Product description
- Images: Array of product images
- ProductCategory: Category this product belongs to

**Product Items (Variants):**

| Method | Purpose | Parameters |
|--------|---------|------------|
| AddProductItem | Create variant | SKU, name, price, quantity, images, attributes |
| GetProductItems | List variants | skip, limit, order |
| RemoveProductItem | Delete variant | ProductItem instance |

**Use Case:** A "T-Shirt" product with items for each size/color combination

---

### ProductItem[AccountID]

**Purpose:** Represents a specific SKU with price, inventory, and attributes

**Properties:**
- ID: Unique identifier
- SKU: Stock Keeping Unit (unique code)
- Name: Item name (e.g., "T-Shirt Large Red")
- Price: Unit price
- QuantityInStock: Available inventory
- Images: Item-specific images
- Attributes: JSON attributes (size, color, etc.)
- Product: Parent product

**Inventory Management:**

| Method | Purpose |
|--------|---------|
| GetQuantityInStock | Get current stock |
| SetQuantityInStock | Set stock level |
| AddQuantityInStock | Adjust stock (positive or negative) |

**Reviews:**

| Method | Purpose |
|--------|---------|
| GetUserReviews | Get reviews for this item |
| CalculateAverageRating | Get average rating from reviews |

**Attributes:** JSON field for flexible properties like {"size": "L", "color": "Red", "material": "Cotton"}

---

## Shopping Cart Contracts

### UserShoppingCartManager[AccountID]

**Purpose:** Manages collection of all shopping carts across users

**Key Methods:**

| Method | Purpose |
|--------|---------|
| GetShoppingCarts | List all carts (admin use) |
| GetShoppingCartBySessionText | Find cart by session |
| RemoveAllShoppingCarts | Delete all carts (cleanup) |

**Note:** Individual users create carts through UserAccount, not this manager

---

### UserShoppingCart[AccountID]

**Purpose:** Represents a user's shopping cart

**Properties:**
- ID: Unique identifier
- UserAccountID: Owner account
- SessionText: Session identifier

**Cart Item Management:**

| Method | Purpose |
|--------|---------|
| NewShoppingCartItem | Add item to cart |
| GetShoppingCartItems | List cart items |
| RemoveShoppingCartItem | Remove item |
| RemoveAllShoppingCartItems | Clear cart |
| GetShoppingCartItemCount | Count items |

**Financial:**

| Method | Purpose | Parameters |
|--------|---------|------------|
| CalculateDept | Calculate cart total | shipping method |

**Ordering:**

| Method | Purpose | Parameters |
|--------|---------|------------|
| Order | Convert cart to order | payment method, address, shipping method, comment |

**Workflow:** Add items → Calculate total → Order → Cart becomes UserOrder

---

### UserShoppingCartItem[AccountID]

**Purpose:** Represents an item in a shopping cart

**Properties:**
- ID: Unique identifier
- UserAccountID: Owner account  
- ShoppingCart: Parent cart
- ProductItem: The product item being purchased
- Quantity: How many units

**Quantity Management:**

| Method | Purpose |
|--------|---------|
| GetQuantity | Get current quantity |
| SetQuantity | Set quantity |
| AddQuantity | Adjust quantity (delta) |

**Financial:**

| Method | Purpose |
|--------|---------|
| CalculateDept | Calculate line total (price × quantity) |

---

## Order Management Contracts

### UserOrderManager[AccountID]

**Purpose:** Manages collection of all orders across users

**Key Methods:**

| Method | Purpose |
|--------|---------|
| GetUserOrders | List all orders (admin use) |
| GetUserOrderCount | Count total orders |
| RemoveAllUserOrders | Delete all orders (cleanup) |

---

### UserOrder[AccountID]

**Purpose:** Represents a completed order

**Properties:**
- ID: Unique identifier
- UserAccountID: Customer account
- OrderDate: When order was placed
- OrderTotal: Total price
- Status: Current order status
- PaymentMethod: Payment method used
- ShippingAddress: Delivery address
- ShippingMethod: Shipping option chosen
- UserComment: Customer notes
- DeliveryDate: When delivered
- DeliveryComment: Delivery notes

**Status Management:**

| Method | Purpose |
|--------|---------|
| GetStatus | Get current status |
| SetStatus | Update status |
| IsDeliveried | Check if delivered |
| Deliver | Mark as delivered with date and comment |

**Product Items:**

| Method | Purpose |
|--------|---------|
| GetProductItems | List items in order (snapshot) |
| SetProductItems | Set order items |

**Note:** Order items are snapshots - changes to product catalog don't affect existing orders

---

### OrderStatusManager

**Purpose:** Manages order status catalog

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewOrderStatus | Create status |
| GetOrderStatusByName | Find by name |
| GetOrderStatuses | List all statuses |
| GetDeliveriedOrderStatus | Get the "delivered" status |
| GetIdleOrderStatus | Get the "idle/pending" status |

**Special Statuses:** System relies on specific statuses for delivered and idle states

---

### OrderStatus

**Purpose:** Represents an order status definition

**Properties:**
- ID: Unique identifier
- Name: Status name (e.g., "Pending", "Shipped", "Delivered")

**Special Method:**
- `IsDeliveried() bool` - Returns true if this is the delivered status

---

## Payment Contracts

### PaymentTypeManager

**Purpose:** Manages payment type catalog (credit card, PayPal, etc.)

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewPaymentType | Create payment type |
| GetPaymentTypeByName | Find by name |
| GetPaymentTypes | List all types |
| ExistsPaymentType | Check existence |

**Usage:** Define types like "Credit Card", "PayPal", "Bank Transfer"

---

### PaymentType

**Purpose:** Represents a payment type definition

**Properties:**
- ID: Unique identifier
- Name: Type name

---

### UserPaymentMethodManager[AccountID]

**Purpose:** Manages collection of all payment methods across users

**Key Methods:**

| Method | Purpose |
|--------|---------|
| GetPaymentMethods | List all payment methods (admin) |
| GetPaymentMethodCount | Count total payment methods |

---

### UserPaymentMethod[AccountID]

**Purpose:** Represents a user's payment method

**Properties:**
- ID: Unique identifier
- UserAccountID: Owner account
- PaymentType: Type of payment
- Provider: Payment provider (e.g., "Visa", "MasterCard")
- AccountNumber: Account/card number (last 4 digits typically)
- ExpiryDate: When payment method expires
- IsDefault: Whether this is the default payment method

**Expiry Tracking:**

| Method | Purpose |
|--------|---------|
| GetExpiryDate | Get expiration date |
| SetExpiryDate | Update expiration |
| IsExpired | Check if expired |

**Default Management:**
- Each user can have one default payment method
- Setting a method as default removes default from others

---

## Address Contracts

### UserAddressManager[AccountID]

**Purpose:** Manages collection of all addresses across users

**Key Methods:**

| Method | Purpose |
|--------|---------|
| GetAddresses | List all addresses (admin) |
| GetAddressCount | Count total addresses |

---

### UserAddress[AccountID]

**Purpose:** Represents a shipping/billing address

**Properties:**
- ID: Unique identifier
- UserAccountID: Owner account
- UnitNumber: Apartment/unit number
- StreetNumber: Street number
- AddressLine1: Primary address line
- AddressLine2: Secondary address line
- City: City name
- Region: State/province/region
- PostalCode: ZIP/postal code
- Country: Country reference
- IsDefault: Whether this is the default address

**Address Components:** Support international address formats with flexible fields

**Default Management:**
- Each user can have one default address
- Used automatically during checkout

---

## Review Contracts

### UserReviewManager[AccountID]

**Purpose:** Manages product reviews

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewUserReview | Create review for product item |
| GetUserReviews | List all reviews |
| GetUserReviewsForProductItem | Get reviews for specific item |
| GetUserReviewsForAccount | Get reviews by specific user |
| RemoveUserReview | Delete review |

**Usage:** Users review ProductItem instances (not Products)

---

### UserReview[AccountID]

**Purpose:** Represents a user's product review

**Properties:**
- ID: Unique identifier
- UserAccountID: Reviewer account
- ProductItem: Item being reviewed
- RatingValue: Numeric rating (e.g., 1-5)
- Comment: Review text

**Methods:**

| Method | Purpose |
|--------|---------|
| GetRatingValue | Get rating |
| SetRatingValue | Update rating |
| GetComment | Get review text |
| SetComment | Update review text |
| GetProductItem | Get item being reviewed |

---

## Reference Data Contracts

### CountryManager

**Purpose:** Manages country reference data

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewCountry | Create country |
| GetCountryByName | Find by name |
| GetCountries | List all countries |
| ExistsCountry | Check existence |

**Usage:** Used by addresses to specify country

---

### Country

**Purpose:** Represents a country definition

**Properties:**
- ID: Unique identifier
- Name: Country name

---

### ShippingMethodManager

**Purpose:** Manages shipping options

**Key Methods:**

| Method | Purpose |
|--------|---------|
| NewShippingMethod | Create shipping option |
| GetShippingMethodByName | Find by name |
| GetShippingMethods | List all methods |

**Usage:** Define options like "Standard", "Express", "Overnight"

---

### ShippingMethod

**Purpose:** Represents a shipping option

**Properties:**
- ID: Unique identifier
- Name: Method name
- Price: Shipping cost

**Usage:** Selected during checkout, price added to order total

---

## Common Patterns Across Contracts

### Pagination Pattern

Many Get methods support pagination:

**Parameters:**
- `skip int64`: Number of records to skip
- `limit int64`: Maximum records to return (capped at 500)
- `queueOrder QueueOrder`: "asc" or "desc"

**Usage:** Efficiently retrieve large result sets in pages

### Form Object Pattern

Entities use form objects for data transfer and caching:

**Methods:**
- `ToFormObject(ctx) (*Form, error)` - Export current state to form
- `ApplyFormObject(ctx, form *Form) error` - Import state from form

**Purpose:** Performance optimization through caching

### Builtin Access Pattern

All managers and entities provide access to builtin implementation:

**Method:** `ToBuiltinObject(ctx) (*Builtin..., error)`

**Usage:** Downcast to concrete type for accessing builtin-specific features

### Generic Type Consistency

All user-related interfaces use the same AccountID type parameter:

- Ensures type safety across user domain
- Prevents mixing incompatible ID types
- Chosen once at App creation time

## Contract Implementation Guidelines

### When Implementing Contracts

1. **Implement All Methods**: Every method in the interface must be implemented
2. **Follow Semantics**: Respect the contract's intended behavior
3. **Handle Errors**: Return meaningful errors for failure cases
4. **Thread Safety**: Ensure concurrent access is safe
5. **Lifecycle**: Implement Init, Pulse, Close appropriately
6. **Context Respect**: Honor context cancellation and timeouts

### Testing Contract Implementations

1. **Interface Compliance**: Verify implementation satisfies interface
2. **Functional Tests**: Test each method's behavior
3. **Edge Cases**: Test boundary conditions
4. **Concurrency**: Test concurrent access
5. **Integration**: Test interaction with other components

## Next Steps

- Learn about implementing database contracts: [Database Integration](database-integration.md)
- Understand builtin implementations: [Managers](managers.md) and [Entities](entities.md)
- See contracts in action: [Examples](examples.md)
- Extend or customize: [Extending Builtin Objects](extending-builtin-objects.md)
