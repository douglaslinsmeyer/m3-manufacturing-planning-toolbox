DROP INDEX IF EXISTS idx_planned_mfg_orders_deleted_remotely;
DROP INDEX IF EXISTS idx_mfg_orders_deleted_remotely;
DROP INDEX IF EXISTS idx_production_orders_deleted_remotely;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS deleted_remotely;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS deleted_remotely;
ALTER TABLE production_orders DROP COLUMN IF EXISTS deleted_remotely;
