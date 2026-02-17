-- Remove pending putaway quantity field from manufacturing_orders table
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS pending_putaway_qty;
