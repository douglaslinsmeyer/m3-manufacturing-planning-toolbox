-- Rollback MITMAS (Item Master) enrichment fields

-- Drop indexes
DROP INDEX IF EXISTS idx_mo_item_type;
DROP INDEX IF EXISTS idx_mo_item_group;
DROP INDEX IF EXISTS idx_pmo_item_type;
DROP INDEX IF EXISTS idx_pmo_item_group;
DROP INDEX IF EXISTS idx_col_item_type;
DROP INDEX IF EXISTS idx_col_item_group;

-- Drop columns from Manufacturing Orders
ALTER TABLE manufacturing_orders
DROP COLUMN IF EXISTS item_type,
DROP COLUMN IF EXISTS item_description,
DROP COLUMN IF EXISTS item_group,
DROP COLUMN IF EXISTS product_group,
DROP COLUMN IF EXISTS procurement_group,
DROP COLUMN IF EXISTS group_technology_class;

-- Drop columns from Planned Manufacturing Orders
ALTER TABLE planned_manufacturing_orders
DROP COLUMN IF EXISTS item_type,
DROP COLUMN IF EXISTS item_description,
DROP COLUMN IF EXISTS item_group,
DROP COLUMN IF EXISTS product_group,
DROP COLUMN IF EXISTS procurement_group,
DROP COLUMN IF EXISTS group_technology_class;

-- Drop columns from Customer Order Lines
ALTER TABLE customer_order_lines
DROP COLUMN IF EXISTS item_type,
DROP COLUMN IF EXISTS item_description_master,
DROP COLUMN IF EXISTS item_group,
DROP COLUMN IF EXISTS product_group,
DROP COLUMN IF EXISTS procurement_group,
DROP COLUMN IF EXISTS group_technology_class;
