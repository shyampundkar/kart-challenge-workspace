# Order Food Service

A RESTful API service for ordering food online, built with Go and the Gin web framework.

## Features

- **HATEOAS REST Level 3 API**: Hypermedia-driven responses with resource navigation links
- **Pagination Support**: Efficient data retrieval with customizable page sizes
- **Promo Code Validation**: Smart validation against multiple data sources
- **Product Management**: List and retrieve product information
- **Order Management**: Create and manage orders with authentication
- **Health Checks**: Readiness and liveness probes for Kubernetes
- **CORS Support**: Cross-origin resource sharing enabled
- **Request Logging**: Structured logging for all requests
- **API Key Authentication**: Secure access to order endpoints
- **Comprehensive Testing**: Unit tests with 32.3% coverage

## API Endpoints

### Health Checks

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### Products

- `GET /api/products` - List all products (supports pagination)
- `GET /api/products/:productId` - Get a specific product

**Query Parameters:**
- `page` - Page number (default: 1)
- `perPage` - Items per page (default: 10, max: 100)

### Orders

- `GET /api/orders` - List all orders (requires authentication, supports pagination)
- `GET /api/orders/:orderId` - Get a specific order (requires authentication)
- `POST /api/orders` - Place an order with optional promo code (requires authentication)

**Query Parameters:**
- `page` - Page number (default: 1)
- `perPage` - Items per page (default: 10, max: 100)

## Authentication

The order endpoint requires an API key in the header:

```
api_key: apitest
```

## Running Locally

### Prerequisites

- Go 1.25.4 or later
- Docker (optional, for containerized deployment)

### Run with Go

```bash
# Install dependencies
go mod download

# Run the application
go run cmd/main.go
```

The server will start on port 8080 by default.

### Run with Docker

```bash
# Build the image
docker build -t order-food:latest .

# Run the container
docker run -p 8080:8080 order-food:latest
```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: orderfood)
- `DB_SSLMODE` - SSL mode (default: disable)

## Example API Calls

### List all products

```bash
curl http://localhost:8080/api/products
```

### List products with pagination

```bash
curl "http://localhost:8080/api/products?page=1&perPage=5"
```

### Get a specific product

```bash
curl http://localhost:8080/api/products/1
```

### List all orders

```bash
curl -H "api_key: apitest" http://localhost:8080/api/orders
```

### Get a specific order

```bash
curl -H "api_key: apitest" http://localhost:8080/api/orders/{orderId}
```

### Place an order

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{
    "items": [
      {
        "productId": "1",
        "quantity": 2
      },
      {
        "productId": "3",
        "quantity": 1
      }
    ],
    "couponCode": "HAPPYHRS"
  }'
```

### Health check

```bash
curl http://localhost:8080/health
```

## Project Structure

```
order-food/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handler/               # HTTP request handlers
│   │   ├── health_handler.go
│   │   ├── order_handler.go
│   │   └── product_handler.go
│   ├── middleware/            # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logger.go
│   ├── models/                # Data models
│   │   ├── order.go
│   │   ├── product.go
│   │   └── response.go
│   ├── repository/            # Data access layer
│   │   ├── order_repository.go
│   │   └── product_repository.go
│   ├── router/                # Route configuration
│   │   └── router.go
│   └── service/               # Business logic
│       ├── order_service.go
│       └── product_service.go
├── api/
│   └── openapi.yaml           # OpenAPI specification
├── helm/                      # Helm chart
├── Dockerfile
├── go.mod
└── README.md
```

## Development

### Add New Products

Edit `internal/repository/product_repository.go` and add products to the `seedData()` method.

### Change API Key

Edit `internal/middleware/auth.go` and update the `ValidAPIKey` constant.

## Kubernetes Deployment

Deploy using Helm:

```bash
helm install order-food ./helm
```

Access the service:

```bash
kubectl port-forward svc/order-food 8080:80
```

## Testing

### Run Unit Tests

**Run all tests with coverage:**
```bash
./test.sh
```

This will:
- Run all tests with race detection
- Generate coverage report
- Create HTML coverage report (coverage.html)
- Display total coverage percentage

**Run tests without coverage:**
```bash
go test ./... -v
```

**Run specific package tests:**
```bash
go test ./internal/handler -v
go test ./internal/middleware -v
go test ./internal/service -v
```

**View coverage in browser:**
```bash
./test.sh
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Test Coverage

Current test coverage: **32.3%**

| Component | Coverage |
|-----------|----------|
| Handlers  | 91.8%    |
| Middleware| 100.0%   |
| Services  | 37.9%    |
| Utils     | 100.0%   |

For detailed testing documentation, see [TESTING.md](TESTING.md).

### API Testing

Test the endpoints using the provided examples or import the OpenAPI specification (`api/openapi.yaml`) into tools like Postman or Swagger UI.

## API Documentation

The API follows the OpenAPI 3.1 specification defined in `api/openapi.yaml`.

## License

MIT
