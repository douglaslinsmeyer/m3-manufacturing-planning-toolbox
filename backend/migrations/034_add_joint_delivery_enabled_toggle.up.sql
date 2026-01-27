-- ========================================
-- Add Enabled Toggle for Joint Delivery Date Mismatch Detector
-- ========================================
-- Adds the missing enabled toggle that other detectors have

INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints) VALUES
    -- TRN Environment
    ('TRN', 'detector_joint_delivery_date_mismatch_enabled', 'true', 'boolean',
     'Enable detection of production orders within same joint delivery group with mismatched delivery dates',
     'detection', '{}'::jsonb),

    -- PRD Environment
    ('PRD', 'detector_joint_delivery_date_mismatch_enabled', 'true', 'boolean',
     'Enable detection of production orders within same joint delivery group with mismatched delivery dates',
     'detection', '{}'::jsonb)

ON CONFLICT (environment, setting_key) DO NOTHING;
