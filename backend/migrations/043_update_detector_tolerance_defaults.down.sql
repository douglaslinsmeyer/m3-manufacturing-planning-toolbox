-- ========================================
-- ROLLBACK: RESTORE DEFAULT TOLERANCE VALUES FROM 1 TO 0
-- ========================================
-- Reverts tolerance_days defaults back to 0 for both detectors
-- Only reverts records that haven't been manually modified

UPDATE system_settings
SET
    setting_value = '{"global": 0, "overrides": []}',
    last_modified_at = CURRENT_TIMESTAMP
WHERE setting_key IN (
    'detector_joint_delivery_date_mismatch_tolerance_days',
    'detector_dlix_date_mismatch_tolerance_days'
)
AND setting_value = '{"global": 1, "overrides": []}'
AND (last_modified_by IS NULL OR last_modified_by = '');
