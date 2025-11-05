# Hybrid Architecture: Init Containers + CronJob

## Overview

The **database-load** module now operates in a **hybrid mode**:

1. **Init Container** (in order-food pod) - For initial data loading
   - Runs once when pod starts
   - **MUST succeed** for pod to become Ready
   - Ensures data exists before application serves traffic

2. **CronJob** (separate resource) - For periodic data refresh
   - Runs on schedule (default: every 6 hours)
   - **Failures do NOT impact** running order-food pods
   - Keeps data fresh without service interruption

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                          │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │                  order-food Pod                           │   │
│  │                                                           │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │ Init #1: database-migration (runs first)        │    │   │
│  │  │ - Runs schema migrations                        │    │   │
│  │  │ - MUST succeed for pod to start                 │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  │                       ↓                                   │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │ Init #2: database-load (runs second)            │    │   │
│  │  │ - Loads initial data                            │    │   │
│  │  │ - MUST succeed for pod to start                 │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  │                       ↓                                   │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │ Main Container: order-food                      │    │   │
│  │  │ - REST API server                               │    │   │
│  │  │ - Starts after all inits succeed                │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  └───────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │              database-load CronJob                        │   │
│  │                                                           │   │
│  │  Schedule: "0 */6 * * *" (every 6 hours)                │   │
│  │  Purpose:  Periodic data refresh                         │   │
│  │  Impact:   Failures don't affect running pods            │   │
│  │                                                           │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │ Job Created Every 6 Hours                       │    │   │
│  │  │ - Runs database-load container                  │    │   │
│  │  │ - Updates data in database                      │    │   │
│  │  │ - order-food continues serving during refresh   │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  └───────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Key Benefits

### ✅ Initial Data Guarantee
- **Init container ensures data exists** before application starts
- No race condition on first deployment
- Application never starts with empty database

### ✅ Isolation of Concerns
- **Init container**: Critical initial setup (must succeed)
- **CronJob**: Non-critical periodic refresh (can fail)
- Failures separated by scope

### ✅ Zero Downtime Data Refresh
- CronJob runs independently
- order-food keeps serving traffic during refresh
- No pod restarts required

### ✅ Resilience
- CronJob failures logged but don't trigger alerts
- Application remains available even if refresh fails
- Can retry on next schedule

## Configuration

### CronJob Schedule

Edit [database-load/helm/values.yaml](database-load/helm/values.yaml:63-72):

```yaml
cronjob:
  enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours
  # Other common schedules:
  # "0 * * * *"     - Every hour
  # "*/30 * * * *"  - Every 30 minutes
  # "0 0 * * *"     - Daily at midnight
  # "0 2 * * 0"     - Weekly on Sunday at 2 AM
  concurrencyPolicy: Forbid  # Don't run concurrent jobs
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  backoffLimit: 2
  restartPolicy: OnFailure
  suspend: false  # Set to true to temporarily disable
```

### Disable CronJob (Use Init Container Only)

```bash
helm upgrade --install database-load ./database-load/helm \
  --set cronjob.enabled=false
```

### Disable Init Container (Use CronJob Only)

```bash
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseLoad.enabled=false
```

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

## Monitoring

### Check CronJob Status

```bash
# View CronJob
kubectl get cronjobs -n default

# Expected output:
# NAME            SCHEDULE        SUSPEND   ACTIVE   LAST SCHEDULE   AGE
# database-load   0 */6 * * *     False     0        2h              1d
```

### View CronJob History

```bash
# List Jobs created by CronJob
kubectl get jobs -n default -l app.kubernetes.io/name=database-load

# View logs from latest Job
kubectl logs -n default -l app.kubernetes.io/name=database-load --tail=100
```

### Manually Trigger CronJob

```bash
# Create a one-off Job from the CronJob
kubectl create job --from=cronjob/database-load manual-load-$(date +%s) -n default

# Watch the Job
kubectl get jobs -n default -w

# View logs
kubectl logs -n default job/manual-load-<timestamp>
```

### View Init Container Logs

```bash
# View database-load init container logs (from pod startup)
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load
```

## Execution Timeline

### Initial Deployment

```
Time: T+0
1. Deploy database-load CronJob
   - CronJob created
   - Waits for schedule (no immediate execution)

2. Deploy order-food
   ↓
   Init Container #1: database-migration runs
   - Migrates database schema
   - Exits with code 0
   ↓
   Init Container #2: database-load runs
   - Loads initial data
   - Exits with code 0
   ↓
   Main Container: order-food starts
   - Application is Ready
   - Begins serving traffic
```

### Periodic Refresh

```
Time: T+6h (first scheduled run)
1. CronJob triggers
   - Creates Job: database-load-<timestamp>
   - Runs database-load container
   - Updates data in database

2. order-food continues running
   - No interruption
   - Serves traffic during refresh
   - Reads fresh data after refresh completes
```

### Failure Scenarios

#### Scenario 1: Init Container Fails
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

**Impact**: Application won't start (CRITICAL)

#### Scenario 2: CronJob Fails
```
CronJob execution at T+6h FAILS
   ↓
Job marked as Failed
   ↓
order-food continues running
   ↓
Next CronJob execution at T+12h
```

**Impact**: No impact on running application (NON-CRITICAL)

## Comparison Matrix

| Aspect | Init Container | CronJob |
|--------|---------------|---------|
| **When** | Pod startup | On schedule |
| **Frequency** | Once per pod | Every N hours |
| **Failure Impact** | Pod won't start | No impact |
| **Criticality** | CRITICAL | NON-CRITICAL |
| **Purpose** | Initial setup | Periodic refresh |
| **Resource** | Part of pod | Separate Job |
| **Logs Location** | Pod init container | Job logs |
| **Retries** | Pod restart | Job backoff |

## Use Cases

### Use Init Container When:
- ✅ Data MUST exist before application starts
- ✅ First-time setup
- ✅ Schema-dependent data
- ✅ Critical for application to function

### Use CronJob When:
- ✅ Data needs periodic refresh
- ✅ Failures acceptable (stale data okay)
- ✅ Long-running refresh operations
- ✅ Independent of application lifecycle

### Use Both (This Architecture) When:
- ✅ Need guaranteed initial data
- ✅ Want periodic updates
- ✅ Refresh failures shouldn't stop app
- ✅ Zero-downtime updates required

## Troubleshooting

### CronJob Not Running

```bash
# Check if CronJob is suspended
kubectl get cronjob database-load -n default -o yaml | grep suspend

# If suspended, enable it
kubectl patch cronjob database-load -n default -p '{"spec":{"suspend":false}}'
```

### CronJob Running Too Often

```bash
# Check concurrency policy
kubectl get cronjob database-load -n default -o yaml | grep concurrencyPolicy

# Should be "Forbid" to prevent concurrent runs
```

### View Failed Jobs

```bash
# List failed Jobs
kubectl get jobs -n default -l app.kubernetes.io/name=database-load | grep -v Complete

# View logs from failed Job
kubectl logs -n default job/<job-name>
```

### Delete Old Jobs

```bash
# Delete completed Jobs older than 1 hour
kubectl delete jobs -n default -l app.kubernetes.io/name=database-load \
  --field-selector status.successful=1

# Delete failed Jobs
kubectl delete jobs -n default -l app.kubernetes.io/name=database-load \
  --field-selector status.successful=0
```

## Tuning Recommendations

### Production Schedule

```yaml
# Conservative: Every 12 hours
schedule: "0 */12 * * *"

# Moderate: Every 6 hours (default)
schedule: "0 */6 * * *"

# Aggressive: Every hour
schedule: "0 * * * *"
```

### Resource Limits

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

### Backoff Limit

```yaml
# Retry failed jobs
backoffLimit: 3  # Retry up to 3 times
```

### History Limits

```yaml
# Keep more history for debugging
successfulJobsHistoryLimit: 5
failedJobsHistoryLimit: 5
```

## Best Practices

### 1. Idempotent Data Loading
```go
// Good: Upsert operation
INSERT INTO products VALUES (...) ON CONFLICT (id) DO UPDATE ...

// Bad: Insert that fails on duplicate
INSERT INTO products VALUES (...)  // Fails if already exists
```

### 2. Monitoring
```bash
# Set up alerts for consecutive failures
FAILURES=$(kubectl get jobs -n default -l app.kubernetes.io/name=database-load \
  --field-selector status.successful=0 | wc -l)

if [ $FAILURES -gt 3 ]; then
  echo "ALERT: Database load has failed $FAILURES times"
fi
```

### 3. Testing
```bash
# Test CronJob before scheduling
kubectl create job --from=cronjob/database-load test-load -n default
kubectl wait --for=condition=complete job/test-load -n default --timeout=5m
kubectl logs job/test-load -n default
kubectl delete job test-load -n default
```

### 4. Graceful Degradation
- Application should handle stale data gracefully
- Consider adding data freshness indicators
- Cache invalidation on refresh

## Migration from Previous Architecture

### Before (Init Container Only)
```yaml
# database-load ONLY as init container
initContainers:
  databaseLoad:
    enabled: true
```

### After (Hybrid)
```yaml
# database-load as BOTH init container and CronJob

# In order-food/helm/values.yaml
initContainers:
  databaseLoad:
    enabled: true  # Still needed for initial load

# In database-load/helm/values.yaml
cronjob:
  enabled: true  # NEW: Periodic refresh
  schedule: "0 */6 * * *"
```

## Summary

This hybrid architecture provides:
- ✅ **Guaranteed initial data** via init container
- ✅ **Periodic refresh** via CronJob
- ✅ **Failure isolation** - CronJob failures don't impact app
- ✅ **Zero downtime** - Application stays running during refresh
- ✅ **Flexibility** - Can disable either component
- ✅ **Observability** - Separate logs for each execution

**Best of both worlds**: Critical initial setup + non-critical periodic updates.
