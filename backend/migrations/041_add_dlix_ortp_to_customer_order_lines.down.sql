-- Rollback Migration 041: Remove DLIX and ORTP columns from customer_order_lines

DROP INDEX IF EXISTS idx_customer_order_lines_dlix;

ALTER TABLE customer_order_lines
DROP COLUMN IF EXISTS ortp,
DROP COLUMN IF EXISTS dlix;
