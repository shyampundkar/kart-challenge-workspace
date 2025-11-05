# Order Food API - Quick Start Guide

## What's Been Implemented

A complete RESTful API service based on the OpenAPI specification with:
- ✅ Gin web framework integration
- ✅ All API endpoints from openapi.yaml
- ✅ API key authentication
- ✅ Health check endpoints
- ✅ CORS support
- ✅ Request logging
- ✅ Clean architecture (handlers, services, repositories)
- ✅ Pre-seeded sample data (10 products)
- ✅ Docker support
- ✅ Kubernetes/Helm deployment
- ✅ Comprehensive testing script

## Quick Commands

### Run Locally (Quick Test)
```bash
# From order-food directory
go run cmd/main.go

# Server starts on http://localhost:8080
```

### Test the API
```bash
# In another terminal
./test-api.sh

# Or test manually
curl http://localhost:8080/api/product
```

### Build and Run with Docker
```bash
# Build
docker build -t order-food:latest .

# Run
docker run -p 8080:8080 order-food:latest
```

### Deploy to Minikube
```bash
# From workspace root
./deploy.sh

# Access the service
kubectl port-forward svc/order-food 8080:80

# Test
cd order-food && ./test-api.sh
```

## API Endpoints

### Public Endpoints (No Auth)
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/product` - List all products
- `GET /api/product/:id` - Get product by ID

### Protected Endpoints (Requires API Key)
- `POST /api/order` - Place an order

**API Key:** Add header `api_key: apitest`

## Example Usage

### List All Products
```bash
curl http://localhost:8080/api/product
```

### Get Product by ID
```bash
curl http://localhost:8080/api/product/1
```

### Place an Order
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

## Available Products

| ID | Name                  | Price  | Category |
|----|-----------------------|--------|----------|
| 1  | Chicken Waffle        | $12.99 | Waffle   |
| 2  | Belgian Waffle        | $10.99 | Waffle   |
| 3  | Blueberry Pancakes    | $9.99  | Pancakes |
| 4  | Chocolate Pancakes    | $11.99 | Pancakes |
| 5  | Caesar Salad          | $8.99  | Salad    |
| 6  | Greek Salad           | $9.49  | Salad    |
| 7  | Margherita Pizza      | $13.99 | Pizza    |
| 8  | Pepperoni Pizza       | $15.99 | Pizza    |
| 9  | Cheeseburger          | $11.49 | Burger   |
| 10 | Veggie Burger         | $10.49 | Burger   |

## Project Structure

```
order-food/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── handler/                   # HTTP handlers
│   │   ├── product_handler.go
│   │   ├── order_handler.go
│   │   └── health_handler.go
│   ├── middleware/                # Middleware
│   │   ├── auth.go               # API key authentication
│   │   ├── cors.go               # CORS support
│   │   └── logger.go             # Request logging
│   ├── models/                    # Data models
│   │   ├── product.go
│   │   ├── order.go
│   │   └── response.go
│   ├── repository/                # Data layer
│   │   ├── product_repository.go
│   │   └── order_repository.go
│   ├── service/                   # Business logic
│   │   ├── product_service.go
│   │   └── order_service.go
│   └── router/                    # Route setup
│       └── router.go
├── api/
│   └── openapi.yaml              # API specification
├── helm/                          # Kubernetes deployment
├── Dockerfile
├── Makefile
├── test-api.sh                    # API test script
├── README.md                      # Full documentation
├── API_IMPLEMENTATION.md          # Implementation details
└── QUICK_START.md                 # This file
```

## Makefile Commands

```bash
make help         # Show all commands
make deps         # Download dependencies
make build        # Build binary
make run          # Run application
make test-api     # Test API endpoints
make docker-build # Build Docker image
make docker-run   # Run in Docker
make clean        # Clean build artifacts
```

## Deployment Architecture

```
┌─────────────────────────────────────────┐
│         Minikube Cluster                │
│                                         │
│  ┌────────────────┐  ┌───────────────┐ │
│  │ database-      │  │ database-load │ │
│  │ migration      │  │ (Job)         │ │
│  │ (Job)          │  └───────────────┘ │
│  └────────────────┘                     │
│                                         │
│  ┌────────────────────────────────────┐ │
│  │        order-food                  │ │
│  │       (Deployment)                 │ │
│  │                                    │ │
│  │  ┌──────────────────────────────┐ │ │
│  │  │  Pod: order-food-xxxxx       │ │ │
│  │  │  - Port 8080                 │ │ │
│  │  │  - Health checks enabled     │ │ │
│  │  └──────────────────────────────┘ │ │
│  └────────────────────────────────────┘ │
│                                         │
│  ┌────────────────────────────────────┐ │
│  │      Service: order-food           │ │
│  │      ClusterIP: xxx.xxx.xxx.xxx    │ │
│  │      Port: 80 → 8080               │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
                    │
                    │ port-forward
                    ▼
            http://localhost:8080
```

## Troubleshooting

### Server won't start
```bash
# Check if port 8080 is in use
lsof -i :8080

# Use different port
PORT=3000 go run cmd/main.go
```

### Dependencies issues
```bash
go mod tidy
go mod download
```

### Docker build fails
```bash
# Clear Docker cache
docker builder prune

# Rebuild
docker build --no-cache -t order-food:latest .
```

### Kubernetes pod not starting
```bash
# Check pod status
kubectl get pods

# View logs
kubectl logs -l app.kubernetes.io/name=order-food

# Describe pod
kubectl describe pod <pod-name>
```

## Next Steps

1. **Test the API** - Use `test-api.sh` or curl commands
2. **Modify Products** - Edit `internal/repository/product_repository.go`
3. **Add Database** - Replace in-memory storage with PostgreSQL/MySQL
4. **Add Tests** - Create unit and integration tests
5. **Add Swagger UI** - Integrate Swagger for interactive API docs
6. **Deploy to Cloud** - Use Helm chart with cloud provider

## Learn More

- [README.md](README.md) - Full documentation
- [API_IMPLEMENTATION.md](API_IMPLEMENTATION.md) - Implementation details
- [api/openapi.yaml](api/openapi.yaml) - OpenAPI specification
- [../DEPLOYMENT.md](../DEPLOYMENT.md) - Deployment guide

## Support

For issues or questions:
1. Check logs: `kubectl logs -l app.kubernetes.io/name=order-food`
2. Verify health: `curl http://localhost:8080/health`
3. Test endpoints: `./test-api.sh`
