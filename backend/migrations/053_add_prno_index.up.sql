-- Migration 053: Add performance indexes for bulk operations with PRNO filtering
-- This migration adds indexes to optimize criteria-based bulk operations on detected issues

-- Index for product number (PRNO) filtering on production_orders
-- Used when filtering issues by product number in bulk operations
CREATE INDEX IF NOT EXISTS idx_production_orders_prno
ON production_orders(prno, environment)
WHERE prno IS NOT NULL;

-- Composite index for common filter combinations on detected_issues
-- Used when filtering issues by detector type, facility, and warehouse
-- Covers the most common query patterns for bulk operations
CREATE INDEX IF NOT EXISTS idx_detected_issues_filter_combo
ON detected_issues(environment, detector_type, facility, warehouse);
