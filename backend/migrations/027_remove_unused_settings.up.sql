-- ========================================
-- Remove Unused System Settings
-- ========================================
-- Removes settings that were defined but never actually used in application logic
-- This cleanup reduces confusion and maintenance burden

-- Remove unused M3 API settings (no code references these timeouts/retries)
DELETE FROM system_settings WHERE setting_key IN (
    'm3_api_timeout',
    'm3_api_retry_attempts',
    'm3_api_retry_delay'
);

-- Remove unused Compass SQL settings (only compass_batch_size, compass_page_size, compass_over_partition_factor are used)
DELETE FROM system_settings WHERE setting_key IN (
    'compass_api_timeout',
    'compass_max_query_records'
);

-- Remove unused rate limiting settings (no rate limiting middleware exists)
DELETE FROM system_settings WHERE setting_key IN (
    'rate_limit_enabled',
    'rate_limit_requests_per_minute'
);

-- Remove unused cache settings (no caching implementation uses these)
DELETE FROM system_settings WHERE setting_key IN (
    'cache_ttl_user_profiles',
    'cache_ttl_m3_context'
);

-- Remove deprecated snapshot settings (superseded by compass_batch_size)
DELETE FROM system_settings WHERE setting_key IN (
    'snapshot_concurrent_batches',
    'snapshot_batch_size'
);

-- Log summary
DO $$
DECLARE
    remaining_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO remaining_count FROM system_settings;
    RAISE NOTICE 'Removed 11 unused settings. % settings remaining.', remaining_count;
END $$;
