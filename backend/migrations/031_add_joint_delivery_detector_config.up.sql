-- ========================================
-- JOINT DELIVERY DATE MISMATCH DETECTOR
-- ========================================

INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints) VALUES

    -- TRN Environment Settings
    ('TRN', 'detector_joint_delivery_date_mismatch_tolerance_days',
     '{"global": 0, "overrides": []}',
     'json',
     'Allow dates within ±N days to match within JDCD group (0 = exact match only, hierarchical)',
     'detection',
     '{"min": 0, "max": 7, "unit": "days", "hierarchical": true}'::jsonb),

    ('TRN', 'detector_joint_delivery_date_mismatch_exclude_mo_statuses',
     '[]',
     'json',
     'Exclude MOs from comparison with these WHST codes',
     'detection',
     '{}'::jsonb),

    ('TRN', 'detector_joint_delivery_date_mismatch_exclude_mop_statuses',
     '[]',
     'json',
     'Exclude MOPs from comparison with these PSTS codes',
     'detection',
     '{}'::jsonb),

    -- PRD Environment Settings
    ('PRD', 'detector_joint_delivery_date_mismatch_tolerance_days',
     '{"global": 0, "overrides": []}',
     'json',
     'Allow dates within ±N days to match within JDCD group (0 = exact match only, hierarchical)',
     'detection',
     '{"min": 0, "max": 7, "unit": "days", "hierarchical": true}'::jsonb),

    ('PRD', 'detector_joint_delivery_date_mismatch_exclude_mo_statuses',
     '[]',
     'json',
     'Exclude MOs from comparison with these WHST codes',
     'detection',
     '{}'::jsonb),

    ('PRD', 'detector_joint_delivery_date_mismatch_exclude_mop_statuses',
     '[]',
     'json',
     'Exclude MOPs from comparison with these PSTS codes',
     'detection',
     '{}'::jsonb)

ON CONFLICT (environment, setting_key) DO NOTHING;
