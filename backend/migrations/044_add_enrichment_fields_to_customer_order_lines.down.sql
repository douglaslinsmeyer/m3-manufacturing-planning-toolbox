DROP INDEX IF EXISTS idx_customer_order_lines_cuno;

ALTER TABLE customer_order_lines
    DROP COLUMN IF EXISTS delivery_method,
    DROP COLUMN IF EXISTS co_type_description,
    DROP COLUMN IF EXISTS customer_name;
