-- Migration 015 DOWN: Remove field descriptions
-- Note: PostgreSQL COMMENT ON COLUMN can be removed by setting to NULL

-- Manufacturing Orders
COMMENT ON COLUMN manufacturing_orders.cono IS NULL;
COMMENT ON COLUMN manufacturing_orders.divi IS NULL;
COMMENT ON COLUMN manufacturing_orders.faci IS NULL;
-- (Add all other fields if rollback needed, but typically comments are harmless)

-- Planned Manufacturing Orders
COMMENT ON COLUMN planned_manufacturing_orders.cono IS NULL;
COMMENT ON COLUMN planned_manufacturing_orders.divi IS NULL;
COMMENT ON COLUMN planned_manufacturing_orders.faci IS NULL;
-- (Add all other fields if rollback needed)

-- Production Orders
COMMENT ON COLUMN production_orders.order_type IS NULL;
COMMENT ON COLUMN production_orders.order_number IS NULL;
-- (Add all other fields if rollback needed)
