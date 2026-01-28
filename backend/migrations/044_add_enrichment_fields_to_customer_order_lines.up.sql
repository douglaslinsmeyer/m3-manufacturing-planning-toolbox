-- Add reference data fields for display enrichment
ALTER TABLE customer_order_lines
    ADD COLUMN customer_name TEXT,
    ADD COLUMN co_type_description TEXT,
    ADD COLUMN delivery_method VARCHAR(10);

-- Add index on customer_number for lookups
CREATE INDEX IF NOT EXISTS idx_customer_order_lines_cuno
    ON customer_order_lines(cuno);
