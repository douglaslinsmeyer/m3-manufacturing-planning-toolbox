-- Make co_id nullable since we're doing production-centric loading
-- We may load CO lines without headers when they're referenced by MOs/MOPs

ALTER TABLE customer_order_lines
ALTER COLUMN co_id DROP NOT NULL;

-- Add index for lines without headers
CREATE INDEX IF NOT EXISTS idx_co_lines_no_header ON customer_order_lines(order_number) WHERE co_id IS NULL;
