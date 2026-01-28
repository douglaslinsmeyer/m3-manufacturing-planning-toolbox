-- ========================================
-- Migration 037 Rollback: Restore Non-Implemented User Settings
-- ========================================
-- Restores the columns removed in the up migration

-- Restore display preference columns
ALTER TABLE user_settings ADD COLUMN items_per_page INTEGER DEFAULT 20 CHECK (items_per_page > 0 AND items_per_page <= 200);
ALTER TABLE user_settings ADD COLUMN theme VARCHAR(20) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto'));
ALTER TABLE user_settings ADD COLUMN date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD';
ALTER TABLE user_settings ADD COLUMN time_format VARCHAR(20) DEFAULT '24h' CHECK (time_format IN ('12h', '24h'));

-- Restore notification columns
ALTER TABLE user_settings ADD COLUMN enable_notifications BOOLEAN DEFAULT true;
ALTER TABLE user_settings ADD COLUMN notification_sound BOOLEAN DEFAULT false;

-- Restore flexible preferences JSONB
ALTER TABLE user_settings ADD COLUMN preferences JSONB DEFAULT '{}'::jsonb;

-- Restore original table comment
COMMENT ON TABLE user_settings IS 'User-specific preferences and default contexts';
