-- Remove pending putaway quantity field from production_orders table
ALTER TABLE production_orders DROP COLUMN IF EXISTS pending_putaway_qty;
