-- ========================================
-- Revert compass_page_size Constraint Fix
-- ========================================

UPDATE system_settings
SET
    description = 'Page size for Data Fabric result fetching per batch (max: 10000 per API limits)',
    constraints = '{"min": 1000, "max": 10000, "unit": "records"}'::jsonb
WHERE setting_key = 'compass_page_size';
