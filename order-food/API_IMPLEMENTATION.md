# Order Food API Implementation

This document describes the implementation of the Order Food API based on the OpenAPI 3.1 specification.

## Implementation Overview

The API has been implemented using:
- **Go 1.25.4** - Programming language
- **Gin Web Framework** - HTTP web framework
- **UUID** - For generating unique order IDs
- **In-memory storage** - For development/testing (can be replaced with database)

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────┐
│   Router    │ - Route configuration and middleware setup
└──────┬──────┘
       │
┌──────▼──────┐
│  Handlers   │ - HTTP request/response handling
└──────┬──────┘
       │
┌──────▼──────┐
│  Services   │ - Business logic
└──────┬──────┘
       │
┌──────▼──────┐
│ Repositories│ - Data access layer
└──────┬──────┘
       │
┌──────▼──────┐
│    Data     │ - In-memory storage
└─────────────┘
```

## Components

### 1. Models (`internal/models/`)

Defines the data structures according to the OpenAPI schema:

- **Product** - Product information (id, name, price, category)
- **OrderItem** - Individual item in an order (productId, quantity)
- **OrderReq** - Order creation request (items, couponCode)
- **Order** - Complete order (id, items, products)
- **ApiResponse** - Standard API response format

### 2. Repositories (`internal/repository/`)

Data access layer with in-memory storage:

- **ProductRepository** - Manages product data
  - Pre-seeded with 10 sample products
  - Thread-safe operations using sync.RWMutex
  - Methods: GetAll(), GetByID(), GetByIDs()

- **OrderRepository** - Manages order data
  - Thread-safe operations
  - Methods: Create(), GetByID(), GetAll()

### 3. Services (`internal/service/`)

Business logic layer:

- **ProductService** - Product-related operations
  - ListProducts() - Returns all products
  - GetProduct(id) - Returns single product

- **OrderService** - Order-related operations
  - PlaceOrder(req) - Creates new order with UUID
  - GetOrder(id) - Retrieves order by ID
  - Validates products exist before creating order

### 4. Handlers (`internal/handler/`)

HTTP request handlers:

- **ProductHandler** - Product endpoints
  - GET /api/product - List all products
  - GET /api/product/:productId - Get specific product

- **OrderHandler** - Order endpoints
  - POST /api/order - Place new order (requires auth)

- **HealthHandler** - Health check endpoints
  - GET /health - Liveness probe
  - GET /ready - Readiness probe

### 5. Middleware (`internal/middleware/`)

HTTP middleware components:

- **AuthMiddleware** - API key authentication
  - Validates "api_key" header
  - Expected value: "apitest"
  - Returns 401 if missing, 403 if invalid

- **CORSMiddleware** - Cross-Origin Resource Sharing
  - Allows all origins (*)
  - Supports all common HTTP methods
  - Includes api_key in allowed headers

- **LoggerMiddleware** - Request logging
  - Logs method, URI, IP, status, duration
  - Useful for debugging and monitoring

### 6. Router (`internal/router/`)

Route configuration and setup:
- Applies global middleware (CORS, Logger)
- Groups routes by functionality
- Applies auth middleware selectively
- Follows RESTful conventions

## API Endpoints Implementation

### Product Endpoints

#### GET /api/product
**OpenAPI Operation:** `listProducts`

**Implementation:**
```go
func (h *ProductHandler) ListProducts(c *gin.Context)
```

- Returns all products from repository
- No authentication required
- Response: Array of Product objects
- Status: 200 OK

#### GET /api/product/:productId
**OpenAPI Operation:** `getProduct`

**Implementation:**
```go
func (h *ProductHandler) GetProduct(c *gin.Context)
```

- Validates productId parameter
- Looks up product in repository
- Response: Single Product object or error
- Status: 200 OK, 400 Bad Request, 404 Not Found

### Order Endpoints

#### POST /api/order
**OpenAPI Operation:** `placeOrder`

**Implementation:**
```go
func (h *OrderHandler) PlaceOrder(c *gin.Context)
```

- Requires "api_key: apitest" header
- Validates request body against OrderReq schema
- Validates all product IDs exist
- Generates UUID for order
- Returns created order with product details
- Status: 200 OK, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 422 Unprocessable Entity

## Authentication Implementation

As per OpenAPI spec, the `/order` endpoint uses API key authentication:

```yaml
security:
  - api_key: ["create_order"]
```

**Implementation:**
- Middleware checks "api_key" header
- Valid key: "apitest"
- Invalid/missing key returns appropriate error response
- Only applied to order creation endpoint

## Sample Data

The ProductRepository is pre-seeded with 10 products:

| ID | Name                  | Price  | Category |
|----|-----------------------|--------|----------|
| 1  | Chicken Waffle        | 12.99  | Waffle   |
| 2  | Belgian Waffle        | 10.99  | Waffle   |
| 3  | Blueberry Pancakes    | 9.99   | Pancakes |
| 4  | Chocolate Pancakes    | 11.99  | Pancakes |
| 5  | Caesar Salad          | 8.99   | Salad    |
| 6  | Greek Salad           | 9.49   | Salad    |
| 7  | Margherita Pizza      | 13.99  | Pizza    |
| 8  | Pepperoni Pizza       | 15.99  | Pizza    |
| 9  | Cheeseburger          | 11.49  | Burger   |
| 10 | Veggie Burger         | 10.49  | Burger   |

## Error Handling

All error responses follow the ApiResponse schema:

```json
{
  "code": 404,
  "type": "error",
  "message": "Product not found"
}
```

Error codes:
- **400** - Invalid input/Bad request
- **401** - Missing API key
- **403** - Invalid API key
- **404** - Resource not found
- **422** - Validation error

## Request/Response Examples

### List Products
```bash
curl http://localhost:8080/api/product
```

Response:
```json
[
  {
    "id": "1",
    "name": "Chicken Waffle",
    "price": 12.99,
    "category": "Waffle"
  },
  ...
]
```

### Get Product
```bash
curl http://localhost:8080/api/product/1
```

Response:
```json
{
  "id": "1",
  "name": "Chicken Waffle",
  "price": 12.99,
  "category": "Waffle"
}
```

### Place Order
```bash
curl -X POST http://localhost:8080/api/order \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{
    "items": [
      {"productId": "1", "quantity": 2},
      {"productId": "3", "quantity": 1}
    ],
    "couponCode": "SAVE10"
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "items": [
    {"productId": "1", "quantity": 2},
    {"productId": "3", "quantity": 1}
  ],
  "products": [
    {
      "id": "1",
      "name": "Chicken Waffle",
      "price": 12.99,
      "category": "Waffle"
    },
    {
      "id": "3",
      "name": "Blueberry Pancakes",
      "price": 9.99,
      "category": "Pancakes"
    }
  ]
}
```

## Testing

A test script is provided: `test-api.sh`

```bash
# Test local server
./test-api.sh

# Test remote server
./test-api.sh http://example.com:8080
```

The script tests:
1. Health endpoints
2. Product listing
3. Product retrieval
4. Order creation with/without auth
5. Error cases (invalid products, missing auth, etc.)

## Building and Running

### Local Development
```bash
# Run directly
go run cmd/main.go

# Or use Makefile
make run
```

### Docker
```bash
# Build
docker build -t order-food:latest .

# Run
docker run -p 8080:8080 order-food:latest
```

### Kubernetes
```bash
# Deploy with Helm
helm install order-food ./helm

# Access service
kubectl port-forward svc/order-food 8080:80
```

## Configuration

Environment variables:
- `PORT` - Server port (default: 8080)

To change API key, edit `internal/middleware/auth.go`:
```go
const ValidAPIKey = "apitest"
```

## Future Enhancements

Potential improvements:
1. Database integration (PostgreSQL, MySQL, MongoDB)
2. JWT-based authentication
3. Order status tracking
4. Payment processing
5. User management
6. Order history
7. Product search and filtering
8. Pagination for product list
9. Rate limiting
10. Metrics and monitoring
11. Swagger UI integration
12. Unit and integration tests

## Compliance with OpenAPI Spec

The implementation fully complies with the OpenAPI 3.1 specification:

✅ All endpoints implemented
✅ Request/response schemas match
✅ Authentication implemented as specified
✅ Error codes match specification
✅ Operation IDs followed
✅ Tags respected (product, order)

## Performance Considerations

Current implementation:
- In-memory storage (fast but not persistent)
- Thread-safe operations using mutexes
- Suitable for development/testing
- For production, replace with:
  - Database for persistence
  - Caching layer for frequently accessed data
  - Load balancing for horizontal scaling
