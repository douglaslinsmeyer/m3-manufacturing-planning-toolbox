-- ========================================
-- Rollback Migration 042: Remove DLIX Date Mismatch Detector configuration settings
-- ========================================

DELETE FROM system_settings
WHERE setting_key IN (
  'detector_dlix_date_mismatch_enabled',
  'detector_dlix_date_mismatch_tolerance_days'
);
