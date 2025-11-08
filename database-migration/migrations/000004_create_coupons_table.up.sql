-- Create coupons table as UNLOGGED for faster bulk loading
-- UNLOGGED tables are not crash-safe but much faster for initial data load
CREATE UNLOGGED TABLE IF NOT EXISTS coupons (
    coupon VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    PRIMARY KEY (coupon, file_name)
);

-- Add comments to table
COMMENT ON TABLE coupons IS 'Stores coupon information';
COMMENT ON COLUMN coupons.coupon IS 'Coupon code or identifier';
COMMENT ON COLUMN coupons.file_name IS 'Associated file name for the coupon';
