# S-Commerce

**A powerful, flexible e-commerce library for Go**

S-Commerce is a comprehensive e-commerce library designed to help Go developers build production-ready online stores and marketplaces. Built with a contract-based architecture, it provides all the core functionality you need while remaining flexible enough to adapt to your specific business requirements.

## Features

### Core E-Commerce Functionality
- **User Management**: Complete account system with authentication, roles, profiles, and wallet management
- **Product Catalog**: Hierarchical categories, products with variants, inventory tracking, and search
- **Shopping Cart**: Session-based carts with multi-item support, **item-level attributes**, and seamless checkout
- **Order Management**: Full order lifecycle from creation to delivery with status tracking and **attribute preservation**
- **Payment Processing**: Multiple payment methods, expiry tracking, and default method management
- **Shipping**: Configurable shipping methods with pricing
- **Reviews & Ratings**: Customer reviews with rating aggregation
- **Address Management**: Multiple addresses per user with default address support
- **Item Customization**: Store user-specific customizations and metadata with cart and order items
- **🆕 Subscriptions**: Recurring billing, automatic renewals, and time-based access control for digital products

### Technical Features
- **Contract-Based Design**: All functionality defined through interfaces for maximum flexibility
- **Database Agnostic**: Implement your own database layer or use the PostgreSQL reference implementation
- **File Storage Abstraction**: Local disk storage included, easily swap for cloud storage (S3, GCS, Azure)
- **Generic Type System**: Use any comparable type for Account IDs (uint64, string, UUID, etc.)
- **Thread-Safe**: Built-in concurrency safety with proper mutex usage
- **Performance Optimized**: Smart caching with form objects to minimize database queries
- **Two-Factor Authentication**: Built-in OTP system for secure authentication flows

## Quick Start

### Installation

Add S-Commerce to your Go project:

```bash
go get github.com/MobinYengejehi/scommerce
```

### Basic Usage

Here's a minimal example to get you started:

**Step 1: Set up your database**

First, implement the database contracts or use the PostgreSQL sample. The library requires a database that implements the `DBApplication` interface.

**Step 2: Initialize the application**

Create your application instance with the required configuration:
- Database connection implementing `DBApplication`
- File storage (local disk or cloud)
- OTP configuration (code length, token length, TTL)

**Step 3: Run initialization**

Call the `Init` method on your application to set up database schemas and prepare all managers.

**Step 4: Start building**

Access managers through your application instance:
- `AccountManager` for user operations
- `ProductManager` for catalog management
- `ShoppingCartManager` for cart operations
- And many more!

**Step 5: Cleanup**

Always call `Close` when shutting down to properly release resources.

### Example Workflow

A typical user registration and product browsing flow:

1. **Request OTP**: Call `AccountManager.RequestTwoFactor` with email/username
2. **Send Code**: Deliver the OTP code to the user (email, SMS, etc.)
3. **Create Account**: Call `AccountManager.NewAccount` with token, password, and OTP code
4. **Browse Products**: Use `ProductManager.SearchForProducts` to find products
5. **View Details**: Retrieve product items with pricing and inventory
6. **Add to Cart**: Create a shopping cart and add items with custom attributes
7. **Checkout**: Convert cart to order with payment method and shipping address
8. **Retrieve Order**: Access order items with preserved customizations

### Item Attributes & Customization

**NEW FEATURE**: Shopping cart items and order items now support custom attributes!

Store user-specific information such as:
- Product customizations (engraving text, custom colors)
- Gift wrapping preferences
- Special delivery instructions per item
- Variant selections (size, color, material)
- Bundle configurations
- Add-on selections

```go
import "encoding/json"

// Define custom attributes for a cart item
attrs := json.RawMessage(`{
    "color": "blue",
    "size": "large",
    "engraving": "Happy Birthday!",
    "gift_wrap": true,
    "delivery_note": "Handle with care"
}`)

// Add item to cart with attributes
cartItem, err := cart.NewShoppingCartItem(ctx, productItem, 2, attrs)
if err != nil {
    // handle error
}

// Retrieve attributes from cart item
itemAttrs, err := cartItem.GetAttributes(ctx)
// itemAttrs contains the JSON: {"color": "blue", "size": "large", ...}

// Update attributes if needed
newAttrs := json.RawMessage(`{"color": "red", "size": "large"}`)
err = cartItem.SetAttributes(ctx, newAttrs)

// When you order, attributes are automatically preserved
order, err := cart.Order(ctx, paymentMethod, address, shippingMethod, "Please deliver by Friday")

// Retrieve order items - attributes are preserved!
orderItems, err := order.GetProductItems(ctx, nil, 0, 10, scommerce.QueueOrderAsc)
for _, item := range orderItems {
    fmt.Printf("Product ID: %d\n", item.ProductItem.ID)
    fmt.Printf("Quantity: %d\n", item.Quantity)
    fmt.Printf("Attributes: %s\n", string(item.Attributes))
    // Attributes: {"color": "blue", "size": "large", "engraving": "Happy Birthday!", ...}
}
```

**How it works:**
1. Attributes are stored as JSONB in the database
2. Cart items preserve attributes in `shopping_cart_items.attributes`
3. When ordering, attributes automatically flow to `orders.product_items`
4. Retrieve order items anytime with full attribute data
5. Use `json.RawMessage` for maximum flexibility

## Architecture

S-Commerce follows a clean, layered architecture:

```
┌─────────────────┐
│   Your App      │
└────────┬────────┘
         │
┌────────▼────────┐
│  App (Core)     │  ◄── Contains all managers
└────────┬────────┘
         │
┌────────▼────────┐
│   Managers      │  ◄── Business logic layer
└────────┬────────┘
         │
┌────────▼────────┐
│   Entities      │  ◄── Domain objects
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼────┐
│  DB  │  │  FS   │  ◄── Abstraction layers
└──────┘  └───────┘
```

**Key Concepts:**
- **Contracts**: Interfaces defining all functionality
- **Managers**: Coordinate entity creation and collection operations
- **Entities**: Individual domain objects (accounts, products, orders)
- **Forms**: Data transfer objects with caching for performance
- **DB Contracts**: Database abstraction layer
- **File Storage**: Abstraction for file operations

## Documentation

### Learning Path

**New to S-Commerce?**
1. Start with [Getting Started](docs/getting-started.md) - Set up your first application
2. Read [Architecture Guide](docs/architecture.md) - Understand the design philosophy
3. Explore [Examples](docs/examples.md) - See common patterns in action

**Implementing Integration?**
1. Review [Database Integration](docs/database-integration.md) - Implement your database layer
2. Check [File Storage Guide](docs/file-storage.md) - Set up file storage
3. Reference [Contracts](docs/contracts.md) - Understand all interfaces

**Customizing Behavior?**
1. Study [Extending Builtin Objects](docs/extending-builtin-objects.md) - Learn extension patterns
2. Review [Managers Reference](docs/managers.md) - Understand manager responsibilities
3. Check [Entities Reference](docs/entities.md) - Learn entity lifecycle

**Looking for Reference?**
- [API Reference](docs/api-reference.md) - Complete method documentation
- [Contracts](docs/contracts.md) - All interface definitions

### Complete Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System design and component relationships |
| [Getting Started](docs/getting-started.md) | Installation and first application |
| [Item Attributes](docs/item-attributes.md) | **NEW!** Store customizations with cart and order items |
| [Product Subscriptions](docs/product-subscriptions.md) | **NEW!** Recurring billing and subscription management |
| [User Factors](docs/user-factors.md) | Invoice and receipt management system |
| [User Discounts](docs/user-discounts.md) | Promotional code and discount management |
| [Contracts](docs/contracts.md) | Complete interface reference |
| [Database Integration](docs/database-integration.md) | Implementing database persistence |
| [File Storage](docs/file-storage.md) | File storage system guide |
| [Managers](docs/managers.md) | All manager subsystems |
| [Entities](docs/entities.md) | Entity objects and lifecycle |
| [Extending Builtin Objects](docs/extending-builtin-objects.md) | Customization guide |
| [Examples](docs/examples.md) | Practical use cases |
| [API Reference](docs/api-reference.md) | Complete API documentation |

## Key Advantages

### Flexibility
- **Database Freedom**: Not locked into a specific database - implement the contracts for any database system
- **Storage Options**: Use local disk, cloud storage, or hybrid approaches
- **Extensible**: Override any behavior through embedding or custom implementations
- **Type Flexibility**: Use your preferred ID type system-wide

### Production Ready
- **Thread-Safe**: Proper concurrency controls throughout
- **Performance**: Smart caching reduces database queries
- **Error Handling**: Clear error patterns and proper error propagation
- **Lifecycle Management**: Clean initialization and shutdown

### Developer Friendly
- **Clear Contracts**: Every interface is well-defined and documented
- **Reference Implementation**: PostgreSQL and local disk implementations included
- **Comprehensive Examples**: Learn from real-world scenarios
- **Type Safety**: Leverage Go's type system fully

## Project Status

S-Commerce is actively maintained and used in production environments. The core API is stable, though we continue to add features and improvements.

## Requirements

- **Go**: 1.18 or higher (for generics support)
- **Database**: Any database you can implement contracts for (PostgreSQL sample included)
- **File System**: Local disk access or cloud storage credentials

## Contributing

Contributions are welcome! Whether you're fixing bugs, improving documentation, or adding features, we appreciate your help.

**Ways to Contribute:**
- Report issues and bugs
- Suggest new features
- Improve documentation
- Submit pull requests
- Share your implementations (alternative databases, storage backends)

## License

S-Commerce is open-source software. Please check the LICENSE file for details.

## Support

- **Documentation**: Comprehensive guides in the `/docs` directory
- **Examples**: See `example/` directory for working code samples
- **Issues**: Report bugs and request features through the issue tracker

## Acknowledgments

S-Commerce is built with best practices from the Go community and modern e-commerce platforms. We thank all contributors and users who help improve the library.

---

**Ready to build your e-commerce platform?** Start with the [Getting Started Guide](docs/getting-started.md)!
