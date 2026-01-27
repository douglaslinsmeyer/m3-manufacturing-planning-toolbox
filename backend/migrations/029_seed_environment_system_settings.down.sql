-- ========================================
-- Migration 029 Rollback: Remove PRD System Settings
-- ========================================
-- Removes PRD environment settings, keeping only TRN

-- Delete all PRD settings
DELETE FROM system_settings WHERE environment = 'PRD';

-- Restore original table comment
COMMENT ON TABLE system_settings IS 'System-wide configuration settings (admin-only)';

-- Verify deletion
DO $$
DECLARE
    prd_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO prd_count FROM system_settings WHERE environment = 'PRD';

    RAISE NOTICE 'PRD environment settings removed';
    RAISE NOTICE '  Remaining PRD settings: %', prd_count;

    IF prd_count > 0 THEN
        RAISE WARNING 'Failed to delete all PRD settings. % settings remain.', prd_count;
    END IF;
END $$;
