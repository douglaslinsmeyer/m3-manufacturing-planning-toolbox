-- ========================================
-- Revert Compass Settings Consolidation
-- ========================================

-- Restore compass_batch_size original description
UPDATE system_settings
SET
    description = 'Target records per batch for parallel processing (Spark-optimized ID range batching)',
    constraints = '{"min": 10000, "max": 100000, "unit": "records"}'::jsonb
WHERE setting_key = 'compass_batch_size';

-- Restore compass_page_size setting
INSERT INTO system_settings (setting_key, setting_value, setting_type, description, category, constraints)
VALUES (
    'compass_page_size', '10000', 'integer',
    'Page size for Data Fabric result fetching per batch (max: 100000 per API limits)',
    'data_refresh',
    '{"min": 1000, "max": 100000, "unit": "records"}'::jsonb
)
ON CONFLICT (environment, setting_key) DO NOTHING;
