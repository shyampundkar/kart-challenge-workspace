# Architecture Documentation

Complete architecture documentation for the kart-challenge-workspace microservices project.

## Table of Contents
1. [System Overview](#system-overview)
2. [Hybrid Architecture](#hybrid-architecture)
3. [Init Container Architecture](#init-container-architecture)
4. [API Implementation](#api-implementation)
5. [Observability](#observability)
6. [Deployment](#deployment)
7. [Best Practices](#best-practices)

---

## System Overview

The kart-challenge-workspace project implements a microservices-based food ordering system with three main components:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Deployment Architecture                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Helm Release #1: database-load (CronJob)                           │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │ CronJob: database-load                                     │   │
│  │ Schedule: Every 6 hours                                    │   │
│  │ Purpose: Periodic data refresh                             │   │
│  │ Failure Impact: NONE - App continues running               │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  Helm Release #2: order-food (Deployment with Init Containers)      │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │ Pod: order-food                                            │   │
│  │                                                            │   │
│  │  Init Container #1: database-migration                     │   │
│  │  - Runs database schema migrations                         │   │
│  │  - CRITICAL: Must succeed for pod to start                 │   │
│  │                                                            │   │
│  │  Init Container #2: database-load                          │   │
│  │  - Loads initial data                                      │   │
│  │  - CRITICAL: Must succeed for pod to start                 │   │
│  │                                                            │   │
│  │  Main Container: order-food                                │   │
│  │  - REST API server                                         │   │
│  │  - Starts after all init containers succeed                │   │
│  └────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Type | Purpose | Criticality |
|-----------|------|---------|-------------|
| **database-migration** | Init Container | Schema migrations | CRITICAL |
| **database-load** | Init Container | Initial data load | CRITICAL |
| **database-load** | CronJob | Periodic refresh | NON-CRITICAL |
| **order-food** | Deployment | API Server | - |

---

## Hybrid Architecture

The system uses a **hybrid architecture** combining init containers and CronJob for optimal reliability and data freshness.

### Key Benefits

#### ✅ Initial Data Guarantee
- Init container ensures data exists before application starts
- No race condition on first deployment
- Application never starts with empty database

#### ✅ Isolation of Concerns
- **Init container**: Critical initial setup (must succeed)
- **CronJob**: Non-critical periodic refresh (can fail)
- Failures separated by scope

#### ✅ Zero Downtime Data Refresh
- CronJob runs independently
- order-food keeps serving traffic during refresh
- No pod restarts required

#### ✅ Resilience
- CronJob failures logged but don't trigger alerts
- Application remains available even if refresh fails
- Can retry on next schedule

### Execution Flow

#### Initial Deployment (T=0)
```
Time    Event                              Impact
────────────────────────────────────────────────────────
T+0s    Deploy database-load CronJob       CronJob created, scheduled
T+0s    Deploy order-food                  Pod creation starts
T+1s    Init: database-migration starts    Pod: Init:0/2
T+3s    Init: database-migration complete  Pod: Init:1/2
T+3s    Init: database-load starts         Pod: Init:1/2
T+5s    Init: database-load complete       Pod: Init:2/2
T+5s    Main: order-food starts            Pod: Running
T+7s    Readiness probe succeeds           Pod: 1/1 Ready
T+7s    Service starts routing traffic     App available ✓
```

#### Periodic Refresh (T=6h)
```
Time    Event                              Impact on order-food
──────────────────────────────────────────────────────────────
T+6h    CronJob triggers                   No impact - keeps running
T+6h    Job: database-load-<ts> starts     No impact - keeps running
T+6h+2s Job updates database               No impact - keeps running
T+6h+3s Job completes successfully         No impact - uses fresh data
```

### Failure Scenarios

#### Scenario 1: Init Container Failure (CRITICAL)
```
Init Container #2: database-load FAILS
   ↓
Pod does NOT start
   ↓
Kubernetes retries (creates new pod)
   ↓
Init containers run again
   ↓
Application starts only if init succeeds
```
**Impact**: Application unavailable until fixed

#### Scenario 2: CronJob Failure (NON-CRITICAL)
```
CronJob execution at T+6h FAILS
   ↓
Job marked as Failed
   ↓
order-food continues running
   ↓
Next CronJob execution at T+12h
```
**Impact**: No impact on running application

### Configuration

#### CronJob Schedule
```yaml
# database-load/helm/values.yaml
cronjob:
  enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  backoffLimit: 2
  restartPolicy: OnFailure
  suspend: false
```

#### Init Containers
```yaml
# order-food/helm/values.yaml
initContainers:
  databaseMigration:
    enabled: true
    image:
      repository: database-migration
      tag: "latest"

  databaseLoad:
    enabled: true
    image:
      repository: database-load
      tag: "latest"
```

---

## Init Container Architecture

Both `database-migration` and `database-load` modules run as **init containers** within the `order-food` pod.

### Architecture Diagram

```
┌───────────────────────────────────────────────────────────────┐
│                    order-food Pod                             │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Init Container #1: database-migration               │    │
│  │  (runs first, must succeed)                          │    │
│  └──────────────────────────────────────────────────────┘    │
│                           ↓                                    │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Init Container #2: database-load                    │    │
│  │  (runs second, must succeed)                         │    │
│  └──────────────────────────────────────────────────────┘    │
│                           ↓                                    │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Main Container: order-food                          │    │
│  │  (starts after all inits succeed)                    │    │
│  └──────────────────────────────────────────────────────┘    │
└───────────────────────────────────────────────────────────────┘
```

### Benefits

1. **Atomic Deployment** - All setup steps and application deployment are tightly coupled
2. **Simplified Operations** - Only need to deploy 1 Helm release instead of 3
3. **Better Resource Utilization** - Init containers terminate after completion
4. **Improved Reliability** - Migration must succeed before application starts

### Viewing Init Container Logs

```bash
# View database-migration init container logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration

# View database-load init container logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load

# View specific pod's init container
kubectl logs -n default <pod-name> -c database-migration
```

### Disabling Init Containers

```bash
# Disable database-migration
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false

# Disable database-load
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseLoad.enabled=false
```

---

## API Implementation

The Order Food API is implemented using Go 1.25.4 and the Gin web framework.

### Application Architecture

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

### Components

#### Models (internal/models/)
- **Product** - Product information (id, name, price, category)
- **OrderItem** - Individual item in an order
- **OrderReq** - Order creation request
- **Order** - Complete order with details
- **APIResponse** - Standard API response format

#### Repositories (internal/repository/)
- **ProductRepository** - Thread-safe product data management
- **OrderRepository** - Thread-safe order data management

#### Services (internal/service/)
- **ProductService** - Product-related operations
- **OrderService** - Order-related operations with validation

#### Handlers (internal/handler/)
- **ProductHandler** - Product endpoints
- **OrderHandler** - Order endpoints
- **HealthHandler** - Health check endpoints

#### Middleware (internal/middleware/)
- **AuthMiddleware** - API key authentication
- **CORSMiddleware** - Cross-Origin Resource Sharing
- **LoggerMiddleware** - Request logging

### API Endpoints

#### GET /api/product
- Returns all products
- No authentication required
- Status: 200 OK

#### GET /api/product/:productId
- Returns specific product
- No authentication required
- Status: 200 OK, 400 Bad Request, 404 Not Found

#### POST /api/order
- Creates new order
- Requires "api_key: apitest" header
- Validates all product IDs exist
- Generates UUID for order
- Status: 200 OK, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 422 Unprocessable Entity

### Authentication

API key authentication is implemented for order creation:
- Header: `api_key: apitest`
- Invalid/missing key returns appropriate error response
- Only applied to order creation endpoint

### Sample Data

The system is pre-seeded with 10 products:

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

### Example Requests

#### List Products
```bash
curl http://localhost:8080/api/product
```

#### Get Product
```bash
curl http://localhost:8080/api/product/1
```

#### Place Order
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

---

## Observability

All modules are instrumented with OpenTelemetry for distributed tracing and metrics.

### Observability Stack

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

### Components

| Module | Tracing | Metrics | Auto-Instrumentation |
|--------|---------|---------|----------------------|
| **database-migration** | ✅ | ❌ | Manual spans |
| **database-load** | ✅ | ❌ | Manual spans |
| **order-food** | ✅ | ✅ | Automatic HTTP tracing |

### Tracing Implementation

#### database-migration
**Spans:**
- `migration.execute` - Root span for entire migration
- `migration.createTables` - Table creation operations
- `migration.createIndexes` - Index creation operations

#### database-load
**Spans:**
- `dataload.execute` - Root span for entire load operation
- `dataload.loadProducts` - Product data loading
- `dataload.loadUsers` - User data loading
- `dataload.loadOrders` - Order data loading

#### order-food
**Automatic Tracing:**
- HTTP requests (method, path, status code)
- Request/response headers
- Duration and errors
- Middleware: `otelgin.Middleware("order-food")`

### Metrics (order-food only)

**Prometheus Metrics:**
- HTTP request count
- Request duration histogram
- Request/response size
- Active requests

**Endpoint:** `GET /metrics`

### Configuration

Environment variables:
- `JAEGER_ENDPOINT` - Jaeger collector URL (default: `http://localhost:14268/api/traces`)
- `ENVIRONMENT` - Environment name (dev/staging/prod)
- `SERVICE_VERSION` - Service version for tracking
- `ENABLE_METRICS` - Enable Prometheus metrics (order-food)

### Viewing Traces

1. Access Jaeger UI: http://localhost:16686
2. Select service: `database-migration`, `database-load`, or `order-food`
3. Click "Find Traces"
4. Click on a trace to see detailed view

### Custom Instrumentation

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
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    return nil
}
```

---

## Deployment

### Quick Deploy

```bash
./deploy.sh
```

This will:
1. Deploy database-load CronJob
2. Deploy order-food with both init containers

### Manual Deployment

```bash
# Build images
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food

# Deploy database-load CronJob
helm upgrade --install database-load ./database-load/helm \
  --set image.pullPolicy=Never \
  --set job.enabled=false \
  --set cronjob.enabled=true

# Deploy order-food (with init containers)
helm upgrade --install order-food ./order-food/helm \
  --set image.pullPolicy=Never \
  --set initContainers.databaseMigration.image.pullPolicy=Never \
  --set initContainers.databaseLoad.image.pullPolicy=Never
```

### Monitoring

#### Check CronJob Status
```bash
kubectl get cronjobs -n default
kubectl get jobs -n default -l app.kubernetes.io/name=database-load
kubectl logs -n default -l app.kubernetes.io/name=database-load --tail=100
```

#### Manually Trigger CronJob
```bash
kubectl create job --from=cronjob/database-load manual-load-$(date +%s) -n default
```

#### View Pod Status
```bash
kubectl get pods -n default -l app.kubernetes.io/name=order-food
kubectl describe pod -n default <pod-name>
```

### Cleanup

```bash
./cleanup.sh
# Or manually:
helm uninstall database-load order-food -n default
```

---

## Best Practices

### 1. Deployment
- Use versioned images in production (not `latest`)
- Set appropriate resource limits
- Configure liveness and readiness probes
- Use rolling updates for zero downtime

### 2. Observability
- Always record errors in spans
- Use meaningful span names
- Add relevant attributes for filtering
- Monitor init container execution time

### 3. Data Management
- Keep migrations fast (< 1 minute)
- Ensure migrations are idempotent
- Use backward-compatible schema changes
- Adjust CronJob schedule based on data size

### 4. Resource Limits

```yaml
# For large datasets
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

### 5. Monitoring & Alerting

```yaml
# Alert on init container failures
- alert: InitContainerFailed
  expr: kube_pod_init_container_status_restarts_total > 3
  severity: critical

# Alert on consecutive CronJob failures
- alert: CronJobConsecutiveFailures
  expr: job_failures{job="database-load"} > 3
  severity: warning
```

---

## Summary

The kart-challenge-workspace architecture provides:

- ✅ **Hybrid approach** - Init containers + CronJob for optimal reliability
- ✅ **Guaranteed initial data** - Via init containers
- ✅ **Periodic refresh** - Via CronJob
- ✅ **Failure isolation** - CronJob failures don't impact app
- ✅ **Zero downtime** - Application stays running during refresh
- ✅ **Full observability** - Distributed tracing and metrics
- ✅ **Production-ready** - Kubernetes-native with Helm charts
- ✅ **Scalable** - Microservices architecture with clean separation

**Status**: ✅ Complete and ready for deployment
**Default CronJob Schedule**: Every 6 hours
