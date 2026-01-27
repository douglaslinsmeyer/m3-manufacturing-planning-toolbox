-- ========================================
-- Fix compass_page_size Constraint
-- ========================================
-- Correct the max limit from 10000 to 100000 for Data Fabric result endpoint

UPDATE system_settings
SET
    description = 'Page size for Data Fabric result fetching per batch (max: 100000 per API limits)',
    constraints = '{"min": 1000, "max": 100000, "unit": "records"}'::jsonb
WHERE setting_key = 'compass_page_size';

-- Note: Data Fabric /jobs/{id}/result/ endpoint supports up to 100K records per call
