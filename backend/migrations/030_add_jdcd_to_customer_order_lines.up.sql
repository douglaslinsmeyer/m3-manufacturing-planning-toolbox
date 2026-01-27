-- Add JDCD (Joint Delivery Code) column to customer_order_lines table
ALTER TABLE customer_order_lines
ADD COLUMN jdcd VARCHAR(50);

-- Add index for efficient JDCD grouping queries
CREATE INDEX idx_customer_order_lines_jdcd ON customer_order_lines(orno, jdcd) WHERE jdcd IS NOT NULL AND jdcd != '';
