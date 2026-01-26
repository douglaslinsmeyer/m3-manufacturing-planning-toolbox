-- ========================================
-- Reorganize System Settings Categories
-- ========================================
-- Move data refresh settings to dedicated 'data_refresh' category
-- for better organization and clarity

UPDATE system_settings
SET category = 'data_refresh'
WHERE setting_key IN (
    -- Compass Data Fabric batching and pagination settings
    'compass_batch_size',
    'compass_page_size',
    'compass_over_partition_factor',
    'compass_api_timeout',
    'compass_max_query_records',

    -- Legacy snapshot settings (deprecated but kept for backward compatibility)
    'snapshot_concurrent_batches',
    'snapshot_batch_size'
);

-- Update table comment to reflect new organization
COMMENT ON TABLE system_settings IS 'System-wide configuration settings organized by functional category';

-- Log the reorganization
DO $$
DECLARE
    refresh_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO refresh_count
    FROM system_settings
    WHERE category = 'data_refresh';

    RAISE NOTICE 'Reorganized % settings into data_refresh category', refresh_count;
END $$;
