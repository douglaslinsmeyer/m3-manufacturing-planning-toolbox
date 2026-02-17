-- Add MITMAS (Item Master) enrichment fields to all production order tables

-- Manufacturing Orders
ALTER TABLE manufacturing_orders
ADD COLUMN item_type VARCHAR(50),
ADD COLUMN item_description TEXT,
ADD COLUMN item_group VARCHAR(50),
ADD COLUMN product_group VARCHAR(50),
ADD COLUMN procurement_group VARCHAR(50),
ADD COLUMN group_technology_class VARCHAR(50);

-- Planned Manufacturing Orders
ALTER TABLE planned_manufacturing_orders
ADD COLUMN item_type VARCHAR(50),
ADD COLUMN item_description TEXT,
ADD COLUMN item_group VARCHAR(50),
ADD COLUMN product_group VARCHAR(50),
ADD COLUMN procurement_group VARCHAR(50),
ADD COLUMN group_technology_class VARCHAR(50);

-- Customer Order Lines
ALTER TABLE customer_order_lines
ADD COLUMN item_type VARCHAR(50),
ADD COLUMN item_description_master TEXT,  -- Separate from ITDS which already exists
ADD COLUMN item_group VARCHAR(50),
ADD COLUMN product_group VARCHAR(50),
ADD COLUMN procurement_group VARCHAR(50),
ADD COLUMN group_technology_class VARCHAR(50);

-- Add indexes for common filter fields
CREATE INDEX idx_mo_item_type ON manufacturing_orders(item_type) WHERE item_type IS NOT NULL;
CREATE INDEX idx_mo_item_group ON manufacturing_orders(item_group) WHERE item_group IS NOT NULL;

CREATE INDEX idx_pmo_item_type ON planned_manufacturing_orders(item_type) WHERE item_type IS NOT NULL;
CREATE INDEX idx_pmo_item_group ON planned_manufacturing_orders(item_group) WHERE item_group IS NOT NULL;

CREATE INDEX idx_col_item_type ON customer_order_lines(item_type) WHERE item_type IS NOT NULL;
CREATE INDEX idx_col_item_group ON customer_order_lines(item_group) WHERE item_group IS NOT NULL;

-- Add column comments for documentation
COMMENT ON COLUMN manufacturing_orders.item_type IS 'MITMAS.ITTY - Item type from Item Master';
COMMENT ON COLUMN manufacturing_orders.item_description IS 'MITMAS.ITDS - Item description from Item Master';
COMMENT ON COLUMN manufacturing_orders.item_group IS 'MITMAS.ITGR - Item group classification';
COMMENT ON COLUMN manufacturing_orders.product_group IS 'MITMAS.ITCL - Product group classification';
COMMENT ON COLUMN manufacturing_orders.procurement_group IS 'MITMAS.PRGP - Procurement group';
COMMENT ON COLUMN manufacturing_orders.group_technology_class IS 'MITMAS.GRTI - Group technology class';

COMMENT ON COLUMN planned_manufacturing_orders.item_type IS 'MITMAS.ITTY - Item type from Item Master';
COMMENT ON COLUMN planned_manufacturing_orders.item_description IS 'MITMAS.ITDS - Item description from Item Master';
COMMENT ON COLUMN planned_manufacturing_orders.item_group IS 'MITMAS.ITGR - Item group classification';
COMMENT ON COLUMN planned_manufacturing_orders.product_group IS 'MITMAS.ITCL - Product group classification';
COMMENT ON COLUMN planned_manufacturing_orders.procurement_group IS 'MITMAS.PRGP - Procurement group';
COMMENT ON COLUMN planned_manufacturing_orders.group_technology_class IS 'MITMAS.GRTI - Group technology class';

COMMENT ON COLUMN customer_order_lines.item_type IS 'MITMAS.ITTY - Item type from Item Master';
COMMENT ON COLUMN customer_order_lines.item_description_master IS 'MITMAS.ITDS - Item description from Item Master (separate from OOLINE.ITDS)';
COMMENT ON COLUMN customer_order_lines.item_group IS 'MITMAS.ITGR - Item group classification';
COMMENT ON COLUMN customer_order_lines.product_group IS 'MITMAS.ITCL - Product group classification';
COMMENT ON COLUMN customer_order_lines.procurement_group IS 'MITMAS.PRGP - Procurement group';
COMMENT ON COLUMN customer_order_lines.group_technology_class IS 'MITMAS.GRTI - Group technology class';
