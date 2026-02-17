-- Remove CFIN indexes

DROP INDEX IF EXISTS idx_planned_mfg_orders_cfin_unlinked;
DROP INDEX IF EXISTS idx_customer_order_lines_cfin;
DROP INDEX IF EXISTS idx_planned_manufacturing_orders_cfin;
DROP INDEX IF EXISTS idx_manufacturing_orders_cfin;
