-- ========================================
-- REMOVE DETECTOR CONFIGURATION SETTINGS
-- Preserves detector enabled toggles
-- ========================================

DELETE FROM system_settings
WHERE setting_key IN (
  'detector_unlinked_orders_exclude_mo_statuses',
  'detector_unlinked_orders_exclude_mop_statuses',
  'detector_unlinked_orders_min_order_age_days',
  'detector_unlinked_orders_exclude_facilities',
  'detector_unlinked_orders_min_quantity_threshold'
);
