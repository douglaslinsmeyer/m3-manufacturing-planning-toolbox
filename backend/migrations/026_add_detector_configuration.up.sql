-- ========================================
-- DETECTOR CONFIGURATION SETTINGS
-- Add hierarchical and global settings for all detectors
-- ========================================

INSERT INTO system_settings (setting_key, setting_value, setting_type, description, category, constraints) VALUES

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
