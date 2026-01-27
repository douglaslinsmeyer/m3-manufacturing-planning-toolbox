-- ========================================
-- REMOVE DETECTOR CONFIGURATION SETTINGS
-- Preserves detector enabled toggles
-- ========================================

DELETE FROM system_settings
WHERE setting_key IN (
  'detector_production_timing_days_early',
  'detector_production_timing_days_late',
  'detector_production_timing_exclude_mo_statuses',
  'detector_production_timing_exclude_mop_statuses',
  'detector_start_date_mismatch_tolerance_days',
  'detector_start_date_mismatch_exclude_mo_statuses',
  'detector_start_date_mismatch_exclude_mop_statuses',
  'detector_start_date_mismatch_min_quantity_threshold',
  'detector_unlinked_orders_exclude_mo_statuses',
  'detector_unlinked_orders_exclude_mop_statuses',
  'detector_unlinked_orders_min_order_age_days',
  'detector_unlinked_orders_exclude_facilities',
  'detector_unlinked_orders_min_quantity_threshold'
);
