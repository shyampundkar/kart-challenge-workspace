# Order Food Service

A RESTful API service for ordering food online, built with Go and the Gin web framework.

## Features

- List all available products
- Get product details by ID
- Place orders with authentication
- Health check endpoints for Kubernetes
- CORS support
- Request logging
- API key authentication

## API Endpoints

### Health Checks

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### Products

- `GET /api/product` - List all products
- `GET /api/product/:productId` - Get a specific product

### Orders

- `POST /api/order` - Place an order (requires authentication)

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

## Example API Calls

### List all products

```bash
curl http://localhost:8080/api/product
```

### Get a specific product

```bash
curl http://localhost:8080/api/product/1
```

### Place an order

```bash
curl -X POST http://localhost:8080/api/order \
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
    "couponCode": "SAVE10"
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

Test the endpoints using the provided examples or import the OpenAPI specification (`api/openapi.yaml`) into tools like Postman or Swagger UI.

## API Documentation

The API follows the OpenAPI 3.1 specification defined in `api/openapi.yaml`.

## License

MIT
