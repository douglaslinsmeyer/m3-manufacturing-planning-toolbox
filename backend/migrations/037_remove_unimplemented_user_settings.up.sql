-- ========================================
-- Migration 037: Remove Non-Implemented User Settings
-- ========================================
-- Removes user_settings columns that have no backend implementation:
-- - items_per_page (frontend-only, no backend pagination)
-- - theme, date_format, time_format (frontend-only display preferences)
-- - enable_notifications, notification_sound (no notification system exists)
-- - preferences (never populated or used)
--
-- Keeps only the implemented settings:
-- - default_warehouse, default_facility, default_division, default_company

-- Drop unused columns
ALTER TABLE user_settings DROP COLUMN IF EXISTS items_per_page;
ALTER TABLE user_settings DROP COLUMN IF EXISTS theme;
ALTER TABLE user_settings DROP COLUMN IF EXISTS date_format;
ALTER TABLE user_settings DROP COLUMN IF EXISTS time_format;
ALTER TABLE user_settings DROP COLUMN IF EXISTS enable_notifications;
ALTER TABLE user_settings DROP COLUMN IF EXISTS notification_sound;
ALTER TABLE user_settings DROP COLUMN IF EXISTS preferences;

-- Update table comment to reflect simplified purpose
COMMENT ON TABLE user_settings IS 'User-specific default context overrides (company, division, facility, warehouse)';
