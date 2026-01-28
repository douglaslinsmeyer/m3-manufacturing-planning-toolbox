-- Migration 041: Add DLIX (Delivery Number) and ORTP (Order Type) to customer_order_lines
-- DLIX comes from M3 delivery tables (MHDISL/MHDISH), ORTP comes from order header (OOHEAD)

ALTER TABLE customer_order_lines
ADD COLUMN dlix VARCHAR(20),
ADD COLUMN ortp VARCHAR(10);

-- Add filtered index for DLIX grouping queries (matches JDCD pattern)
CREATE INDEX idx_customer_order_lines_dlix
ON customer_order_lines(orno, dlix)
WHERE dlix IS NOT NULL AND dlix != '';

-- Add documentation comments
COMMENT ON COLUMN customer_order_lines.dlix IS 'M3 Delivery Number from MHDISL - groups CO lines delivered together in same shipment';
COMMENT ON COLUMN customer_order_lines.ortp IS 'M3 Order Type from OOHEAD - categorizes customer order (standard, rush, return, etc)';
