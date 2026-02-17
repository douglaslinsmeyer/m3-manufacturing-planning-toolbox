-- Remove CO quantity mismatch detector settings
DELETE FROM system_settings
WHERE setting_key LIKE 'detector_co_quantity_mismatch_%';
