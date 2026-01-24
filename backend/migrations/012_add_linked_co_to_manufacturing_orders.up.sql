-- Migration 012: Add linked_co columns to manufacturing_orders
-- Add MPREAL-linked customer order fields to manufacturing_orders table

ALTER TABLE manufacturing_orders
    ADD COLUMN linked_co_number VARCHAR(50),
    ADD COLUMN linked_co_line INTEGER,
    ADD COLUMN linked_co_suffix INTEGER,
    ADD COLUMN allocated_qty DECIMAL(15,6);

-- Create index for CO lookups
CREATE INDEX idx_mo_linked_co ON manufacturing_orders(linked_co_number, linked_co_line);

-- Add comments explaining the source
COMMENT ON COLUMN manufacturing_orders.linked_co_number IS 'Customer order number from MPREAL.DRDN join';
COMMENT ON COLUMN manufacturing_orders.linked_co_line IS 'Customer order line from MPREAL.DRDL join';
COMMENT ON COLUMN manufacturing_orders.linked_co_suffix IS 'Customer order line suffix from MPREAL.DRDX join';
COMMENT ON COLUMN manufacturing_orders.allocated_qty IS 'Allocated quantity from MPREAL.PQTY join';
