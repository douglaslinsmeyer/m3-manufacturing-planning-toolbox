-- ========================================
-- Drop Duplicate Production Order Constraint
-- ========================================
-- Remove the incorrect unique_production_order constraint
-- Keep the correct unique_order_number constraint (environment, order_number, order_type)

ALTER TABLE production_orders DROP CONSTRAINT IF EXISTS unique_production_order;

-- unique_order_number (environment, order_number, order_type) is the correct constraint
