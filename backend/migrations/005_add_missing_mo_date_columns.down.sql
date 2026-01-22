-- Rollback missing M3 date columns

DROP INDEX IF EXISTS idx_mop_stdt_fidt;
DROP INDEX IF EXISTS idx_mo_stdt_fidt;

ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS pldt;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS mfti;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS msti;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS reld;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS fidt;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS stdt;

ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rpdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS refd;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rsdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS ffid;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS fstd;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS mfti;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS msti;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS fidt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS stdt;
