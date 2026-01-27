-- ========================================
-- Restore Duplicate Production Order Constraint
-- ========================================

ALTER TABLE production_orders ADD CONSTRAINT unique_production_order UNIQUE (order_number);
