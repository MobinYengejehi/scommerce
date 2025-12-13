# Changelog

## [Latest] - Item Attributes Feature

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
