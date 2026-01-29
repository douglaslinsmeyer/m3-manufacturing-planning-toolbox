-- Add delivery_method_description column to customer_order_lines table
-- This will store the human-readable description from CSYTAB for delivery method codes (MODL)

ALTER TABLE customer_order_lines
ADD COLUMN delivery_method_description VARCHAR(15);

-- No index needed - this is a display-only field used for enrichment, not filtering
