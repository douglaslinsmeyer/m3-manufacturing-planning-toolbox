-- ========================================
-- Consolidate Compass Settings
-- ========================================
-- Remove compass_page_size and update compass_batch_size to handle both partitioning and result fetching
-- Both Data Fabric endpoints support up to 100K records, so one setting is sufficient

-- Update compass_batch_size description and constraints
UPDATE system_settings
SET
    description = 'Records per batch for parallel processing. Controls both ID range partitioning and result fetch size (max: 100000)',
    constraints = '{"min": 10000, "max": 100000, "unit": "records"}'::jsonb
WHERE setting_key = 'compass_batch_size';

-- Remove compass_page_size (no longer needed)
DELETE FROM system_settings WHERE setting_key = 'compass_page_size';

-- Note: Both /jobs/ (submit) and /jobs/{id}/result/ (fetch) endpoints support up to 100K records
