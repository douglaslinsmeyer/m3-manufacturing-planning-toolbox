-- Remove API throttle and bulk operation settings
DELETE FROM system_settings
WHERE setting_key IN (
    'api_throttle_requests_per_second',
    'api_throttle_burst_size',
    'bulk_operation_batch_size'
);
