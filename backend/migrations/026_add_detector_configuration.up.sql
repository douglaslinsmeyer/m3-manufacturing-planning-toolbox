-- ========================================
-- DETECTOR CONFIGURATION SETTINGS
-- Add hierarchical and global settings for all detectors
-- ========================================

INSERT INTO system_settings (setting_key, setting_value, setting_type, description, category, constraints) VALUES

-- ========================================
-- PRODUCTION TIMING DETECTOR
-- ========================================

    -- Hierarchical Thresholds
    ('detector_production_timing_days_early',
     '{"global": 3, "overrides": []}',
     'json',
     'Days before delivery to flag as too early (hierarchical: supports warehouse/facility/mo_type overrides)',
     'detection',
     '{"min": 0, "max": 30, "unit": "days", "hierarchical": true}'::jsonb),

    ('detector_production_timing_days_late',
     '{"global": 0, "overrides": []}',
     'json',
     'Days after delivery to flag as too late (hierarchical)',
     'detection',
     '{"min": 0, "max": 30, "unit": "days", "hierarchical": true}'::jsonb),

    -- Global Filters
    ('detector_production_timing_exclude_mo_statuses',
     '[]',
     'json',
     'Exclude MOs with these WHST status codes (e.g., ["10", "90"])',
     'detection',
     '{}'::jsonb),

    ('detector_production_timing_exclude_mop_statuses',
     '[]',
     'json',
     'Exclude MOPs with these PSTS status codes',
     'detection',
     '{}'::jsonb),

-- ========================================
-- START DATE MISMATCH DETECTOR
-- ========================================

    -- Hierarchical Threshold
    ('detector_start_date_mismatch_tolerance_days',
     '{"global": 0, "overrides": []}',
     'json',
     'Allow dates within ±N days to match (0 = exact match only, hierarchical)',
     'detection',
     '{"min": 0, "max": 7, "unit": "days", "hierarchical": true}'::jsonb),

    -- Global Filters
    ('detector_start_date_mismatch_exclude_mo_statuses',
     '[]',
     'json',
     'Exclude MOs from comparison with these WHST codes',
     'detection',
     '{}'::jsonb),

    ('detector_start_date_mismatch_exclude_mop_statuses',
     '[]',
     'json',
     'Exclude MOPs from comparison with these PSTS codes',
     'detection',
     '{}'::jsonb),

    ('detector_start_date_mismatch_min_quantity_threshold',
     '0',
     'float',
     'Only flag mismatches when order quantity >= threshold',
     'detection',
     '{"min": 0, "unit": "quantity"}'::jsonb),

-- ========================================
-- UNLINKED PRODUCTION ORDERS DETECTOR
-- ========================================

    -- Global Filters Only (no hierarchical thresholds)
    ('detector_unlinked_orders_exclude_mo_statuses',
     '["10"]',
     'json',
     'Exclude MOs with these WHST codes (default: skip preliminary)',
     'detection',
     '{}'::jsonb),

    ('detector_unlinked_orders_exclude_mop_statuses',
     '[]',
     'json',
     'Exclude MOPs with these PSTS codes',
     'detection',
     '{}'::jsonb),

    ('detector_unlinked_orders_min_order_age_days',
     '0',
     'integer',
     'Only flag unlinked orders older than N days (0 = all orders)',
     'detection',
     '{"min": 0, "max": 365, "unit": "days"}'::jsonb),

    ('detector_unlinked_orders_exclude_facilities',
     '[]',
     'json',
     'Exclude these facilities entirely (e.g., ["AZ2", "TX1"])',
     'detection',
     '{}'::jsonb),

    ('detector_unlinked_orders_min_quantity_threshold',
     '0',
     'float',
     'Only flag unlinked orders with quantity >= threshold (0 = all)',
     'detection',
     '{"min": 0, "unit": "quantity"}'::jsonb)

ON CONFLICT (setting_key) DO NOTHING;

-- ========================================
-- COMMENTS
-- ========================================
COMMENT ON TABLE system_settings IS 'System-wide configuration including hierarchical detector thresholds stored as JSONB';
