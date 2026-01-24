-- Migration 018: Add linked CO fields to production_orders
-- These fields are present in both manufacturing_orders and planned_manufacturing_orders
-- and should be included in the unified production_orders view

ALTER TABLE production_orders
    ADD COLUMN IF NOT EXISTS linked_co_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS linked_co_line VARCHAR(50),
    ADD COLUMN IF NOT EXISTS linked_co_suffix VARCHAR(50),
    ADD COLUMN IF NOT EXISTS allocated_qty VARCHAR(30);

-- Create index for CO lookups
CREATE INDEX IF NOT EXISTS idx_prod_linked_co ON production_orders(linked_co_number, linked_co_line);

-- Add comments
COMMENT ON COLUMN production_orders.linked_co_number IS 'Customer order number from MPREAL join (inherited from MO/MOP)';
COMMENT ON COLUMN production_orders.linked_co_line IS 'Customer order line from MPREAL join (inherited from MO/MOP)';
COMMENT ON COLUMN production_orders.linked_co_suffix IS 'Customer order line suffix from MPREAL join (inherited from MO/MOP)';
COMMENT ON COLUMN production_orders.allocated_qty IS 'Allocated quantity from MPREAL join (inherited from MO/MOP)';
