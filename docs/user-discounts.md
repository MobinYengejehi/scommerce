# User Discount System

## Overview

The **User Discount System** provides a complete promotional code and discount management solution for e-commerce platforms. This system allows merchants to create, distribute, and track discount codes that customers can apply to their purchases. Each discount code is unique, trackable, and can be configured with usage limits and monetary values.

## What is a User Discount?

A User Discount is a promotional code entity that contains:
- **Unique Code**: Auto-generated alphanumeric code for redemption
- **Value**: Discount amount (flat amount or percentage)
- **Valid Count**: Maximum number of times the code can be used
- **Owner**: The account that created/owns the discount
- **Usage Tracking**: List of users who have already used the code

### Use Cases

- 💳 **Promotional Campaigns**: Create limited-time discount codes for marketing
- 🎁 **Referral Programs**: Generate unique codes for user referrals
- 🎉 **Special Events**: Holiday sales, flash sales, seasonal promotions
- 👥 **Influencer Marketing**: Unique codes for influencers to share
- 🆕 **New Customer Incentives**: Welcome discounts for first-time buyers
- 🔄 **Loyalty Rewards**: Codes for repeat customers
- 📧 **Email Marketing**: Personalized discount codes in campaigns
- 🎯 **Targeted Promotions**: Codes for specific user segments

## Architecture

### Core Components

#### 1. UserDiscountManager
Central manager for all discount operations at the system level.

**Key Responsibilities**:
- Create new discounts with auto-generated codes
- Retrieve discounts by code
- Manage all discounts globally
- Check code existence
- Generate unique discount codes

#### 2. UserDiscount
Individual discount entity representing a single promotional code.

**Key Properties**:
- `ID`: Unique discount identifier
- `UserAccountID`: Owner/creator of the discount
- `Code`: Unique alphanumeric redemption code
- `Value`: Discount amount
- `ValidCount`: Remaining usage count
- `UsedBy`: Array of accounts that used this code

#### 3. UserAccount Integration
Discounts are fully integrated into user accounts.

**Account Methods**:
- `NewDiscount()`: Create discount for this account
- `GetDiscounts()`: Retrieve account's discounts
- `GetDiscountCount()`: Count discounts owned
- `RemoveDiscount()`: Delete a specific discount
- `RemoveAllDiscounts()`: Delete all account discounts

## Database Schema

### Discounts Table

```sql
CREATE TABLE discounts (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    code            VARCHAR(255) NOT NULL UNIQUE,
    value           DOUBLE PRECISION NOT NULL,
    valid_count     BIGINT NOT NULL DEFAULT 0,
    used_by         BIGINT[] DEFAULT '{}',
    
    -- Indexes for performance
    INDEX idx_discounts_user_id (user_id),
    INDEX idx_discounts_code (code),
    INDEX idx_discounts_valid_count (valid_count)
);
```

### Field Descriptions

- **id**: Auto-incrementing unique identifier
- **user_id**: Foreign key to the user who owns/created this discount
- **code**: Unique alphanumeric code (e.g., "SAVE20XYZ")
- **value**: Discount amount (could be flat amount or percentage based on implementation)
- **valid_count**: Number of remaining uses (-1 for unlimited, 0 for exhausted)
- **used_by**: Array of user IDs who have already redeemed this code

## Usage Examples

### Basic Setup

```go
import (
    "context"
    "github.com/MobinYengejehi/scommerce/scommerce"
)

// Get discount manager from app
discountManager := app.GetUserDiscountManager()

// Initialize (creates database tables)
if err := discountManager.Init(ctx); err != nil {
    return err
}
```

### Creating Discounts

#### Via Manager (System-Level)

```go
// Create a discount owned by a specific user
discount, err := discountManager.NewUserDiscount(
    ctx,
    ownerAccount,   // User account that owns this discount
    25.00,          // Discount value ($25 or 25%)
    100,            // Valid for 100 uses
)
if err != nil {
    return err
}

// Get the generated code
code, err := discount.GetCode(ctx)
fmt.Printf("Discount Code: %s\n", code) // e.g., "SAVE25ABC"
```

#### Via User Account

```go
// User creates their own discount
discount, err := userAccount.NewDiscount(
    ctx,
    50.00,  // $50 discount
    10,     // Limited to 10 uses
)
if err != nil {
    return err
}

code, _ := discount.GetCode(ctx)
fmt.Printf("Your discount code: %s\n", code)
```

### Retrieving Discounts

#### By Code (Public Lookup)

```go
// Customer enters discount code at checkout
code := "SAVE25ABC"

// Check if code exists
exists, err := discountManager.ExistsUserDiscountCode(ctx, code)
if !exists {
    fmt.Println("Invalid discount code")
    return
}

// Retrieve the discount
discount, err := discountManager.GetUserDiscountByCode(ctx, code)
if err != nil {
    return err
}

// Get discount details
value, _ := discount.GetValue(ctx)
validCount, _ := discount.GetValidCount(ctx)

fmt.Printf("Discount: $%.2f\n", value)
fmt.Printf("Uses remaining: %d\n", validCount)
```

#### User's Own Discounts

```go
// Get all discounts created by a user
discounts, err := userAccount.GetDiscounts(
    ctx,
    nil,    // Output slice (nil = create new)
    0,      // Skip
    20,     // Limit
    scommerce.QueueOrderDesc,  // Newest first
)
if err != nil {
    return err
}

// Display user's discount codes
for _, discount := range discounts {
    code, _ := discount.GetCode(ctx)
    value, _ := discount.GetValue(ctx)
    validCount, _ := discount.GetValidCount(ctx)
    
    fmt.Printf("Code: %s - $%.2f - %d uses left\n", 
        code, value, validCount)
}
```

#### Count User's Discounts

```go
count, err := userAccount.GetDiscountCount(ctx)
fmt.Printf("You have created %d discount codes\n", count)
```

### Applying Discounts at Checkout

```go
func applyDiscountToOrder(ctx context.Context, discountCode string, orderTotal float64, customer UserAccount[int64]) (float64, error) {
    // Retrieve discount by code
    discount, err := discountManager.GetUserDiscountByCode(ctx, discountCode)
    if err != nil {
        return orderTotal, fmt.Errorf("invalid discount code")
    }
    
    // Check if already used by this customer
    hasUsed, err := discount.HasUserUsed(ctx, customer)
    if err != nil {
        return orderTotal, err
    }
    if hasUsed {
        return orderTotal, fmt.Errorf("you have already used this code")
    }
    
    // Check if code still has valid uses
    validCount, err := discount.GetValidCount(ctx)
    if err != nil {
        return orderTotal, err
    }
    if validCount <= 0 {
        return orderTotal, fmt.Errorf("discount code has been exhausted")
    }
    
    // Get discount value
    discountValue, err := discount.GetValue(ctx)
    if err != nil {
        return orderTotal, err
    }
    
    // Apply discount (ensure not negative)
    newTotal := orderTotal - discountValue
    if newTotal < 0 {
        newTotal = 0
    }
    
    // Mark as used by customer
    if err := discount.MarkAsUsedBy(ctx, customer); err != nil {
        return orderTotal, err
    }
    
    // Decrement valid count
    if err := discount.DecrementValidCount(ctx); err != nil {
        return orderTotal, err
    }
    
    return newTotal, nil
}
```

### Managing Discount Properties

#### Update Discount Value

```go
// Change discount amount
newValue := 30.00
if err := discount.SetValue(ctx, newValue); err != nil {
    return err
}
```

#### Update Valid Count

```go
// Add more uses to the discount
newCount := 50
if err := discount.SetValidCount(ctx, newCount); err != nil {
    return err
}

// Or decrement when used
if err := discount.DecrementValidCount(ctx); err != nil {
    return err
}
```

#### Change Discount Code

```go
// Update the code (must remain unique)
newCode := "SUMMER2024"
if err := discount.SetCode(ctx, newCode); err != nil {
    return err
}
```

### Tracking Usage

```go
// Get list of users who used this discount
usedByIDs, err := discount.GetUsedBy(ctx)
fmt.Printf("%d users have used this code\n", len(usedByIDs))

// Check if specific user used it
hasUsed, err := discount.HasUserUsed(ctx, customerAccount)
if hasUsed {
    fmt.Println("This customer already used this code")
}
```

### Deleting Discounts

```go
// Remove specific discount
if err := userAccount.RemoveDiscount(ctx, discount); err != nil {
    return err
}

// Or via manager
if err := discountManager.RemoveUserDiscount(ctx, discount); err != nil {
    return err
}

// Remove all discounts for a user
if err := userAccount.RemoveAllDiscounts(ctx); err != nil {
    return err
}

// Remove all discounts in system (admin only)
if err := discountManager.RemoveAllUserDiscounts(ctx); err != nil {
    return err
}
```

## Integration Patterns

### Pattern 1: Referral Program

```go
func createReferralDiscount(ctx context.Context, referrer UserAccount[int64]) (string, error) {
    // Create discount for referrer to share
    discount, err := referrer.NewDiscount(
        ctx,
        10.00,  // $10 off
        50,     // Can be used 50 times
    )
    if err != nil {
        return "", err
    }
    
    // Get the unique code
    code, err := discount.GetCode(ctx)
    if err != nil {
        return "", err
    }
    
    // Share this code with friends
    return code, nil
}

func trackReferralRedemptions(ctx context.Context, code string) (int, error) {
    discount, err := discountManager.GetUserDiscountByCode(ctx, code)
    if err != nil {
        return 0, err
    }
    
    usedBy, err := discount.GetUsedBy(ctx)
    if err != nil {
        return 0, err
    }
    
    // Count successful referrals
    return len(usedBy), nil
}
```

### Pattern 2: Influencer Campaign

```go
func createInfluencerCodes(ctx context.Context, influencers []UserAccount[int64]) (map[string]string, error) {
    codes := make(map[string]string)
    
    for _, influencer := range influencers {
        // Create unique discount for each influencer
        discount, err := influencer.NewDiscount(
            ctx,
            15.00,  // 15% or $15 off
            1000,   // High usage limit for influencer audience
        )
        if err != nil {
            return nil, err
        }
        
        code, _ := discount.GetCode(ctx)
        influencerName, _ := influencer.GetFirstName(ctx)
        
        codes[influencerName] = code
    }
    
    return codes, nil
}
```

### Pattern 3: Limited-Time Flash Sale

```go
func createFlashSaleDiscount(ctx context.Context, adminAccount UserAccount[int64]) (UserDiscount[int64], error) {
    // Create high-value, limited-use discount
    discount, err := adminAccount.NewDiscount(
        ctx,
        100.00,  // $100 off!
        50,      // Only 50 people can use it
    )
    if err != nil {
        return nil, err
    }
    
    // Announce the code publicly
    code, _ := discount.GetCode(ctx)
    fmt.Printf("🔥 FLASH SALE: Use code %s for $100 off!\n", code)
    
    // Monitor usage in real-time
    go monitorFlashSale(ctx, discount)
    
    return discount, nil
}

func monitorFlashSale(ctx context.Context, discount UserDiscount[int64]) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        validCount, _ := discount.GetValidCount(ctx)
        usedBy, _ := discount.GetUsedBy(ctx)
        
        fmt.Printf("Flash sale: %d uses remaining, %d claimed\n", 
            validCount, len(usedBy))
        
        if validCount <= 0 {
            fmt.Println("⚡ Flash sale sold out!")
            break
        }
    }
}
```

### Pattern 4: Email Campaign Codes

```go
func generatePersonalizedCodes(ctx context.Context, customers []UserAccount[int64], campaign string) error {
    for _, customer := range customers {
        // Create unique discount for each customer
        discount, err := customer.NewDiscount(
            ctx,
            20.00,  // $20 off
            1,      // Single use per customer
        )
        if err != nil {
            continue
        }
        
        code, _ := discount.GetCode(ctx)
        email, _ := customer.GetEmail(ctx)
        
        // Send personalized email with unique code
        sendCampaignEmail(email, code, campaign)
    }
    
    return nil
}
```

### Pattern 5: Loyalty Tiers

```go
func createLoyaltyDiscount(ctx context.Context, customer UserAccount[int64], tier string) (UserDiscount[int64], error) {
    var value float64
    var validCount int64
    
    switch tier {
    case "bronze":
        value = 5.00
        validCount = 3
    case "silver":
        value = 10.00
        validCount = 5
    case "gold":
        value = 25.00
        validCount = 10
    case "platinum":
        value = 50.00
        validCount = 20
    default:
        return nil, fmt.Errorf("invalid tier")
    }
    
    discount, err := customer.NewDiscount(ctx, value, validCount)
    if err != nil {
        return nil, err
    }
    
    return discount, nil
}
```

## Advanced Features

### Automatic Code Generation

The system automatically generates unique alphanumeric codes:

```go
// Code generation happens automatically
discount, _ := userAccount.NewDiscount(ctx, 15.00, 100)
code, _ := discount.GetCode(ctx)
// Returns something like: "X7K9M2PQ4R"
```

**Code Properties**:
- Configurable length (set in manager initialization)
- Alphanumeric characters for easy typing
- Uniqueness guaranteed (up to 1000 retry attempts)
- Thread-safe generation with mutex locking

### Usage Tracking with User Lists

```go
// Get all users who redeemed a code
usedByIDs, err := discount.GetUsedBy(ctx)

// Convert to user accounts for analysis
var customers []UserAccount[int64]
for _, uid := range usedByIDs {
    customer, err := accountManager.GetUserAccountByID(ctx, uid)
    if err == nil {
        customers = append(customers, customer)
    }
}

// Analyze customer data
fmt.Printf("Code used by %d customers\n", len(customers))
```

### Preventing Abuse

```go
func validateDiscountRedemption(ctx context.Context, code string, customer UserAccount[int64]) error {
    discount, err := discountManager.GetUserDiscountByCode(ctx, code)
    if err != nil {
        return fmt.Errorf("invalid code")
    }
    
    // Check 1: Already used by this customer?
    hasUsed, _ := discount.HasUserUsed(ctx, customer)
    if hasUsed {
        return fmt.Errorf("you already used this code")
    }
    
    // Check 2: Code exhausted?
    validCount, _ := discount.GetValidCount(ctx)
    if validCount <= 0 {
        return fmt.Errorf("code no longer valid")
    }
    
    // Check 3: Customer account status
    isActive, _ := customer.IsActive(ctx)
    if !isActive {
        return fmt.Errorf("account not active")
    }
    
    return nil
}
```

### Bulk Operations

```go
// Get all discounts in system
allDiscounts, err := discountManager.GetUserDiscounts(
    ctx,
    nil,
    0,
    1000,
    scommerce.QueueOrderAsc,
)

// Filter by criteria
var activeDiscounts []UserDiscount[int64]
for _, discount := range allDiscounts {
    validCount, _ := discount.GetValidCount(ctx)
    if validCount > 0 {
        activeDiscounts = append(activeDiscounts, discount)
    }
}

fmt.Printf("Found %d active discount codes\n", len(activeDiscounts))
```

## Best Practices

### 1. Validate Before Redemption

```go
// Always check code validity before applying
func safeApplyDiscount(ctx context.Context, code string, customer UserAccount[int64]) error {
    if code == "" {
        return fmt.Errorf("no code provided")
    }
    
    exists, err := discountManager.ExistsUserDiscountCode(ctx, code)
    if err != nil || !exists {
        return fmt.Errorf("invalid discount code")
    }
    
    discount, err := discountManager.GetUserDiscountByCode(ctx, code)
    if err != nil {
        return err
    }
    
    // Validate all conditions
    if err := validateDiscountRedemption(ctx, code, customer); err != nil {
        return err
    }
    
    // Apply discount
    // ...
    
    return nil
}
```

### 2. Set Appropriate Usage Limits

```go
// Use case-specific limits
func createDiscount(ctx context.Context, user UserAccount[int64], purpose string) (UserDiscount[int64], error) {
    var validCount int64
    
    switch purpose {
    case "personal":
        validCount = 1  // Single use
    case "friends":
        validCount = 10  // Share with friends
    case "influencer":
        validCount = 1000  // Large audience
    case "unlimited":
        validCount = -1  // No limit (if supported)
    default:
        validCount = 5  // Default
    }
    
    return user.NewDiscount(ctx, 10.00, validCount)
}
```

### 3. Track and Analyze Usage

```go
func analyzeDiscountPerformance(ctx context.Context, discount UserDiscount[int64]) {
    code, _ := discount.GetCode(ctx)
    value, _ := discount.GetValue(ctx)
    usedBy, _ := discount.GetUsedBy(ctx)
    validCount, _ := discount.GetValidCount(ctx)
    
    totalUses := len(usedBy)
    totalDiscounted := float64(totalUses) * value
    
    fmt.Printf("Code: %s\n", code)
    fmt.Printf("Times used: %d\n", totalUses)
    fmt.Printf("Remaining uses: %d\n", validCount)
    fmt.Printf("Total discounted: $%.2f\n", totalDiscounted)
}
```

### 4. Clean Up Expired Codes

```go
func cleanupExpiredDiscounts(ctx context.Context) error {
    // Get all discounts
    allDiscounts, err := discountManager.GetUserDiscounts(ctx, nil, 0, 10000, scommerce.QueueOrderAsc)
    if err != nil {
        return err
    }
    
    // Remove exhausted codes
    for _, discount := range allDiscounts {
        validCount, _ := discount.GetValidCount(ctx)
        if validCount <= 0 {
            discountManager.RemoveUserDiscount(ctx, discount)
        }
    }
    
    return nil
}
```

### 5. Implement Discount Stacking Rules

```go
func applyMultipleDiscounts(ctx context.Context, codes []string, orderTotal float64, customer UserAccount[int64]) (float64, error) {
    // Define stacking policy
    const maxDiscounts = 2  // Maximum 2 codes per order
    
    if len(codes) > maxDiscounts {
        return orderTotal, fmt.Errorf("maximum %d discount codes allowed", maxDiscounts)
    }
    
    finalTotal := orderTotal
    appliedCodes := 0
    
    for _, code := range codes {
        discount, err := discountManager.GetUserDiscountByCode(ctx, code)
        if err != nil {
            continue  // Skip invalid codes
        }
        
        // Validate
        if err := validateDiscountRedemption(ctx, code, customer); err != nil {
            continue
        }
        
        // Apply discount
        value, _ := discount.GetValue(ctx)
        finalTotal -= value
        
        // Mark as used
        discount.MarkAsUsedBy(ctx, customer)
        discount.DecrementValidCount(ctx)
        
        appliedCodes++
    }
    
    if finalTotal < 0 {
        finalTotal = 0
    }
    
    fmt.Printf("Applied %d discount codes\n", appliedCodes)
    return finalTotal, nil
}
```

## Performance Considerations

### Database Indexing

The system includes important indexes:

1. **idx_discounts_user_id**: Fast lookup by owner
2. **idx_discounts_code**: Fast code validation (UNIQUE)
3. **idx_discounts_valid_count**: Filter active codes

### Code Generation Performance

```go
// Codes are generated with retry mechanism
// Up to 1000 attempts to find unique code
// Mutex-protected for thread safety
```

**Tips**:
- Use longer codes for lower collision probability
- Monitor code generation failures
- Consider pre-generating code pools for high traffic

### Caching Strategies

```go
type DiscountCache struct {
    codes map[string]*BuiltinUserDiscount[int64]
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *DiscountCache) Get(code string) (*BuiltinUserDiscount[int64], bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    discount, ok := c.codes[code]
    return discount, ok
}

func (c *DiscountCache) Set(code string, discount *BuiltinUserDiscount[int64]) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.codes[code] = discount
    
    // Auto-expire
    time.AfterFunc(c.ttl, func() {
        c.mu.Lock()
        delete(c.codes, code)
        c.mu.Unlock()
    })
}
```

## Troubleshooting

### Issue: "Code Already Used"

**Symptoms**: Customer reports code doesn't work

**Solutions**:
1. Check if customer already redeemed: `discount.HasUserUsed(ctx, customer)`
2. Verify code still has uses: `discount.GetValidCount(ctx)`
3. Check usage list: `discount.GetUsedBy(ctx)`

### Issue: "Invalid Code"

**Symptoms**: Code not found in system

**Solutions**:
1. Verify code exists: `discountManager.ExistsUserDiscountCode(ctx, code)`
2. Check for typos (codes are case-sensitive)
3. Ensure code wasn't deleted

### Issue: Code Generation Fails

**Symptoms**: `ErrExceededMaxRetries` when creating discount

**Solutions**:
1. Increase code length in manager initialization
2. Clean up old/expired codes
3. Check database for unique constraint issues

```go
// Increase code length
manager := NewBuiltinUserDiscountManager(db, 12) // 12 characters instead of 8
```

## API Reference

### UserDiscountManager Methods

| Method | Description |
|--------|-------------|
| `NewUserDiscount(ctx, owner, value, count)` | Create new discount |
| `GetUserDiscountByCode(ctx, code)` | Retrieve by code |
| `ExistsUserDiscountCode(ctx, code)` | Check code existence |
| `GetUserDiscounts(ctx, discounts, skip, limit, order)` | List all discounts |
| `GetUserDiscountCount(ctx)` | Count all discounts |
| `RemoveUserDiscount(ctx, discount)` | Delete discount |
| `RemoveAllUserDiscounts(ctx)` | Delete all discounts |
| `Init(ctx)` | Initialize database schema |

### UserDiscount Methods

| Method | Description |
|--------|-------------|
| `GetID(ctx)` | Get discount ID |
| `GetUserAccountID(ctx)` | Get owner ID |
| `GetCode(ctx)` | Get discount code |
| `SetCode(ctx, code)` | Update code |
| `GetValue(ctx)` | Get discount value |
| `SetValue(ctx, value)` | Update value |
| `GetValidCount(ctx)` | Get remaining uses |
| `SetValidCount(ctx, count)` | Update use count |
| `DecrementValidCount(ctx)` | Decrease by one |
| `GetUsedBy(ctx)` | Get user list |
| `HasUserUsed(ctx, account)` | Check if user used |
| `MarkAsUsedBy(ctx, account)` | Mark as used |

### UserAccount Discount Methods

| Method | Description |
|--------|-------------|
| `NewDiscount(ctx, value, validCount)` | Create discount |
| `GetDiscounts(ctx, discounts, skip, limit, order)` | Get account's discounts |
| `GetDiscountCount(ctx)` | Count account's discounts |
| `RemoveDiscount(ctx, discount)` | Remove specific discount |
| `RemoveAllDiscounts(ctx)` | Remove all discounts |

## Summary

The User Discount System provides comprehensive promotional code management:

✅ **Auto-Generated Codes**: Unique alphanumeric codes automatically created  
✅ **Usage Tracking**: Track who used each code and how many times  
✅ **Flexible Limits**: Configure usage limits per code  
✅ **User Integration**: Fully integrated with user accounts  
✅ **Thread-Safe**: Concurrent access protection with mutexes  
✅ **Abuse Prevention**: Prevent duplicate use by same customer  
✅ **Manager + Entity**: Dual-level API for system and user operations  
✅ **Performance Optimized**: Indexed queries and smart caching  

Perfect for e-commerce platforms needing robust promotional code functionality! 🎯

## See Also

- [User Accounts](getting-started.md#accounts) - User management
- [Orders](getting-started.md#orders) - Order processing with discounts
- [Shopping Cart](getting-started.md#shopping-carts) - Apply codes at checkout
- [User Factors](user-factors.md) - Invoice generation with discount tracking
