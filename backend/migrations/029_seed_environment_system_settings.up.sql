-- ========================================
-- Migration 029: Seed Environment-Specific System Settings
-- ========================================
-- Duplicates all existing system_settings for both TRN and PRD environments
-- This ensures both environments start with the same configuration baseline

-- First, ensure all existing settings are assigned to TRN environment
-- (This is a no-op since DEFAULT 'TRN' was set in migration 028, but explicit for clarity)
UPDATE system_settings SET environment = 'TRN' WHERE environment = 'TRN';

-- Then duplicate all TRN settings for PRD environment
INSERT INTO system_settings (
    environment,
    setting_key,
    setting_value,
    setting_type,
    description,
    category,
    constraints,
    last_modified_by,
    last_modified_at,
    created_at
)
SELECT
    'PRD' as environment,
    setting_key,
    setting_value,
    setting_type,
    description,
    category,
    constraints,
    last_modified_by,
    NOW() as last_modified_at,
    NOW() as created_at
FROM system_settings
WHERE environment = 'TRN'
ON CONFLICT (environment, setting_key) DO NOTHING;

-- Update table comment to reflect environment-specific nature
COMMENT ON TABLE system_settings IS 'Environment-specific system configuration (TRN/PRD isolated) - detector thresholds, API timeouts, and performance settings';

-- Verify duplication
DO $$
DECLARE
    trn_count INTEGER;
    prd_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO trn_count FROM system_settings WHERE environment = 'TRN';
    SELECT COUNT(*) INTO prd_count FROM system_settings WHERE environment = 'PRD';

    RAISE NOTICE 'System settings seeded:';
    RAISE NOTICE '  TRN environment: % settings', trn_count;
    RAISE NOTICE '  PRD environment: % settings', prd_count;

    IF trn_count != prd_count THEN
        RAISE WARNING 'TRN and PRD setting counts differ. Expected equal counts.';
    END IF;
END $$;
