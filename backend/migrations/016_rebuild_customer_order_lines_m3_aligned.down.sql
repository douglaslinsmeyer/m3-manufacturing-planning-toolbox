-- Migration 016 DOWN: Rollback customer_order_lines rebuild
-- WARNING: This drops all customer order line data

DROP TABLE IF EXISTS customer_order_lines CASCADE;

-- To restore previous schema, re-run migrations 001-015
