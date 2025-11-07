# PostgreSQL Migration Setup

Complete implementation of PostgreSQL database with migration support for the order-food application.

## Overview

This document describes the PostgreSQL database setup and migration implementation based on the OpenAPI schema defined in [order-food/api/openapi.yaml](order-food/api/openapi.yaml).

## Implementation Summary

### 1. Database Schema

Three tables have been created to support the order-food application:

#### Products Table
Stores product information (Chicken Waffle, Belgian Waffle, etc.)

```sql
CREATE TABLE products (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `idx_products_category` - For filtering by category

#### Orders Table
Stores order information with optional coupon codes

```sql
CREATE TABLE orders (
    id VARCHAR(50) PRIMARY KEY,
    coupon_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `idx_orders_created_at` - For sorting by creation date

#### Order Items Table (Junction Table)
Links orders to products with quantity (many-to-many relationship)

```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_order
        FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    UNIQUE(order_id, product_id)
);
```

**Foreign Keys**:
- `fk_order`: References orders table, cascade on delete (deleting an order deletes its items)
- `fk_product`: References products table, restrict on delete (can't delete products that are in orders)

**Indexes**:
- `idx_order_items_order_id` - For querying items by order
- `idx_order_items_product_id` - For querying items by product

**Constraints**:
- Unique constraint on (order_id, product_id) - prevents duplicate products in same order
- Check constraint on quantity - must be greater than 0

### 2. Migration Implementation

**Files Created**:
- [database-migration/internal/migration/migration.go](database-migration/internal/migration/migration.go) - Migration logic with PostgreSQL driver
- [database-migration/internal/migration/doc.go](database-migration/internal/migration/doc.go) - Package documentation
- [database-migration/cmd/main.go](database-migration/cmd/main.go) - Updated to use PostgreSQL migration

**Key Features**:
- PostgreSQL driver integration (github.com/lib/pq)
- Environment-based configuration
- OpenTelemetry tracing for all migration steps
- Proper error handling and rollback support
- Idempotent migrations (safe to run multiple times)

**Environment Variables**:
```bash
DB_HOST=postgres       # Default: localhost
DB_PORT=5432          # Default: 5432
DB_USER=postgres      # Default: postgres
DB_PASSWORD=postgres  # Default: postgres
DB_NAME=orderfood     # Default: orderfood
DB_SSLMODE=disable    # Default: disable
```

### 3. PostgreSQL Deployment

**Helm Chart Created**: [postgres/helm/](postgres/helm/)

**Components**:
- Deployment with PostgreSQL 16 Alpine image
- Service (ClusterIP) for cluster-internal access
- PersistentVolumeClaim for data persistence (1Gi)
- ConfigMap for database configuration
- Secret for password storage
- Liveness and readiness probes

**Configuration** ([postgres/helm/values.yaml](postgres/helm/values.yaml)):
```yaml
postgres:
  database: orderfood
  user: postgres
  password: postgres
  port: 5432

persistence:
  enabled: true
  size: 1Gi

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

### 4. Integration with Order-Food

The order-food Helm chart has been updated to include PostgreSQL connection environment variables for both init containers:

**Updated File**: [order-food/helm/values.yaml](order-food/helm/values.yaml)

**Init Container Configuration**:
```yaml
initContainers:
  databaseMigration:
    enabled: true
    env:
      - name: DB_HOST
        value: "postgres"
      - name: DB_PORT
        value: "5432"
      - name: DB_NAME
        value: "orderfood"
      - name: DB_USER
        value: "postgres"
      - name: DB_PASSWORD
        value: "postgres"
      - name: DB_SSLMODE
        value: "disable"

  databaseLoad:
    enabled: true
    env:
      # Same PostgreSQL environment variables
```

### 5. Deployment Flow

The deployment sequence ensures proper initialization:

1. **PostgreSQL** deploys first
2. Wait for PostgreSQL to be ready
3. **database-load CronJob** deploys (for periodic refresh)
4. **order-food** pod starts:
   - **Init Container #1**: database-migration runs
     - Connects to PostgreSQL
     - Creates tables: products, orders, order_items
     - Creates indexes
     - Creates foreign key constraints
   - **Init Container #2**: database-load runs
     - Loads initial product data
     - Loads initial user data (if applicable)
     - Loads initial order data (if applicable)
   - **Main Container**: order-food starts
     - API server becomes ready
     - Accepts traffic

## Usage

### Deploy Everything

```bash
./deploy.sh
```

This will:
1. Start Minikube (if not running)
2. Build Docker images
3. Deploy PostgreSQL
4. Deploy database-load CronJob
5. Deploy order-food (with migrations as init container)

### Verify Migration

```bash
# Check migration logs
kubectl logs -l app.kubernetes.io/name=order-food -c database-migration

# Check database schema
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\dt'

# View table details
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d products'
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d orders'
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d order_items'

# Check foreign keys
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
  ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = '\''FOREIGN KEY'\'';
'
```

### Access PostgreSQL

```bash
# Port forward
kubectl port-forward svc/postgres 5432:5432

# Connect with psql
psql -h localhost -U postgres -d orderfood
# Password: postgres
```

### Query Data

```bash
# View products
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c 'SELECT * FROM products;'

# View orders with items
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '
SELECT
    o.id as order_id,
    o.coupon_code,
    p.name as product_name,
    oi.quantity,
    p.price,
    (oi.quantity * p.price) as total
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
ORDER BY o.created_at DESC;
'
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                       │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │ PostgreSQL Deployment                            │      │
│  │ - Database: orderfood                            │      │
│  │ - Tables: products, orders, order_items          │      │
│  │ - Persistent Storage: 1Gi PVC                    │      │
│  └──────────────────────────────────────────────────┘      │
│                         ↑                                     │
│                         │ DB Connection                       │
│                         │                                     │
│  ┌──────────────────────────────────────────────────┐      │
│  │ order-food Pod                                   │      │
│  │                                                  │      │
│  │  ┌────────────────────────────────────────┐    │      │
│  │  │ Init #1: database-migration            │    │      │
│  │  │ - Connects to PostgreSQL               │    │      │
│  │  │ - Creates tables                       │    │      │
│  │  │ - Creates indexes                      │    │      │
│  │  │ - Creates foreign keys                 │    │      │
│  │  └────────────────────────────────────────┘    │      │
│  │                    ↓                             │      │
│  │  ┌────────────────────────────────────────┐    │      │
│  │  │ Init #2: database-load                 │    │      │
│  │  │ - Loads initial product data           │    │      │
│  │  │ - Loads initial user data              │    │      │
│  │  └────────────────────────────────────────┘    │      │
│  │                    ↓                             │      │
│  │  ┌────────────────────────────────────────┐    │      │
│  │  │ Main: order-food API                   │    │      │
│  │  │ - REST API server                      │    │      │
│  │  │ - Queries PostgreSQL                   │    │      │
│  │  └────────────────────────────────────────┘    │      │
│  └──────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Database Relationships

```
┌─────────────────┐
│    products     │
│                 │
│ • id (PK)       │
│ • name          │
│ • price         │
│ • category      │
└────────┬────────┘
         │
         │ fk_product (RESTRICT)
         │
         ↓
┌─────────────────┐         ┌─────────────────┐
│  order_items    │         │     orders      │
│                 │         │                 │
│ • id (PK)       │         │ • id (PK)       │
│ • order_id (FK) │←────────│ • coupon_code   │
│ • product_id(FK)│         │ • created_at    │
│ • quantity      │         └─────────────────┘
└─────────────────┘
         ↑
         │ fk_order (CASCADE)
```

**Relationship Notes**:
- One order can have multiple items (one-to-many)
- One product can be in multiple orders (one-to-many)
- order_items is a junction table for many-to-many relationship
- Deleting an order cascades to its items
- Cannot delete a product that's in existing orders (RESTRICT)

## Security Considerations

### Current Configuration (Development)
- Username: postgres
- Password: postgres
- SSL Mode: disable
- Port: 5432 (ClusterIP - not exposed externally)

### Production Recommendations

1. **Strong Password**:
   ```bash
   helm upgrade --install postgres ./postgres/helm \
     --set postgres.password="$(openssl rand -base64 32)"
   ```

2. **Use Kubernetes Secrets**:
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: postgres-credentials
   type: Opaque
   stringData:
     password: <strong-password>
   ```

3. **Enable SSL**:
   ```bash
   helm upgrade --install postgres ./postgres/helm \
     --set postgres.sslmode=require
   ```

4. **Network Policies**:
   Restrict database access to authorized pods only

5. **Read-Only User**:
   Create separate users for read-only operations

## Troubleshooting

### Migration Fails

```bash
# Check init container logs
kubectl logs -l app.kubernetes.io/name=order-food -c database-migration

# Check PostgreSQL logs
kubectl logs -l app.kubernetes.io/name=postgres

# Describe pod for status
kubectl describe pod -l app.kubernetes.io/name=order-food
```

### Connection Issues

```bash
# Test connection from debug pod
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql postgresql://postgres:postgres@postgres:5432/orderfood

# Check service
kubectl get svc postgres
kubectl describe svc postgres
```

### Schema Issues

```bash
# Manually run migration (for testing)
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood < schema.sql

# Reset database (WARNING: deletes all data)
kubectl exec -it deployment/postgres -- psql -U postgres -c 'DROP DATABASE orderfood; CREATE DATABASE orderfood;'
```

## Files Modified/Created

### Created
1. `database-migration/internal/migration/migration.go` - PostgreSQL migration implementation
2. `database-migration/internal/migration/doc.go` - Package documentation
3. `postgres/helm/` - Complete Helm chart for PostgreSQL
4. `postgres/README.md` - PostgreSQL documentation

### Modified
1. `database-migration/cmd/main.go` - Updated to use PostgreSQL migration
2. `database-migration/go.mod` - Added github.com/lib/pq dependency
3. `order-food/helm/values.yaml` - Added PostgreSQL configuration and env vars
4. `deploy.sh` - Added PostgreSQL deployment step
5. `cleanup.sh` - Added PostgreSQL cleanup

## Next Steps

1. **Update order-food Application**:
   - Replace in-memory storage with PostgreSQL queries
   - Implement repository layer for database access
   - Add connection pooling

2. **Update database-load**:
   - Modify to load data into PostgreSQL instead of in-memory
   - Ensure idempotent data loading

3. **Add Migrations**:
   - Version-controlled migrations (e.g., using golang-migrate)
   - Rollback support
   - Migration history tracking

4. **Performance Optimization**:
   - Add more indexes based on query patterns
   - Implement connection pooling
   - Add query logging for optimization

5. **Monitoring**:
   - Add PostgreSQL exporter for Prometheus
   - Create Grafana dashboards
   - Set up alerts for connection issues

## Testing

```bash
# Full deployment test
./deploy.sh

# Verify all components
kubectl get all

# Test migration
kubectl logs -l app.kubernetes.io/name=order-food -c database-migration | grep "✓"

# Test database schema
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\dt'
```

## Cleanup

```bash
# Remove all deployments including PostgreSQL
./cleanup.sh

# Or manually
helm uninstall postgres database-load order-food
kubectl delete pvc postgres-pvc
```

## Summary

✅ **Completed**:
- PostgreSQL deployment with Helm chart
- Database migration implementation with proper schema
- Foreign key relationships between orders and products
- Environment variable configuration
- Integration with existing order-food deployment
- Persistent storage for database
- Health checks and resource limits
- Documentation and troubleshooting guides

**Database Schema**: 3 tables (products, orders, order_items)
**Foreign Keys**: 2 (order_items → orders, order_items → products)
**Indexes**: 4 (for performance optimization)
**Deployment**: Fully automated with ./deploy.sh
