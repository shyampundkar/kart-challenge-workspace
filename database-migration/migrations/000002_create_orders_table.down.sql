-- Drop index first
DROP INDEX IF EXISTS idx_orders_created_at;

-- Drop orders table (will fail if order_items references it)
DROP TABLE IF EXISTS orders CASCADE;
