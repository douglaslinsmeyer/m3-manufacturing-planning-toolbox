-- Migration 011: Standardize column names to M3 conventions
-- Renames human-readable column names to match M3 field names
-- Exception: Keep aliased join columns (linked_co_*) human-readable

-- ============================================================================
-- Manufacturing Orders
-- ============================================================================

ALTER TABLE manufacturing_orders
    RENAME COLUMN facility TO faci;

ALTER TABLE manufacturing_orders
    RENAME COLUMN mo_number TO mfno;

ALTER TABLE manufacturing_orders
    RENAME COLUMN product_number TO prno;

ALTER TABLE manufacturing_orders
    RENAME COLUMN item_number TO itno;

-- Update constraint name to match
ALTER TABLE manufacturing_orders
    RENAME CONSTRAINT unique_mo_number TO unique_mfno;

-- Update indexes
DROP INDEX IF EXISTS idx_mo_facility;
DROP INDEX IF EXISTS idx_mo_item;

CREATE INDEX idx_mo_faci ON manufacturing_orders(faci);
CREATE INDEX idx_mo_itno ON manufacturing_orders(itno);

-- ============================================================================
-- Planned Manufacturing Orders
-- ============================================================================

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN facility TO faci;

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN product_number TO prno;

ALTER TABLE planned_manufacturing_orders
    RENAME COLUMN item_number TO itno;

-- Note: Keep linked_co_number, linked_co_line, linked_co_suffix, allocated_qty
-- These are aliased columns from MPREAL join and should remain human-readable

-- Update indexes
DROP INDEX IF EXISTS idx_mop_facility;
DROP INDEX IF EXISTS idx_mop_item;

CREATE INDEX idx_mop_faci ON planned_manufacturing_orders(faci);
CREATE INDEX idx_mop_itno ON planned_manufacturing_orders(itno);

-- ============================================================================
-- Production Orders (Unified View)
-- ============================================================================

ALTER TABLE production_orders
    RENAME COLUMN facility TO faci;

ALTER TABLE production_orders
    RENAME COLUMN product_number TO prno;

ALTER TABLE production_orders
    RENAME COLUMN item_number TO itno;

-- Note: Keep order_number, order_type, mo_id, mop_id as abstractions

-- Update indexes
DROP INDEX IF EXISTS idx_prod_facility;
DROP INDEX IF EXISTS idx_prod_item;

CREATE INDEX idx_prod_faci ON production_orders(faci);
CREATE INDEX idx_prod_itno ON production_orders(itno);
