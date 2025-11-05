# Quick Start: Init Container Setup

## What Changed?

`database-migration` now runs as an **init container** inside the `order-food` pod, ensuring migrations complete before the app starts.

## Quick Deploy

### With Docker Compose
```bash
docker-compose up --build
```

The migration runs automatically before order-food starts.

### With Minikube
```bash
./deploy.sh
```

This will:
1. Build `database-migration` image
2. Build `database-load` image
3. Build `order-food` image
4. Deploy `database-load` as a Job
5. Deploy `order-food` with `database-migration` as init container

## Verify Init Container Execution

### Check Pod Status
```bash
kubectl get pods -n default -l app.kubernetes.io/name=order-food
```

Expected output:
```
NAME                          READY   STATUS    RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    1/1     Running   0          2m
```

### View Init Container Logs
```bash
# View migration logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration

# View specific pod
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration
```

### View Main Container Logs
```bash
# View order-food logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food

# Follow logs in real-time
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food -f
```

## Pod Startup Sequence

```
1. Pod Created
   ↓
2. Init: database-migration starts
   ↓
3. Migration runs (visible in logs)
   ↓
4. Init: database-migration completes (Exit 0)
   ↓
5. Main: order-food starts
   ↓
6. Pod becomes Ready (1/1)
```

## Check Init Container Details

```bash
# Describe pod to see init container status
kubectl describe pod -n default -l app.kubernetes.io/name=order-food

# Look for:
# Init Containers:
#   database-migration:
#     State:          Terminated
#       Reason:       Completed
#       Exit Code:    0
```

## Configuration

### Enable/Disable Init Container

**Disable** (use for development if needed):
```bash
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false
```

**Enable** (default):
```bash
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=true \
  --set initContainers.databaseMigration.image.pullPolicy=Never
```

### Custom Values File

Create `custom-values.yaml`:
```yaml
initContainers:
  databaseMigration:
    enabled: true
    image:
      repository: database-migration
      tag: "v1.0.0"  # Use specific version
      pullPolicy: Always
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
      - name: ENVIRONMENT
        value: "production"
    resources:
      limits:
        cpu: 1000m
        memory: 512Mi
```

Deploy with:
```bash
helm upgrade --install order-food ./order-food/helm -f custom-values.yaml
```

## Troubleshooting

### Init Container Failed

**Symptoms:**
```
NAME                          READY   STATUS                  RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:Error              0          1m
```

**Solution:**
```bash
# 1. View logs
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration

# 2. Check events
kubectl get events -n default --sort-by='.lastTimestamp'

# 3. Describe pod
kubectl describe pod -n default order-food-7d8f9b5c4-xyz12
```

### Init Container Stuck

**Symptoms:**
```
NAME                          READY   STATUS        RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:0/1      0          5m
```

**Solution:**
```bash
# Watch logs in real-time
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration -f

# Check what it's waiting for
kubectl describe pod -n default order-food-7d8f9b5c4-xyz12
```

### Init Container CrashLoopBackOff

**Symptoms:**
```
NAME                          READY   STATUS                   RESTARTS   AGE
order-food-7d8f9b5c4-xyz12    0/1     Init:CrashLoopBackOff    3          2m
```

**Solution:**
```bash
# View logs from previous attempt
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration --previous

# Delete pod to force recreation (if needed)
kubectl delete pod -n default order-food-7d8f9b5c4-xyz12
```

## Observability

### Jaeger Traces

1. Access Jaeger UI:
   ```bash
   # With Docker Compose
   open http://localhost:16686

   # With Minikube (after port-forward)
   kubectl port-forward -n default svc/jaeger 16686:16686
   open http://localhost:16686
   ```

2. Select service: `database-migration`
3. View traces showing migration execution

### Prometheus Metrics

Access metrics endpoint:
```bash
# Port-forward order-food service
kubectl port-forward -n default svc/order-food 8080:80

# View metrics
curl http://localhost:8080/metrics
```

## Key Differences from Previous Setup

| Before | After |
|--------|-------|
| 3 Helm releases | 2 Helm releases |
| database-migration as Job | database-migration as init container |
| `helm uninstall database-migration database-load order-food` | `helm uninstall database-load order-food` |
| Migration runs independently | Migration runs with each pod |
| Manual ordering with dependencies | Automatic ordering (init runs first) |

## Best Practices

1. **Keep migrations fast**: Init containers block pod startup
2. **Make migrations idempotent**: Safe to run multiple times
3. **Use specific image tags**: Don't rely on `latest` in production
4. **Monitor init container duration**: Set up alerts for slow migrations
5. **Test rollback scenarios**: Ensure migrations can be reverted if needed

## Common Tasks

### Update Migration Code

```bash
# 1. Update code in database-migration/
# 2. Rebuild image
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration

# 3. Restart order-food pods to run new migration
kubectl rollout restart deployment/order-food -n default

# 4. Watch rollout
kubectl rollout status deployment/order-food -n default

# 5. Check migration logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration
```

### Scale order-food

```bash
# Scale to 3 replicas
kubectl scale deployment/order-food --replicas=3 -n default

# Each new pod will run the migration init container
kubectl get pods -n default -l app.kubernetes.io/name=order-food

# All pods should show 1/1 Ready
```

### Force Migration Re-run

```bash
# Delete all order-food pods
kubectl delete pods -n default -l app.kubernetes.io/name=order-food

# Kubernetes will recreate them, running migration again
kubectl get pods -n default -w
```

## Access Application

### With Minikube
```bash
# Port-forward to access
kubectl port-forward -n default svc/order-food 8080:80

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/product
```

### With Docker Compose
```bash
# Already exposed on 8080
curl http://localhost:8080/health
curl http://localhost:8080/api/product
```

## Clean Up

### Minikube
```bash
./cleanup.sh
# or manually:
helm uninstall database-load order-food -n default
```

### Docker Compose
```bash
docker-compose down
# Remove volumes (if any)
docker-compose down -v
```

## Next Steps

1. ✅ Init container setup complete
2. 📊 Monitor migration execution in Jaeger
3. 🔄 Set up CI/CD to build and push images
4. 🏷️ Use versioned tags instead of `latest`
5. 📈 Add alerts for init container failures
6. 🔐 Add database credentials via Secrets
7. 💾 Set up persistent volumes for database

## Resources

- [INIT_CONTAINER_ARCHITECTURE.md](INIT_CONTAINER_ARCHITECTURE.md) - Detailed architecture docs
- [DEPLOYMENT.md](DEPLOYMENT.md) - Full deployment guide
- [OBSERVABILITY.md](OBSERVABILITY.md) - Observability setup
- [Kubernetes Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
