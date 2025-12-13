# Item Attributes Guide

## Overview

Shopping cart items and order items support **custom attributes** - flexible JSON data that allows you to store user-specific customizations, variant selections, and metadata alongside each item.

## What Are Item Attributes?

Attributes are arbitrary JSON data stored with each cart item and preserved through the order process. They enable:

- **Product Customizations**: Engraving text, custom colors, personalization
- **Variant Selections**: Size, color, material, style choices
- **Add-ons & Options**: Gift wrapping, expedited handling, special packaging
- **User Preferences**: Delivery notes, handling instructions per item
- **Bundle Configurations**: Component selections for configurable products
- **Metadata**: Any custom data your application needs per item

## Technical Implementation

### Data Type
Attributes are stored as `json.RawMessage` in Go and `JSONB` in PostgreSQL, providing:
- **Flexibility**: Store any valid JSON structure
- **Performance**: Efficient storage and retrieval
- **Queryability**: PostgreSQL JSONB supports indexing and queries (if needed)
- **Type Safety**: Validated JSON at the database level

### Data Flow
```
1. User selects customizations
   ↓
2. Add to cart with attributes
   ↓ (stored in shopping_cart_items.attributes)
3. Cart item preserves attributes
   ↓
4. Order placed from cart
   ↓ (PostgreSQL function reads attributes)
5. Attributes saved in orders.product_items
   ↓
6. Retrieve order items with attributes
```

## Usage Examples

### Basic Example: Product Variants

```go
import (
    "context"
    "encoding/json"
    "github.com/MobinYengejehi/scommerce/scommerce"
)

// Define attributes for a t-shirt
attrs := json.RawMessage(`{
    "size": "large",
    "color": "blue",
    "style": "v-neck"
}`)

// Add to cart with attributes
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, 2, attrs)
if err != nil {
    return err
}
```

### Advanced Example: Custom Engraving

```go
type EngravingOptions struct {
    Text     string `json:"text"`
    Font     string `json:"font"`
    Location string `json:"location"`
}

engraving := EngravingOptions{
    Text:     "Happy Anniversary!",
    Font:     "Cursive",
    Location: "center",
}

attrsBytes, err := json.Marshal(engraving)
if err != nil {
    return err
}

attrs := json.RawMessage(attrsBytes)
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, 1, attrs)
```

### Complex Example: Gift Bundle

```go
type GiftBundleAttrs struct {
    GiftWrap        bool   `json:"gift_wrap"`
    GiftMessage     string `json:"gift_message"`
    DeliveryDate    string `json:"delivery_date"`
    RecipientName   string `json:"recipient_name"`
    SpecialHandling bool   `json:"special_handling"`
}

attrs := GiftBundleAttrs{
    GiftWrap:        true,
    GiftMessage:     "Wishing you all the best!",
    DeliveryDate:    "2024-12-25",
    RecipientName:   "John Doe",
    SpecialHandling: true,
}

attrsJSON, _ := json.Marshal(attrs)
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, 1, json.RawMessage(attrsJSON))
```

## Working with Attributes

### Adding Items with Attributes

```go
// Simple inline JSON
attrs := json.RawMessage(`{"color": "red", "size": "M"}`)
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity, attrs)

// Or use nil for no attributes
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, quantity, nil)
```

### Retrieving Attributes from Cart Items

```go
// Get attributes from a cart item
attrs, err := cartItem.GetAttributes(ctx)
if err != nil {
    return err
}

// Parse the JSON
var customization struct {
    Color string `json:"color"`
    Size  string `json:"size"`
}
if err := json.Unmarshal(attrs, &customization); err != nil {
    return err
}

fmt.Printf("Color: %s, Size: %s\n", customization.Color, customization.Size)
```

### Updating Attributes

```go
// Update existing attributes
newAttrs := json.RawMessage(`{"color": "green", "size": "L"}`)
err := cartItem.SetAttributes(ctx, newAttrs)
if err != nil {
    return err
}
```

### Retrieving Attributes from Orders

```go
// Get order items with attributes
orderItems, err := order.GetProductItems(ctx, nil, 0, 100, scommerce.QueueOrderAsc)
if err != nil {
    return err
}

for _, item := range orderItems {
    fmt.Printf("Product: %d\n", item.ProductItem.ID)
    fmt.Printf("Quantity: %d\n", item.Quantity)
    
    // Parse attributes
    if len(item.Attributes) > 0 {
        var attrs map[string]interface{}
        if err := json.Unmarshal(item.Attributes, &attrs); err == nil {
            fmt.Printf("Customizations: %+v\n", attrs)
        }
    }
}
```

## Real-World Use Cases

### 1. Clothing Store - Size & Color Selection

```go
attrs := json.RawMessage(`{
    "size": "XL",
    "color": "navy blue",
    "fit": "regular"
}`)
```

### 2. Jewelry Store - Engraving Service

```go
attrs := json.RawMessage(`{
    "engraving_text": "Forever Yours",
    "engraving_font": "Script",
    "gift_box": true,
    "certificate_of_authenticity": true
}`)
```

### 3. Electronics - Configuration Options

```go
attrs := json.RawMessage(`{
    "ram": "16GB",
    "storage": "512GB SSD",
    "color": "Space Gray",
    "warranty_extension": "2 years",
    "setup_service": true
}`)
```

### 4. Bakery - Custom Cake Order

```go
attrs := json.RawMessage(`{
    "flavor": "chocolate",
    "frosting": "vanilla buttercream",
    "cake_text": "Happy Birthday Sarah!",
    "decoration_theme": "unicorn",
    "allergen_free": ["gluten", "nuts"],
    "pickup_date": "2024-06-15",
    "special_instructions": "Please make extra colorful"
}`)
```

### 5. Furniture - Assembly & Delivery Options

```go
attrs := json.RawMessage(`{
    "fabric": "velvet",
    "color": "emerald green",
    "assembly_required": false,
    "white_glove_delivery": true,
    "delivery_floor": 3,
    "old_furniture_removal": true,
    "delivery_window": "morning"
}`)
```

### 6. Gift Service - Complete Customization

```go
attrs := json.RawMessage(`{
    "gift_wrap": {
        "style": "premium",
        "color": "gold",
        "ribbon": true
    },
    "card": {
        "message": "Congratulations on your new home!",
        "signature": "From the Smiths"
    },
    "delivery": {
        "recipient": "Jane Doe",
        "leave_at_door": false,
        "signature_required": true
    }
}`)
```

## Best Practices

### 1. Define Schemas
Create Go structs for your attribute types:

```go
type ProductAttributes struct {
    Color      string   `json:"color,omitempty"`
    Size       string   `json:"size,omitempty"`
    Material   string   `json:"material,omitempty"`
    CustomText string   `json:"custom_text,omitempty"`
    Options    []string `json:"options,omitempty"`
}
```

### 2. Validate Attributes
Validate before storing:

```go
func ValidateAttributes(attrs json.RawMessage) error {
    var parsed ProductAttributes
    if err := json.Unmarshal(attrs, &parsed); err != nil {
        return fmt.Errorf("invalid attributes JSON: %w", err)
    }
    
    // Add business logic validation
    if parsed.Size != "" {
        validSizes := []string{"S", "M", "L", "XL", "XXL"}
        if !contains(validSizes, parsed.Size) {
            return fmt.Errorf("invalid size: %s", parsed.Size)
        }
    }
    
    return nil
}
```

### 3. Handle Missing Attributes
Always check for empty/null attributes:

```go
if len(item.Attributes) > 0 && string(item.Attributes) != "null" {
    // Parse and use attributes
    var attrs ProductAttributes
    json.Unmarshal(item.Attributes, &attrs)
}
```

### 4. Use Type-Safe Marshaling
Marshal structs instead of raw strings when possible:

```go
// Good
attrs := ProductAttributes{Color: "red", Size: "L"}
attrsJSON, _ := json.Marshal(attrs)
cartItem, err := cart.NewShoppingCartItem(ctx, item, 1, attrsJSON)

// Less ideal
attrs := json.RawMessage(`{"color":"red","size":"L"}`)
```

### 5. Document Expected Schema
Document what attributes your application supports:

```go
// Expected attributes for clothing products:
// {
//   "size": "S" | "M" | "L" | "XL" | "XXL",
//   "color": string,
//   "fit": "slim" | "regular" | "relaxed",
//   "length": "short" | "regular" | "long" (optional)
// }
```

## Database Schema

### shopping_cart_items Table
```sql
CREATE TABLE shopping_cart_items (
    id              BIGINT PRIMARY KEY,
    cart_id         BIGINT NOT NULL,
    product_item_id BIGINT NOT NULL,
    quantity        BIGINT NOT NULL,
    attributes      JSONB  -- Stores custom attributes
);
```

### orders.product_items Column
```json
[
    {
        "product_item_id": 123,
        "quantity": 2,
        "attributes": {
            "color": "blue",
            "size": "large",
            "engraving": "Happy Birthday!"
        }
    }
]
```

## Performance Considerations

### Storage
- JSONB is efficiently stored in PostgreSQL
- Typical attributes (< 1KB) have negligible storage impact
- Large attributes (> 10KB) should be avoided

### Queries
- Attributes don't affect standard query performance
- PostgreSQL JSONB supports indexing if you need to query by attributes
- Consider GIN indexes for frequently queried attribute fields

### Example: Adding an index
```sql
-- If you frequently search by color attribute
CREATE INDEX idx_cart_items_color ON shopping_cart_items 
USING GIN ((attributes -> 'color'));
```

## Troubleshooting

### Issue: Attributes Not Saved
**Cause**: Passing `nil` or invalid JSON
**Solution**: Always validate JSON before storing
```go
if _, err := json.Marshal(attrs); err != nil {
    return fmt.Errorf("invalid JSON: %w", err)
}
```

### Issue: Attributes Lost After Order
**Cause**: Check PostgreSQL function is updated
**Solution**: Ensure `order_shopping_cart` function includes attributes in `jsonb_build_object`

### Issue: Cannot Parse Attributes
**Cause**: Schema mismatch
**Solution**: Use flexible parsing
```go
var attrs map[string]interface{}
if err := json.Unmarshal(item.Attributes, &attrs); err != nil {
    // Handle gracefully
}
```

## Migration Guide

If you have an existing system without attributes:

### 1. Database Migration
```sql
-- Add attributes column to existing table
ALTER TABLE shopping_cart_items 
ADD COLUMN attributes JSONB;

-- Update existing rows (optional)
UPDATE shopping_cart_items 
SET attributes = 'null'::jsonb 
WHERE attributes IS NULL;
```

### 2. Update Code
```go
// Before
cartItem, err := cart.NewShoppingCartItem(ctx, item, quantity)

// After (pass nil for no attributes)
cartItem, err := cart.NewShoppingCartItem(ctx, item, quantity, nil)
```

### 3. Recreate order_shopping_cart Function
Run the updated `InitUserShoppingCartManager` to recreate the PostgreSQL function with attributes support.

## API Reference

### UserShoppingCartItem Methods
- `GetAttributes(ctx) (json.RawMessage, error)` - Retrieve item attributes
- `SetAttributes(ctx, attrs json.RawMessage) error` - Update item attributes

### UserShoppingCart Methods
- `NewShoppingCartItem(ctx, item ProductItem, count int64, attrs json.RawMessage) (UserShoppingCartItem, error)` - Add item with attributes

### UserOrderProductItem Struct
```go
type UserOrderProductItem struct {
    ProductItem *BuiltinProductItem `json:"product_item"`
    Quantity    uint64              `json:"quantity"`
    Attributes  json.RawMessage     `json:"attributes,omitempty"`
}
```

## Summary

Item attributes provide a powerful, flexible way to store customizations and metadata with cart and order items. Key points:

✅ Store any valid JSON structure  
✅ Attributes automatically flow from cart → order  
✅ Retrieve attributes anytime from cart or order items  
✅ Use `json.RawMessage` for type safety  
✅ JSONB storage is efficient and queryable  
✅ Perfect for product variants, customizations, and metadata  

Start using attributes today to build richer e-commerce experiences!
