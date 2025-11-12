# Extending Builtin Objects

This guide shows you how to customize S-Commerce for your specific needs. Learn to extend or replace builtin implementations while maintaining compatibility.

## Table of Contents

- [Extension Philosophy](#extension-philosophy)
- [Extension Strategies](#extension-strategies)
- [Extending Entities](#extending-entities)
- [Extending Managers](#extending-managers)
- [Custom Database Implementation](#custom-database-implementation)
- [Custom File Storage Implementation](#custom-file-storage-implementation)
- [Best Practices](#best-practices)
- [Common Pitfalls](#common-pitfalls)

## Extension Philosophy

### Why Extend?

S-Commerce provides complete, production-ready implementations. But every business has unique needs:

**Common Customization Needs:**
- Additional validation rules
- Business-specific calculations
- Integration with existing systems
- Custom storage backends
- Specialized search capabilities
- Audit logging
- Performance optimizations

### Contract-Based Design Enables Extension

Because S-Commerce uses interfaces for everything, you can:

**Replace Components:** Implement interfaces from scratch

**Extend Components:** Embed builtin types and override specific methods

**Wrap Components:** Add behavior before/after calling builtin

**Mix Approaches:** Use custom for some components, builtin for others

### When to Extend vs. Replace

**Extend (Embedding) When:**
- You want most builtin behavior
- Only need to change a few methods
- Want to add extra functionality
- Need to intercept specific operations

**Replace (Custom Implementation) When:**
- Fundamentally different logic needed
- Integrating with existing system
- Using different data model
- Need complete control

## Extension Strategies

### Strategy 1: Embedding Builtin Objects

**Concept:** Create new struct embedding builtin, override specific methods

**Pattern:**
```
CustomType struct {
    *scommerce.BuiltinType    // Embedded builtin
    AdditionalField           // Your extra fields
}
```

**Benefits:**
- Inherit all builtin behavior
- Override only what you need
- Simple implementation
- Full access to builtin methods

**When to Use:**
- Adding validation
- Adding logging
- Implementing hooks
- Extending functionality

---

### Strategy 2: Custom Implementation

**Concept:** Implement contract interface completely from scratch

**Pattern:**
```
CustomType struct {
    // Your fields
}

// Implement all interface methods
```

**Benefits:**
- Complete control
- No dependency on builtin
- Optimized for your use case
- Clean separation

**When to Use:**
- Different business logic
- Integration with existing systems
- Performance requirements
- Fundamentally different approach

---

### Strategy 3: Wrapper Pattern

**Concept:** Implement interface, delegate to builtin with interceptors

**Pattern:**
```
WrapperType struct {
    wrapped scommerce.InterfaceType
}

// Implement methods:
// - Pre-processing
// - Call wrapped
// - Post-processing
```

**Benefits:**
- Add behavior without modifying builtin
- Maintain compatibility
- Easy to add logging, metrics
- Chain multiple wrappers

**When to Use:**
- Cross-cutting concerns (logging, metrics)
- Authorization/authentication
- Caching
- Auditing

## Extending Entities

### Example: Custom User Account with Email Validation

**Scenario:** Enforce email format for account tokens, add email verification

**Approach: Embedding**

**Step 1: Define Custom Account**

```
CustomUserAccount struct:
- Embed *BuiltinUserAccount[uint64]
- Add EmailVerified field
- Add VerificationToken field
```

**Step 2: Override SetToken**

```
Implementation concept:
1. Validate email format using regex
2. If invalid, return error
3. If valid, call embedded.SetToken
4. Generate verification token
5. Store verification token
```

**Step 3: Add Email Verification Method**

```
VerifyEmail method:
1. Check verification token matches
2. If matches, set EmailVerified = true
3. Update database
4. Return success
```

**Step 4: Modify Authentication**

```
Override authentication to check EmailVerified:
1. Call embedded.Authenticate
2. Check if EmailVerified is true
3. If not, return error requiring verification
4. If verified, return account
```

**Usage:**

Replace BuiltinUserAccount with CustomUserAccount in your manager or use factory pattern to create custom accounts.

**Benefits:**
- Email validation enforced
- Verification workflow added
- All other functionality preserved
- Minimal code duplication

---

### Example: Adding Loyalty Points to User Account

**Scenario:** Calculate loyalty points based on order history

**Approach: Embedding with Computed Properties**

**Step 1: Extend Form Object**

```
LoyaltyUserAccountForm struct:
- Embed UserAccountForm[uint64]
- Add LoyaltyPoints *int64
```

**Step 2: Create Custom Account**

```
LoyaltyUserAccount struct:
- Embed *BuiltinUserAccount[uint64]
- Override form to use LoyaltyUserAccountForm
- Add OrderManager reference
```

**Step 3: Add GetLoyaltyPoints Method**

```
Implementation concept:
1. Check if cached in custom form
2. If cached, return immediately
3. If not, query order history via OrderManager
4. Calculate points: sum of order totals × multiplier
5. Cache in form
6. Return points
```

**Step 4: Add EarnPoints Method**

```
Implementation when order is placed:
1. Calculate points for order
2. Add to total loyalty points
3. Store in database
4. Invalidate cache
```

**Usage:**

Use LoyaltyUserAccount wherever you need loyalty functionality. Can coexist with regular UserAccount for users without loyalty program.

**Benefits:**
- Loyalty logic separated
- No changes to core account
- Can be toggled per user
- Backwards compatible

---

### Example: Product Item with Composite Pricing

**Scenario:** Price based on quantity (bulk discounts) and user level

**Approach: Embedding with Override**

**Step 1: Create Custom Product Item**

```
BulkPriceProductItem struct:
- Embed *BuiltinProductItem[uint64]
- Add PriceTiers []PriceTier
- Add UserLevelDiscounts map[int64]float64
```

**Step 2: Define Price Tier**

```
PriceTier struct:
- MinQuantity int64
- UnitPrice float64
```

**Step 3: Add GetPriceForQuantity Method**

```
Implementation concept:
1. Find applicable price tier for quantity
2. Get base price from tier
3. Apply user level discount if applicable
4. Return final unit price
```

**Step 4: Override AddProductItem in Category**

```
In custom category:
1. Create BulkPriceProductItem instead of BuiltinProductItem
2. Set price tiers from parameters
3. Return custom instance
```

**Usage:**

Shopping cart calculates price using GetPriceForQuantity instead of simple GetPrice.

**Benefits:**
- Flexible pricing
- Quantity discounts
- User-specific pricing
- Original pricing preserved

## Extending Managers

### Example: Product Manager with Elasticsearch

**Scenario:** Use Elasticsearch for product search, keep database for CRUD

**Approach: Wrapper with Selective Override**

**Step 1: Create Custom Manager**

```
ElasticsearchProductManager struct:
- builtin *BuiltinProductManager[uint64]
- elasticsearch *elasticsearch.Client
}
```

**Step 2: Implement ProductManager Interface**

Delegate most methods to builtin:
```
GetProductCategories: return builtin.GetProductCategories(...)
NewProductCategory: return builtin.NewProductCategory(...)
etc.
```

**Step 3: Override Search Methods**

```
SearchForProducts implementation:
1. Query Elasticsearch with search text
2. Get product IDs from results
3. For each ID, get Product from builtin
4. Return array of products

Index synchronization:
- When product created/updated, index in Elasticsearch
- When product deleted, remove from Elasticsearch
```

**Step 4: Configure Elasticsearch**

```
Setup:
1. Create Elasticsearch client
2. Define product index mapping
3. Index existing products on initialization
```

**Usage:**

Use ElasticsearchProductManager in App instead of BuiltinProductManager.

**Benefits:**
- Fast full-text search
- Rich query capabilities
- CRUD uses proven database logic
- Search specialized for performance

---

### Example: Account Manager with Admin Approval

**Scenario:** New accounts require admin approval before activation

**Approach: Embedding with Workflow Override**

**Step 1: Create Custom Manager**

```
ApprovalUserAccountManager struct:
- Embed *BuiltinUserAccountManager[uint64]
- Add PendingAccounts storage
- Add notification service
```

**Step 2: Override NewAccount**

```
Implementation:
1. Call embedded.NewAccount to create account
2. Set account as inactive (SetActive(false))
3. Add to pending accounts list
4. Notify admins of pending account
5. Return account with pending status
```

**Step 3: Add Approval Method**

```
ApproveAccount:
1. Get account from pending list
2. Call account.SetActive(true)
3. Remove from pending list
4. Notify user of approval
```

**Step 4: Add Rejection Method**

```
RejectAccount:
1. Get account from pending list
2. Delete account via RemoveAccount
3. Notify user of rejection
```

**Usage:**

Replace BuiltinUserAccountManager with ApprovalUserAccountManager in App.

**Benefits:**
- Account screening
- Spam prevention
- Quality control
- Existing functionality preserved

## Custom Database Implementation

### Example: MongoDB Database

**Scenario:** Use MongoDB instead of PostgreSQL

**Approach: Implement DBApplication from Scratch**

**Step 1: Create MongoDB Wrapper**

```
MongoDatabase struct:
- client *mongo.Client
- database *mongo.Database
- collections map[string]*mongo.Collection
```

**Step 2: Implement DBApplication Interface**

Must implement all sub-interfaces:
- DBUserAccountManager
- DBUserAccount
- DBProductManager
- DBProduct
- ... (all DB* interfaces)

**Step 3: Implement Manager Methods**

Example: NewUserAccount
```
Implementation concept:
1. Create MongoDB document
2. Insert into users collection
3. Get inserted ID
4. Populate form object from document
5. Return ID
```

**Step 4: Implement Entity Methods**

Example: GetUserAccountFirstName
```
Implementation concept:
1. Query users collection by ID
2. Extract first_name field
3. Populate form.FirstName
4. Return value
```

**Step 5: Handle Relationships**

MongoDB strategies:
```
Embedded documents for one-to-few:
- Profile images as array in user document

References for one-to-many:
- Orders as separate collection with user_id

Denormalization for performance:
- Product name in order items
```

**Step 6: Implement Pagination**

```
Using MongoDB skip/limit:
- collection.find().skip(skip).limit(limit).sort(order)
```

**Step 7: Handle Transactions**

```
MongoDB transactions:
- Start session
- Execute operations in transaction
- Commit or abort
- Requires replica set
```

**Challenges:**
- No joins (requires multiple queries or aggregation)
- Different transaction model
- Schema flexibility (validate in code)
- Mapping relational patterns to documents

**Benefits:**
- Document model flexibility
- Horizontal scalability
- JSON-like storage
- No schema migrations

---

### Example: Multi-Database Architecture

**Scenario:** PostgreSQL for orders, MongoDB for products, Redis for carts

**Approach: Composite DBApplication**

**Step 1: Create Composite Wrapper**

```
MultiDatabase struct:
- postgres *PostgresDatabase
- mongodb *MongoDatabase
- redis *RedisDatabase
```

**Step 2: Implement DBApplication**

Delegate to appropriate backend:
```
User/Order methods → PostgreSQL
Product methods → MongoDB
Cart methods → Redis
```

**Step 3: Handle Cross-Database Operations**

Example: Creating order from cart
```
1. Get cart from Redis
2. Get product details from MongoDB
3. Create order in PostgreSQL
4. Clear cart in Redis
All in application-level transaction with compensation
```

**Benefits:**
- Optimize each domain separately
- Use best tool for each job
- Scale components independently

**Challenges:**
- Transaction coordination
- Data consistency
- Complexity
- Testing

## Custom File Storage Implementation

### Example: AWS S3 Storage

**Scenario:** Store files in S3 instead of local disk

**Approach: Implement FileStorage Interface**

**Step 1: Create S3 Storage Struct**

```
S3FileStorage struct:
- s3Client *s3.Client
- bucketName string
- region string
- urlExpiration time.Duration
```

**Step 2: Implement Connect**

```
Implementation:
1. Initialize AWS SDK client
2. Configure with credentials/role
3. Verify bucket exists and accessible
4. Set up any required configuration
```

**Step 3: Implement Create**

```
Implementation:
1. Generate unique token (UUID)
2. Return S3FileIO for writing
3. S3FileIO buffers writes
4. On Close, upload to S3 with token as key
```

**Step 4: Implement Open**

```
Implementation:
1. Verify object exists (HeadObject)
2. Create S3FileIO for reading
3. On Read, stream from S3
4. Support seeking with range requests
```

**Step 5: Implement Delete**

```
Implementation:
1. Call DeleteObject with bucket and key
2. Handle errors (not found is success)
```

**Step 6: Implement S3FileIO**

```
S3FileIO struct:
- s3Client *s3.Client
- bucket string
- key string (token)
- buffer *bytes.Buffer
- position int64

Read:
- Download object or use GetObject
- Buffer data
- Return from buffer

Write:
- Append to buffer

Seek:
- Update position in buffer
- Or use range requests for reads

Close:
- If writing, upload buffer to S3
- If reading, cleanup
```

**Benefits:**
- Scalable storage
- No local disk needed
- Built-in redundancy
- CDN integration

**Considerations:**
- Latency higher than local
- Cost per operation
- Upload/download bandwidth
- Credentials management

---

### Example: Hybrid Storage (Local + Cloud)

**Scenario:** Thumbnails local, originals in cloud

**Approach: Conditional Delegation**

**Step 1: Create Hybrid Storage**

```
HybridFileStorage struct:
- local *LocalDiskFileStorage
- cloud *S3FileStorage
- sizeThreshold int64
}
```

**Step 2: Implement Create**

```
Implementation:
1. If file size < threshold:
   - Use local.Create
   - Prefix token with "local-"
2. Else:
   - Use cloud.Create
   - Prefix token with "cloud-"
3. Return appropriate FileIO
```

**Step 3: Implement Open**

```
Implementation:
1. Check token prefix
2. If "local-", use local.Open
3. If "cloud-", use cloud.Open
4. Return appropriate FileIO
```

**Benefits:**
- Fast access for small files
- Scalable storage for large files
- Cost optimization
- Performance tuning

## Best Practices

### Respecting Contracts

**Critical Rules:**

1. **Implement All Methods:** Every method in interface must be implemented

2. **Maintain Semantics:** Methods should do what the contract implies

3. **Error Handling:** Return errors, don't panic

4. **Context Respect:** Honor context cancellation and timeouts

5. **Thread Safety:** Ensure concurrent access is safe

### Maintaining Thread Safety

**When Extending:**

1. **Use Existing Locks:** If embedding, use embedded object's mutex

2. **Add New Locks:** For new fields, add appropriate synchronization

3. **Lock Order:** Prevent deadlocks with consistent lock ordering

4. **Test Concurrency:** Test with concurrent access

### Following Form Object Pattern

**Cache Management:**

1. **Populate Forms:** Always populate forms in database methods

2. **Invalidate Correctly:** Invalidate related fields on updates

3. **Clone for Safety:** Use form.Clone when needed

4. **Respect Optionals:** Don't assume pointers are non-nil

### Documenting Custom Behavior

**Documentation Needs:**

1. **Different Behavior:** Document any deviations from builtin

2. **Configuration:** Explain new configuration options

3. **Dependencies:** List additional dependencies

4. **Migration:** Provide migration guide from builtin

### Testing Extensions

**Test Strategy:**

1. **Interface Compliance:** Verify implements interface

2. **Functional Tests:** Test all methods

3. **Integration Tests:** Test with real dependencies

4. **Regression Tests:** Ensure doesn't break existing code

5. **Concurrent Tests:** Test thread safety

## Common Pitfalls

### Breaking Interface Contracts

**Problem:** Method doesn't satisfy contract semantics

**Example:** GetAccount returns different user than requested

**Solution:** Carefully read contract documentation, test thoroughly

---

### Ignoring Thread Safety

**Problem:** Concurrent access causes data races

**Example:** Updating field without lock

**Solution:** Use mutexes, test with race detector

---

### Not Invalidating Form Caches

**Problem:** Stale data returned after updates

**Example:** Update database but don't invalidate form

**Solution:** Invalidate related form fields on every update

---

### Circular Dependencies

**Problem:** Manager A depends on Manager B which depends on Manager A

**Example:** Custom managers referencing each other

**Solution:** Break circular dependency with interfaces or third component

---

### Forgetting Lifecycle Methods

**Problem:** Init/Pulse/Close not implemented or called

**Example:** Custom manager doesn't initialize database schema

**Solution:** Implement all lifecycle methods, call in correct order

---

### Incomplete Database Implementation

**Problem:** Implementing only some DB* interfaces

**Example:** Missing entity-level methods

**Solution:** Implement complete DBApplication interface

---

### Token Collisions in File Storage

**Problem:** Different files get same token

**Example:** Non-unique filename generation

**Solution:** Use UUIDs, check Exists, implement collision handling

## Advanced Patterns

### Factory Pattern for Custom Entities

**Scenario:** Create different entity types based on configuration

**Pattern:**

```
Entity factory function:
- Check configuration
- Return CustomEntity or BuiltinEntity
- Caller uses interface, doesn't care which
```

**Benefits:** Runtime flexibility, easy testing, gradual migration

---

### Decorator Chain

**Scenario:** Add multiple cross-cutting concerns

**Pattern:**

```
Chain wrappers:
- LoggingWrapper(MetricsWrapper(ValidationWrapper(Builtin)))
- Each adds behavior
- Delegates to next in chain
```

**Benefits:** Modular, composable, maintainable

---

### Strategy Pattern for Business Logic

**Scenario:** Different pricing strategies, shipping calculations

**Pattern:**

```
Define strategy interface
Implement multiple strategies
Choose strategy at runtime
Inject into entity/manager
```

**Benefits:** Testable, flexible, extensible

## Migration Strategies

### Gradual Migration from Builtin

**Approach:**

1. **Start:** Use all builtin

2. **Identify:** Find components needing customization

3. **Extend:** Create custom version of one component

4. **Test:** Verify behavior

5. **Deploy:** Use custom alongside builtin

6. **Repeat:** Migrate next component

**Benefits:** Low risk, incremental, reversible

---

### Testing Strategy During Extension

**Phases:**

1. **Unit Test:** Test custom logic in isolation

2. **Integration Test:** Test with real database/storage

3. **Compatibility Test:** Ensure works with builtin components

4. **Load Test:** Verify performance

5. **Regression Test:** Ensure no broken functionality

## Next Steps

- Review contracts to extend: [Contracts](contracts.md)
- See builtin implementations: [Managers](managers.md) and [Entities](entities.md)
- Practice with examples: [Examples](examples.md)
- Understand database layer: [Database Integration](database-integration.md)
- Learn file storage: [File Storage](file-storage.md)
