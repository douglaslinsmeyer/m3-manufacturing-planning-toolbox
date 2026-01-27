-- ========================================
-- Migration 028: Add Environment Column to All Data Tables
-- ========================================
-- This migration adds complete environment isolation by adding an environment
-- column to all snapshot, settings, and audit tables.
--
-- Tables affected:
-- - Snapshot tables (MOPs, MOs, COs, CO lines, production_orders, detected_issues, ignored_issues)
-- - Child tables (mo_operations, mo_materials, deliveries)
-- - Settings (user_settings, system_settings)
-- - Audit (audit_log, issue_detection_jobs)

-- ========================================
-- STEP 1: Add environment column to all tables
-- ========================================

-- Snapshot tables
ALTER TABLE planned_manufacturing_orders ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE manufacturing_orders ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE customer_orders ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE customer_order_lines ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE production_orders ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE detected_issues ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE ignored_issues ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';

-- Child tables (inherit from parent relationships, but need explicit column)
ALTER TABLE mo_operations ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE mo_materials ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE deliveries ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';

-- Settings tables
ALTER TABLE user_settings ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';
ALTER TABLE system_settings ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';

-- Audit and tracking (audit_log is nullable for backwards compatibility)
ALTER TABLE audit_log ADD COLUMN environment VARCHAR(10);
ALTER TABLE issue_detection_jobs ADD COLUMN environment VARCHAR(10) NOT NULL DEFAULT 'TRN';

-- ========================================
-- STEP 2: Create indexes for efficient environment filtering
-- ========================================

CREATE INDEX idx_planned_mfg_orders_env ON planned_manufacturing_orders(environment);
CREATE INDEX idx_manufacturing_orders_env ON manufacturing_orders(environment);
CREATE INDEX idx_customer_orders_env ON customer_orders(environment);
CREATE INDEX idx_customer_order_lines_env ON customer_order_lines(environment);
CREATE INDEX idx_production_orders_env ON production_orders(environment);
CREATE INDEX idx_detected_issues_env ON detected_issues(environment);
CREATE INDEX idx_ignored_issues_env ON ignored_issues(environment);
CREATE INDEX idx_mo_operations_env ON mo_operations(environment);
CREATE INDEX idx_mo_materials_env ON mo_materials(environment);
CREATE INDEX idx_deliveries_env ON deliveries(environment);
CREATE INDEX idx_user_settings_env ON user_settings(environment);
CREATE INDEX idx_system_settings_env ON system_settings(environment);
CREATE INDEX idx_audit_log_env ON audit_log(environment) WHERE environment IS NOT NULL;
CREATE INDEX idx_issue_detection_jobs_env ON issue_detection_jobs(environment);

-- ========================================
-- STEP 3: Update unique constraints to include environment
-- ========================================

-- planned_manufacturing_orders: (plpn) -> (environment, plpn)
ALTER TABLE planned_manufacturing_orders DROP CONSTRAINT IF EXISTS unique_mop_number;
ALTER TABLE planned_manufacturing_orders ADD CONSTRAINT unique_mop_number UNIQUE (environment, plpn);

-- manufacturing_orders: (faci, mfno) -> (environment, faci, mfno)
ALTER TABLE manufacturing_orders DROP CONSTRAINT IF EXISTS unique_mo_number;
ALTER TABLE manufacturing_orders ADD CONSTRAINT unique_mo_number UNIQUE (environment, faci, mfno);

-- customer_orders: (orno) -> (environment, orno)
ALTER TABLE customer_orders DROP CONSTRAINT IF EXISTS unique_customer_order_number;
ALTER TABLE customer_orders ADD CONSTRAINT unique_customer_order_number UNIQUE (environment, orno);

-- customer_order_lines: (orno, ponr, posx) -> (environment, orno, ponr, posx)
ALTER TABLE customer_order_lines DROP CONSTRAINT IF EXISTS unique_order_line;
ALTER TABLE customer_order_lines ADD CONSTRAINT unique_order_line UNIQUE (environment, orno, ponr, posx);

-- production_orders: (order_number, order_type) -> (environment, order_number, order_type)
ALTER TABLE production_orders DROP CONSTRAINT IF EXISTS unique_order_number;
ALTER TABLE production_orders ADD CONSTRAINT unique_order_number UNIQUE (environment, order_number, order_type);

-- mo_operations: (facility, mo_number, operation_number) -> (environment, facility, mo_number, operation_number)
ALTER TABLE mo_operations DROP CONSTRAINT IF EXISTS unique_mo_operation;
ALTER TABLE mo_operations ADD CONSTRAINT unique_mo_operation UNIQUE (environment, facility, mo_number, operation_number);

-- ignored_issues: (facility, detector_type, issue_key, production_order_number) -> add environment
ALTER TABLE ignored_issues DROP CONSTRAINT IF EXISTS ignored_issues_facility_detector_type_issue_key_production_or_key;
ALTER TABLE ignored_issues ADD CONSTRAINT unique_ignored_issue UNIQUE (environment, facility, detector_type, issue_key, production_order_number);

-- user_settings: (user_id) -> (environment, user_id)
ALTER TABLE user_settings DROP CONSTRAINT IF EXISTS user_settings_pkey;
ALTER TABLE user_settings ADD PRIMARY KEY (environment, user_id);

-- system_settings: (setting_key) -> (environment, setting_key)
ALTER TABLE system_settings DROP CONSTRAINT IF EXISTS system_settings_setting_key_key;
ALTER TABLE system_settings ADD CONSTRAINT unique_system_setting UNIQUE (environment, setting_key);

-- ========================================
-- STEP 4: Add comments for documentation
-- ========================================

COMMENT ON COLUMN planned_manufacturing_orders.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN manufacturing_orders.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN customer_orders.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN customer_order_lines.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN production_orders.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN detected_issues.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN ignored_issues.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN mo_operations.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN mo_materials.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN deliveries.environment IS 'M3 environment (TRN or PRD)';
COMMENT ON COLUMN user_settings.environment IS 'M3 environment (TRN or PRD) for environment-specific user preferences';
COMMENT ON COLUMN system_settings.environment IS 'M3 environment (TRN or PRD) for environment-specific system configuration';
COMMENT ON COLUMN audit_log.environment IS 'M3 environment where action occurred (nullable for backwards compatibility)';
COMMENT ON COLUMN issue_detection_jobs.environment IS 'M3 environment (TRN or PRD)';
