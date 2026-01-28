-- ========================================
-- UPDATE DEFAULT TOLERANCE VALUES FROM 0 TO 1
-- ========================================
-- Updates both Joint Delivery and DLIX detector tolerance_days defaults from 0 to 1
-- Only updates records that are still at the original default (global: 0, no overrides, not modified by users)

UPDATE system_settings
SET
    setting_value = '{"global": 1, "overrides": []}',
    last_modified_at = CURRENT_TIMESTAMP
WHERE setting_key IN (
    'detector_joint_delivery_date_mismatch_tolerance_days',
    'detector_dlix_date_mismatch_tolerance_days'
)
AND setting_value = '{"global": 0, "overrides": []}'
AND (last_modified_by IS NULL OR last_modified_by = '');
