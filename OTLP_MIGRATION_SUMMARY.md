# OTLP Migration Summary

## Overview

Successfully migrated from the deprecated Jaeger exporter to the recommended OTLP (OpenTelemetry Protocol) HTTP exporter across all microservices.

## Why This Change?

The `go.opentelemetry.io/otel/exporters/jaeger` package was deprecated in July 2023. OpenTelemetry officially recommends using OTLP as the standard protocol for sending telemetry data. Jaeger v1.35+ natively supports OTLP, making this migration seamless.

## What Changed

### 1. Telemetry Code Updates

Updated all three microservices to use OTLP HTTP exporter:

- **[database-migration/internal/telemetry/telemetry.go](database-migration/internal/telemetry/telemetry.go)**
- **[database-load/internal/telemetry/telemetry.go](database-load/internal/telemetry/telemetry.go)**
- **[order-food/internal/telemetry/telemetry.go](order-food/internal/telemetry/telemetry.go)**
- **[order-food/cmd/main.go](order-food/cmd/main.go#L60)** - Updated logging to use OTLPEndpoint

#### Before (Deprecated):
```go
import (
    "go.opentelemetry.io/otel/exporters/jaeger"
)

// Create Jaeger exporter
exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
```

#### After (Recommended):
```go
import (
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// Create OTLP HTTP exporter for Jaeger
exp, err := otlptracehttp.New(context.Background(),
    otlptracehttp.WithEndpoint(config.OTLPEndpoint),
    otlptracehttp.WithInsecure(), // Use insecure for local development
)
```

### 2. Configuration Changes

#### Config Struct
- **Before**: `JaegerEndpoint string`
- **After**: `OTLPEndpoint string`

#### Environment Variable
- **Before**: `JAEGER_ENDPOINT` (default: `http://localhost:14268/api/traces`)
- **After**: `OTLP_ENDPOINT` (default: `localhost:4318`)

#### Jaeger OTLP Endpoints
- **HTTP**: Port `4318` (used in this migration)
- **gRPC**: Port `4317` (alternative, not used)

### 3. Dependency Updates

All `go.mod` files updated to use OTLP exporter:

#### database-migration/go.mod
```go
require (
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
    go.opentelemetry.io/otel/sdk v1.38.0
    go.opentelemetry.io/otel/trace v1.38.0
)
```

#### database-load/go.mod
```go
require (
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
    go.opentelemetry.io/otel/sdk v1.38.0
    go.opentelemetry.io/otel/trace v1.38.0
)
```

#### order-food/go.mod
```go
require (
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
    go.opentelemetry.io/otel/exporters/prometheus v0.46.0
    go.opentelemetry.io/otel/sdk v1.38.0
    go.opentelemetry.io/otel/sdk/metric v1.38.0
)
```

**Version Upgrades**:
- OpenTelemetry: `v1.24.0` → `v1.38.0`
- Removed: `go.opentelemetry.io/otel/exporters/jaeger v1.17.0`
- Added: `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0`

### 4. Helm Values Updates

Updated [order-food/helm/values.yaml](order-food/helm/values.yaml#L112-L133) to use OTLP endpoint:

#### Before:
```yaml
initContainers:
  databaseMigration:
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
  databaseLoad:
    env:
      - name: JAEGER_ENDPOINT
        value: "http://jaeger:14268/api/traces"
```

#### After:
```yaml
initContainers:
  databaseMigration:
    env:
      - name: OTLP_ENDPOINT
        value: "jaeger:4318"
  databaseLoad:
    env:
      - name: OTLP_ENDPOINT
        value: "jaeger:4318"
```

## Benefits

1. **Future-Proof**: Uses the official OpenTelemetry standard protocol
2. **Better Support**: OTLP is actively maintained and recommended
3. **Flexibility**: Can send traces to any OTLP-compatible backend (Jaeger, Tempo, etc.)
4. **Modern Stack**: Aligned with latest OpenTelemetry best practices
5. **No Deprecation Warnings**: Removes staticcheck warnings

## Jaeger Compatibility

Jaeger supports OTLP natively since v1.35:

| Protocol | Port | Endpoint Path |
|----------|------|---------------|
| OTLP HTTP | 4318 | `/v1/traces` |
| OTLP gRPC | 4317 | N/A |
| Legacy HTTP | 14268 | `/api/traces` (deprecated) |

## Testing

To verify the migration:

1. **Build and Deploy**:
   ```bash
   ./deploy.sh
   ```

2. **Check Traces**:
   ```bash
   # Forward Jaeger UI port
   kubectl port-forward -n default svc/jaeger 16686:16686

   # Open browser to http://localhost:16686
   # You should see traces from all services
   ```

3. **Verify Logs**:
   ```bash
   # Check that telemetry initializes correctly
   kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food | grep "OpenTelemetry tracer initialized"
   kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration | grep "OpenTelemetry tracer initialized"
   kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load | grep "OpenTelemetry tracer initialized"
   ```

## Troubleshooting

### Issue: Traces Not Appearing in Jaeger

**Check OTLP Endpoint**:
```bash
# Verify Jaeger is listening on port 4318
kubectl get svc -n default jaeger
kubectl port-forward -n default svc/jaeger 4318:4318

# Test OTLP endpoint
curl -v http://localhost:4318/v1/traces
```

**Check Application Logs**:
```bash
kubectl logs -n default -l app.kubernetes.io/name=order-food
```

### Issue: Connection Refused

**Verify Jaeger Service**:
```bash
# Check if Jaeger pod is running
kubectl get pods -n default | grep jaeger

# Check Jaeger logs
kubectl logs -n default <jaeger-pod-name>
```

**Check Environment Variables**:
```bash
# Verify OTLP_ENDPOINT is set correctly
kubectl exec -n default <order-food-pod> -- env | grep OTLP
```

## Rollback (If Needed)

If you need to rollback to the old Jaeger exporter:

1. **Revert Telemetry Code**:
   ```bash
   git checkout HEAD~1 -- database-migration/internal/telemetry/telemetry.go
   git checkout HEAD~1 -- database-load/internal/telemetry/telemetry.go
   git checkout HEAD~1 -- order-food/internal/telemetry/telemetry.go
   ```

2. **Revert Dependencies**:
   ```bash
   cd database-migration && go get go.opentelemetry.io/otel/exporters/jaeger@v1.17.0 && go mod tidy
   cd ../database-load && go get go.opentelemetry.io/otel/exporters/jaeger@v1.17.0 && go mod tidy
   cd ../order-food && go get go.opentelemetry.io/otel/exporters/jaeger@v1.17.0 && go mod tidy
   ```

3. **Revert Helm Values**:
   ```bash
   git checkout HEAD~1 -- order-food/helm/values.yaml
   ```

## Migration Checklist

- [x] Update database-migration telemetry code
- [x] Update database-load telemetry code
- [x] Update order-food telemetry code
- [x] Update database-migration go.mod
- [x] Update database-load go.mod
- [x] Update order-food go.mod
- [x] Update Helm values (OTLP_ENDPOINT)
- [x] Remove all JAEGER_ENDPOINT references
- [x] Test build and deployment
- [x] Verify traces in Jaeger UI

## Summary

This migration successfully modernizes the telemetry stack by:
- Removing deprecated Jaeger exporter
- Adopting OTLP as the standard protocol
- Upgrading OpenTelemetry to v1.38.0
- Maintaining backward compatibility with existing Jaeger infrastructure

**Result**: All microservices now use the recommended OTLP HTTP exporter, ensuring long-term supportability and compatibility with the OpenTelemetry ecosystem.

## Related Documentation

- [OpenTelemetry OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
- [Jaeger OTLP Support](https://www.jaegertracing.io/docs/latest/apis/#opentelemetry-protocol-stable)
- [Go OTLP Exporter Docs](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp)
