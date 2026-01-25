-- Add orty (order type) column to production_orders table
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS orty VARCHAR(10);

COMMENT ON COLUMN production_orders.orty IS 'Manufacturing/Customer order type code (ORTY from M3)';

-- Create index for filtering
CREATE INDEX IF NOT EXISTS idx_production_orders_orty ON production_orders(orty);
