-- Rollback anomaly detector settings

DELETE FROM system_settings
WHERE category = 'anomaly_detection'
  AND setting_key LIKE 'anomaly_%';
