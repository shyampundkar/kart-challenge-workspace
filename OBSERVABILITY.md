# Observability with OpenTelemetry

Complete guide to distributed tracing and metrics in the kart-challenge-workspace project.

## Overview

All three modules are fully instrumented with OpenTelemetry for production-grade observability:

| Module | Tracing | Metrics | Auto-Instrumentation |
|--------|---------|---------|----------------------|
| **database-migration** | ✅ | ❌ | Manual spans |
| **database-load** | ✅ | ❌ | Manual spans |
| **order-food** | ✅ | ✅ | Automatic HTTP tracing |

## Architecture

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  database-       │────▶│  database-load   │────▶│   order-food     │
│  migration       │     │                  │     │   (API Server)   │
│                  │     │  Traces:         │     │                  │
│  Traces:         │     │  - dataload.*    │     │  Traces:         │
│  - migration.*   │     └──────────────────┘     │  - HTTP requests │
└──────────────────┘              │                │  - Spans         │
         │                        │                │                  │
         │  Jaeger Export         │                │  Metrics:        │
         │                        │                │  - Prometheus    │
         └────────────────────────┴────────────────┤  - /metrics      │
                                  │                └──────────────────┘
                                  ▼
                          ┌──────────────┐
                          │   Jaeger     │
                          │  All-in-One  │
                          └──────────────┘
                                  │
                                  ▼
                        http://localhost:16686
                         (Jaeger UI)
```

## Quick Start

### 1. Start with Docker Compose

```bash
# Start all services with Jaeger
docker-compose up --build

# Services available:
# - Order Food API: http://localhost:8080
# - Jaeger UI: http://localhost:16686
# - Prometheus Metrics: http://localhost:8080/metrics
```

### 2. Generate Some Traces

```bash
# List products
curl http://localhost:8080/api/product

# Get specific product
curl http://localhost:8080/api/product/1

# Place an order (generates authenticated trace)
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

### 3. View Traces in Jaeger

1. Open http://localhost:16686
2. Select service from dropdown:
   - `database-migration`
   - `database-load`
   - `order-food`
3. Click "Find Traces"
4. Click on a trace to see detailed view

### 4. View Metrics

```bash
# Prometheus metrics endpoint
curl http://localhost:8080/metrics
```

## Components

### 1. Telemetry Package

Each module has `internal/telemetry/telemetry.go`:

**Features:**
- Tracer initialization with Jaeger exporter
- Metrics initialization with Prometheus (order-food only)
- Resource configuration (service name, version, environment)
- Graceful shutdown handlers
- Environment-based configuration

**Usage:**
```go
import "github.com/shyampundkar/kart-challenge-workspace/order-food/internal/telemetry"

// Get configuration from environment
config := telemetry.GetConfig("service-name")

// Initialize telemetry
shutdown, err := telemetry.InitTelemetry(config)
if err != nil {
    log.Fatal(err)
}
defer telemetry.GracefulShutdown(shutdown, 5*time.Second)
```

### 2. Tracing Implementation

#### database-migration

**Spans:**
- `migration.execute` - Root span for entire migration
- `migration.createTables` - Table creation operations
- `migration.createIndexes` - Index creation operations

**Events:**
- Tables created successfully
- Indexes created successfully

#### database-load

**Spans:**
- `dataload.execute` - Root span for entire load operation
- `dataload.loadProducts` - Product data loading
- `dataload.loadUsers` - User data loading
- `dataload.loadOrders` - Order data loading

**Events:**
- Loaded X products
- Loaded X users
- Loaded X orders

#### order-food

**Automatic Tracing:**
- HTTP requests (method, path, status code)
- Request/response headers
- Duration
- Errors

**Middleware:** `otelgin.Middleware("order-food")`

### 3. Metrics (order-food only)

**Prometheus Metrics:**
- HTTP request count
- Request duration histogram
- Request size
- Response size
- Active requests

**Endpoint:** `GET /metrics`

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JAEGER_ENDPOINT` | `http://localhost:14268/api/traces` | Jaeger collector URL |
| `ENVIRONMENT` | `development` | Environment name (dev/staging/prod) |
| `SERVICE_VERSION` | `1.0.0` | Service version for tracking |
| `ENABLE_METRICS` | `true` | Enable Prometheus metrics (order-food) |
| `PORT` | `8080` | Server port (order-food) |

### Docker Compose Configuration

```yaml
environment:
  - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
  - ENVIRONMENT=docker
  - SERVICE_VERSION=1.0.0
  - ENABLE_METRICS=true
```

### Kubernetes/Helm Configuration

Update Helm values:

```yaml
env:
  - name: JAEGER_ENDPOINT
    value: "http://jaeger-collector:14268/api/traces"
  - name: ENVIRONMENT
    value: "production"
  - name: SERVICE_VERSION
    value: "1.0.0"
  - name: ENABLE_METRICS
    value: "true"
```

## Local Development

### Run Jaeger Locally

```bash
# Start Jaeger all-in-one
docker run -d --name jaeger \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:latest
```

### Run Services Locally

```bash
# Terminal 1: database-migration
cd database-migration
JAEGER_ENDPOINT=http://localhost:14268/api/traces go run cmd/main.go

# Terminal 2: database-load
cd database-load
JAEGER_ENDPOINT=http://localhost:14268/api/traces go run cmd/main.go

# Terminal 3: order-food
cd order-food
JAEGER_ENDPOINT=http://localhost:14268/api/traces go run cmd/main.go
```

## Jaeger UI Guide

### Finding Traces

1. **Service:** Select which service to search
2. **Operation:** Filter by specific operation (optional)
3. **Tags:** Add filters like `http.status_code=200`
4. **Lookback:** Time range to search
5. **Min/Max Duration:** Filter by trace duration
6. **Limit Results:** Number of traces to return

### Trace View

**Timeline:**
- Shows all spans in chronological order
- Colored bars represent different services
- Length indicates duration

**Span Details:**
- Operation name
- Duration
- Tags (key-value pairs)
- Logs/Events
- Process information

**Service Dependencies:**
- Visual graph of service interactions
- Request counts
- Error rates
- Average latencies

## Trace Examples

### Successful Order Flow

```
order-food: HTTP POST /api/order [200ms]
  ├─ order-food: middleware.auth [2ms]
  ├─ order-food: handler.PlaceOrder [150ms]
  │   ├─ order-food: service.ValidateProducts [50ms]
  │   │   └─ order-food: repository.GetProducts [30ms]
  │   └─ order-food: service.CreateOrder [100ms]
  │       └─ order-food: repository.SaveOrder [80ms]
  └─ order-food: middleware.logger [1ms]
```

### Migration with Database Load

```
database-migration: migration.execute [250ms]
  ├─ migration.createTables [150ms]
  │   └─ Event: "Tables created successfully"
  └─ migration.createIndexes [100ms]
      └─ Event: "Indexes created successfully"

database-load: dataload.execute [500ms]
  ├─ dataload.loadProducts [200ms]
  │   └─ Event: "Loaded 100 products"
  ├─ dataload.loadUsers [150ms]
  │   └─ Event: "Loaded 50 users"
  └─ dataload.loadOrders [150ms]
      └─ Event: "Loaded 200 orders"
```

## Metrics Examples

### Prometheus Metrics Output

```
# HTTP request count
http_requests_total{method="GET",path="/api/product",status="200"} 42

# Request duration histogram
http_request_duration_seconds_bucket{le="0.1",method="GET",path="/api/product"} 38
http_request_duration_seconds_bucket{le="0.5",method="GET",path="/api/product"} 42
http_request_duration_seconds_sum{method="GET",path="/api/product"} 2.5
http_request_duration_seconds_count{method="GET",path="/api/product"} 42

# Response size
http_response_size_bytes_sum{method="GET",path="/api/product"} 8400
http_response_size_bytes_count{method="GET",path="/api/product"} 42
```

## Custom Instrumentation

### Adding Custom Spans

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func MyFunction(ctx context.Context) error {
    tracer := otel.Tracer("my-service")

    // Start a span
    ctx, span := tracer.Start(ctx, "my-operation")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("user.id", "123"),
        attribute.Int("item.count", 5),
    )

    // Add events
    span.AddEvent("processing started")

    // Do work...
    result, err := doWork(ctx)

    if err != nil {
        // Record errors
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.AddEvent("processing completed")
    return nil
}
```

### Adding Custom Metrics

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
)

func setupMetrics() {
    meter := otel.Meter("my-service")

    // Counter
    orderCounter, _ := meter.Int64Counter(
        "orders.placed",
        metric.WithDescription("Number of orders placed"),
    )

    // Histogram
    orderValue, _ := meter.Float64Histogram(
        "order.value",
        metric.WithDescription("Order value in dollars"),
        metric.WithUnit("USD"),
    )

    // Use them
    orderCounter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("payment.method", "credit_card"),
    ))

    orderValue.Record(ctx, 99.99, metric.WithAttributes(
        attribute.String("currency", "USD"),
    ))
}
```

## Best Practices

### 1. Span Naming

✅ **Good:**
- `database.query`
- `http.request`
- `service.processOrder`
- `repository.findUser`

❌ **Bad:**
- `function1`
- `process`
- `/api/users/123` (use attributes instead)

### 2. Attributes vs Events

**Use Attributes for:**
- Structured data
- Filtering in Jaeger UI
- HTTP status codes
- User IDs
- Error types

**Use Events for:**
- Significant moments in time
- Log-like messages
- Debugging information
- State changes

### 3. Sampling

**Development:**
```go
trace.WithSampler(trace.AlwaysSample())
```

**Production:**
```go
trace.WithSampler(trace.TraceIDRatioBased(0.1)) // 10% sampling
```

### 4. Error Handling

Always record errors in spans:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

## Troubleshooting

### Traces Not Appearing

**Check:**
1. Jaeger is running: `curl http://localhost:16686`
2. Service can reach Jaeger: `curl http://localhost:14268/api/traces`
3. Environment variable is set: `echo $JAEGER_ENDPOINT`
4. Check service logs for telemetry initialization
5. Verify no firewall blocking

**Debug:**
```bash
# Check Jaeger health
docker logs jaeger

# Check service logs
docker logs order-food
```

### Metrics Not Available

**Check:**
1. `ENABLE_METRICS=true` is set
2. Metrics endpoint is accessible: `curl http://localhost:8080/metrics`
3. Prometheus exporter initialized (check logs)

### High Overhead

**Solutions:**
1. Reduce sampling rate in production
2. Limit attribute sizes
3. Use batch export
4. Adjust buffer sizes

## Production Deployment

### With Jaeger Operator (Kubernetes)

```bash
# Install Jaeger Operator
kubectl create namespace observability
kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.51.0/jaeger-operator.yaml

# Deploy Jaeger instance
kubectl apply -f - <<EOF
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: jaeger
  namespace: observability
spec:
  strategy: production
  storage:
    type: elasticsearch
EOF
```

### With External Jaeger

Update Helm values:

```yaml
env:
  - name: JAEGER_ENDPOINT
    value: "https://jaeger.example.com/api/traces"
```

### With OpenTelemetry Collector

For centralized telemetry processing:

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "http://otel-collector:4317"
```

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [otelgin Middleware](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin)

## Support

For issues:
1. Check service logs
2. Verify Jaeger connectivity
3. Review environment variables
4. Check this documentation
5. Create issue with trace IDs
