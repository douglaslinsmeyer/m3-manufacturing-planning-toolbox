-- Remove unused global filter settings for joint delivery detector
-- These settings were defined but never implemented in the detector code
DELETE FROM system_settings
WHERE setting_key IN (
    'detector_joint_delivery_date_mismatch_exclude_mo_statuses',
    'detector_joint_delivery_date_mismatch_exclude_mop_statuses'
)
AND environment IN ('TRN', 'PRD');
