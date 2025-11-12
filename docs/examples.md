# Examples

This document provides practical examples of common S-Commerce operations. Each example shows the complete workflow with explanations.

## Table of Contents

- [Basic Setup](#basic-setup)
- [User Management](#user-management)
- [Product Catalog](#product-catalog)
- [Shopping and Checkout](#shopping-and-checkout)
- [Order Management](#order-management)
- [Advanced Scenarios](#advanced-scenarios)

## Basic Setup

### Initializing the Application

**Scenario:** Set up S-Commerce for the first time

**Prerequisites:**
- PostgreSQL running
- Empty database created
- Go 1.18+ installed

**Steps:**

**Step 1: Create Database Connection**

Parse your PostgreSQL connection string and create a database instance using the provided sample implementation. The connection string format is: `postgresql://username:password@host:port/database?sslmode=disable`

**Step 2: Set Up File Storage**

Create a directory for file storage (e.g., `./scommerce-files`) and initialize LocalDiskFileStorage pointing to this directory.

**Step 3: Configure Application**

Create an AppConfig with:
- Your database instance
- Your file storage instance
- OTP code length: 8 digits
- OTP token length: 32 characters
- OTP TTL: 2 minutes

Use uint64 as AccountID type for simplicity.

**Step 4: Create App Instance**

Call NewBuiltinApplication with your configuration to get an App instance with all managers initialized.

**Step 5: Initialize Database**

Call app.Init to create database schemas and set up all managers. This creates tables, indexes, and default data.

**Step 6: Set Up Periodic Maintenance**

Create a goroutine or use a ticker to call app.Pulse every minute. This handles OTP cleanup and other maintenance.

**Step 7: Use the Application**

Access managers through app (e.g., app.AccountManager, app.ProductManager) to perform operations.

**Step 8: Graceful Shutdown**

On application shutdown, call app.Close to clean up resources, close connections, and release memory.

**Key Points:**
- Init must be called before using managers
- Pulse should run periodically for maintenance
- Close ensures clean shutdown
- All operations use context for cancellation/timeouts

---

## User Management

### User Registration with OTP

**Scenario:** New user wants to create an account with two-factor authentication

**Prerequisites:**
- Initialized application
- Email/SMS service for sending OTP codes

**Flow:**

**Step 1: Request OTP Code**

User provides username/email. Call AccountManager.RequestTwoFactor with this token. This returns an OTP code (e.g., "12345678").

**Step 2: Deliver Code to User**

Send the code via email or SMS. The code is valid for 2 minutes (based on TTL configuration).

**Step 3: User Submits Registration**

User provides username, password, and the OTP code they received.

**Step 4: Create Account**

Call AccountManager.NewAccount with:
- Token: username/email
- Password: user's password (hash this first in production!)
- TwoFactor: the OTP code

If the code is valid and not expired, a UserAccount instance is returned.

**Step 5: OTP Session Cleanup**

The OTP session is automatically deleted after successful use. No manual cleanup needed.

**Error Handling:**
- Invalid OTP code: Return error asking user to check code
- Expired OTP: Request new code via RequestTwoFactor
- Duplicate username: Database returns unique constraint error

**Extension Ideas:**
- Add email verification link
- Implement password strength requirements
- Add CAPTCHA for security
- Rate limit OTP requests

---

### User Authentication (Login)

**Scenario:** Existing user wants to log in

**Flow:**

**Step 1: Request OTP (if using two-factor for login)**

Call AccountManager.RequestTwoFactor with user's username/email. Send code to user.

**Step 2: User Submits Credentials**

User provides username, password, and OTP code.

**Step 3: Authenticate**

Call AccountManager.Authenticate with credentials. If successful, returns UserAccount instance representing logged-in user.

**Step 4: Check Account Status**

After authentication, check if account is active and not banned:
- Call account.IsActive to verify account is active
- Call account.IsBanned to check ban status

**Step 5: Create Session**

Store account ID in session/JWT for subsequent requests.

**Security Notes:**
- Always hash passwords before storage
- Use HTTPS for all authentication
- Implement rate limiting
- Log failed attempts

---

### Updating User Profile

**Scenario:** User wants to update their profile information

**Prerequisites:**
- Authenticated user
- UserAccount instance

**Operations:**

**Update Basic Info:**
- Call account.SetFirstName with new first name
- Call account.SetLastName with new last name
- Call account.SetBio with profile biography

**Upload Profile Images:**
- Prepare image FileReaders (from uploaded files)
- Call account.SetProfileImages with array of readers
- Images are uploaded to file storage, tokens stored in database

**Change Password:**
- Verify current password with account.ValidatePassword
- Call account.SetPassword with new password (hashed)

**Update Role (Admin Only):**
- Retrieve desired role with RoleManager.GetUserRoleByName
- Call account.SetRole with the role instance

**Key Points:**
- Each Set method updates database immediately
- No transaction needed for single property
- Multiple updates can be done sequentially
- Cache is automatically invalidated

---

### Managing User Wallet

**Scenario:** Handle user account balance and transactions

**Operations:**

**Check Balance:**
- Call account.GetWalletCurrency to get current balance
- Negative balance indicates debt/penalty

**Add Funds:**
- Call account.ChargeWallet with amount to add
- Can be from payment, refund, admin credit

**Transfer Between Users:**
- Get source and destination UserAccount instances
- Call source.TransferCurrency with destination account and amount
- Verifies sufficient balance before transfer

**Apply Penalty/Fine:**
- Call account.Fine with amount to deduct
- Creates negative balance if insufficient funds

**Check Debt:**
- Call account.CalculateTotalDepts to get total debt including penalties
- Call account.CalculateTotalDeptsWithoutPenalty for cart debt only

**Financial Best Practices:**
- Use transactions for transfers
- Log all financial operations
- Implement balance checks before operations
- Set limits on transfers/charges

---

### Ban and Unban Users

**Scenario:** Moderator needs to ban user for policy violation

**Banning:**
- Call account.Ban with duration (e.g., 24 hours) and reason
- User cannot trade or perform sensitive operations while banned
- Ban reason is stored and retrievable

**Checking Ban:**
- Call account.IsBanned returns reason string
- Empty string means not banned
- Non-empty string is the ban reason

**Unbanning:**
- Call account.Unban to remove ban
- User regains full access

**Trading Restrictions:**
- Call account.AllowTrading(false) to disable trading
- Call account.IsTradingAllowed to check status
- Independent of ban status

**Use Cases:**
- Temporary suspensions
- Fraud prevention
- Policy enforcement
- Account restrictions

---

## Product Catalog

### Creating Product Categories

**Scenario:** Build a hierarchical product catalog

**Root Category:**

**Step 1: Create Top-Level Category**

Call ProductManager.NewProductCategory with:
- Name: "Electronics"
- ParentCategory: nil (indicates root category)

Returns a ProductCategory instance.

**Child Categories:**

**Step 2: Create Subcategory**

Call ProductManager.NewProductCategory with:
- Name: "Smartphones"
- ParentCategory: the Electronics category

**Step 3: Create Deeper Nesting**

Call ProductManager.NewProductCategory with:
- Name: "Apple"
- ParentCategory: the Smartphones category

**Result:** Electronics → Smartphones → Apple hierarchy

**Browsing Categories:**

Call ProductManager.GetProductCategories with pagination to list categories.

**Searching:**

Call ProductManager.SearchForProductCategories with:
- Search text: "phone"
- Deep search: true (searches entire hierarchy)

**Key Points:**
- Categories form tree structure
- Products belong to one category
- Search can traverse hierarchy
- Reorganize by changing parent

---

### Adding Products

**Scenario:** Add a product with images to the catalog

**Prerequisites:**
- Product category exists
- Product images prepared (optional)

**Steps:**

**Step 1: Get Category**

Retrieve or create the ProductCategory where product belongs.

**Step 2: Prepare Images (Optional)**

If you have images:
- Create FileReaders from image files
- Put in array

**Step 3: Create Product**

Call category.NewProduct with:
- Name: "iPhone 14"
- Description: "Latest Apple smartphone with advanced features"
- Images: array of FileReaders (or nil)

Returns a Product instance.

**Step 4: Update Product (Later)**

To change details:
- Call product.SetName for new name
- Call product.SetDescription for new description
- Call product.SetImages to replace images
- Call product.SetProductCategory to move to different category

**Listing Products:**

Call category.GetProducts with pagination to see products in category.

**Searching Products:**

Call ProductManager.SearchForProducts with:
- Search text: "iPhone"
- Deep search: true
- Category filter: specific category or nil for all

---

### Adding Product Items (Variants)

**Scenario:** Add variants of a product with different sizes, colors, prices

**Prerequisites:**
- Product exists

**Creating Variants:**

**Variant 1: iPhone 14 128GB Black**

Call product.AddProductItem with:
- SKU: "IPHONE14-128GB-BLK"
- Name: "iPhone 14 128GB Black"
- Price: 999.99
- Quantity: 50 (initial stock)
- Images: variant-specific images (or nil)
- Attrs: JSON like `{"color": "black", "storage": "128GB"}`

**Variant 2: iPhone 14 256GB White**

Call product.AddProductItem with different parameters for white 256GB variant.

**Attributes Format:**

The Attrs field is JSON for flexible properties:
- Color, size, material, etc.
- Custom attributes per product type
- Used for filtering and display

**Inventory Management:**

**Check Stock:**
- Call item.GetQuantityInStock

**Update Stock:**
- Call item.SetQuantityInStock to set absolute quantity
- Call item.AddQuantityInStock with positive delta (restock) or negative (sale)

**Pricing:**
- Call item.SetPrice to update price
- Price changes don't affect existing orders

**Searching Items:**

Call ProductManager.SearchForProductItems with:
- Search text: "128GB"
- Product filter: specific product or nil
- Category filter: specific category or nil

**Key Points:**
- One product, many items (variants)
- Each item has unique SKU
- Inventory tracked per item
- Attributes are flexible JSON

---

## Shopping and Checkout

### Creating Shopping Cart

**Scenario:** User browses products and adds items to cart

**Prerequisites:**
- User authenticated
- UserAccount instance
- Products exist in catalog

**Step 1: Create Cart**

Call account.NewShoppingCart with:
- SessionText: unique session identifier (e.g., "session-" + UUID)

Returns UserShoppingCart instance.

**Session Management:**

For anonymous users before login:
- Generate random session text
- Store in cookie/local storage
- After login, associate cart with user account
- Retrieve cart with ShoppingCartManager.GetShoppingCartBySessionText

**Step 2: Add Items to Cart**

**Find Product Item:**
- Search for product items using ProductManager.SearchForProductItems
- Get specific item by ID or SKU

**Add to Cart:**
- Call cart.NewShoppingCartItem with:
  - Item: the ProductItem instance
  - Count: quantity (e.g., 2)

Returns UserShoppingCartItem instance.

**Step 3: Update Quantities**

**Increase Quantity:**
- Call cartItem.AddQuantity with positive delta

**Decrease Quantity:**
- Call cartItem.AddQuantity with negative delta

**Set Exact Quantity:**
- Call cartItem.SetQuantity with desired amount

**Remove Item:**
- Call cart.RemoveShoppingCartItem with the item

**Step 4: Calculate Total**

Before checkout, calculate cart total:
- Get shipping method with ShippingMethodManager.GetShippingMethodByName
- Call cart.CalculateDept with shipping method
- Returns total including items and shipping

**Cart Management:**

**List Cart Items:**
- Call cart.GetShoppingCartItems with pagination

**Clear Cart:**
- Call cart.RemoveAllShoppingCartItems

**Multiple Carts:**
- Users can have multiple carts (wishlists, saved carts)
- List with account.GetShoppingCarts

---

### Checkout Process

**Scenario:** Convert shopping cart to order

**Prerequisites:**
- Cart with items
- User has address
- User has payment method
- Shipping method selected

**Complete Checkout Flow:**

**Step 1: Verify Cart**

Ensure cart has items and stock is available for each item.

**Step 2: Get/Create Address**

**Option A: Use Default**
- Call account.GetDefaultAddress

**Option B: Create New**
- Get country with CountryManager.GetCountryByName
- Call account.NewAddress with:
  - Unit number, street number
  - Address lines 1 and 2
  - City, region, postal code
  - Country instance
  - isDefault: true/false

**Step 3: Get/Create Payment Method**

**Option A: Use Default**
- Call account.GetDefaultPaymentMethod

**Option B: Create New**
- Get payment type with PaymentTypeManager.GetPaymentTypeByName
- Call account.NewPaymentMethod with:
  - PaymentType instance
  - Provider: "Visa", "MasterCard", etc.
  - Account number: last 4 digits only
  - Expiry date
  - isDefault: true/false

**Step 4: Select Shipping Method**

Get shipping method:
- Call ShippingMethodManager.GetShippingMethodByName
- Or list all with GetShippingMethods

**Step 5: Place Order**

Call cart.Order with:
- Payment method instance
- Shipping address instance
- Shipping method instance
- User comment: optional order notes

Returns UserOrder instance.

**Step 6: Cart After Order**

After successful order:
- Cart items are copied to order (snapshot)
- Cart remains but typically cleared
- Order is independent of cart changes

**Payment Processing:**

S-Commerce doesn't process payments directly. Your application should:
1. Get payment method details from order
2. Call payment gateway (Stripe, PayPal, etc.)
3. Update order status based on payment result

**Error Handling:**
- Insufficient stock: Return error, ask user to adjust
- Payment declined: Keep cart, notify user
- Invalid address: Ask for correction
- Out of stock: Remove item or update quantity

---

## Order Management

### Viewing Orders

**Scenario:** User wants to see their order history

**User's Orders:**

Call account.GetOrders with pagination to get list of UserOrder instances.

**Order Details:**

For each order:
- Call order.GetOrderDate for when placed
- Call order.GetOrderTotal for total price
- Call order.GetStatus for current status
- Call order.GetShippingAddress for delivery address
- Call order.GetPaymentMethod for payment used
- Call order.GetProductItems for what was ordered

**Admin View:**

Call OrderManager.GetUserOrders with pagination to see all orders across all users.

**Filtering:**

Filter by status:
- Get all orders
- For each, call GetStatus
- Filter in application code

Or implement custom database query in your DBApplication.

---

### Updating Order Status

**Scenario:** Admin or system updates order as it progresses

**Status Flow:**

Typical e-commerce statuses:
1. Idle/Pending (initial)
2. Processing
3. Shipped
4. Delivered

**Update Status:**

**Step 1: Get Status**

Get desired status:
- Call OrderStatusManager.GetOrderStatusByName

**Step 2: Update Order**

Call order.SetStatus with the status instance.

**Special: Mark as Delivered:**

Call order.Deliver with:
- Delivery date (time.Time)
- Delivery comment: notes from courier

This sets:
- Delivery date
- Delivery comment  
- Status to "Delivered" automatically

**Check if Delivered:**

Call order.IsDeliveried which returns true if order has been delivered.

**Status Notifications:**

In your application:
- Listen for status changes
- Send emails/SMS to user
- Update tracking information
- Trigger fulfillment processes

---

### Reviewing Products

**Scenario:** User wants to review a product they purchased

**Prerequisites:**
- User has received order
- User authenticated

**Steps:**

**Step 1: Get Product Item**

From order or by searching:
- Call order.GetProductItems to see what was ordered
- Or search ProductManager

**Step 2: Submit Review**

Call UserReviewManager.NewUserReview with:
- Account: the user's account instance
- ProductItem: item being reviewed
- Rating value: 1-5 (or your scale)
- Comment: review text

Returns UserReview instance.

**Step 3: Update Review**

To modify later:
- Call review.SetRatingValue for new rating
- Call review.SetComment for new text

**Viewing Reviews:**

**For Product Item:**
- Call item.GetUserReviews with pagination
- Call item.CalculateAverageRating for average

**For User:**
- Call account.GetUserReviews to see user's reviews

**All Reviews:**
- Call UserReviewManager.GetUserReviews for all reviews

**Moderation:**

In your application:
- Review comments before publishing
- Allow reporting inappropriate reviews
- Admin can delete with UserReviewManager.RemoveUserReview

---

## Advanced Scenarios

### Complete E-Commerce Workflow

**Scenario:** End-to-end shopping experience

**Flow:**

1. **User Registration**
   - Request OTP, user receives code
   - Create account with credentials and OTP

2. **Browse Catalog**
   - Search products by category
   - View product details
   - Check product item availability

3. **Add to Cart**
   - Create shopping cart
   - Add multiple items with quantities
   - Update quantities as needed

4. **Prepare Checkout**
   - Create shipping address
   - Add payment method
   - Select shipping option

5. **Place Order**
   - Review cart total
   - Submit order
   - Process payment (external)

6. **Track Order**
   - View order status
   - Receive status updates
   - Confirm delivery

7. **Post-Purchase**
   - Submit product review
   - Check order history
   - Request support (future feature)

**Integration Points:**
- Email service for OTP and notifications
- Payment gateway for processing
- Shipping provider for tracking
- Review moderation system

---

### Multi-User Currency Transfer

**Scenario:** Users can transfer currency between accounts

**Steps:**

**Step 1: Get Both Accounts**

Retrieve sender and recipient:
- Call AccountManager.GetAccount for each user

**Step 2: Verify Balance**

Check sender has sufficient funds:
- Call sender.GetWalletCurrency
- Compare to transfer amount

**Step 3: Execute Transfer**

Call sender.TransferCurrency with:
- Recipient account instance
- Amount to transfer

**Step 4: Verify Results**

Check balances updated:
- Sender balance decreased
- Recipient balance increased

**Security:**

- Authenticate sender
- Verify amount is positive
- Limit transfer amounts
- Log all transfers
- Implement approval for large amounts

**Use Cases:**
- Peer-to-peer marketplace
- Gift giving
- Refunds
- Credits

---

### Inventory Management

**Scenario:** Track and manage product inventory

**Operations:**

**Receiving Stock:**

When new inventory arrives:
- Get ProductItem
- Call item.AddQuantityInStock with quantity received

**Selling Items:**

When order is placed:
- For each item in order
- Call item.AddQuantityInStock with negative quantity (e.g., -2)

**Stock Checks:**

Before checkout:
- For each cart item
- Get product item quantity
- Verify sufficient stock
- Return error if insufficient

**Low Stock Alerts:**

In your application:
- Periodically check item.GetQuantityInStock
- Alert when below threshold
- Trigger reorder process

**Inventory Reports:**

Query database or:
- List all product items
- Check quantities
- Generate reports

---

### Reference Data Management

**Scenario:** Set up reference data (countries, payment types, shipping methods)

**Countries:**

Create common countries:
- Call CountryManager.NewCountry for each country

**Payment Types:**

Create payment types:
- "Credit Card"
- "PayPal"
- "Bank Transfer"
- "Cash on Delivery"

**Shipping Methods:**

Create shipping options:
- "Standard" - $5.00
- "Express" - $15.00
- "Overnight" - $25.00

**Order Statuses:**

Create statuses:
- "Idle"
- "Processing"
- "Shipped"
- "Delivered"
- "Cancelled"

**User Roles:**

Create roles:
- "Customer"
- "Admin"
- "Moderator"
- "Vendor"

**Best Practices:**
- Create during application initialization
- Check existence before creating
- Use consistent naming
- Document special statuses (delivered, idle)

---

## Testing Tips

### Unit Testing

Test individual operations:
- Mock database and file storage
- Test manager methods
- Verify error handling
- Check edge cases

### Integration Testing

Test with real database:
- Use test database
- Create test data
- Run full workflows
- Clean up after tests

### End-to-End Testing

Test complete scenarios:
- Register user
- Create products
- Complete checkout
- Verify order created

## Next Steps

- Understand underlying contracts: [Contracts](contracts.md)
- Implement database layer: [Database Integration](database-integration.md)
- Set up file storage: [File Storage](file-storage.md)
- Learn customization: [Extending Builtin Objects](extending-builtin-objects.md)
