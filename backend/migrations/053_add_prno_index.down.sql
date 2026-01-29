-- Migration 053 rollback: Remove bulk operations indexes

DROP INDEX IF EXISTS idx_production_orders_prno;
DROP INDEX IF EXISTS idx_detected_issues_filter_combo;
