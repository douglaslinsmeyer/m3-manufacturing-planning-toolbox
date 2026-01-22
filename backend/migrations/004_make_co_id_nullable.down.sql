-- Rollback making co_id nullable

DROP INDEX IF EXISTS idx_co_lines_no_header;

-- Note: Cannot make NOT NULL again if there are NULL values
-- Would need to delete orphaned records first
ALTER TABLE customer_order_lines
ALTER COLUMN co_id SET NOT NULL;
