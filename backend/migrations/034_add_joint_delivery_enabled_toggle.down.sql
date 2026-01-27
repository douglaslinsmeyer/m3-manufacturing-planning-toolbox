-- ========================================
-- Revert Joint Delivery Date Mismatch Detector Enabled Toggle
-- ========================================

DELETE FROM system_settings
WHERE setting_key = 'detector_joint_delivery_date_mismatch_enabled';
