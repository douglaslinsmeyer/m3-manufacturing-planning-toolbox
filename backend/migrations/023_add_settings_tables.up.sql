-- ========================================
-- USER SETTINGS TABLE
-- ========================================
-- Stores user-specific preferences and defaults
CREATE TABLE IF NOT EXISTS user_settings (
    user_id VARCHAR(100) PRIMARY KEY, -- Matches user_profiles.user_id

    -- Context Defaults
    default_warehouse VARCHAR(10),
    default_facility VARCHAR(10),
    default_division VARCHAR(10),
    default_company VARCHAR(10),

    -- Display Preferences
    items_per_page INTEGER DEFAULT 20 CHECK (items_per_page > 0 AND items_per_page <= 200),
    theme VARCHAR(20) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    time_format VARCHAR(20) DEFAULT '24h' CHECK (time_format IN ('12h', '24h')),

    -- Notification Preferences
    enable_notifications BOOLEAN DEFAULT true,
    notification_sound BOOLEAN DEFAULT false,

    -- Additional flexible settings (for future expansion)
    preferences JSONB DEFAULT '{}'::jsonb,

    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for quick lookups
CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);

-- Comments
COMMENT ON TABLE user_settings IS 'User-specific preferences and default contexts';
COMMENT ON COLUMN user_settings.preferences IS 'Flexible JSONB for additional user preferences';

-- ========================================
-- SYSTEM SETTINGS TABLE
-- ========================================
-- Stores system-wide configuration
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    setting_type VARCHAR(20) NOT NULL CHECK (setting_type IN ('string', 'integer', 'float', 'boolean', 'json')),
    description TEXT,
    category VARCHAR(50) NOT NULL, -- 'integration', 'performance', 'security', etc.

    -- Validation constraints (stored as JSONB)
    constraints JSONB, -- e.g., {"min": 1, "max": 300, "unit": "seconds"}

    -- Audit fields
    last_modified_by VARCHAR(100), -- User ID who last modified
    last_modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for category-based queries
CREATE INDEX idx_system_settings_category ON system_settings(category);
CREATE INDEX idx_system_settings_key ON system_settings(setting_key);

-- Comments
COMMENT ON TABLE system_settings IS 'System-wide configuration settings (admin-only)';
COMMENT ON COLUMN system_settings.setting_type IS 'Data type for proper parsing: string, integer, float, boolean, json';
COMMENT ON COLUMN system_settings.constraints IS 'Validation rules stored as JSONB';

-- ========================================
-- SEED SYSTEM SETTINGS (Default Values)
-- ========================================
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

-- ========================================
-- UPDATE TRIGGER FOR user_settings
-- ========================================
CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
