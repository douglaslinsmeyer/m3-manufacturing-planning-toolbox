-- ========================================
-- Rollback: Remove Batching and Detector Settings
-- ========================================

DELETE FROM system_settings WHERE setting_key IN (
    -- Batching settings
    'compass_batch_size',
    'compass_page_size',
    'compass_over_partition_factor',

    -- Detector toggles
    'detector_unlinked_production_orders_enabled',
    'detector_start_date_mismatch_enabled',
    'detector_production_timing_enabled'
);
