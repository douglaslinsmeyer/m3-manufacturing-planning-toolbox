-- Add missing M3 date field columns for manufacturing_orders

-- Planning dates (as integers in YYYYMMDD format from M3)
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS stdt INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS fidt INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS msti INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS mfti INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS fstd INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS ffid INTEGER;

-- Actual dates (as integers in YYYYMMDD format from M3)
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rsdt INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS refd INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rpdt INTEGER;

-- Add missing columns for planned_manufacturing_orders
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS stdt INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS fidt INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS reld INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS msti INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS mfti INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS pldt INTEGER;

-- Create indexes on M3 date fields for faster queries
CREATE INDEX IF NOT EXISTS idx_mo_stdt_fidt ON manufacturing_orders(stdt, fidt) WHERE stdt IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mop_stdt_fidt ON planned_manufacturing_orders(stdt, fidt) WHERE stdt IS NOT NULL;
