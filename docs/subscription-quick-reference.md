# Product Subscriptions - Quick Reference

## 🚀 Quick Start

```go
// 1. Get subscription manager
subManager := app.GetProductItemSubscriptionManager()

// 2. Initialize (creates database tables)
subManager.Init(ctx)

// 3. Create a subscription
subscription, err := subManager.NewSubscription(
    ctx,
    userAccount,        // User to subscribe
    productItem,        // Product to subscribe to
    30*24*time.Hour,    // Duration (30 days)
    "monthly",          // Type
    true,               // Auto-renew enabled
)

// 4. Process renewals periodically
subManager.Pulse(ctx) // Call hourly via cron/scheduler
```

## 📋 Common Operations

### Create Subscription
```go
sub, err := manager.NewSubscription(ctx, user, product, duration, "type", autoRenew)
```

### Check Status
```go
isActive, _ := subscription.IsActive(ctx)
isExpired, _ := subscription.IsExpired(ctx)
expiresAt, _ := subscription.GetExpiresAt(ctx)
```

### Get User's Subscriptions
```go
subs, err := manager.GetUserSubscriptions(ctx, user, nil, 0, 10, QueueOrderDesc)
```

### Cancel Subscription
```go
err := subscription.Cancel(ctx)
// Sets AutoRenew=false and IsActive=false
```

### Extend Subscription
```go
currentExpiry, _ := subscription.GetExpiresAt(ctx)
subscription.SetExpiresAt(ctx, currentExpiry.Add(7*24*time.Hour))
```

### Toggle Auto-Renew
```go
subscription.SetAutoRenew(ctx, false) // Disable
subscription.SetAutoRenew(ctx, true)  // Enable
```

## 🔄 Renewal Handler Examples

### Default (Built-in)
```go
// Checks wallet → Extends subscription → Returns price to charge
// Auto-used if no custom handler set
```

### Custom: Loyalty Discount
```go
func loyaltyHandler(ctx, sub, user, product) (bool, float64, error) {
    price, _ := product.GetPrice(ctx)
    
    // Check subscription age
    subscribedAt, _ := sub.GetSubscribedAt(ctx)
    age := time.Since(subscribedAt)
    
    // Apply discount
    discount := 0.0
    if age > 365*24*time.Hour {
        discount = 0.20 // 20% off for 1+ year
    }
    
    finalPrice := price * (1.0 - discount)
    
    // Verify funds
    balance, _ := user.GetWalletCurrency(ctx)
    if balance < finalPrice {
        return false, 0, nil
    }
    
    // Extend subscription
    expires, _ := sub.GetExpiresAt(ctx)
    duration, _ := sub.GetDuration(ctx)
    sub.SetExpiresAt(ctx, expires.Add(duration))
    
    return true, finalPrice, nil
}

// Set custom handler
manager.SetRenewalHandler(ctx, loyaltyHandler)
```

### Custom: Free Trial → Paid
```go
func trialHandler(ctx, sub, user, product) (bool, float64, error) {
    subType, _ := sub.GetSubscriptionType(ctx)
    
    if subType == "trial" {
        // First renewal: trial → paid
        sub.SetSubscriptionType(ctx, "paid")
        price, _ := product.GetPrice(ctx)
        balance, _ := user.GetWalletCurrency(ctx)
        
        if balance < price {
            return false, 0, nil
        }
        
        expires, _ := sub.GetExpiresAt(ctx)
        duration, _ := sub.GetDuration(ctx)
        sub.SetExpiresAt(ctx, expires.Add(duration))
        
        return true, price, nil
    }
    
    // Regular renewal - use default logic
    price, _ := product.GetPrice(ctx)
    balance, _ := user.GetWalletCurrency(ctx)
    
    if balance < price {
        return false, 0, nil
    }
    
    expires, _ := sub.GetExpiresAt(ctx)
    duration, _ := sub.GetDuration(ctx)
    sub.SetExpiresAt(ctx, expires.Add(duration))
    
    return true, price, nil
}
```

## 🔍 Query Patterns

### All Subscriptions
```go
all, err := manager.GetSubscriptions(ctx, nil, 0, 100, QueueOrderDesc)
```

### User's Active Subscriptions
```go
userSubs, _ := manager.GetUserSubscriptions(ctx, user, nil, 0, 100, QueueOrderDesc)

for _, sub := range userSubs {
    if active, _ := sub.IsActive(ctx); active {
        if expired, _ := sub.IsExpired(ctx); !expired {
            // This is an active, non-expired subscription
        }
    }
}
```

### Product's Subscribers
```go
productSubs, err := manager.GetProductItemSubscriptions(ctx, product, nil, 0, 100, QueueOrderAsc)
count, _ := manager.GetProductItemSubscriptionCount(ctx, product)
```

## ⚙️ Integration Patterns

### Subscription-Gated Content
```go
func hasAccess(ctx context.Context, user UserAccount, content ProductItem) bool {
    subs, _ := manager.GetUserSubscriptions(ctx, user, nil, 0, 100, QueueOrderDesc)
    contentID, _ := content.GetID(ctx)
    
    for _, sub := range subs {
        item, _ := sub.GetProductItem(ctx)
        itemID, _ := item.GetID(ctx)
        
        if itemID == contentID {
            active, _ := sub.IsActive(ctx)
            expired, _ := sub.IsExpired(ctx)
            return active && !expired
        }
    }
    
    return false
}
```

### Auto-Subscribe on Purchase
```go
func purchaseWithSubscription(ctx, user, product, duration) error {
    // Charge for purchase
    price, _ := product.GetPrice(ctx)
    user.ChargeWallet(ctx, -price)
    
    // Create subscription
    _, err := manager.NewSubscription(ctx, user, product, duration, "purchased", true)
    return err
}
```

### Schedule Periodic Renewals
```go
func startRenewalScheduler(ctx context.Context, manager *BuiltinProductItemSubscriptionManager) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := manager.Pulse(ctx); err != nil {
                log.Printf("Renewal error: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

// Start in background
go startRenewalScheduler(context.Background(), subManager)
```

## 🎯 Subscription Types

### Common Types
- `"trial"` - Free trial period
- `"basic"` - Basic tier
- `"premium"` - Premium tier
- `"annual"` - Annual subscription
- `"monthly"` - Monthly subscription
- `"lifetime"` - One-time purchase with lifetime access
- `"promotional"` - Special offer subscription

### Usage
```go
// Create different tiers
basicSub, _ := manager.NewSubscription(ctx, user, basicProduct, 30*24*time.Hour, "basic", true)
premiumSub, _ := manager.NewSubscription(ctx, user, premiumProduct, 30*24*time.Hour, "premium", true)
annualSub, _ := manager.NewSubscription(ctx, user, product, 365*24*time.Hour, "annual", true)
```

## ⏱️ Duration Examples

```go
// Common durations
daily := 24 * time.Hour
weekly := 7 * 24 * time.Hour
monthly := 30 * 24 * time.Hour
quarterly := 90 * 24 * time.Hour
annual := 365 * 24 * time.Hour

// Custom duration
custom := 14 * 24 * time.Hour // 2 weeks
```

## 🛠️ Troubleshooting

### Renewals Not Processing
```go
// Check if Pulse() is being called
subManager.Pulse(ctx)

// Or manually process
subManager.ProcessExpiredSubscriptions(ctx)
```

### Check Expiration Status
```go
expiresAt, _ := subscription.GetExpiresAt(ctx)
now := time.Now()
fmt.Printf("Expires in: %v\n", expiresAt.Sub(now))

isExpired, _ := subscription.IsExpired(ctx)
fmt.Printf("Is expired: %v\n", isExpired)
```

### Verify Auto-Renew
```go
autoRenew, _ := subscription.IsAutoRenew(ctx)
fmt.Printf("Auto-renew: %v\n", autoRenew)

// Enable if disabled
if !autoRenew {
    subscription.SetAutoRenew(ctx, true)
}
```

## 📊 Renewal Return Values

| Success | Amount | Meaning |
|---------|--------|---------|
| `true` | `> 0` | Renewal succeeded, charge user wallet |
| `true` | `= 0` | Renewal succeeded, no charge (free) |
| `false` | `0` | Renewal failed (insufficient funds, etc.) |
| `error` | - | System error occurred |

## 💡 Best Practices

1. **Schedule Renewals**: Call `Pulse()` hourly via cron
2. **Monitor Failed Renewals**: Log when renewals fail
3. **Notify Users**: Send email before expiration
4. **Index Database**: Ensure `expires_at` is indexed
5. **Batch Processing**: Process in batches of 1000
6. **Custom Handlers**: Implement business-specific logic
7. **Error Handling**: Handle failed renewals gracefully
8. **Test Edge Cases**: Test insufficient funds, expired cards, etc.

## 🔗 Related Documentation

- [Full Guide](product-subscriptions.md) - Complete documentation
- [Item Attributes](item-attributes.md) - Store metadata with subscriptions
- [User Accounts](getting-started.md#accounts) - Wallet management
- [Products](getting-started.md#products) - Subscription products

## 📝 API Summary

| Method | Purpose |
|--------|---------|
| `NewSubscription` | Create subscription |
| `GetUserSubscriptions` | Get user's subscriptions |
| `GetProductItemSubscriptions` | Get product's subscribers |
| `SetRenewalHandler` | Set custom renewal logic |
| `ProcessExpiredSubscriptions` | Process renewals manually |
| `Pulse` | Periodic task (calls ProcessExpiredSubscriptions) |
| `Cancel` | Cancel subscription |
| `IsActive` | Check if active |
| `IsExpired` | Check if expired |
| `SetAutoRenew` | Toggle auto-renewal |

---

**Need more details?** See the [complete documentation](product-subscriptions.md)!
