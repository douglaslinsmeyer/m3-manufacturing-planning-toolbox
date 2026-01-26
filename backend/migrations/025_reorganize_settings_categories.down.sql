-- ========================================
-- Rollback: Restore Original Categories
-- ========================================
-- Restore data refresh settings to their original categories

-- Restore Compass batching settings to 'performance'
UPDATE system_settings
SET category = 'performance'
WHERE setting_key IN (
    'compass_batch_size',
    'compass_page_size',
    'compass_over_partition_factor',
    'snapshot_concurrent_batches',
    'snapshot_batch_size'
);

-- Restore Compass API settings to 'integration'
UPDATE system_settings
SET category = 'integration'
WHERE setting_key IN (
    'compass_api_timeout',
    'compass_max_query_records'
);

-- Restore original table comment
COMMENT ON TABLE system_settings IS 'System-wide configuration settings with batching and detector controls';
