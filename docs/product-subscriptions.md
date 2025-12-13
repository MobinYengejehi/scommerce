# Product Item Subscriptions

## Overview

The **Product Item Subscription** system enables recurring billing and time-based access control for digital goods, memberships, and SaaS-style products in your e-commerce platform. This feature automatically handles subscription lifecycles, renewals, payment processing, and expiration management.

## What Are Product Item Subscriptions?

Product item subscriptions allow customers to:
- **Subscribe to products** for a specified duration (monthly, yearly, etc.)
- **Automatic renewal** when subscriptions expire (if enabled)
- **Flexible billing** with custom renewal handlers
- **Time-based access** to digital products or services
- **Subscription management** (cancel, renew, check status)

### Use Cases

- 🎵 **Music/Video Streaming**: Monthly or annual access to content libraries
- 📰 **Digital Magazines**: Recurring access to publications
- 🎮 **Gaming Memberships**: Premium features or content access
- 💼 **SaaS Products**: Software-as-a-Service subscriptions
- 🏋️ **Fitness Programs**: Recurring access to workout content
- 📚 **Online Courses**: Time-limited access to educational materials
- ☁️ **Cloud Storage**: Tiered storage plans with recurring billing

## Architecture

### Core Components

#### 1. ProductItemSubscriptionManager
The central manager that handles all subscription operations.

**Key Responsibilities**:
- Create new subscriptions
- Process expired subscriptions for renewal
- Manage subscription lifecycle
- Configure custom renewal logic

#### 2. ProductItemSubscription
Individual subscription entity representing a user's subscription to a product.

**Key Properties**:
- `ID`: Unique subscription identifier
- `UserAccountID`: The subscribing user
- `ProductItem`: The subscribed product
- `SubscribedAt`: Initial subscription timestamp
- `ExpiresAt`: When the subscription expires
- `Duration`: Subscription period length
- `SubscriptionType`: Category (e.g., "monthly", "annual", "premium")
- `AutoRenew`: Whether to automatically renew on expiration
- `IsActive`: Current subscription status

#### 3. RenewalHandler
Customizable function that determines renewal eligibility and charges.

**Signature**:
```go
func(ctx, subscription, account, productItem) (success bool, amountCharged float64, err error)
```

## Technical Implementation

### Database Schema

Subscriptions are stored with the following structure:

```sql
CREATE TABLE product_item_subscriptions (
    id              BIGINT PRIMARY KEY,
    user_account_id BIGINT NOT NULL,
    product_item_id BIGINT NOT NULL,
    subscribed_at   TIMESTAMP NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    duration        BIGINT NOT NULL,  -- Duration in nanoseconds
    subscription_type TEXT,
    auto_renew      BOOLEAN DEFAULT true,
    is_active       BOOLEAN DEFAULT true,
    
    FOREIGN KEY (user_account_id) REFERENCES user_accounts(id),
    FOREIGN KEY (product_item_id) REFERENCES product_items(id)
);

-- Index for finding expired subscriptions
CREATE INDEX idx_subscriptions_expires_at ON product_item_subscriptions(expires_at);
CREATE INDEX idx_subscriptions_user ON product_item_subscriptions(user_account_id);
CREATE INDEX idx_subscriptions_product ON product_item_subscriptions(product_item_id);
```

### Subscription Lifecycle

```
┌─────────────────┐
│  User Subscribes │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Active         │◄──────┐
│  Subscription   │       │
└────────┬────────┘       │
         │                │
         │ (Time passes)  │
         ▼                │
┌─────────────────┐       │
│  Expired        │       │
└────────┬────────┘       │
         │                │
         │ AutoRenew?     │
         ├─Yes────────────┤
         │                │
         │   ┌──────────────────┐
         │   │ RenewalHandler   │
         │   │ - Check funds    │
         │   │ - Calculate fee  │
         │   └────────┬─────────┘
         │            │
         │     Success & amount > 0?
         │            │
         │      ┌─────┴─────┐
         │      │           │
         │     Yes         No
         │      │           │
         │  ┌───▼───┐   ┌───▼────┐
         │  │Charge │   │ Skip   │
         │  │Wallet │   │Charge  │
         │  └───┬───┘   └────────┘
         │      │
         │   Renewed
         │      │
         └──────┘
         │
        No
         │
         ▼
┌─────────────────┐
│  Inactive       │
│  (Cancelled)    │
└─────────────────┘
```

## Usage Examples

### Basic Subscription Creation

```go
import (
    "context"
    "time"
    "github.com/MobinYengejehi/scommerce/scommerce"
)

// Create subscription manager
manager := scommerce.NewBuiltinProductItemSubscriptionManager(db, fileStorage, nil)

// Initialize the manager (creates database tables/functions)
if err := manager.Init(ctx); err != nil {
    return err
}

// Subscribe user to a product for 30 days
subscription, err := manager.NewSubscription(
    ctx,
    userAccount,
    productItem,
    30*24*time.Hour,      // Duration: 30 days
    "monthly",            // Subscription type
    true,                 // Auto-renew enabled
)
if err != nil {
    return err
}

fmt.Printf("Subscription created! Expires at: %v\n", subscription.GetExpiresAt(ctx))
```

### Checking Subscription Status

```go
// Check if subscription is active
isActive, err := subscription.IsActive(ctx)
if err != nil {
    return err
}

// Check if subscription has expired
isExpired, err := subscription.IsExpired(ctx)
if err != nil {
    return err
}

// Get expiration date
expiresAt, err := subscription.GetExpiresAt(ctx)
if err != nil {
    return err
}

fmt.Printf("Active: %v, Expired: %v, Expires: %v\n", isActive, isExpired, expiresAt)
```

### Managing Subscriptions

```go
// Get all user's subscriptions
userSubs, err := manager.GetUserSubscriptions(
    ctx,
    userAccount,
    nil,    // Output slice (nil = create new)
    0,      // Skip
    10,     // Limit
    scommerce.QueueOrderDesc,
)

// Get subscriptions for a specific product
productSubs, err := manager.GetProductItemSubscriptions(
    ctx,
    productItem,
    nil,
    0,
    100,
    scommerce.QueueOrderAsc,
)

// Cancel a subscription
err = subscription.Cancel(ctx)
// This sets AutoRenew=false and IsActive=false
```

### Manual Subscription Updates

```go
// Extend subscription manually
currentExpiry, _ := subscription.GetExpiresAt(ctx)
newExpiry := currentExpiry.Add(7 * 24 * time.Hour) // Add 7 days
err := subscription.SetExpiresAt(ctx, newExpiry)

// Change subscription type
err = subscription.SetSubscriptionType(ctx, "premium")

// Toggle auto-renewal
err = subscription.SetAutoRenew(ctx, false)

// Manually activate/deactivate
err = subscription.SetActive(ctx, true)
```

## Renewal System

### How Automatic Renewal Works

The renewal system runs through the `ProcessExpiredSubscriptions` method, typically called via `Pulse()`:

```go
// In your application's periodic task (e.g., every hour)
func periodicTask(ctx context.Context) {
    if err := manager.Pulse(ctx); err != nil {
        log.Printf("Subscription renewal error: %v", err)
    }
}
```

**Renewal Process**:

1. **Find Expired Subscriptions**: Queries database for subscriptions where `expires_at < NOW()` and `auto_renew = true`
2. **For Each Subscription**:
   - Load user account and product item
   - Call `RenewalHandler` to determine renewal eligibility
   - If successful and amount > 0, charge user's wallet
   - Update subscription's `expires_at` to new expiration date
3. **Concurrent Processing**: Handles up to 10 renewals simultaneously using goroutines
4. **Error Handling**: Failed renewals remain in expired state for retry on next pulse

### Default Renewal Handler

The built-in renewal handler implements standard subscription renewal logic:

```go
func defaultRenewalHandler(ctx, subscription, account, productItem) (bool, float64, error) {
    // 1. Get product price
    price, err := productItem.GetPrice(ctx)
    if err != nil {
        return false, 0, err
    }

    // 2. Check if user has sufficient funds
    walletBalance, err := account.GetWalletCurrency(ctx)
    if err != nil {
        return false, 0, err
    }

    if walletBalance < price {
        // Insufficient funds - renewal fails but no error
        return false, 0, nil
    }

    // 3. Update expiration date
    expiresAt, _ := subscription.GetExpiresAt(ctx)
    duration, _ := subscription.GetDuration(ctx)
    newExpiresAt := expiresAt.Add(duration)
    
    if err := subscription.SetExpiresAt(ctx, newExpiresAt); err != nil {
        return false, 0, err
    }

    // 4. Return success with amount to charge
    return true, price, nil
}
```

**Payment Deduction**:
After the renewal handler returns:
- If `success = true` and `amountCharged > 0`: Deduct from wallet
- If `success = true` and `amountCharged = 0`: No charge (free renewal)
- If `success = false`: No charge, subscription remains expired

### Custom Renewal Handlers

You can implement custom renewal logic for:
- **Discounts**: Reduced pricing for loyal customers
- **Promotional pricing**: Special offers during renewal
- **Tiered pricing**: Different rates based on subscription type
- **Grace periods**: Allow temporary access even without payment
- **Credits/Vouchers**: Apply account credits before charging

#### Example: Discounted Renewal for Long-Term Customers

```go
func loyaltyDiscountHandler(ctx context.Context, subscription ProductItemSubscription[int64], account UserAccount[int64], productItem ProductItem[int64]) (bool, float64, error) {
    // Get original price
    price, err := productItem.GetPrice(ctx)
    if err != nil {
        return false, 0, err
    }

    // Check how long user has been subscribed
    subscribedAt, err := subscription.GetSubscribedAt(ctx)
    if err != nil {
        return false, 0, err
    }

    subscriptionAge := time.Since(subscribedAt)
    discount := 0.0

    // Apply loyalty discount
    if subscriptionAge > 365*24*time.Hour {
        discount = 0.20 // 20% off for 1+ year subscribers
    } else if subscriptionAge > 180*24*time.Hour {
        discount = 0.10 // 10% off for 6+ month subscribers
    }

    discountedPrice := price * (1.0 - discount)

    // Check wallet balance
    walletBalance, err := account.GetWalletCurrency(ctx)
    if err != nil {
        return false, 0, err
    }

    if walletBalance < discountedPrice {
        return false, 0, nil
    }

    // Update expiration
    expiresAt, _ := subscription.GetExpiresAt(ctx)
    duration, _ := subscription.GetDuration(ctx)
    newExpiresAt := expiresAt.Add(duration)
    
    if err := subscription.SetExpiresAt(ctx, newExpiresAt); err != nil {
        return false, 0, err
    }

    return true, discountedPrice, nil
}

// Set custom handler
manager.SetRenewalHandler(ctx, loyaltyDiscountHandler)
```

#### Example: Free Trial with Paid Renewal

```go
func freeTrialHandler(ctx context.Context, subscription ProductItemSubscription[int64], account UserAccount[int64], productItem ProductItem[int64]) (bool, float64, error) {
    subscriptionType, err := subscription.GetSubscriptionType(ctx)
    if err != nil {
        return false, 0, err
    }

    // First renewal from "trial" to "paid"
    if subscriptionType == "trial" {
        // Convert to paid subscription
        if err := subscription.SetSubscriptionType(ctx, "paid"); err != nil {
            return false, 0, err
        }

        // Get price
        price, err := productItem.GetPrice(ctx)
        if err != nil {
            return false, 0, err
        }

        // Check funds
        walletBalance, err := account.GetWalletCurrency(ctx)
        if err != nil {
            return false, 0, err
        }

        if walletBalance < price {
            return false, 0, nil
        }

        // Extend subscription
        expiresAt, _ := subscription.GetExpiresAt(ctx)
        duration, _ := subscription.GetDuration(ctx)
        subscription.SetExpiresAt(ctx, expiresAt.Add(duration))

        return true, price, nil
    }

    // Regular paid renewal - use standard logic
    price, _ := productItem.GetPrice(ctx)
    walletBalance, _ := account.GetWalletCurrency(ctx)
    
    if walletBalance < price {
        return false, 0, nil
    }

    expiresAt, _ := subscription.GetExpiresAt(ctx)
    duration, _ := subscription.GetDuration(ctx)
    subscription.SetExpiresAt(ctx, expiresAt.Add(duration))

    return true, price, nil
}
```

#### Example: Credit-Based Renewal

```go
func creditBasedHandler(ctx context.Context, subscription ProductItemSubscription[int64], account UserAccount[int64], productItem ProductItem[int64]) (bool, float64, error) {
    price, err := productItem.GetPrice(ctx)
    if err != nil {
        return false, 0, err
    }

    // Check user's account credits (custom field)
    // In real implementation, you'd have a custom UserAccount extension
    accountCredits := getAccountCredits(account) // Your custom function
    
    chargeAmount := price
    
    // Apply credits first
    if accountCredits >= price {
        // Fully covered by credits
        deductCredits(account, price) // Your custom function
        chargeAmount = 0
    } else if accountCredits > 0 {
        // Partially covered by credits
        deductCredits(account, accountCredits)
        chargeAmount = price - accountCredits
    }

    // Check wallet for remaining amount
    if chargeAmount > 0 {
        walletBalance, _ := account.GetWalletCurrency(ctx)
        if walletBalance < chargeAmount {
            return false, 0, nil
        }
    }

    // Extend subscription
    expiresAt, _ := subscription.GetExpiresAt(ctx)
    duration, _ := subscription.GetDuration(ctx)
    subscription.SetExpiresAt(ctx, expiresAt.Add(duration))

    return true, chargeAmount, nil
}
```

## Advanced Features

### Subscription Types

Use subscription types to categorize and handle different tiers:

```go
// Create different subscription tiers
basicSub, _ := manager.NewSubscription(
    ctx, user, product, 30*24*time.Hour, "basic", true,
)

premiumSub, _ := manager.NewSubscription(
    ctx, user, product, 30*24*time.Hour, "premium", true,
)

annualSub, _ := manager.NewSubscription(
    ctx, user, product, 365*24*time.Hour, "annual", true,
)
```

Then query by type:

```go
subscriptions, _ := manager.GetUserSubscriptions(ctx, user, nil, 0, 100, scommerce.QueueOrderDesc)

for _, sub := range subscriptions {
    subType, _ := sub.GetSubscriptionType(ctx)
    switch subType {
    case "basic":
        // Grant basic features
    case "premium":
        // Grant premium features
    case "annual":
        // Grant annual benefits
    }
}
```

### Concurrent Renewal Processing

The system uses controlled concurrency to process renewals efficiently:

```go
// ProcessExpiredSubscriptions processes up to 1000 expired subscriptions
// with a concurrency limit of 10 simultaneous renewals

sem := make(chan struct{}, 10)  // Semaphore for 10 concurrent workers
errChan := make(chan error, len(ids))
var wg sync.WaitGroup

for i := range subscriptions {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        sem <- struct{}{}        // Acquire semaphore
        defer func() { <-sem }() // Release semaphore
        
        // Process renewal for subscription[idx]
    }(i)
}

wg.Wait()
```

### Error Handling

```go
// Renewal processing collects all errors
var errors []error
for err := range errChan {
    errors = append(errors, err)
}

// Returns first error if any occurred
if len(errors) > 0 {
    return errors[0]
}
```

**Error Scenarios**:
- **Insufficient funds**: Renewal fails gracefully, subscription stays expired for retry
- **Database errors**: Propagated to caller, transaction rolled back
- **Network errors**: Retry on next pulse
- **Handler errors**: Logged and subscription remains unchanged

## Integration Patterns

### Pattern 1: Subscription-Gated Content Access

```go
func getProtectedContent(ctx context.Context, user UserAccount[int64], contentProductItem ProductItem[int64]) ([]byte, error) {
    // Get user's subscriptions
    subs, err := manager.GetUserSubscriptions(ctx, user, nil, 0, 100, scommerce.QueueOrderDesc)
    if err != nil {
        return nil, err
    }

    // Check if user has active subscription to this content
    productItemID, _ := contentProductItem.GetID(ctx)
    
    for _, sub := range subs {
        subProductItem, _ := sub.GetProductItem(ctx)
        subProductID, _ := subProductItem.GetID(ctx)
        
        if subProductID == productItemID {
            isActive, _ := sub.IsActive(ctx)
            isExpired, _ := sub.IsExpired(ctx)
            
            if isActive && !isExpired {
                // User has valid subscription - grant access
                return loadProtectedContent(contentProductItem), nil
            }
        }
    }

    return nil, errors.New("subscription required")
}
```

### Pattern 2: Automatic Subscription on Purchase

```go
func purchaseSubscriptionProduct(ctx context.Context, user UserAccount[int64], product ProductItem[int64], duration time.Duration) error {
    // Charge user for initial purchase
    price, _ := product.GetPrice(ctx)
    if err := user.ChargeWallet(ctx, -price); err != nil {
        return err
    }

    // Create subscription
    subscription, err := manager.NewSubscription(
        ctx,
        user,
        product,
        duration,
        "purchased",
        true, // Auto-renew
    )
    
    return err
}
```

### Pattern 3: Upgrade/Downgrade Subscriptions

```go
func upgradeSubscription(ctx context.Context, currentSub ProductItemSubscription[int64], newProductItem ProductItem[int64]) error {
    // Calculate prorated refund/charge
    remainingTime, _ := calculateRemainingTime(currentSub)
    currentPrice, _ := getCurrentProductPrice(currentSub)
    newPrice, _ := newProductItem.GetPrice(ctx)
    
    proratedRefund := (currentPrice / 30) * float64(remainingTime.Hours()/24)
    proratedCharge := (newPrice / 30) * float64(remainingTime.Hours()/24)
    difference := proratedCharge - proratedRefund

    // Charge/refund difference
    if difference > 0 {
        user, _ := getUserForSubscription(currentSub)
        if err := user.ChargeWallet(ctx, -difference); err != nil {
            return err
        }
    } else if difference < 0 {
        user, _ := getUserForSubscription(currentSub)
        user.ChargeWallet(ctx, -difference) // Positive refund
    }

    // Update subscription
    currentSub.SetProductItem(ctx, newProductItem)
    
    return nil
}
```

## Best Practices

### 1. Schedule Regular Renewal Processing

Run `ProcessExpiredSubscriptions` regularly (recommended: hourly):

```go
// Using a cron job or scheduled task
func subscriptionRenewalJob(ctx context.Context, manager *BuiltinProductItemSubscriptionManager[int64]) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := manager.ProcessExpiredSubscriptions(ctx); err != nil {
                log.Printf("Renewal processing error: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### 2. Implement Retry Logic

```go
func processWithRetry(ctx context.Context, manager *BuiltinProductItemSubscriptionManager[int64], maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = manager.ProcessExpiredSubscriptions(ctx)
        if err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
    }
    return err
}
```

### 3. Monitor Subscription Health

```go
func getSubscriptionMetrics(ctx context.Context, manager *BuiltinProductItemSubscriptionManager[int64]) {
    totalSubs, _ := manager.GetSubscriptionCount(ctx)
    
    // Custom queries for metrics
    activeSubs := countActiveSubscriptions(ctx, db)
    expiredSubs := countExpiredSubscriptions(ctx, db)
    renewalRate := calculateRenewalRate(ctx, db)
    
    log.Printf("Total: %d, Active: %d, Expired: %d, Renewal Rate: %.2f%%", 
        totalSubs, activeSubs, expiredSubs, renewalRate)
}
```

### 4. Notify Users Before Expiration

```go
func sendExpirationNotifications(ctx context.Context) {
    // Find subscriptions expiring in 3 days
    soonToExpire := findSubscriptionsExpiringWithin(ctx, 3*24*time.Hour)
    
    for _, sub := range soonToExpire {
        user, _ := getUserForSubscription(sub)
        sendEmail(user, "Your subscription expires in 3 days!")
    }
}
```

### 5. Handle Failed Payments Gracefully

```go
func customRenewalWithNotification(ctx context.Context, subscription ProductItemSubscription[int64], account UserAccount[int64], productItem ProductItem[int64]) (bool, float64, error) {
    price, _ := productItem.GetPrice(ctx)
    balance, _ := account.GetWalletCurrency(ctx)
    
    if balance < price {
        // Send notification about insufficient funds
        sendPaymentFailureEmail(account, price, balance)
        return false, 0, nil
    }

    // Process renewal normally
    expiresAt, _ := subscription.GetExpiresAt(ctx)
    duration, _ := subscription.GetDuration(ctx)
    subscription.SetExpiresAt(ctx, expiresAt.Add(duration))
    
    return true, price, nil
}
```

## Performance Considerations

### Database Indexing

Ensure proper indexes for optimal query performance:

```sql
-- Essential indexes
CREATE INDEX idx_subs_expires_renew ON product_item_subscriptions(expires_at, auto_renew) 
    WHERE auto_renew = true;

CREATE INDEX idx_subs_user_active ON product_item_subscriptions(user_account_id, is_active);

CREATE INDEX idx_subs_product_active ON product_item_subscriptions(product_item_id, is_active);
```

### Batch Processing

Limit batch size to prevent overwhelming the system:

```go
// ProcessExpiredSubscriptions processes up to 1000 subscriptions at once
// For larger systems, run multiple batches

func processAllExpiredInBatches(ctx context.Context, manager *BuiltinProductItemSubscriptionManager[int64]) error {
    for {
        err := manager.ProcessExpiredSubscriptions(ctx)
        if err != nil {
            return err
        }
        
        // Check if more subscriptions need processing
        hasMore := checkForMoreExpiredSubscriptions(ctx)
        if !hasMore {
            break
        }
        
        time.Sleep(1 * time.Second) // Rate limiting
    }
    return nil
}
```

### Caching Active Subscriptions

Cache frequently accessed subscription data:

```go
type SubscriptionCache struct {
    cache map[int64][]ProductItemSubscription[int64]
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *SubscriptionCache) GetUserSubscriptions(ctx context.Context, userID int64) ([]ProductItemSubscription[int64], bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    subs, ok := c.cache[userID]
    return subs, ok
}

func (c *SubscriptionCache) SetUserSubscriptions(userID int64, subs []ProductItemSubscription[int64]) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache[userID] = subs
    
    // Schedule cache invalidation
    time.AfterFunc(c.ttl, func() {
        c.mu.Lock()
        delete(c.cache, userID)
        c.mu.Unlock()
    })
}
```

## Troubleshooting

### Issue: Renewals Not Processing

**Symptoms**: Expired subscriptions remain expired despite auto-renew being enabled

**Solutions**:
1. Check if `Pulse()` or `ProcessExpiredSubscriptions()` is being called regularly
2. Verify renewal handler is set correctly
3. Check database indexes on `expires_at` column
4. Review error logs for renewal handler failures

### Issue: Duplicate Charges

**Symptoms**: Users charged multiple times for same renewal

**Solutions**:
1. Ensure `ProcessExpiredSubscriptions()` isn't called concurrently
2. Use database transactions for renewal operations
3. Add unique constraints on subscription renewals
4. Implement idempotency keys for payment processing

### Issue: Memory Leaks in Long-Running Processes

**Symptoms**: Memory usage grows over time

**Solutions**:
1. Limit batch size in `ProcessExpiredSubscriptions()`
2. Close database connections properly
3. Clear subscription form caches periodically
4. Use context timeouts for goroutines

### Issue: Slow Renewal Processing

**Symptoms**: Renewals take too long to process

**Solutions**:
1. Add database indexes on frequently queried columns
2. Increase concurrency limit (currently 10 concurrent renewals)
3. Optimize renewal handler logic
4. Use connection pooling for database

## API Reference

### ProductItemSubscriptionManager Methods

| Method | Description |
|--------|-------------|
| `NewSubscription(ctx, account, productItem, duration, type, autoRenew)` | Create new subscription |
| `RemoveSubscription(ctx, subscription)` | Delete subscription |
| `RemoveAllSubscriptions(ctx)` | Delete all subscriptions |
| `GetSubscriptions(ctx, subs, skip, limit, order)` | Get all subscriptions |
| `GetSubscriptionCount(ctx)` | Count total subscriptions |
| `GetUserSubscriptions(ctx, account, subs, skip, limit, order)` | Get user's subscriptions |
| `GetUserSubscriptionCount(ctx, account)` | Count user's subscriptions |
| `GetProductItemSubscriptions(ctx, item, subs, skip, limit, order)` | Get product's subscriptions |
| `GetProductItemSubscriptionCount(ctx, item)` | Count product's subscriptions |
| `SetRenewalHandler(ctx, handler)` | Set custom renewal function |
| `ProcessExpiredSubscriptions(ctx)` | Process all expired subscriptions |
| `Init(ctx)` | Initialize database schema |
| `Pulse(ctx)` | Periodic task (calls ProcessExpiredSubscriptions) |
| `Close(ctx)` | Cleanup resources |

### ProductItemSubscription Methods

| Method | Description |
|--------|-------------|
| `GetID(ctx)` | Get subscription ID |
| `GetUserAccountID(ctx)` | Get subscriber's account ID |
| `GetProductItem(ctx)` | Get subscribed product |
| `SetProductItem(ctx, item)` | Change subscribed product |
| `GetSubscribedAt(ctx)` | Get subscription start time |
| `SetSubscribedAt(ctx, time)` | Set subscription start time |
| `GetExpiresAt(ctx)` | Get expiration time |
| `SetExpiresAt(ctx, time)` | Set expiration time |
| `GetDuration(ctx)` | Get subscription duration |
| `SetDuration(ctx, duration)` | Set subscription duration |
| `GetSubscriptionType(ctx)` | Get subscription type |
| `SetSubscriptionType(ctx, type)` | Set subscription type |
| `IsAutoRenew(ctx)` | Check if auto-renew enabled |
| `SetAutoRenew(ctx, enabled)` | Enable/disable auto-renew |
| `IsActive(ctx)` | Check if subscription is active |
| `SetActive(ctx, active)` | Activate/deactivate subscription |
| `IsExpired(ctx)` | Check if subscription has expired |
| `Cancel(ctx)` | Cancel subscription (sets AutoRenew=false, IsActive=false) |

## Migration Guide

### Adding Subscriptions to Existing Products

```go
// 1. Initialize subscription manager
subManager := NewBuiltinProductItemSubscriptionManager(db, fs, nil)
if err := subManager.Init(ctx); err != nil {
    log.Fatal(err)
}

// 2. Convert existing purchases to subscriptions (one-time migration)
func migrateExistingPurchases(ctx context.Context) error {
    orders, _ := orderManager.GetOrders(ctx, nil, 0, 1000, QueueOrderAsc)
    
    for _, order := range orders {
        items, _ := order.GetProductItems(ctx, nil, 0, 100, QueueOrderAsc)
        
        for _, item := range items {
            // Create subscription for each purchased item
            userID, _ := order.GetUserAccountID(ctx)
            account, _ := accountManager.GetAccountByID(ctx, userID)
            
            _, err := subManager.NewSubscription(
                ctx,
                account,
                item.ProductItem,
                30*24*time.Hour, // Default 30 days
                "migrated",
                true,
            )
            if err != nil {
                log.Printf("Migration error for order %d: %v", order.GetID(ctx), err)
            }
        }
    }
    
    return nil
}
```

## Summary

The Product Item Subscription system provides a complete solution for recurring billing in e-commerce:

✅ **Flexible subscription management** with customizable durations and types  
✅ **Automatic renewal processing** with configurable retry logic  
✅ **Custom billing logic** via renewal handler functions  
✅ **Concurrent processing** for scalability  
✅ **Wallet-based payments** with conditional charging  
✅ **Comprehensive API** for all subscription operations  
✅ **Database-agnostic design** following the library's architecture  

Start building subscription-based products today with minimal code!

## See Also

- [Item Attributes Guide](item-attributes.md) - Store custom metadata with subscriptions
- [Order Management](getting-started.md#orders) - Handle subscription-related orders
- [User Accounts](getting-started.md#user-accounts) - Wallet management for subscriptions
- [Product Management](getting-started.md#products) - Configure subscription products
