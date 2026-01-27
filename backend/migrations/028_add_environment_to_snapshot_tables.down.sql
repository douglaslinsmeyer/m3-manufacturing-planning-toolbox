-- ========================================
-- Migration 028 Rollback: Remove Environment Column
-- ========================================
-- Reverses all changes from the up migration
-- Order is important to avoid constraint violations

-- ========================================
-- STEP 1: Restore original unique constraints
-- ========================================

-- system_settings
ALTER TABLE system_settings DROP CONSTRAINT IF EXISTS unique_system_setting;
ALTER TABLE system_settings ADD CONSTRAINT system_settings_setting_key_key UNIQUE (setting_key);

-- user_settings
ALTER TABLE user_settings DROP CONSTRAINT IF EXISTS user_settings_pkey;
ALTER TABLE user_settings ADD PRIMARY KEY (user_id);

-- ignored_issues
ALTER TABLE ignored_issues DROP CONSTRAINT IF EXISTS unique_ignored_issue;
ALTER TABLE ignored_issues ADD CONSTRAINT ignored_issues_facility_detector_type_issue_key_production_or_key
    UNIQUE (facility, detector_type, issue_key, production_order_number);

-- mo_operations
ALTER TABLE mo_operations DROP CONSTRAINT IF EXISTS unique_mo_operation;
ALTER TABLE mo_operations ADD CONSTRAINT unique_mo_operation UNIQUE (faci, mfno, opno);

-- production_orders
ALTER TABLE production_orders DROP CONSTRAINT IF EXISTS unique_order_number;
ALTER TABLE production_orders ADD CONSTRAINT unique_order_number UNIQUE (order_number, order_type);

-- customer_order_lines
ALTER TABLE customer_order_lines DROP CONSTRAINT IF EXISTS unique_order_line;
ALTER TABLE customer_order_lines ADD CONSTRAINT unique_order_line UNIQUE (orno, ponr, posx);

-- customer_orders
ALTER TABLE customer_orders DROP CONSTRAINT IF EXISTS unique_customer_order_number;
ALTER TABLE customer_orders ADD CONSTRAINT unique_customer_order_number UNIQUE (orno);

-- manufacturing_orders
ALTER TABLE manufacturing_orders DROP CONSTRAINT IF EXISTS unique_mo_number;
ALTER TABLE manufacturing_orders ADD CONSTRAINT unique_mo_number UNIQUE (faci, mfno);

-- planned_manufacturing_orders
ALTER TABLE planned_manufacturing_orders DROP CONSTRAINT IF EXISTS unique_mop_number;
ALTER TABLE planned_manufacturing_orders ADD CONSTRAINT unique_mop_number UNIQUE (plpn);

-- ========================================
-- STEP 2: Drop indexes
-- ========================================

DROP INDEX IF EXISTS idx_issue_detection_jobs_env;
DROP INDEX IF EXISTS idx_audit_log_env;
DROP INDEX IF EXISTS idx_system_settings_env;
DROP INDEX IF EXISTS idx_user_settings_env;
DROP INDEX IF EXISTS idx_deliveries_env;
DROP INDEX IF EXISTS idx_mo_materials_env;
DROP INDEX IF EXISTS idx_mo_operations_env;
DROP INDEX IF EXISTS idx_ignored_issues_env;
DROP INDEX IF EXISTS idx_detected_issues_env;
DROP INDEX IF EXISTS idx_production_orders_env;
DROP INDEX IF EXISTS idx_customer_order_lines_env;
DROP INDEX IF EXISTS idx_customer_orders_env;
DROP INDEX IF EXISTS idx_manufacturing_orders_env;
DROP INDEX IF EXISTS idx_planned_mfg_orders_env;

-- ========================================
-- STEP 3: Remove environment column from all tables
-- ========================================

-- Audit and tracking
ALTER TABLE issue_detection_jobs DROP COLUMN IF EXISTS environment;
ALTER TABLE audit_log DROP COLUMN IF EXISTS environment;

-- Settings tables
ALTER TABLE system_settings DROP COLUMN IF EXISTS environment;
ALTER TABLE user_settings DROP COLUMN IF EXISTS environment;

-- Child tables
ALTER TABLE deliveries DROP COLUMN IF EXISTS environment;
ALTER TABLE mo_materials DROP COLUMN IF EXISTS environment;
ALTER TABLE mo_operations DROP COLUMN IF EXISTS environment;

-- Snapshot tables
ALTER TABLE ignored_issues DROP COLUMN IF EXISTS environment;
ALTER TABLE detected_issues DROP COLUMN IF EXISTS environment;
ALTER TABLE production_orders DROP COLUMN IF EXISTS environment;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS environment;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS environment;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS environment;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS environment;
