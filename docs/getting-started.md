# Getting Started with S-Commerce

This guide will help you set up S-Commerce and build your first e-commerce application. By the end, you'll have a working application that can manage users, products, and orders.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Basic Configuration](#basic-configuration)
- [First Application](#first-application)
- [Understanding the Flow](#understanding-the-flow)
- [Next Steps](#next-steps)

## Prerequisites

Before starting with S-Commerce, ensure you have:

### Required

**Go 1.18 or Higher**

S-Commerce uses Go generics, which require Go 1.18+. Check your version:

```bash
go version
```

If you need to upgrade, visit [golang.org](https://golang.org/dl/)

**Database System**

S-Commerce requires a database for persistence. While the library is database-agnostic, we recommend starting with PostgreSQL:

- PostgreSQL 12+ (recommended for beginners)
- Or any database you can implement contracts for (MySQL, MongoDB, etc.)

**File System Access**

For storing product images, user profile pictures, and other files:

- Local disk storage (builtin)
- Or cloud storage (S3, GCS, Azure Blob) with custom implementation

### Recommended

- Basic understanding of Go interfaces and generics
- Familiarity with e-commerce concepts (users, products, orders, etc.)
- Experience with your chosen database system

## Installation

### Step 1: Create Your Project

Create a new Go module for your e-commerce application:

```bash
mkdir my-ecommerce-app
cd my-ecommerce-app
go mod init github.com/yourusername/my-ecommerce-app
```

### Step 2: Add S-Commerce Dependency

Install the S-Commerce library:

```bash
go get github.com/MobinYengejehi/scommerce
```

This will download S-Commerce and its dependencies.

### Step 3: Install Database Driver

For PostgreSQL, add the pgx driver:

```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

For other databases, install the appropriate driver.

### Step 4: Verify Installation

Create a simple test file to verify:

```bash
touch main.go
```

Add imports to verify packages are available:
- `github.com/MobinYengejehi/scommerce/scommerce`
- `github.com/jackc/pgx/v5/pgxpool`

## Basic Configuration

### Database Setup

#### PostgreSQL Configuration

**Create Database**:

```sql
CREATE DATABASE ecommerce_db;
CREATE USER ecommerce_user WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE ecommerce_db TO ecommerce_user;
```

**Connection String Format**:
```
postgresql://username:password@host:port/database?sslmode=disable
```

**Example**:
```
postgresql://ecommerce_user:your_password@localhost:5432/ecommerce_db?sslmode=disable
```

### File Storage Setup

#### Local Disk Storage

Create a directory for file storage:

```bash
mkdir scommerce-files
```

This directory will store:
- User profile images
- Product images
- Product item images

**Permissions**: Ensure your application has read/write access to this directory.

### OTP Configuration

S-Commerce includes built-in two-factor authentication using One-Time Passwords (OTP). Configure:

**OTP Code Length**: How many digits in the code (recommended: 6 or 8)
- Example: 6 digits = codes like "123456"

**OTP Token Length**: Length of the session token (recommended: 32)
- Used internally to track OTP sessions

**OTP Time-To-Live**: How long codes remain valid (recommended: 2-5 minutes)
- Balance security (shorter) vs user experience (longer)

## First Application

Let's build a complete application that demonstrates the core features.

### Application Structure

Create the following structure:

```
my-ecommerce-app/
  ├── main.go                    # Application entry point
  ├── scommerce-files/           # File storage directory
  └── go.mod                     # Module definition
```

### Implementation Overview

Your application will:

1. **Initialize Database**: Set up PostgreSQL connection
2. **Configure File Storage**: Point to local directory
3. **Create App Instance**: Initialize S-Commerce with configuration
4. **Run Init**: Set up database schemas and managers
5. **Perform Operations**: Create users, products, orders
6. **Cleanup**: Properly close connections

### Step-by-Step Implementation

#### Step 1: Import Required Packages

You'll need:
- `context` for managing request context
- `github.com/MobinYengejehi/scommerce/scommerce` for core library
- `github.com/MobinYengejehi/scommerce/scommerce/db_samples/postgresql` for database
- `github.com/MobinYengejehi/scommerce/scommerce/fs` for file storage
- `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL connection
- `time` for durations and timestamps

#### Step 2: Create Main Function

Set up the application context and defer cleanup.

#### Step 3: Configure PostgreSQL Connection

Parse the connection string using pgxpool.ParseConfig, providing your database credentials.

Example connection string:
```
postgresql://ecommerce_user:your_password@localhost:5432/ecommerce_db?sslmode=disable
```

#### Step 4: Create Database Instance

Use the PostgreSQL sample implementation:
- Call NewPostgreDatabase with context and configuration
- Defer database Close to ensure cleanup

#### Step 5: Configure File Storage

Create local disk file storage:
- Use NewLocalDiskFileStorage with your directory path
- Defer Close for cleanup

#### Step 6: Create App Configuration

Build the AppConfig structure with:
- DB: Your database instance
- FileStorage: Your file storage instance
- OTPCodeLength: 8 (8-digit codes)
- OTPTokenLength: 32 (32-character tokens)
- OTPTTL: 2 minutes (time.Minute * 2)

Choose your AccountID type (uint64 is simplest to start).

#### Step 7: Create App Instance

Call NewBuiltinApplication with your configuration. This creates an App instance with all managers initialized.

#### Step 8: Initialize the Application

Call app.Init(ctx) to set up database schemas and initialize all managers. This must be done before using any managers.

#### Step 9: Set Up Periodic Maintenance

Create a ticker to call app.Pulse(ctx) every minute. This handles OTP cleanup and other maintenance tasks.

#### Step 10: Use the Application

Now you can use the managers to perform operations!

### Example Operations

#### Creating a User Account

**Workflow**:

1. **Request OTP Code**: Call AccountManager.RequestTwoFactor with username/email
   - Returns: OTP code (deliver this to user via email/SMS)

2. **User Receives Code**: In real application, send via email/SMS
   - For testing: Print the code or use a known value

3. **Create Account**: Call AccountManager.NewAccount with:
   - Token: username or email
   - Password: user's password (hash this in production!)
   - TwoFactor: the OTP code

4. **Result**: Returns UserAccount instance if successful

**Important Notes**:
- OTP codes expire after TTL (2 minutes in our config)
- If code expires, request a new one
- OTP session is deleted after successful use

#### Creating Product Categories

**Workflow**:

1. **Create Root Category**: Call ProductManager.NewProductCategory with:
   - Name: "Electronics"
   - ParentCategory: nil (for root category)

2. **Create Child Category**: Call ProductManager.NewProductCategory with:
   - Name: "Smartphones"
   - ParentCategory: the Electronics category

**Result**: Hierarchical category structure

#### Creating Products

**Workflow**:

1. **Get Category**: Retrieve or create a ProductCategory

2. **Prepare Product Images**: Create file readers for product images
   - For testing: You can skip images or use empty arrays

3. **Create Product**: Call category.NewProduct with:
   - Name: "iPhone 14"
   - Description: "Latest iPhone model"
   - Images: array of FileReader (or nil)

**Result**: Returns Product instance

#### Adding Product Items (Variants)

**Workflow**:

1. **Get Product**: Retrieve the Product instance

2. **Create Product Item**: Call product.AddProductItem with:
   - SKU: "IPHONE14-128GB-BLK"
   - Name: "iPhone 14 128GB Black"
   - Price: 999.99
   - Quantity: 50 (stock level)
   - Images: array of FileReader (or nil)
   - Attrs: JSON attributes like {"color": "black", "storage": "128GB"}

3. **Repeat**: Create more variants (different colors, sizes, etc.)

**Result**: Multiple product items for the same product

#### Creating Shopping Cart and Ordering

**Workflow**:

1. **Get User Account**: Retrieve or create UserAccount

2. **Create Shopping Cart**: Call account.NewShoppingCart with:
   - SessionText: "session-" + random string

3. **Add Items to Cart**: 
   - Get ProductItem instance
   - Call cart.NewShoppingCartItem with:
     - Item: the ProductItem
     - Count: quantity (e.g., 2)

4. **Create Address**:
   - Get or create Country
   - Call account.NewAddress with address details
   - Mark as default if desired

5. **Create Payment Method**:
   - Get or create PaymentType
   - Call account.NewPaymentMethod with payment details
   - Mark as default if desired

6. **Get Shipping Method**:
   - Retrieve shipping method from ShippingMethodManager

7. **Place Order**: Call cart.Order with:
   - PaymentMethod: user's payment method
   - Address: shipping address
   - ShippingMethod: chosen shipping option
   - UserComment: optional order notes

**Result**: Shopping cart converted to UserOrder

### Error Handling

Always check errors returned by methods:

- **Database Errors**: Connection failures, query errors
- **Validation Errors**: Invalid email, expired OTP, insufficient stock
- **Not Found Errors**: Entity doesn't exist
- **Constraint Errors**: Duplicate entries, foreign key violations

**Best Practice**: Log errors and return meaningful messages to users.

### Testing Your Application

**Verify**:

1. **Database Tables**: Check that tables were created by Init
2. **User Creation**: Try creating a user with OTP
3. **Product Catalog**: Create categories and products
4. **Order Flow**: Complete end-to-end ordering process
5. **File Storage**: Verify files are created in storage directory

## Understanding the Flow

### Application Lifecycle

```mermaid
graph LR
    A[Create Config] --> B[New App]
    B --> C[Init]
    C --> D[Use Managers]
    D --> E[Pulse Periodically]
    E --> D
    D --> F[Close]
```

**Explanation**:

1. **Create Config**: Set up database, file storage, OTP settings
2. **New App**: Create App instance with builtin managers
3. **Init**: Initialize database schemas and managers
4. **Use Managers**: Perform business operations
5. **Pulse Periodically**: Run maintenance tasks
6. **Close**: Clean shutdown

### Request Flow Example: User Registration

```mermaid
graph TB
    A[Client Requests Registration] --> B[Request OTP]
    B --> C[Generate OTP Code]
    C --> D[Store OTP Session]
    D --> E[Send Code to User]
    E --> F[User Submits Code + Password]
    F --> G[Validate OTP]
    G --> H[Create User Account]
    H --> I[Delete OTP Session]
    I --> J[Return Success]
```

**Key Points**:

- OTP provides two-factor security
- OTP session is temporary (expires after TTL)
- Same flow for login (Authenticate instead of NewAccount)

### Manager and Entity Interaction

```mermaid
graph LR
    A[Your Code] --> B[Manager]
    B --> C[Create Entity]
    C --> D[Initialize Entity]
    D --> E[Return Entity]
    E --> A
    A --> F[Entity Method]
    F --> G[Database]
    G --> F
    F --> A
```

**Pattern**:

1. Your code calls manager to create entity
2. Manager handles database creation
3. Manager initializes entity instance
4. You use entity methods for operations
5. Entity methods interact with database directly

## Next Steps

### Immediate Next Steps

**Explore Documentation**:

1. **Architecture Guide**: Understand the design patterns - [Architecture](architecture.md)
2. **Contracts Reference**: Learn all available interfaces - [Contracts](contracts.md)
3. **Examples**: See more complex scenarios - [Examples](examples.md)

**Enhance Your Application**:

1. Add proper error handling and logging
2. Implement email/SMS delivery for OTP codes
3. Add password hashing (use bcrypt)
4. Create REST API endpoints
5. Add authentication middleware
6. Implement proper transaction handling

### Database Implementation

If the PostgreSQL sample doesn't meet your needs:

1. **Study the Sample**: Review db_samples/postgresql implementation
2. **Understand Contracts**: Read all DB* interfaces in db_contracts.go
3. **Implement Your Database**: Create your own DBApplication implementation
4. **Follow Patterns**: Use the PostgreSQL sample as a reference

See: [Database Integration Guide](database-integration.md)

### File Storage Implementation

To use cloud storage instead of local disk:

1. **Study LocalDiskFileStorage**: Review scommerce/fs/local_disk.go
2. **Implement FileStorage**: Create your cloud storage implementation
3. **Implement FileIO**: Create your file handle type
4. **Handle Tokens**: Use cloud object keys as tokens

See: [File Storage Guide](file-storage.md)

### Customization

To customize behavior:

1. **Identify Component**: Which manager or entity needs customization?
2. **Choose Strategy**: Embed builtin or implement from scratch?
3. **Implement Interface**: Satisfy the contract
4. **Test Thoroughly**: Ensure correct behavior

See: [Extending Builtin Objects](extending-builtin-objects.md)

### Production Readiness

Before going to production:

**Security**:
- Hash passwords (use bcrypt or argon2)
- Use HTTPS/TLS for all communications
- Implement rate limiting for OTP requests
- Add CSRF protection
- Validate all user inputs

**Performance**:
- Set up database indexes for common queries
- Implement connection pooling
- Add caching where appropriate
- Monitor query performance
- Use CDN for static files

**Reliability**:
- Implement proper error handling
- Add logging and monitoring
- Set up database backups
- Test failure scenarios
- Implement graceful shutdown

**Scalability**:
- Consider database replication
- Use message queues for async tasks
- Implement horizontal scaling
- Add load balancing
- Monitor resource usage

### Learning Resources

**Recommended Reading Order**:

1. [Architecture](architecture.md) - Understand the design
2. [Contracts](contracts.md) - Learn all interfaces
3. [Managers](managers.md) - Explore manager capabilities
4. [Entities](entities.md) - Understand entity lifecycle
5. [Examples](examples.md) - See practical patterns
6. [Database Integration](database-integration.md) - Implement database
7. [File Storage](file-storage.md) - Implement file storage
8. [Extending Builtin Objects](extending-builtin-objects.md) - Customize behavior
9. [API Reference](api-reference.md) - Complete method reference

### Community and Support

**Get Help**:
- Check documentation for answers
- Review examples for common patterns
- Look at sample implementations
- Ask questions in community channels

**Contribute**:
- Report bugs and issues
- Suggest improvements
- Share your implementations
- Improve documentation

## Troubleshooting

### Common Issues

**Issue: "Cannot create app - interface conversion error"**
- **Cause**: Database doesn't implement all required interfaces
- **Solution**: Ensure your database implements complete DBApplication interface

**Issue: "OTP validation failed"**
- **Cause**: Code expired or incorrect
- **Solution**: Check OTP TTL, verify code delivery, ensure correct code entry

**Issue: "Database tables not created"**
- **Cause**: Init not called or database errors
- **Solution**: Call app.Init(ctx), check database logs for errors

**Issue: "Files not found"**
- **Cause**: File storage directory doesn't exist or no permissions
- **Solution**: Create directory, check permissions, verify path

**Issue: "Generic type error"**
- **Cause**: Mixing AccountID types
- **Solution**: Use same AccountID type throughout application

### Debug Tips

**Enable Logging**: Add logging to database and file storage operations

**Check Database**: Verify tables created, data inserted correctly

**Inspect Forms**: Forms show cached vs database state

**Test Incrementally**: Test each component separately before integration

**Use Context Timeouts**: Add timeouts to prevent hanging operations

## Summary

You now have:

✅ S-Commerce installed and configured
✅ Database connection established
✅ File storage configured
✅ First application running
✅ Understanding of core concepts
✅ Knowledge of next steps

**Continue Learning**: Explore the documentation to build more sophisticated e-commerce features!
