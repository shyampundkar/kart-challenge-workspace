# CronJob + Init Container Summary

## What Changed

Added **database-load CronJob** for periodic data refresh while keeping the init container for initial data loading.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  order-food Pod (Initial Deployment)                   │
│  ┌────────────────────────────────────────────────┐   │
│  │ Init #1: database-migration                    │   │
│  └────────────────────────────────────────────────┘   │
│                       ↓                                 │
│  ┌────────────────────────────────────────────────┐   │
│  │ Init #2: database-load ← MUST SUCCEED          │   │
│  └────────────────────────────────────────────────┘   │
│                       ↓                                 │
│  ┌────────────────────────────────────────────────┐   │
│  │ Main: order-food                               │   │
│  └────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  database-load CronJob (Periodic Refresh)               │
│  Schedule: Every 6 hours                                │
│  Failures: DO NOT impact order-food ✓                   │
│  ┌────────────────────────────────────────────────┐   │
│  │ Job runs independently                         │   │
│  │ Updates data while app runs                    │   │
│  └────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Files Created/Modified

### New Files
1. **[database-load/helm/templates/cronjob.yaml](database-load/helm/templates/cronjob.yaml)** - CronJob template

### Modified Files
1. **[database-load/helm/values.yaml](database-load/helm/values.yaml:63-72)** - Added cronjob configuration
2. **[deploy.sh](deploy.sh:162-171)** - Deploy CronJob before order-food
3. **[cleanup.sh](cleanup.sh:48-55)** - Cleanup CronJob
4. **[deploy.sh](deploy.sh:223-232)** - Updated display info with CronJob commands

## Key Features

### Init Container (database-load)
- **Purpose**: Initial data loading
- **When**: Pod startup
- **Failure Impact**: **CRITICAL** - Pod won't start
- **Use Case**: First-time setup

### CronJob (database-load)
- **Purpose**: Periodic data refresh
- **When**: Every 6 hours (configurable)
- **Failure Impact**: **NONE** - App continues running
- **Use Case**: Keep data fresh

## Configuration

### Default Schedule (Every 6 Hours)
```yaml
cronjob:
  schedule: "0 */6 * * *"
```

### Common Schedules
```yaml
# Every hour
schedule: "0 * * * *"

# Every 30 minutes
schedule: "*/30 * * * *"

# Daily at 2 AM
schedule: "0 2 * * *"

# Weekly on Sunday at midnight
schedule: "0 0 * * 0"
```

### Disable CronJob
```yaml
cronjob:
  enabled: false
```

## Deployment

### Quick Deploy
```bash
./deploy.sh
```

### Manual Deploy
```bash
# Deploy CronJob
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

### Check CronJob
```bash
# View CronJob
kubectl get cronjobs -n default

# View Job history
kubectl get jobs -n default -l app.kubernetes.io/name=database-load

# View logs
kubectl logs -n default -l app.kubernetes.io/name=database-load
```

### Manually Trigger
```bash
kubectl create job --from=cronjob/database-load manual-load-$(date +%s) -n default
```

### Suspend/Resume
```bash
# Suspend (stop scheduling)
kubectl patch cronjob database-load -n default -p '{"spec":{"suspend":true}}'

# Resume
kubectl patch cronjob database-load -n default -p '{"spec":{"suspend":false}}'
```

## Benefits

| Feature | Init Container | CronJob |
|---------|---------------|---------|
| **Initial Data** | ✅ Guaranteed | ❌ |
| **Periodic Refresh** | ❌ | ✅ Scheduled |
| **Failure Impact** | 🔴 Critical | 🟢 None |
| **Zero Downtime** | ❌ Blocks startup | ✅ Independent |

## Use Cases

### When Init Container Runs
1. **New deployment** - Fresh pod creation
2. **Pod restart** - After crash or eviction
3. **Scaling up** - New replica added
4. **Rolling update** - New version deployed

### When CronJob Runs
1. **Scheduled time** - Every 6 hours (default)
2. **Manual trigger** - When explicitly created
3. **After failure** - Retries per backoffLimit

## Failure Scenarios

### Init Container Fails ❌
```
Result: Pod does NOT start
Action: Kubernetes retries (creates new pod)
Impact: CRITICAL - Service unavailable
```

### CronJob Fails ❌
```
Result: Job marked as failed
Action: Wait for next schedule
Impact: NONE - Service continues
```

## Cleanup

```bash
# Quick cleanup
./cleanup.sh

# Manual cleanup
helm uninstall database-load order-food -n default
```

## Documentation

- **[HYBRID_ARCHITECTURE.md](HYBRID_ARCHITECTURE.md)** - Detailed architecture docs
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - General deployment guide
- **[database-load/helm/values.yaml](database-load/helm/values.yaml)** - Configuration options

## Summary

**Two deployment modes**:
1. **Init Container** - Critical initial setup (MUST succeed)
2. **CronJob** - Non-critical periodic refresh (can fail)

**Result**: Guaranteed initial data + periodic updates without risking application availability.

---

**Status**: ✅ Implemented and ready for deployment
**Default Schedule**: Every 6 hours
**Failure Isolation**: CronJob failures don't impact running pods
