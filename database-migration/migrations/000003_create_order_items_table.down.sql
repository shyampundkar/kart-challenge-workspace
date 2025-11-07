-- Drop indexes first
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;

-- Drop order_items table (foreign keys will be dropped automatically)
DROP TABLE IF EXISTS order_items CASCADE;
