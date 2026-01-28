-- ========================================
-- DLIX DATE MISMATCH DETECTOR
-- ========================================
-- Adds configuration settings for the DLIX (Delivery Number) Date Mismatch Detector
-- This detector identifies production orders within the same delivery (DLIX) that have misaligned start dates

INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints) VALUES

    -- TRN Environment Settings
    ('TRN', 'detector_dlix_date_mismatch_enabled',
     'true',
     'boolean',
     'Enable detection of production orders within same delivery (DLIX) with mismatched start dates',
     'detection',
     '{}'::jsonb),

    ('TRN', 'detector_dlix_date_mismatch_tolerance_days',
     '{"global": 0, "overrides": []}',
     'json',
     'Allow dates within ±N days to match within DLIX group (0 = exact match only, hierarchical)',
     'detection',
     '{"min": 0, "max": 7, "unit": "days", "hierarchical": true}'::jsonb),

    -- PRD Environment Settings
    ('PRD', 'detector_dlix_date_mismatch_enabled',
     'true',
     'boolean',
     'Enable detection of production orders within same delivery (DLIX) with mismatched start dates',
     'detection',
     '{}'::jsonb),

    ('PRD', 'detector_dlix_date_mismatch_tolerance_days',
     '{"global": 0, "overrides": []}',
     'json',
     'Allow dates within ±N days to match within DLIX group (0 = exact match only, hierarchical)',
     'detection',
     '{"min": 0, "max": 7, "unit": "days", "hierarchical": true}'::jsonb)

ON CONFLICT (environment, setting_key) DO NOTHING;
