-- Migration 014 DOWN: Revert to typed schema
-- WARNING: This drops all data - restore from migration 010 by re-running migrations

DROP TABLE IF EXISTS production_orders CASCADE;
DROP TABLE IF EXISTS manufacturing_orders CASCADE;
DROP TABLE IF EXISTS planned_manufacturing_orders CASCADE;

-- To restore: re-run migrations 010-013
