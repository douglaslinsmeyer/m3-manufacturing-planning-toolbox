-- Migration 011 DOWN: Revert to human-readable column names

-- ============================================================================
-- Manufacturing Orders
-- ============================================================================

ALTER TABLE manufacturing_orders
    RENAME COLUMN faci TO facility;

ALTER TABLE manufacturing_orders
    RENAME COLUMN mfno TO mo_number;

ALTER TABLE manufacturing_orders
    RENAME COLUMN prno TO product_number;

ALTER TABLE manufacturing_orders
    RENAME COLUMN itno TO item_number;

ALTER TABLE manufacturing_orders
    RENAME CONSTRAINT unique_mfno TO unique_mo_number;

DROP INDEX IF EXISTS idx_mo_faci;
DROP INDEX IF EXISTS idx_mo_itno;

CREATE INDEX idx_mo_facility ON manufacturing_orders(facility);
CREATE INDEX idx_mo_item ON manufacturing_orders(item_number);

-- ============================================================================
-- Planned Manufacturing Orders
-- ============================================================================

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN faci TO facility;

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN prno TO product_number;

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN itno TO item_number;

DROP INDEX IF EXISTS idx_mop_faci;
DROP INDEX IF EXISTS idx_mop_itno;

CREATE INDEX idx_mop_facility ON planned_manufacturing_orders(facility);
CREATE INDEX idx_mop_item ON planned_manufacturing_orders(item_number);

-- ============================================================================
-- Production Orders
-- ============================================================================

ALTER TABLE production_orders
    RENAME COLUMN faci TO facility;

ALTER TABLE production_orders
    RENAME COLUMN prno TO product_number;

ALTER TABLE production_orders
    RENAME COLUMN itno TO item_number;

DROP INDEX IF EXISTS idx_prod_faci;
DROP INDEX IF EXISTS idx_prod_itno;

CREATE INDEX idx_prod_facility ON production_orders(facility);
CREATE INDEX idx_prod_item ON production_orders(item_number);
