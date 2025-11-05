# Init Container Architecture

## Overview

Both `database-migration` and `database-load` modules now run as **init containers** within the `order-food` pod. This ensures that:
1. Database migrations are executed first
2. Initial data is loaded second
3. The main application starts last

All steps must complete successfully in sequence before the application becomes ready.

## Architecture Changes

### Before
```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│ database-migration  │────▶│  database-load      │────▶│    order-food       │
│  (Kubernetes Job)   │     │  (Kubernetes Job)   │     │  (Deployment)       │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘
```

### After
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

## Benefits

### 1. Atomic Deployment
- All setup steps and application deployment are tightly coupled
- Order-food pod won't start until all init containers succeed
- No race conditions between migration, data loading, and application startup
- Guaranteed execution order

### 2. Simplified Operations
- Only need to deploy 1 Helm release instead of 3
- Single unit of deployment and management
- Automatic retry if any init container fails (Kubernetes will restart the pod)

### 3. Better Resource Utilization
- Init containers terminate after completion
- No separate Job resources to clean up
- Setup runs on-demand when pods are created/restarted

### 4. Improved Reliability
- Migration must succeed before application starts
- Prevents application from starting with outdated schema
- Clear failure visibility (pod won't be ready if init fails)

## Implementation Details

### Helm Chart Configuration

**[order-food/helm/values.yaml](order-food/helm/values.yaml)**
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
```

**[order-food/helm/templates/deployment.yaml](order-food/helm/templates/deployment.yaml:30-43)**
```yaml
{{- if .Values.initContainers.databaseMigration.enabled }}
initContainers:
- name: database-migration
  image: "{{ .Values.initContainers.databaseMigration.image.repository }}:{{ .Values.initContainers.databaseMigration.image.tag }}"
  imagePullPolicy: {{ .Values.initContainers.databaseMigration.image.pullPolicy }}
  {{- with .Values.initContainers.databaseMigration.env }}
  env:
    {{- toYaml . | nindent 12 }}
  {{- end }}
  {{- with .Values.initContainers.databaseMigration.resources }}
  resources:
    {{- toYaml . | nindent 12 }}
  {{- end }}
{{- end }}
```

### Deployment Process

1. **Build Phase**
   ```bash
   docker build -t database-migration:latest ./database-migration
   docker build -t order-food:latest ./order-food
   ```

2. **Deploy Phase**
   ```bash
   helm upgrade --install order-food ./order-food/helm \
     --set image.pullPolicy=Never \
     --set initContainers.databaseMigration.image.pullPolicy=Never
   ```

3. **Execution Flow**
   ```
   Pod Created
      ↓
   Init Container: database-migration starts
      ↓
   Migration executes (with OpenTelemetry tracing)
      ↓
   Init Container: database-migration completes successfully
      ↓
   Main Container: order-food starts
      ↓
   Pod becomes Ready
   ```

## Disabling the Init Container

If you need to disable the init container (for example, in development):

```bash
helm upgrade --install order-food ./order-food/helm \
  --set initContainers.databaseMigration.enabled=false
```

Or create a custom values file:

```yaml
# custom-values.yaml
initContainers:
  databaseMigration:
    enabled: false
```

```bash
helm upgrade --install order-food ./order-food/helm -f custom-values.yaml
```

## Viewing Init Container Logs

### Check Init Container Status
```bash
kubectl get pods -n default -l app.kubernetes.io/name=order-food
```

### View Init Container Logs
```bash
# View database-migration init container logs
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration

# View specific pod's init container
kubectl logs -n default <pod-name> -c database-migration
```

### Describe Pod (shows init container state)
```bash
kubectl describe pod -n default <pod-name>
```

Output shows:
```
Init Containers:
  database-migration:
    State:          Terminated
      Reason:       Completed
      Exit Code:    0
```

## Troubleshooting

### Init Container Failed

If the init container fails, the pod won't start:

```bash
# Check pod status
kubectl get pods -n default

# You'll see something like:
# NAME                          READY   STATUS                  RESTARTS   AGE
# order-food-7d8f9b5c4-xyz12    0/1     Init:Error              0          1m
```

**To debug:**

1. View init container logs:
   ```bash
   kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration
   ```

2. Describe the pod for detailed error:
   ```bash
   kubectl describe pod -n default order-food-7d8f9b5c4-xyz12
   ```

3. Check events:
   ```bash
   kubectl get events -n default --sort-by='.lastTimestamp'
   ```

### Init Container Stuck

If init container is running too long:

```bash
# Check if it's still running
kubectl get pods -n default

# You'll see:
# NAME                          READY   STATUS           RESTARTS   AGE
# order-food-7d8f9b5c4-xyz12    0/1     Init:0/1         0          5m

# View logs to see what it's doing
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration -f
```

### Init Container CrashLoopBackOff

If init container keeps failing and restarting:

```bash
# View logs from previous attempt
kubectl logs -n default order-food-7d8f9b5c4-xyz12 -c database-migration --previous
```

## Docker Compose

In Docker Compose, we still use `depends_on` to ensure proper startup order:

**[docker-compose.yml](docker-compose.yml:68-71)**
```yaml
order-food:
  depends_on:
    - jaeger
    - database-migration  # Ensures migration runs first
    - database-load
```

## Observability

Both init container and main container are instrumented with OpenTelemetry:

### Tracing

**Init Container (database-migration):**
- Service: `database-migration`
- Spans: `migration.execute`, `migration.createTables`, `migration.createIndexes`

**Main Container (order-food):**
- Service: `order-food`
- Automatic HTTP tracing via otelgin middleware
- Metrics exposed at `/metrics`

### Viewing Traces in Jaeger

1. Access Jaeger UI: http://localhost:16686
2. Select service: `database-migration` or `order-food`
3. View traces showing the complete execution flow

## Migration Strategy

### Zero-Downtime Updates

When deploying updates to order-food:

1. New pod is created
2. Init container runs migration
3. If migration succeeds, new pod starts
4. Old pod is terminated only after new pod is ready
5. Rolling update ensures no downtime

### Schema Compatibility

Ensure migrations are backward-compatible:
- New columns should be nullable or have defaults
- Don't drop columns used by running instances
- Use multi-step migrations for breaking changes

## Best Practices

1. **Keep Migrations Fast**: Init containers block pod startup
2. **Idempotent Migrations**: Migrations should be safe to run multiple times
3. **Error Handling**: Fail fast with clear error messages
4. **Resource Limits**: Set appropriate CPU/memory limits for init container
5. **Monitoring**: Monitor init container execution time and failures
6. **Version Control**: Tag migration images with versions

## Comparison with Job-Based Approach

| Aspect | Init Container | Separate Job |
|--------|---------------|--------------|
| Coupling | Tight - runs with each pod | Loose - runs independently |
| Timing | Before pod starts | Scheduled separately |
| Cleanup | Automatic | Manual or with TTL |
| Failure Handling | Pod won't start | Job fails, manual intervention |
| Resource Usage | Ephemeral | Persistent until cleaned |
| Visibility | Pod logs | Separate Job logs |
| Complexity | Lower | Higher |
| Use Case | Per-deployment migrations | One-time setup |

## When to Use Init Containers vs Jobs

### Use Init Container When:
- ✅ Migration must run before application starts
- ✅ Migration is idempotent
- ✅ Migration is fast (< 1 minute)
- ✅ You want atomic deployment

### Use Separate Job When:
- ✅ Migration is a one-time setup
- ✅ Migration takes a long time
- ✅ Migration requires special permissions
- ✅ You need to run migrations independently

## Future Enhancements

Potential improvements:

1. **Migration Versioning**: Track applied migrations
2. **Health Checks**: Add liveness probes to init container
3. **Rollback Support**: Automated rollback on failure
4. **Shared Volume**: Use volume for migration state
5. **Notification**: Alert on migration failures
6. **Metrics**: Track migration duration and success rate

## References

- [Kubernetes Init Containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Database Migration Patterns](https://martinfowler.com/articles/evodb.html)

## Summary

The init container approach provides:
- **Better reliability**: Migration must succeed before app starts
- **Simpler operations**: Fewer resources to manage
- **Atomic deployment**: App and migration are tightly coupled
- **Clear failure semantics**: Pod won't be ready if migration fails
- **Automatic cleanup**: Init container terminates after completion

This architecture is ideal for microservices where database migrations are part of the application deployment lifecycle.
