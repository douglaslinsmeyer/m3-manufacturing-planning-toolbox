-- Remove JDCD index
DROP INDEX IF EXISTS idx_customer_order_lines_jdcd;

-- Remove JDCD column
ALTER TABLE customer_order_lines
DROP COLUMN IF EXISTS jdcd;
