# Database Migrations

This directory contains database migration files managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Files

Migrations are organized in numbered pairs of up/down SQL files:

### 000001 - Create Products Table
- **Up**: [000001_create_products_table.up.sql](000001_create_products_table.up.sql)
  - Creates `products` table with columns: id, name, price, category, timestamps
  - Creates index on `category` column
  - Adds table and column comments

- **Down**: [000001_create_products_table.down.sql](000001_create_products_table.down.sql)
  - Drops `idx_products_category` index
  - Drops `products` table

### 000002 - Create Orders Table
- **Up**: [000002_create_orders_table.up.sql](000002_create_orders_table.up.sql)
  - Creates `orders` table with columns: id, coupon_code, timestamps
  - Creates index on `created_at` column
  - Adds table and column comments

- **Down**: [000002_create_orders_table.down.sql](000002_create_orders_table.down.sql)
  - Drops `idx_orders_created_at` index
  - Drops `orders` table (with CASCADE to handle foreign keys)

### 000003 - Create Order Items Table
- **Up**: [000003_create_order_items_table.up.sql](000003_create_order_items_table.up.sql)
  - Creates `order_items` junction table with columns: id, order_id, product_id, quantity, created_at
  - Creates foreign keys:
    - `fk_order`: order_id → orders.id (CASCADE delete)
    - `fk_product`: product_id → products.id (RESTRICT delete)
  - Creates unique constraint on (order_id, product_id)
  - Creates indexes on both foreign key columns
  - Adds table, column, and constraint comments

- **Down**: [000003_create_order_items_table.down.sql](000003_create_order_items_table.down.sql)
  - Drops indexes: `idx_order_items_product_id`, `idx_order_items_order_id`
  - Drops `order_items` table (foreign keys dropped automatically)

## Migration Naming Convention

Migration files follow this naming pattern:
```
{version}_{description}.{direction}.sql
```

Examples:
- `000001_create_products_table.up.sql` - Migration version 1, going up
- `000001_create_products_table.down.sql` - Migration version 1, going down

## Usage

### Apply All Migrations (Up)

The init container automatically runs all pending migrations:

```bash
# This happens automatically when order-food pod starts
kubectl logs -l app.kubernetes.io/name=order-food -c database-migration
```

### Rollback Last Migration (Down)

To rollback the last migration, you can modify the database-migration code to call `Down()` instead of `Run()`.

### Check Current Version

```bash
# Connect to PostgreSQL
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood

# Check schema_migrations table
SELECT * FROM schema_migrations;
```

The `schema_migrations` table is automatically created by golang-migrate and tracks:
- `version` - Current migration version
- `dirty` - Whether a migration is in progress or failed

## Migration Details

### Products Table Schema

```sql
CREATE TABLE products (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
```

**Purpose**: Stores product catalog
**Constraints**: Price must be non-negative
**Indexes**: category (for filtering)

### Orders Table Schema

```sql
CREATE TABLE orders (
    id VARCHAR(50) PRIMARY KEY,
    coupon_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
```

**Purpose**: Stores order information
**Indexes**: created_at (for sorting)

### Order Items Table Schema

```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    UNIQUE(order_id, product_id)
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
```

**Purpose**: Junction table for order-product many-to-many relationship
**Foreign Keys**:
- `order_id` → `orders.id` (CASCADE: deleting order removes its items)
- `product_id` → `products.id` (RESTRICT: can't delete products in orders)
**Constraints**:
- quantity > 0
- Unique (order_id, product_id) - no duplicate products per order
**Indexes**: Both foreign key columns for join performance

## Relationships

```
products (1) ←──────────────────────→ (M) order_items
                    fk_product
                    (RESTRICT)

orders (1) ←────────────────────────→ (M) order_items
                    fk_order
                    (CASCADE)
```

- One order can have multiple items
- One product can be in multiple orders
- `order_items` is the junction table implementing the many-to-many relationship

## Creating New Migrations

To create a new migration:

1. **Name the file correctly**:
   ```
   000004_add_user_table.up.sql
   000004_add_user_table.down.sql
   ```

2. **Write the up migration** (applies changes):
   ```sql
   CREATE TABLE users (
       id VARCHAR(50) PRIMARY KEY,
       email VARCHAR(255) UNIQUE NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
   );
   ```

3. **Write the down migration** (reverts changes):
   ```sql
   DROP TABLE IF EXISTS users CASCADE;
   ```

4. **Test the migration**:
   ```bash
   # Build and deploy
   ./deploy.sh

   # Check logs
   kubectl logs -l app.kubernetes.io/name=order-food -c database-migration
   ```

## Best Practices

### Up Migrations

✅ **Do**:
- Use `IF NOT EXISTS` for tables and indexes
- Add comments to document schema
- Create indexes for frequently queried columns
- Use meaningful constraint names
- Include foreign keys for referential integrity

❌ **Don't**:
- Assume specific data exists
- Use database-specific features without consideration
- Create tables without primary keys
- Forget to add down migrations

### Down Migrations

✅ **Do**:
- Drop dependencies before tables (indexes, foreign keys)
- Use `CASCADE` when appropriate
- Test rollback in development
- Document destructive operations

❌ **Don't**:
- Forget to drop indexes
- Leave orphaned data
- Assume rollback will always work

## Troubleshooting

### Migration Failed (Dirty State)

If a migration fails halfway, golang-migrate marks the version as "dirty":

```bash
# Check version and dirty state
kubectl exec -it deployment/postgres -- psql -U postgres -d orderfood -c \
  'SELECT * FROM schema_migrations;'
```

**Solution**: Force the version and retry:
```go
// In migration code (use with caution!)
migrator.Force(2) // Force to version 2
migrator.Run(ctx, tracer) // Retry migration
```

### Migration Won't Apply

**Check if already applied**:
```sql
SELECT version, dirty FROM schema_migrations;
```

**Check migration files**:
```bash
kubectl exec -it deployment/database-migration -- ls -la migrations/
```

### Rollback Failed

**Check for dependent data**:
```sql
-- Check if tables have data
SELECT COUNT(*) FROM products;
SELECT COUNT(*) FROM orders;
SELECT COUNT(*) FROM order_items;
```

**Check for foreign key violations**:
```sql
-- List all foreign keys
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY';
```

## Version Control

**Current Schema Version**: 3

Each migration increments the version:
- Version 0: No migrations applied (empty database)
- Version 1: Products table created
- Version 2: Orders table created
- Version 3: Order items table created (current)

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [OpenAPI Schema](../../../order-food/api/openapi.yaml)
