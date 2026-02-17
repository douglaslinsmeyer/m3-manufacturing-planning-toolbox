-- Add pending putaway quantity field to manufacturing_orders table
-- This field is populated from MPTAWY.TRQT during data loading
-- Represents the quantity of manufactured items awaiting warehouse putaway

ALTER TABLE manufacturing_orders
ADD COLUMN IF NOT EXISTS pending_putaway_qty VARCHAR(30);

-- Add index for filtering/querying on non-null pending putaways
CREATE INDEX IF NOT EXISTS idx_mos_pending_putaway
    ON manufacturing_orders(environment, pending_putaway_qty)
    WHERE pending_putaway_qty IS NOT NULL AND pending_putaway_qty != '';

COMMENT ON COLUMN manufacturing_orders.pending_putaway_qty IS
    'Pending putaway quantity from MPTAWY.TRQT - quantity awaiting warehouse putaway';
