-- Drop unused tables that were created but never implemented
-- These tables have no active code references and are not part of the application functionality

-- Drop co_id foreign key column from customer_order_lines (references customer_orders which is being dropped)
-- First drop the indexes related to co_id
DROP INDEX IF EXISTS idx_co_lines_no_header; -- Created in migration 004
DROP INDEX IF EXISTS idx_co_lines_co_id;     -- Created in migration 016
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS co_id;

-- Drop unused parent table for customer orders
DROP TABLE IF EXISTS customer_orders CASCADE;

-- Drop unused delivery tracking table
DROP TABLE IF EXISTS deliveries CASCADE;

-- Drop unused manufacturing order operation details table
DROP TABLE IF EXISTS mo_operations CASCADE;

-- Drop unused manufacturing order material requirements table
DROP TABLE IF EXISTS mo_materials CASCADE;

-- Drop superseded snapshot metadata table (replaced by refresh_jobs)
DROP TABLE IF EXISTS snapshot_metadata CASCADE;
