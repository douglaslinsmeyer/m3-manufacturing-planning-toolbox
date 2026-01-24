-- Migration 010 DOWN: Rollback to previous schema
-- WARNING: This migration drops all manufacturing order data
-- To rollback safely, you must restore from backup or re-run migrations 001-009

DROP TABLE IF EXISTS production_orders CASCADE;
DROP TABLE IF EXISTS manufacturing_orders CASCADE;
DROP TABLE IF EXISTS planned_manufacturing_orders CASCADE;

-- NOTE: To restore the previous schema, you must:
-- 1. Restore database from backup, OR
-- 2. Re-run migrations 001-009 from scratch
--
-- This migration intentionally does not recreate the old schema
-- to avoid potential data corruption issues from schema mismatches.
