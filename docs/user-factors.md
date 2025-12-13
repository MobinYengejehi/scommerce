# User Factors (Invoices/Receipts)

## Overview

The **User Factor** system provides invoice and receipt management functionality for e-commerce transactions. In Persian business terminology, a "factor" (فاکتور) refers to an invoice or receipt that documents a purchase transaction. This system allows you to create, manage, and retrieve detailed purchase records for users, including product information, discounts, taxes, and payment amounts.

## What is a User Factor?

A User Factor is a financial document that records:
- **Products purchased**: Detailed JSON data about items in the transaction
- **Discount applied**: Any discounts given to the customer
- **Tax charged**: Tax amount for the transaction
- **Amount paid**: Total amount actually paid by the customer

### Use Cases

- 📄 **Invoice Generation**: Create formal invoices for purchases
- 🧾 **Receipt Management**: Store digital receipts for customer reference
- 📊 **Transaction History**: Track all financial transactions per user
- 💰 **Accounting Records**: Maintain detailed financial records for reporting
- 🔍 **Purchase Tracking**: Enable customers to view their purchase history
- 📈 **Financial Analytics**: Analyze sales, discounts, and tax data
- 🧮 **Tax Reporting**: Generate tax reports from stored factor data

## Architecture

### Core Components

#### 1. UserFactorManager
The central manager for all factor operations.

**Key Responsibilities**:
- Retrieve user factors
- Count factors per user
- Manage factor lifecycle
- Query factor data

#### 2. UserFactor
Individual factor entity representing a single invoice/receipt.

**Key Properties**:
- `ID`: Unique factor identifier
- `UserAccountID`: The user who made the purchase
- `Products`: JSON data containing product details
- `Discount`: Discount amount applied
- `Tax`: Tax amount charged
- `AmountPaid`: Total amount paid by customer

## Database Schema

### Factors Table

```sql
CREATE TABLE factors (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    products        JSONB NOT NULL,
    discount        DOUBLE PRECISION NOT NULL DEFAULT 0,
    tax             DOUBLE PRECISION NOT NULL DEFAULT 0,
    amount_paid     DOUBLE PRECISION NOT NULL,
    
    -- Indexes for performance
    INDEX idx_factors_user_id (user_id),
    INDEX idx_factors_amount_paid (amount_paid)
);
```

### Products JSON Structure

The `products` field stores a flexible JSONB structure containing product details:

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
    },
    {
        "product_id": 456,
        "product_name": "USB Cable",
        "sku": "CABLE-USB-C",
        "quantity": 1,
        "unit_price": 9.99,
        "subtotal": 9.99
    }
]
```

## Usage Examples

### Basic Factor Creation

**Note**: The current implementation provides read and update operations for factors. Factor creation typically happens through integration with the order system or a custom implementation.

```go
import (
    "context"
    "encoding/json"
    "github.com/MobinYengejehi/scommerce/scommerce"
)

// Get factor manager
factorManager := app.GetUserFactorManager()

// Initialize (creates database tables)
if err := factorManager.Init(ctx); err != nil {
    return err
}
```

### Retrieving User Factors

```go
// Get all factors for a user
userFactors, err := factorManager.GetUserFactors(
    ctx,
    userAccount,    // User account object
    nil,            // Output slice (nil = create new)
    0,              // Skip
    10,             // Limit
    scommerce.QueueOrderDesc, // Newest first
)
if err != nil {
    return err
}

// Iterate through factors
for _, factor := range userFactors {
    factorID, _ := factor.GetID(ctx)
    amountPaid, _ := factor.GetAmountPaid(ctx)
    discount, _ := factor.GetDiscount(ctx)
    tax, _ := factor.GetTax(ctx)
    
    fmt.Printf("Factor #%d: Paid: $%.2f, Discount: $%.2f, Tax: $%.2f\n",
        factorID, amountPaid, discount, tax)
}
```

### Counting User Factors

```go
// Get total count of factors for a user
count, err := factorManager.GetUserFactorCount(ctx, userAccount)
if err != nil {
    return err
}

fmt.Printf("User has %d factors\n", count)
```

### Retrieving Factor Details

```go
// Get factor ID
factorID, err := factor.GetID(ctx)

// Get products JSON
products, err := factor.GetProducts(ctx)
if err != nil {
    return err
}

// Parse products
var productList []map[string]interface{}
if err := json.Unmarshal(products, &productList); err != nil {
    return err
}

// Display products
for _, product := range productList {
    fmt.Printf("Product: %s, Quantity: %v, Price: %v\n",
        product["product_name"],
        product["quantity"],
        product["unit_price"])
}
```

### Updating Factor Information

```go
// Update discount
newDiscount := 15.50
if err := factor.SetDiscount(ctx, newDiscount); err != nil {
    return err
}

// Update tax
newTax := 8.75
if err := factor.SetTax(ctx, newTax); err != nil {
    return err
}

// Update amount paid
newAmountPaid := 150.00
if err := factor.SetAmountPaid(ctx, newAmountPaid); err != nil {
    return err
}

// Update products
updatedProducts := json.RawMessage(`[
    {
        "product_id": 123,
        "product_name": "Updated Product",
        "quantity": 3,
        "unit_price": 50.00,
        "subtotal": 150.00
    }
]`)
if err := factor.SetProducts(ctx, updatedProducts); err != nil {
    return err
}
```

## Integration Patterns

### Pattern 1: Create Factor from Order

```go
func createFactorFromOrder(ctx context.Context, order UserOrder[int64]) error {
    // Get order details
    orderItems, _ := order.GetProductItems(ctx, nil, 0, 100, scommerce.QueueOrderAsc)
    total, _ := order.GetTotal(ctx)
    
    // Build products JSON
    type FactorProduct struct {
        ProductID   uint64          `json:"product_id"`
        ProductName string          `json:"product_name"`
        Quantity    uint64          `json:"quantity"`
        UnitPrice   float64         `json:"unit_price"`
        Subtotal    float64         `json:"subtotal"`
        Attributes  json.RawMessage `json:"attributes,omitempty"`
    }
    
    var products []FactorProduct
    var subtotal float64
    
    for _, item := range orderItems {
        productID, _ := item.ProductItem.GetID(ctx)
        productName, _ := item.ProductItem.GetName(ctx)
        price, _ := item.ProductItem.GetPrice(ctx)
        itemSubtotal := float64(item.Quantity) * price
        subtotal += itemSubtotal
        
        products = append(products, FactorProduct{
            ProductID:   productID,
            ProductName: productName,
            Quantity:    item.Quantity,
            UnitPrice:   price,
            Subtotal:    itemSubtotal,
            Attributes:  item.Attributes,
        })
    }
    
    productsJSON, _ := json.Marshal(products)
    
    // Calculate discount and tax
    discount := 0.0  // Apply your discount logic
    tax := subtotal * 0.08  // 8% tax
    amountPaid := subtotal - discount + tax
    
    // In a real implementation, you would create the factor in the database
    // This would require adding a NewUserFactor method to the manager
    
    return nil
}
```

### Pattern 2: Generate PDF Invoice

```go
func generateInvoicePDF(ctx context.Context, factor UserFactor[int64]) ([]byte, error) {
    // Get factor details
    factorID, _ := factor.GetID(ctx)
    products, _ := factor.GetProducts(ctx)
    discount, _ := factor.GetDiscount(ctx)
    tax, _ := factor.GetTax(ctx)
    amountPaid, _ := factor.GetAmountPaid(ctx)
    
    // Parse products
    var productList []map[string]interface{}
    json.Unmarshal(products, &productList)
    
    // Generate PDF using your preferred PDF library
    // Example structure:
    /*
    
    INVOICE #factorID
    ==================
    
    Products:
    - Product 1: $XX.XX x 2 = $XX.XX
    - Product 2: $XX.XX x 1 = $XX.XX
    
    Subtotal:     $XX.XX
    Discount:     -$XX.XX
    Tax:          +$XX.XX
    ----------------
    Total Paid:   $XX.XX
    
    */
    
    // Return PDF bytes
    return pdfBytes, nil
}
```

### Pattern 3: Sales Analytics

```go
func analyzeSalesData(ctx context.Context, userAccount UserAccount[int64]) error {
    // Get all factors
    factors, err := factorManager.GetUserFactors(
        ctx, userAccount, nil, 0, 1000, scommerce.QueueOrderAsc,
    )
    if err != nil {
        return err
    }
    
    var totalRevenue float64
    var totalDiscount float64
    var totalTax float64
    productCount := make(map[string]int)
    
    for _, factor := range factors {
        // Accumulate financial data
        amountPaid, _ := factor.GetAmountPaid(ctx)
        discount, _ := factor.GetDiscount(ctx)
        tax, _ := factor.GetTax(ctx)
        
        totalRevenue += amountPaid
        totalDiscount += discount
        totalTax += tax
        
        // Parse products for analytics
        products, _ := factor.GetProducts(ctx)
        var productList []map[string]interface{}
        json.Unmarshal(products, &productList)
        
        for _, product := range productList {
            productName := product["product_name"].(string)
            productCount[productName]++
        }
    }
    
    fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)
    fmt.Printf("Total Discounts: $%.2f\n", totalDiscount)
    fmt.Printf("Total Tax: $%.2f\n", totalTax)
    fmt.Printf("Product Sales: %+v\n", productCount)
    
    return nil
}
```

### Pattern 4: Customer Purchase History

```go
func displayPurchaseHistory(ctx context.Context, userAccount UserAccount[int64]) error {
    // Get factors with pagination
    page := 0
    pageSize := 10
    
    for {
        factors, err := factorManager.GetUserFactors(
            ctx,
            userAccount,
            nil,
            int64(page*pageSize),
            int64(pageSize),
            scommerce.QueueOrderDesc,
        )
        if err != nil {
            return err
        }
        
        if len(factors) == 0 {
            break
        }
        
        fmt.Printf("\n=== Page %d ===\n", page+1)
        
        for _, factor := range factors {
            factorID, _ := factor.GetID(ctx)
            amountPaid, _ := factor.GetAmountPaid(ctx)
            discount, _ := factor.GetDiscount(ctx)
            
            // Get products
            products, _ := factor.GetProducts(ctx)
            var productList []map[string]interface{}
            json.Unmarshal(products, &productList)
            
            fmt.Printf("\nInvoice #%d\n", factorID)
            fmt.Printf("Items: %d\n", len(productList))
            fmt.Printf("Discount: $%.2f\n", discount)
            fmt.Printf("Total: $%.2f\n", amountPaid)
        }
        
        page++
    }
    
    return nil
}
```

## Advanced Features

### Querying by Amount

```go
// Get factors within a price range
// Note: This requires a custom database query implementation
func getFactorsByAmountRange(ctx context.Context, minAmount, maxAmount float64) ([]UserFactor[int64], error) {
    // This would require extending the database contract with a custom query
    // The idx_factors_amount_paid index makes this efficient
    
    // Example custom query:
    // SELECT * FROM factors 
    // WHERE amount_paid BETWEEN $1 AND $2
    // ORDER BY amount_paid DESC
    
    return factors, nil
}
```

### Product Analysis from Factors

```go
type ProductSalesData struct {
    ProductID   uint64
    ProductName string
    TotalSold   int
    Revenue     float64
}

func analyzeProductSales(ctx context.Context) ([]ProductSalesData, error) {
    // Get all factors
    factors, err := factorManager.GetUserFactors(
        ctx, userAccount, nil, 0, 10000, scommerce.QueueOrderAsc,
    )
    if err != nil {
        return nil, err
    }
    
    salesData := make(map[uint64]*ProductSalesData)
    
    for _, factor := range factors {
        products, _ := factor.GetProducts(ctx)
        
        var productList []struct {
            ProductID   uint64  `json:"product_id"`
            ProductName string  `json:"product_name"`
            Quantity    int     `json:"quantity"`
            Subtotal    float64 `json:"subtotal"`
        }
        json.Unmarshal(products, &productList)
        
        for _, product := range productList {
            if _, exists := salesData[product.ProductID]; !exists {
                salesData[product.ProductID] = &ProductSalesData{
                    ProductID:   product.ProductID,
                    ProductName: product.ProductName,
                }
            }
            
            salesData[product.ProductID].TotalSold += product.Quantity
            salesData[product.ProductID].Revenue += product.Subtotal
        }
    }
    
    // Convert to slice
    var results []ProductSalesData
    for _, data := range salesData {
        results = append(results, *data)
    }
    
    return results, nil
}
```

## Best Practices

### 1. Consistent Product JSON Schema

Define a standard structure for products:

```go
type FactorProductItem struct {
    ProductID    uint64          `json:"product_id"`
    ProductName  string          `json:"product_name"`
    SKU          string          `json:"sku"`
    Quantity     int             `json:"quantity"`
    UnitPrice    float64         `json:"unit_price"`
    Subtotal     float64         `json:"subtotal"`
    Discount     float64         `json:"discount,omitempty"`
    Attributes   json.RawMessage `json:"attributes,omitempty"`
}
```

### 2. Validate Financial Data

```go
func validateFactorFinancials(products json.RawMessage, discount, tax, amountPaid float64) error {
    var productList []struct {
        Subtotal float64 `json:"subtotal"`
    }
    
    if err := json.Unmarshal(products, &productList); err != nil {
        return err
    }
    
    var subtotal float64
    for _, product := range productList {
        subtotal += product.Subtotal
    }
    
    expectedTotal := subtotal - discount + tax
    
    if math.Abs(expectedTotal-amountPaid) > 0.01 {
        return fmt.Errorf("amount mismatch: expected %.2f, got %.2f", 
            expectedTotal, amountPaid)
    }
    
    return nil
}
```

### 3. Archive Old Factors

```go
func archiveOldFactors(ctx context.Context, cutoffDate time.Time) error {
    // Move old factors to archive table for better performance
    // This would require extending the database contract
    
    // Example:
    // INSERT INTO factors_archive 
    // SELECT * FROM factors 
    // WHERE created_at < $1
    
    return nil
}
```

### 4. Index Optimization

Ensure proper indexes for common queries:

```sql
-- Already created by Init:
CREATE INDEX idx_factors_user_id ON factors(user_id);
CREATE INDEX idx_factors_amount_paid ON factors(amount_paid);

-- Additional useful indexes:
CREATE INDEX idx_factors_created_at ON factors(created_at);
CREATE INDEX idx_factors_user_amount ON factors(user_id, amount_paid);
```

### 5. Caching Factor Data

```go
type FactorCache struct {
    factors map[uint64]*BuiltinUserFactor[int64]
    mu      sync.RWMutex
    ttl     time.Duration
}

func (c *FactorCache) Get(factorID uint64) (*BuiltinUserFactor[int64], bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    factor, ok := c.factors[factorID]
    return factor, ok
}

func (c *FactorCache) Set(factorID uint64, factor *BuiltinUserFactor[int64]) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.factors[factorID] = factor
    
    // Auto-expire after TTL
    time.AfterFunc(c.ttl, func() {
        c.mu.Lock()
        delete(c.factors, factorID)
        c.mu.Unlock()
    })
}
```

## Performance Considerations

### Database Indexing

The system includes two important indexes:

1. **idx_factors_user_id**: Fast user factor lookups
2. **idx_factors_amount_paid**: Efficient amount-based queries

### JSONB Performance

PostgreSQL's JSONB type offers:
- ✅ Efficient storage and retrieval
- ✅ Indexing capabilities on JSON fields
- ✅ Fast JSON operations

**Tip**: For frequently queried JSON fields, consider GIN indexes:

```sql
CREATE INDEX idx_factors_products_gin ON factors USING GIN (products);
```

### Query Optimization

```go
// Good: Use pagination for large result sets
factors, _ := factorManager.GetUserFactors(ctx, user, nil, 0, 50, QueueOrderDesc)

// Bad: Loading all factors at once
factors, _ := factorManager.GetUserFactors(ctx, user, nil, 0, 10000, QueueOrderDesc)
```

## Troubleshooting

### Issue: Factors Not Found

**Symptoms**: `GetUserFactors` returns empty slice

**Solutions**:
1. Verify user account ID is correct
2. Check that factors exist in database
3. Ensure Init() was called to create tables

### Issue: JSON Parsing Errors

**Symptoms**: Errors when unmarshaling products

**Solutions**:
1. Validate JSON structure before storing
2. Use consistent product schema
3. Handle null/missing fields gracefully

```go
var productList []map[string]interface{}
if err := json.Unmarshal(products, &productList); err != nil {
    // Handle invalid JSON
    log.Printf("Invalid products JSON: %v", err)
    return err
}
```

### Issue: Slow Queries

**Symptoms**: Factor retrieval is slow

**Solutions**:
1. Verify indexes exist: `SELECT * FROM pg_indexes WHERE tablename = 'factors'`
2. Use pagination instead of loading all factors
3. Add additional indexes for common queries
4. Consider archiving old factors

## API Reference

### UserFactorManager Methods

| Method | Description |
|--------|-------------|
| `GetUserFactors(ctx, account, factors, skip, limit, order)` | Get user's factors |
| `GetUserFactorCount(ctx, account)` | Count user's factors |
| `RemoveAllUserFactors(ctx)` | Delete all factors |
| `Init(ctx)` | Initialize database schema |
| `Pulse(ctx)` | Periodic maintenance (no-op) |
| `Close(ctx)` | Cleanup resources |

### UserFactor Methods

| Method | Description |
|--------|-------------|
| `GetID(ctx)` | Get factor ID |
| `GetUserAccountID(ctx)` | Get user account ID |
| `GetProducts(ctx)` | Get products JSON |
| `SetProducts(ctx, products)` | Update products |
| `GetDiscount(ctx)` | Get discount amount |
| `SetDiscount(ctx, discount)` | Set discount amount |
| `GetTax(ctx)` | Get tax amount |
| `SetTax(ctx, tax)` | Set tax amount |
| `GetAmountPaid(ctx)` | Get total paid |
| `SetAmountPaid(ctx, amount)` | Set total paid |

## Summary

The User Factor system provides comprehensive invoice and receipt management:

✅ **Flexible Storage**: JSONB for product data  
✅ **Financial Tracking**: Discount, tax, and payment amounts  
✅ **Performance**: Indexed queries for fast retrieval  
✅ **User-Centric**: Easy factor lookup per user  
✅ **Thread-Safe**: Concurrent access protection  
✅ **Extensible**: JSON structure supports any product schema  
✅ **Analytics**: Enables sales and revenue analysis  

Perfect for e-commerce platforms requiring detailed transaction records and invoice generation!

## See Also

- [User Accounts](getting-started.md#accounts) - User management
- [Orders](getting-started.md#orders) - Order processing
- [Item Attributes](item-attributes.md) - Product customization
- [Product Management](getting-started.md#products) - Product catalog
