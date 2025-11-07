-- Create order_items junction table
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key to orders table (CASCADE delete)
    CONSTRAINT fk_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,

    -- Foreign key to products table (RESTRICT delete)
    CONSTRAINT fk_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT,

    -- Ensure no duplicate products in same order
    UNIQUE(order_id, product_id)
);

-- Create indexes for foreign keys to improve join performance
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- Add comments to table
COMMENT ON TABLE order_items IS 'Junction table linking orders to products (many-to-many relationship)';
COMMENT ON COLUMN order_items.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN order_items.order_id IS 'Reference to orders table';
COMMENT ON COLUMN order_items.product_id IS 'Reference to products table';
COMMENT ON COLUMN order_items.quantity IS 'Number of items ordered (must be > 0)';
COMMENT ON CONSTRAINT fk_order ON order_items IS 'Deleting an order cascades to its items';
COMMENT ON CONSTRAINT fk_product ON order_items IS 'Cannot delete products that are in existing orders';
