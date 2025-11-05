# Complete Init Container Architecture

## Summary

**ALL** setup modules (`database-migration` and `database-load`) now run as **init containers** within the `order-food` pod, creating a fully atomic deployment.

## Architecture

```
┌───────────────────────────────────────────────────────────────────────────┐
│                           order-food Pod                                  │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Init Container #1: database-migration                          │   │
│  │  - Runs database schema migrations                              │   │
│  │  - Must complete successfully (Exit 0)                          │   │
│  │  - Instrumented with OpenTelemetry                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                  ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Init Container #2: database-load                               │   │
│  │  - Loads initial data into database                             │   │
│  │  - Runs ONLY after migration succeeds                           │   │
│  │  - Must complete successfully (Exit 0)                          │   │
│  │  - Instrumented with OpenTelemetry                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                  ↓                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Main Container: order-food                                     │   │
│  │  - REST API server                                              │   │
│  │  - Starts ONLY after all init containers succeed               │   │
│  │  - Exposed on port 8080                                         │   │
│  │  - Instrumented with OpenTelemetry + Prometheus metrics         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

## Execution Flow

```
1. Pod Created
   └─> Kubernetes scheduler assigns pod to node

2. Init Container #1 (database-migration) starts
   ├─> Connects to database
   ├─> Runs schema migrations
   ├─> Sends traces to Jaeger
   └─> Exits with code 0 (success)

3. Init Container #2 (database-load) starts
   ├─> Waits for migration to complete
   ├─> Connects to database
   ├─> Loads initial data (products, users, orders)
   ├─> Sends traces to Jaeger
   └─> Exits with code 0 (success)

4. Main Container (order-food) starts
   ├─> Waits for all init containers to complete
   ├─> Starts HTTP server on port 8080
   ├─> Exposes /health, /ready, /metrics endpoints
   ├─> Sends traces to Jaeger
   ├─> Exposes Prometheus metrics
   └─> Pod becomes Ready (1/1)

5. Service routes traffic to pod
```

## Key Benefits

### ✅ Single Deployment Unit
- **Before**: 3 separate Helm releases (database-migration, database-load, order-food)
- **After**: 1 Helm release (order-food with 2 init containers)
- **Result**: Simplified deployment and management

### ✅ Guaranteed Order
- Init containers run **sequentially** in the order defined
- Migration ALWAYS runs before data load
- Data load ALWAYS runs before application
- No race conditions possible

### ✅ Atomic Deployment
- All setup and application in a single pod
- Pod won't be Ready until everything succeeds
- Rollback is automatic if any step fails

### ✅ Automatic Retry
- If any init container fails, Kubernetes restarts the pod
- All init containers run again from the beginning
- Ensures idempotent operations

### ✅ Clean Resource Model
- Init containers terminate after completion
- No orphaned Jobs to clean up
- Logs accessible via pod container selector

## Configuration

### Helm Values ([order-food/helm/values.yaml](order-food/helm/values.yaml:103-144))

```yaml
initContainers:
  databaseMigration:
    enabled: true
    image:
      repository: database-migration
      tag: "latest"
      pullPolicy: IfNotPresent
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
      - name: ENVIRONMENT
        value: "kubernetes"
      - name: SERVICE_VERSION
        value: "1.0.0"
    resources:
      limits:
        cpu: 500m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi

  databaseLoad:
    enabled: true
    image:
      repository: database-load
      tag: "latest"
      pullPolicy: IfNotPresent
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
      - name: ENVIRONMENT
        value: "kubernetes"
      - name: SERVICE_VERSION
        value: "1.0.0"
    resources:
      limits:
        cpu: 500m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi
```

## Deployment

### Quick Deploy
```bash
./deploy.sh
```

### Manual Deploy
```bash
# Build images
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food

# Deploy
helm upgrade --install order-food ./order-food/helm \
  --set image.pullPolicy=Never \
  --set initContainers.databaseMigration.image.pullPolicy=Never \
  --set initContainers.databaseLoad.image.pullPolicy=Never \
  --wait \
  --timeout 10m
```

## Monitoring

### Check Pod Status
```bash
kubectl get pods -n default -l app.kubernetes.io/name=order-food
```

**Example output:**
```
NAME                          READY   STATUS    RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    1/1     Running   0          2m
```

**Status meanings:**
- `Init:0/2` - Running first init container (database-migration)
- `Init:1/2` - Running second init container (database-load)
- `Running` - All init containers complete, main container running
- `1/1` - Pod is Ready

### View Init Container Logs

```bash
# Migration logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration

# Data load logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load

# Application logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food

# All logs from specific pod
POD=$(kubectl get pods -n default -l app.kubernetes.io/name=order-food -o jsonpath='{.items[0].metadata.name}')
kubectl logs -n default $POD -c database-migration
kubectl logs -n default $POD -c database-load
kubectl logs -n default $POD -c order-food
```

### View Init Container Status
```bash
kubectl describe pod -n default -l app.kubernetes.io/name=order-food
```

**Look for:**
```
Init Containers:
  database-migration:
    State:          Terminated
      Reason:       Completed
      Exit Code:    0
  database-load:
    State:          Terminated
      Reason:       Completed
      Exit Code:    0

Containers:
  order-food:
    State:          Running
      Started:      <timestamp>
```

## Troubleshooting

### Init Container Failed

**Symptoms:**
```bash
NAME                          READY   STATUS                  RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:Error              0          1m
```

**Diagnosis:**
```bash
# Check which init container failed
kubectl describe pod -n default order-food-7d8f9b5c4-xyz12

# View logs from failed container
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration
# or
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-load
```

### Init Container Stuck

**Symptoms:**
```bash
NAME                          READY   STATUS        RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:1/2      0          5m
```

**Diagnosis:**
```bash
# Check which container is running
kubectl describe pod -n default order-food-7d8f9b5c4-xyz12

# Watch logs in real-time
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-load -f
```

### Init Container CrashLoopBackOff

**Symptoms:**
```bash
NAME                          READY   STATUS                   RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:CrashLoopBackOff    3          2m
```

**Diagnosis:**
```bash
# View logs from previous attempt
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration --previous

# Check events
kubectl get events -n default --sort-by='.lastTimestamp' | grep order-food
```

## Observability

### Jaeger Traces

All three containers send traces to Jaeger:

1. **Access Jaeger UI:**
   ```bash
   # Port-forward Jaeger (if running in cluster)
   kubectl port-forward -n default svc/jaeger 16686:16686
   open http://localhost:16686

   # Or with Docker Compose
   open http://localhost:16686
   ```

2. **View traces by service:**
   - `database-migration` - Migration operations
   - `database-load` - Data loading operations
   - `order-food` - HTTP API requests

3. **Trace timeline shows sequential execution:**
   ```
   database-migration: migration.execute [250ms]
       ↓
   database-load: dataload.execute [500ms]
       ↓
   order-food: HTTP GET /api/product [50ms]
   ```

### Prometheus Metrics

Only the main container exposes metrics:

```bash
# Port-forward to access metrics
kubectl port-forward -n default svc/order-food 8080:80

# View metrics
curl http://localhost:8080/metrics
```

## Disabling Init Containers

If needed (e.g., for development):

```bash
# Disable migration only
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false

# Disable data load only
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseLoad.enabled=false

# Disable both
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false \
  --set initContainers.databaseLoad.enabled=false
```

## Comparison: Before vs After

| Aspect | Before | After |
|--------|--------|-------|
| **Helm Releases** | 3 (migration, load, app) | 1 (app with inits) |
| **Resources** | 2 Jobs + 1 Deployment | 1 Deployment |
| **Execution** | Sequential via dependencies | Sequential via init containers |
| **Cleanup** | Manual Job cleanup | Automatic (init terminates) |
| **Logs** | 3 separate resources | 3 containers in 1 pod |
| **Deployment Command** | 3 helm install commands | 1 helm install command |
| **Uninstall Command** | `helm uninstall migration load app` | `helm uninstall app` |
| **Failure Handling** | Job fails, manual intervention | Pod restarts automatically |
| **Resource Usage** | Jobs persist until deleted | Init containers terminate |
| **Coupling** | Loose (separate resources) | Tight (single pod) |

## Best Practices

### 1. Keep Init Containers Fast
- Migrations should complete in < 2 minutes
- Data loading should complete in < 5 minutes
- Consider timeout increase for large datasets

### 2. Make Operations Idempotent
```go
// Good: Idempotent migration
CREATE TABLE IF NOT EXISTS users (...);

// Bad: Non-idempotent migration
CREATE TABLE users (...);  // Fails on retry
```

### 3. Resource Limits
- Set appropriate CPU/memory limits
- Init containers can use more resources than main container
- Consider burst requirements

### 4. Error Handling
```go
// Always fail fast with clear errors
if err != nil {
    log.Fatalf("Migration failed: %v", err)  // Exit code 1
    return
}
```

### 5. Logging
- Log each step clearly
- Include timestamps
- Send structured logs for parsing

## Docker Compose

In Docker Compose, we still use `depends_on` for ordering:

```yaml
order-food:
  depends_on:
    - jaeger
    - database-migration  # Runs first
    - database-load       # Runs second
```

Note: Docker Compose doesn't have init containers, so we use separate services with dependencies.

## Production Considerations

### 1. Image Versioning
```yaml
initContainers:
  databaseMigration:
    image:
      tag: "v1.2.3"  # Use specific versions in production
```

### 2. Resource Limits
```yaml
initContainers:
  databaseMigration:
    resources:
      limits:
        cpu: 1000m
        memory: 512Mi
```

### 3. Timeout Configuration
```bash
# Increase timeout for large migrations
helm upgrade --install order-food ./order-food/helm \
  --wait \
  --timeout 20m
```

### 4. Monitoring
- Set up alerts for init container failures
- Monitor init container duration
- Track success/failure rates

### 5. Rolling Updates
- New pods run migrations before starting
- Ensure migrations are backward-compatible
- Old pods remain running until new pods are ready

## Next Steps

1. ✅ **Complete**: Both init containers configured
2. 📝 **Documentation**: Comprehensive guides created
3. 🧪 **Testing**: Test deployment with `./deploy.sh`
4. 📊 **Monitoring**: Verify traces in Jaeger
5. 🔄 **CI/CD**: Update pipelines to build all three images
6. 🏷️ **Versioning**: Use specific image tags in production
7. 🔐 **Secrets**: Move sensitive config to Kubernetes Secrets
8. 📈 **Metrics**: Add custom metrics for init container duration

## References

- [INIT_CONTAINER_ARCHITECTURE.md](INIT_CONTAINER_ARCHITECTURE.md) - Detailed architecture
- [QUICK_START_INIT_CONTAINER.md](QUICK_START_INIT_CONTAINER.md) - Quick reference
- [DEPLOYMENT.md](DEPLOYMENT.md) - Full deployment guide
- [OBSERVABILITY.md](OBSERVABILITY.md) - Observability setup
- [Kubernetes Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)

## Summary

This architecture provides:
- ✅ **Single deployment unit** - All components in one Helm release
- ✅ **Guaranteed ordering** - Migration → Data Load → Application
- ✅ **Atomic deployment** - All or nothing
- ✅ **Automatic retry** - Kubernetes handles failures
- ✅ **Clean resource model** - No orphaned resources
- ✅ **Full observability** - Traces and logs for all steps
- ✅ **Production-ready** - Configurable, scalable, monitorable

**Result**: A robust, maintainable, and observable microservice deployment architecture.
