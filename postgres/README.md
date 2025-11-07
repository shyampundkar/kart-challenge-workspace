# PostgreSQL Database

PostgreSQL database deployment for the order-food application.

## Overview

This Helm chart deploys a PostgreSQL 16 database instance configured for the order-food application with persistent storage.

## Database Schema

The database contains three main tables:

### Products Table
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

### Orders Table
```sql
CREATE TABLE orders (
    id VARCHAR(50) PRIMARY KEY,
    coupon_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Order Items Table (Junction Table)
```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    UNIQUE(order_id, product_id)
);
```

## Configuration

Default configuration (defined in [helm/values.yaml](helm/values.yaml)):

- **Database**: orderfood
- **User**: postgres
- **Password**: postgres (change in production!)
- **Port**: 5432
- **Storage**: 1Gi persistent volume

## Deployment

PostgreSQL is automatically deployed as part of the main deployment:

```bash
./deploy.sh
```

Or deploy manually:

```bash
helm upgrade --install postgres ./postgres/helm --wait
```

## Accessing PostgreSQL

### From within the cluster

```bash
# Service DNS name
postgres:5432

# Connection string
postgresql://postgres:postgres@postgres:5432/orderfood
```

### From local machine

```bash
# Port forward
kubectl port-forward svc/postgres 5432:5432

# Connect with psql
psql -h localhost -U postgres -d orderfood
# Password: postgres

# Or use connection string
psql postgresql://postgres:postgres@localhost:5432/orderfood
```

## Database Operations

### View Tables

```bash
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\dt'
```

### View Table Schema

```bash
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d products'
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d orders'
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '\d order_items'
```

### Query Data

```bash
# View all products
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c 'SELECT * FROM products;'

# View all orders
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c 'SELECT * FROM orders;'

# View order items with product details
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c '
SELECT o.id as order_id, p.name, oi.quantity, p.price
FROM order_items oi
JOIN orders o ON oi.order_id = o.id
JOIN products p ON oi.product_id = p.id;
'
```

## Environment Variables

The following environment variables are used by the database-migration init container:

- `DB_HOST`: postgres
- `DB_PORT`: 5432
- `DB_NAME`: orderfood
- `DB_USER`: postgres
- `DB_PASSWORD`: postgres
- `DB_SSLMODE`: disable

## Persistence

Data is persisted using a Kubernetes PersistentVolumeClaim (PVC):

- **Storage Class**: Default storage class
- **Access Mode**: ReadWriteOnce
- **Size**: 1Gi

To disable persistence (not recommended for production):

```bash
helm upgrade --install postgres ./postgres/helm \
  --set persistence.enabled=false
```

## Health Checks

The deployment includes liveness and readiness probes:

```bash
# Liveness probe
pg_isready -U postgres

# Check health
kubectl get pods -l app.kubernetes.io/name=postgres
```

## Resource Limits

Default resource allocation:

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

## Migration

Database schema is created automatically by the **database-migration** init container when order-food is deployed. The migration:

1. Creates tables (products, orders, order_items)
2. Sets up foreign key constraints
3. Creates indexes for performance

## Backup and Restore

### Backup

```bash
kubectl exec -it deployment/postgres -- pg_dump -U postgres orderfood > backup.sql
```

### Restore

```bash
kubectl exec -i deployment/postgres -- psql -U postgres orderfood < backup.sql
```

## Security Considerations

For production deployments:

1. **Change default password**:
   ```bash
   helm upgrade --install postgres ./postgres/helm \
     --set postgres.password=<strong-password>
   ```

2. **Enable SSL**:
   ```bash
   helm upgrade --install postgres ./postgres/helm \
     --set postgres.sslmode=require
   ```

3. **Use Kubernetes Secrets**:
   Store credentials in Kubernetes Secrets instead of values.yaml

4. **Network Policies**:
   Restrict database access to only authorized pods

## Troubleshooting

### Pod not starting

```bash
kubectl describe pod -l app.kubernetes.io/name=postgres
kubectl logs -l app.kubernetes.io/name=postgres
```

### Connection issues

```bash
# Test from another pod
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql postgresql://postgres:postgres@postgres:5432/orderfood
```

### Persistence issues

```bash
# Check PVC status
kubectl get pvc
kubectl describe pvc postgres-pvc
```

## Cleanup

```bash
# Uninstall PostgreSQL
helm uninstall postgres

# Delete PVC (if needed)
kubectl delete pvc postgres-pvc
```
