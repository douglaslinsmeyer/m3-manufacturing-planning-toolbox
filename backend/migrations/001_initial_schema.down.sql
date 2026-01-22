-- Rollback initial schema

DROP TRIGGER IF EXISTS update_deliveries_updated_at ON deliveries;
DROP TRIGGER IF EXISTS update_customer_order_lines_updated_at ON customer_order_lines;
DROP TRIGGER IF EXISTS update_customer_orders_updated_at ON customer_orders;
DROP TRIGGER IF EXISTS update_planned_manufacturing_orders_updated_at ON planned_manufacturing_orders;
DROP TRIGGER IF EXISTS update_mo_materials_updated_at ON mo_materials;
DROP TRIGGER IF EXISTS update_mo_operations_updated_at ON mo_operations;
DROP TRIGGER IF EXISTS update_manufacturing_orders_updated_at ON manufacturing_orders;
DROP TRIGGER IF EXISTS update_production_orders_updated_at ON production_orders;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS snapshot_metadata;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS customer_order_lines;
DROP TABLE IF EXISTS customer_orders;
DROP TABLE IF EXISTS planned_manufacturing_orders;
DROP TABLE IF EXISTS mo_materials;
DROP TABLE IF EXISTS mo_operations;
DROP TABLE IF EXISTS manufacturing_orders;
DROP TABLE IF EXISTS production_orders;
