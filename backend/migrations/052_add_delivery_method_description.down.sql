-- Rollback: Remove delivery_method_description column from customer_order_lines table

ALTER TABLE customer_order_lines
DROP COLUMN delivery_method_description;
