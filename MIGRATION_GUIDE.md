# Migration Guide: Init Container Architecture

## Overview

This guide helps you migrate from the old architecture (3 separate Helm releases) to the new architecture (1 Helm release with 2 init containers).

## What Changed

### Old Architecture
```
3 Separate Helm Releases:
1. database-migration (Job)
2. database-load (Job)
3. order-food (Deployment)

Deploy command:
  helm install database-migration ./database-migration/helm
  helm install database-load ./database-load/helm
  helm install order-food ./order-food/helm

Uninstall command:
  helm uninstall database-migration database-load order-food
```

### New Architecture
```
1 Helm Release with Init Containers:
- order-food (Deployment)
  ├─ Init Container #1: database-migration
  ├─ Init Container #2: database-load
  └─ Main Container: order-food

Deploy command:
  helm install order-food ./order-food/helm \
    --set initContainers.databaseMigration.image.pullPolicy=Never \
    --set initContainers.databaseLoad.image.pullPolicy=Never

Uninstall command:
  helm uninstall order-food
```

## Migration Steps

### Step 1: Backup Current State (Optional)

If you have important data:

```bash
# Export Helm releases
helm get values database-migration -n default > backup-migration-values.yaml
helm get values database-load -n default > backup-load-values.yaml
helm get values order-food -n default > backup-order-food-values.yaml

# Backup any persistent data (if applicable)
kubectl get all -n default -o yaml > backup-all-resources.yaml
```

### Step 2: Uninstall Old Releases

```bash
# Stop all existing releases
helm uninstall database-migration database-load order-food -n default

# Verify cleanup
helm list -n default
kubectl get all -n default

# Wait for all resources to be deleted
kubectl wait --for=delete pod -l app.kubernetes.io/name=order-food --timeout=60s
```

### Step 3: Build New Images

```bash
# Set Docker environment to use minikube
eval $(minikube docker-env)

# Build all three images
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food

# Verify images
docker images | grep -E "database-migration|database-load|order-food"
```

### Step 4: Deploy New Architecture

```bash
# Option 1: Use deployment script (recommended)
./deploy.sh

# Option 2: Manual deployment
helm upgrade --install order-food ./order-food/helm \
  --namespace default \
  --set image.pullPolicy=Never \
  --set initContainers.databaseMigration.image.pullPolicy=Never \
  --set initContainers.databaseLoad.image.pullPolicy=Never \
  --wait \
  --timeout 10m
```

### Step 5: Verify Deployment

```bash
# Check pod status
kubectl get pods -n default -l app.kubernetes.io/name=order-food

# Expected output:
# NAME                          READY   STATUS    RESTARTS   AGE
# order-food-7d8f9b5c4-xyz12    1/1     Running   0          2m

# Check init container logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load

# Check main container logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food

# Test the API
kubectl port-forward -n default svc/order-food 8080:80 &
curl http://localhost:8080/health
curl http://localhost:8080/api/product
```

### Step 6: Verify Observability

```bash
# Check Jaeger traces (if running)
# Port-forward Jaeger
kubectl port-forward -n default svc/jaeger 16686:16686 &
open http://localhost:16686

# Look for traces from:
# - database-migration service
# - database-load service
# - order-food service

# Check Prometheus metrics
curl http://localhost:8080/metrics
```

## Rollback Plan

If something goes wrong:

### Quick Rollback

```bash
# Uninstall new release
helm uninstall order-food -n default

# Redeploy old architecture
helm install database-migration ./database-migration/helm
helm install database-load ./database-load/helm
helm install order-food ./order-food/helm
```

### Restore from Backup

```bash
# Restore Helm values
helm install database-migration ./database-migration/helm -f backup-migration-values.yaml
helm install database-load ./database-load/helm -f backup-load-values.yaml
helm install order-food ./order-food/helm -f backup-order-food-values.yaml
```

## Common Issues

### Issue 1: Init Container Fails

**Symptoms:**
```
order-food-7d8f9b5c4-xyz12    0/1     Init:Error    0    1m
```

**Solution:**
```bash
# Check which init container failed
kubectl describe pod -n default -l app.kubernetes.io/name=order-food

# View logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load

# Common causes:
# - Image not built locally (use: eval $(minikube docker-env) && docker build ...)
# - Environment variables incorrect
# - Resource limits too low
```

### Issue 2: Image Pull Error

**Symptoms:**
```
Failed to pull image "database-migration:latest": rpc error: code = Unknown desc = Error response from daemon: pull access denied
```

**Solution:**
```bash
# Ensure you're using minikube's Docker daemon
eval $(minikube docker-env)

# Rebuild images
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food

# Verify images exist
docker images | grep database-migration

# Redeploy with Never pull policy
helm upgrade --install order-food ./order-food/helm \
  --set image.pullPolicy=Never \
  --set initContainers.databaseMigration.image.pullPolicy=Never \
  --set initContainers.databaseLoad.image.pullPolicy=Never
```

### Issue 3: Pod Stuck in Init

**Symptoms:**
```
order-food-7d8f9b5c4-xyz12    0/1     Init:1/2    0    5m
```

**Solution:**
```bash
# Check which init container is running
kubectl describe pod -n default -l app.kubernetes.io/name=order-food

# Watch logs in real-time
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load -f

# Check events
kubectl get events -n default --sort-by='.lastTimestamp' | grep order-food

# If stuck too long, delete pod to force restart
kubectl delete pod -n default -l app.kubernetes.io/name=order-food
```

## Configuration Changes

### Environment Variables

Old way (separate Helm values):
```bash
helm install database-migration ./database-migration/helm \
  --set env.JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

New way (init container values):
```bash
helm install order-food ./order-food/helm \
  --set initContainers.databaseMigration.env[0].value=http://jaeger:14268/api/traces
```

Or use values file:
```yaml
# custom-values.yaml
initContainers:
  databaseMigration:
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
      - name: ENVIRONMENT
        value: "production"
```

```bash
helm install order-food ./order-food/helm -f custom-values.yaml
```

### Resource Limits

Old way:
```yaml
# database-migration/helm/values.yaml
resources:
  limits:
    cpu: 500m
    memory: 256Mi
```

New way:
```yaml
# order-food/helm/values.yaml
initContainers:
  databaseMigration:
    resources:
      limits:
        cpu: 500m
        memory: 256Mi
```

## Testing Checklist

After migration, verify:

- [ ] Pod starts successfully (STATUS: Running)
- [ ] Init containers complete (Init:2/2)
- [ ] Migration logs show success
- [ ] Data load logs show success
- [ ] Application responds to /health
- [ ] Application responds to /api/product
- [ ] Jaeger shows traces from all three containers
- [ ] Prometheus metrics are exposed at /metrics
- [ ] Pod restarts work correctly
- [ ] Scaling works (replicas > 1)
- [ ] Rolling updates work

## Performance Comparison

| Metric | Old Architecture | New Architecture |
|--------|------------------|------------------|
| **Deployment Time** | ~3-5 minutes | ~2-3 minutes |
| **Resource Count** | 2 Jobs + 1 Deployment | 1 Deployment |
| **Cleanup Time** | ~30 seconds | ~10 seconds |
| **Helm Releases** | 3 | 1 |
| **Log Commands** | 3 different selectors | 1 pod, 3 containers |
| **Failure Recovery** | Manual Job restart | Automatic pod restart |

## Best Practices

### 1. Always Use Deployment Script

```bash
# Recommended
./deploy.sh

# Instead of manual helm commands
```

### 2. Check Logs After Deployment

```bash
# Check all init containers completed successfully
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration | tail
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load | tail
```

### 3. Monitor Init Container Duration

```bash
# Add this to your monitoring
kubectl get pods -n default -l app.kubernetes.io/name=order-food \
  -o jsonpath='{range .items[*].status.initContainerStatuses[*]}{.name}{"\t"}{.state.terminated.finishedAt}{"\n"}{end}'
```

### 4. Use Specific Image Tags in Production

```yaml
# Development
initContainers:
  databaseMigration:
    image:
      tag: "latest"

# Production
initContainers:
  databaseMigration:
    image:
      tag: "v1.2.3"
```

## FAQ

### Q: Do I need to keep the old Helm charts?

**A:** Keep the database-migration and database-load directories for:
- Building the Docker images
- Reference and documentation
- Potential future standalone use

But you no longer deploy them as separate Helm releases.

### Q: Can I disable init containers for development?

**A:** Yes:

```bash
helm install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false \
  --set initContainers.databaseLoad.enabled=false
```

### Q: What happens if migration fails?

**A:** The pod won't start. Init containers run again when the pod restarts. Make migrations idempotent!

### Q: Can I run only migration without data load?

**A:** Yes:

```bash
helm install order-food ./order-food/helm \
  --set initContainers.databaseLoad.enabled=false
```

### Q: How do I update just the migration?

**A:**

```bash
# Rebuild migration image
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration

# Restart pods to run new migration
kubectl rollout restart deployment/order-food -n default
```

## Cleanup Old Resources

After successful migration, clean up old Helm chart files if desired:

```bash
# Optional: Remove old Job-based Helm charts
# (Keep directories for Docker builds)

# Remove old Helm templates
rm -rf database-migration/helm/templates/job.yaml
rm -rf database-load/helm/templates/job.yaml

# Or keep them for reference/documentation
```

## Documentation Updates

Updated documentation:
- ✅ [DEPLOYMENT.md](DEPLOYMENT.md) - Reflects init container architecture
- ✅ [INIT_CONTAINER_ARCHITECTURE.md](INIT_CONTAINER_ARCHITECTURE.md) - Detailed architecture docs
- ✅ [INIT_CONTAINERS_COMPLETE.md](INIT_CONTAINERS_COMPLETE.md) - Complete reference
- ✅ [QUICK_START_INIT_CONTAINER.md](QUICK_START_INIT_CONTAINER.md) - Quick reference

## Support

If you encounter issues:

1. Check logs: `kubectl logs -n default -l app.kubernetes.io/name=order-food --all-containers`
2. Describe pod: `kubectl describe pod -n default -l app.kubernetes.io/name=order-food`
3. Check events: `kubectl get events -n default --sort-by='.lastTimestamp'`
4. Refer to troubleshooting section above
5. Roll back if needed

## Summary

Migration is straightforward:
1. ✅ Uninstall old releases (3 helm uninstalls)
2. ✅ Build new images
3. ✅ Deploy new architecture (1 helm install)
4. ✅ Verify all init containers succeed
5. ✅ Test application functionality

**Result**: Simpler, more robust deployment with guaranteed execution order.
