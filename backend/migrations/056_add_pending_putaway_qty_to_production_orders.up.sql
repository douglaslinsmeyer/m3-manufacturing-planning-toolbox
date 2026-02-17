-- Add pending putaway quantity field to production_orders table
-- This field is populated from manufacturing_orders.pending_putaway_qty
-- Only applies to MOs (not MOPs)

ALTER TABLE production_orders
ADD COLUMN IF NOT EXISTS pending_putaway_qty VARCHAR(30);

COMMENT ON COLUMN production_orders.pending_putaway_qty IS
    'Pending putaway quantity from MPTAWY.TRQT (MOs only) - quantity awaiting warehouse putaway';
