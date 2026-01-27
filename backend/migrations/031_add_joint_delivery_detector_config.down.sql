-- Remove Joint Delivery Date Mismatch detector configuration
DELETE FROM system_settings
WHERE setting_key LIKE 'detector_joint_delivery_date_mismatch_%';
