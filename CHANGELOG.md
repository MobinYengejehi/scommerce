# Changelog

## [Latest] - User Discount System Documentation

### 📚 Documentation: User Discounts (Promotional Codes)

Comprehensive documentation added for the User Discount system!

#### What's Documented

**Complete Guide**: [docs/user-discounts.md](docs/user-discounts.md)
- **Overview**: Promotional code and discount management
- **Architecture**: UserDiscountManager and UserDiscount components  
- **Database Schema**: Complete schema with unique code indexing
- **Usage Examples**: 30+ code examples for real-world scenarios
- **Integration Patterns**: 5 ready-to-use patterns
- **Best Practices**: 5 recommended practices
- **API Reference**: Complete method documentation

#### Key Features

**Discount Capabilities**:
- 💳 Auto-generated unique discount codes
- 🎁 Configurable usage limits (single-use, multi-use, unlimited)
- 🎯 Usage tracking (who used each code)
- 📊 Prevent duplicate redemptions
- 🔒 Thread-safe code generation
- ✅ Flexible discount values

**Use Cases**:
- Referral programs
- Influencer campaigns  
- Flash sales
- Email marketing codes
- Loyalty rewards
- Seasonal promotions

#### Integration Examples

1. **Referral Program**: Generate unique codes for users to share
2. **Influencer Campaign**: Create high-limit codes for influencers
3. **Flash Sale**: Limited-use, high-value discount codes
4. **Email Campaigns**: Personalized codes for each customer
5. **Loyalty Tiers**: Different discount values based on customer tier

#### API Highlights

**Manager Methods**:
- `NewUserDiscount(ctx, owner, value, validCount)` - Create discount with auto-generated code
- `GetUserDiscountByCode(ctx, code)` - Retrieve discount by redemption code
- `ExistsUserDiscountCode(ctx, code)` - Check if code exists

**Discount Methods**:
- `GetCode(ctx)` / `SetCode(ctx, code)` - Manage discount code
- `GetValue(ctx)` / `SetValue(ctx, value)` - Manage discount amount
- `GetValidCount(ctx)` / `DecrementValidCount(ctx)` - Track remaining uses
- `HasUserUsed(ctx, account)` - Check if user already redeemed
- `MarkAsUsedBy(ctx, account)` - Mark code as used

**Account Integration**:
- `account.NewDiscount(ctx, value, validCount)` - User creates own discount
- `account.GetDiscounts(ctx, ...)` - List user's discounts
- `account.GetDiscountCount(ctx)` - Count user's discounts

### Files Added

- `docs/user-discounts.md` - Complete documentation (898 lines)
- `README.md` - Updated with user discounts link

---

## [Previous] - User Factor Documentation

### 📚 Documentation: User Factors (Invoices/Receipts)

Comprehensive documentation added for the User Factor system!

#### What's Documented

**Complete Guide**: [docs/user-factors.md](docs/user-factors.md)
- **Overview**: What user factors are (فاکتور - invoices/receipts)
- **Architecture**: UserFactorManager and UserFactor components
- **Database Schema**: Complete schema with indexes
- **Usage Examples**: 25+ code examples for common scenarios
- **Integration Patterns**: 4 ready-to-use patterns
- **Best Practices**: 5 recommended practices
- **Performance Guide**: Optimization tips and indexing strategies
- **Troubleshooting**: Common issues and solutions
- **API Reference**: Complete method documentation

#### Key Topics Covered

**What You Can Do**:
- 📄 Generate formal invoices for purchases
- 🧾 Store digital receipts for customer reference
- 📊 Track all financial transactions per user
- 💰 Maintain detailed financial records for reporting
- 🔍 Enable customers to view purchase history
- 📈 Analyze sales, discounts, and tax data
- 🧮 Generate tax reports from stored data

**Technical Features**:
- JSONB storage for flexible product data
- Financial tracking (discount, tax, amount paid)
- Indexed queries for fast retrieval
- User-centric factor lookup
- Thread-safe operations
- Extensible JSON structure
- Sales and revenue analytics

#### Integration Examples

**Pattern 1**: Create Factor from Order
```go
func createFactorFromOrder(ctx context.Context, order UserOrder[int64]) error {
    // Build products JSON from order items
    // Calculate discount and tax
    // Store factor in database
}
```

**Pattern 2**: Generate PDF Invoice
```go
func generateInvoicePDF(ctx context.Context, factor UserFactor[int64]) ([]byte, error) {
    // Get factor details
    // Generate formatted invoice PDF
}
```

**Pattern 3**: Sales Analytics
```go
func analyzeSalesData(ctx context.Context, userAccount UserAccount[int64]) error {
    // Aggregate revenue, discounts, and tax
    // Track product sales counts
}
```

**Pattern 4**: Customer Purchase History
```go
func displayPurchaseHistory(ctx context.Context, userAccount UserAccount[int64]) error {
    // Paginated factor retrieval
    // Display invoice history
}
```

#### Database Schema

```sql
CREATE TABLE factors (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    products        JSONB NOT NULL,
    discount        DOUBLE PRECISION NOT NULL DEFAULT 0,
    tax             DOUBLE PRECISION NOT NULL DEFAULT 0,
    amount_paid     DOUBLE PRECISION NOT NULL,
    
    INDEX idx_factors_user_id (user_id),
    INDEX idx_factors_amount_paid (amount_paid)
);
```

#### Products JSON Structure

```json
[
    {
        "product_id": 123,
        "product_name": "Wireless Mouse",
        "sku": "MOUSE-001",
        "quantity": 2,
        "unit_price": 29.99,
        "subtotal": 59.98,
        "attributes": {
            "color": "black",
            "warranty": "1 year"
        }
    }
]
```

### API Reference

**UserFactorManager Methods**:
- `GetUserFactors(ctx, account, factors, skip, limit, order)` - Retrieve user's factors
- `GetUserFactorCount(ctx, account)` - Count user's factors
- `RemoveAllUserFactors(ctx)` - Delete all factors
- `Init(ctx)` - Initialize database schema

**UserFactor Methods**:
- `GetProducts(ctx)` / `SetProducts(ctx, products)` - Manage products JSON
- `GetDiscount(ctx)` / `SetDiscount(ctx, discount)` - Manage discount
- `GetTax(ctx)` / `SetTax(ctx, tax)` - Manage tax
- `GetAmountPaid(ctx)` / `SetAmountPaid(ctx, amount)` - Manage amount paid

### Files Added

- `docs/user-factors.md` - Complete documentation (715 lines)
- `README.md` - Updated with user factors link

---

## [Previous] - Product Subscriptions & Payment Improvements

### 🎉 New Feature: Product Item Subscriptions

Complete subscription management system for recurring billing and time-based access control!

#### What's New

**Subscription System**:
- **Recurring Billing**: Automatic subscription renewals with customizable periods
- **Time-Based Access**: Grant time-limited access to products and services
- **Flexible Durations**: Support any subscription period (daily, weekly, monthly, annual, custom)
- **Auto-Renewal**: Configurable automatic renewal on expiration
- **Subscription Types**: Categorize subscriptions (basic, premium, trial, annual, etc.)
- **Wallet Integration**: Seamless payment processing using user wallets

**Manager Features**:
- Create and manage subscriptions for users and products
- Process expired subscriptions automatically via `Pulse()` or `ProcessExpiredSubscriptions()`
- Query subscriptions by user, product, or globally
- Custom renewal handlers for flexible billing logic
- Concurrent renewal processing (up to 10 simultaneous renewals)

**Subscription Entity**:
- Track subscription status (active, expired, cancelled)
- Flexible expiration dates and durations
- Auto-renew toggles
- Product item associations
- Subscription type categorization

#### Payment Logic Enhancement

**Conditional Wallet Deduction**:
- Renewal handlers now return `(success bool, amountCharged float64, error)`
- Only deduct from wallet if `amountCharged > 0`
- Support free renewals with `amountCharged = 0`
- Enable promotional pricing, discounts, and credits

**Default Renewal Handler**:
```go
// Checks wallet balance, validates funds, extends subscription
// Returns: (true, price, nil) on success
//          (false, 0, nil) on insufficient funds
//          (false, 0, error) on errors
```

#### API Changes

**New Interfaces**:
- `ProductItemSubscriptionManager[AccountID]`
- `ProductItemSubscription[AccountID]`
- `RenewalHandlerFunc[AccountID]`

**New Methods**:
```go
// Manager methods
NewSubscription(ctx, account, productItem, duration, type, autoRenew)
GetUserSubscriptions(ctx, account, subs, skip, limit, order)
GetProductItemSubscriptions(ctx, item, subs, skip, limit, order)
SetRenewalHandler(ctx, handler)
ProcessExpiredSubscriptions(ctx)

// Subscription methods
GetExpiresAt(ctx) / SetExpiresAt(ctx, time)
GetDuration(ctx) / SetDuration(ctx, duration)
IsAutoRenew(ctx) / SetAutoRenew(ctx, enabled)
IsActive(ctx) / SetActive(ctx, active)
IsExpired(ctx)
Cancel(ctx)
```

### Use Cases

**Perfect For**:
- 🎵 Music/video streaming services
- 📰 Digital magazine subscriptions
- 🎮 Gaming memberships and season passes
- 💼 SaaS product billing
- 🏋️ Fitness program access
- 📚 Online course memberships
- ☁️ Cloud storage tiers

### Technical Details

**Database**:
- New `product_item_subscriptions` table
- Indexed on `expires_at`, `user_account_id`, `product_item_id`
- Supports concurrent renewal processing

**Performance**:
- Concurrent processing with semaphore (10 workers)
- Batch processing up to 1000 subscriptions
- Indexed queries for efficient expiration checks

**Error Handling**:
- Failed renewals remain expired for retry
- Graceful handling of insufficient funds
- Proper error propagation and logging

**Files Modified**:
- `scommerce/contracts.go` - Added subscription interfaces
- `scommerce/db_contracts.go` - Added database contracts
- `scommerce/product_item_subscription.go` - Core implementation
- `scommerce/account.go` - Interface compatibility updates
- `docs/product-subscriptions.md` - Complete documentation

### Migration

**For Existing Projects**:

1. **Database Migration**: Run `Init()` on subscription manager to create tables
2. **Update Code**: Access via `app.SubscriptionManager`
3. **Schedule Renewals**: Call `manager.Pulse()` periodically (recommended: hourly)

```go
// Initialize subscription manager
subManager := app.GetProductItemSubscriptionManager()
if err := subManager.Init(ctx); err != nil {
    return err
}

// Schedule periodic renewal processing
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        subManager.Pulse(ctx)
    }
}()
```

### Documentation

- **Complete Guide**: [docs/product-subscriptions.md](docs/product-subscriptions.md)
- **README Updated**: Added subscription features to main README
- **Examples**: 15+ code examples for common scenarios
- **API Reference**: Full method documentation

### Breaking Changes

⚠️ **RenewalHandlerFunc Signature Changed**:
```go
// Old:
func(ctx, subscription, account, productItem) (bool, error)

// New:
func(ctx, subscription, account, productItem) (success bool, amountCharged float64, err error)
```

If you implemented custom renewal handlers, update them to return the charge amount.

### Performance Impact

- ✅ Minimal overhead - subscriptions processed in background
- ✅ Indexed queries for fast expiration lookups
- ✅ Controlled concurrency prevents system overload
- ✅ Batch processing limits resource usage

---

## [Previous] - Item Attributes Feature

### 🎉 New Feature: Item-Level Attributes

Shopping cart items and order items now support custom attributes! This powerful feature allows you to store user-specific customizations, variant selections, and metadata with each item.

### What's New

#### Shopping Cart Items
- **New Parameter**: `NewShoppingCartItem()` now accepts `attrs json.RawMessage`
- **New Methods**: 
  - `GetAttributes(ctx) (json.RawMessage, error)`
  - `SetAttributes(ctx, attrs json.RawMessage) error`
- **Database**: `shopping_cart_items.attributes` JSONB column added

#### Order Items
- **Automatic Preservation**: Attributes flow automatically from cart to order
- **Structure Updated**: `UserOrderProductItem` now includes `Attributes json.RawMessage`
- **Retrieval**: `GetProductItems()` returns items with full attribute data

### Use Cases

Store any customization data with items:
- Product variants (size, color, material)
- Engraving and personalization
- Gift wrapping and messages
- Delivery preferences per item
- Bundle configurations
- Add-on selections

### Migration

#### For Existing Projects

**Database Migration:**
```sql
-- PostgreSQL migration
ALTER TABLE shopping_cart_items ADD COLUMN attributes JSONB;
```

**Code Updates:**
```go
// Before
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity)

// After (nil for no attributes)
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity, nil)

// Or with attributes
attrs := json.RawMessage(`{"color": "red", "size": "large"}`)
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity, attrs)
```

**Re-initialize:** Run your application's `Init()` to update the PostgreSQL functions.

### Example Usage

```go
import "encoding/json"

// Add item with customization
attrs := json.RawMessage(`{
    "size": "large",
    "color": "blue",
    "engraving": "Happy Birthday!",
    "gift_wrap": true
}`)

cartItem, err := cart.NewShoppingCartItem(ctx, productItem, 2, attrs)

// Attributes are preserved when ordering
order, err := cart.Order(ctx, paymentMethod, address, shippingMethod, "comment")

// Retrieve order items with attributes
items, err := order.GetProductItems(ctx, nil, 0, 10, scommerce.QueueOrderAsc)
for _, item := range items {
    fmt.Printf("Attributes: %s\n", string(item.Attributes))
}
```

### Technical Details

**Type**: `json.RawMessage` (Go) / `JSONB` (PostgreSQL)

**Data Flow**:
1. Add to cart with attributes → Stored in `shopping_cart_items.attributes`
2. Order placed → PostgreSQL function reads attributes
3. Saved in `orders.product_items` JSON array
4. Retrieved with order items

**Files Modified**:
- `scommerce/shopping_cart_item.go`
- `scommerce/shopping_cart.go`
- `scommerce/order.go`
- `scommerce/contracts.go`
- `scommerce/db_contracts.go`
- `scommerce/db_samples/postgresql/shopping_cart_item.go`
- `scommerce/db_samples/postgresql/shopping_cart.go`
- `scommerce/db_samples/postgresql/order.go`

### Documentation

- **Complete Guide**: See [docs/item-attributes.md](docs/item-attributes.md)
- **README Updated**: Examples added to main README
- **API Reference**: All new methods documented

### Breaking Changes

⚠️ **Signature Change**: `NewShoppingCartItem()` now requires an additional `attrs json.RawMessage` parameter.

**Migration**: Pass `nil` if you don't need attributes:
```go
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity, nil)
```

### Performance Impact

- ✅ Minimal storage overhead (JSONB is efficient)
- ✅ No performance impact on existing queries
- ✅ Optional indexing available for attribute-based queries

### Future Enhancements

Potential future additions:
- Attribute validation middleware
- Attribute templates for common use cases
- Built-in attribute types for common scenarios
- Attribute-based search and filtering

---

## Previous Versions

See git history for earlier changes.
