-- ========================================
-- Rollback: Restore Removed System Settings
-- ========================================
-- Restores settings that were removed in the up migration
-- Uses original values from migration 023_add_settings_tables.up.sql

INSERT INTO system_settings (setting_key, setting_value, setting_type, description, category, constraints) VALUES
    -- M3 API Settings
    ('m3_api_timeout', '30', 'integer', 'M3 API request timeout in seconds', 'integration', '{"min": 5, "max": 120, "unit": "seconds"}'::jsonb),
    ('m3_api_retry_attempts', '3', 'integer', 'Number of retry attempts for failed M3 API calls', 'integration', '{"min": 0, "max": 5}'::jsonb),
    ('m3_api_retry_delay', '1000', 'integer', 'Delay between retry attempts in milliseconds', 'integration', '{"min": 100, "max": 5000, "unit": "milliseconds"}'::jsonb),

    -- Compass SQL Settings
    ('compass_api_timeout', '60', 'integer', 'Compass SQL query timeout in seconds', 'integration', '{"min": 10, "max": 300, "unit": "seconds"}'::jsonb),
    ('compass_max_query_records', '100000', 'integer', 'Maximum records per Compass SQL query', 'integration', '{"min": 1000, "max": 500000}'::jsonb),

    -- Rate Limiting
    ('rate_limit_enabled', 'true', 'boolean', 'Enable API rate limiting', 'performance', '{}'::jsonb),
    ('rate_limit_requests_per_minute', '60', 'integer', 'Maximum API requests per user per minute', 'performance', '{"min": 10, "max": 1000}'::jsonb),

    -- Snapshot/Refresh Settings
    ('snapshot_concurrent_batches', '3', 'integer', 'Number of concurrent batches during snapshot refresh', 'performance', '{"min": 1, "max": 10}'::jsonb),
    ('snapshot_batch_size', '1000', 'integer', 'Records per batch during snapshot refresh', 'performance', '{"min": 100, "max": 5000}'::jsonb),

    -- Cache Settings
    ('cache_ttl_user_profiles', '900', 'integer', 'User profile cache TTL in seconds', 'performance', '{"min": 300, "max": 3600, "unit": "seconds"}'::jsonb),
    ('cache_ttl_m3_context', '3600', 'integer', 'M3 context cache TTL in seconds', 'performance', '{"min": 600, "max": 86400, "unit": "seconds"}'::jsonb)
ON CONFLICT (setting_key) DO NOTHING;

-- Update categories for legacy settings (from migration 025)
UPDATE system_settings
SET category = 'data_refresh'
WHERE setting_key IN (
    'compass_api_timeout',
    'compass_max_query_records',
    'snapshot_concurrent_batches',
    'snapshot_batch_size'
);

-- Log summary
DO $$
DECLARE
    total_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_count FROM system_settings;
    RAISE NOTICE 'Restored 11 settings. Total settings: %', total_count;
END $$;
