-- Global API throttling settings
INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints, created_at)
VALUES
    -- TRN environment
    ('TRN', 'api_throttle_requests_per_second', '20', 'integer', 'Maximum M3 API requests per second across all workers', 'api_throttling', '{"min": 1, "max": 100}', NOW()),
    ('TRN', 'api_throttle_burst_size', '10', 'integer', 'Burst size for token bucket rate limiter', 'api_throttling', '{"min": 1, "max": 50}', NOW()),
    ('TRN', 'bulk_operation_batch_size', '50', 'integer', 'Number of items per batch in bulk operations', 'bulk_operations', '{"min": 1, "max": 100}', NOW()),

    -- PRD environment (more conservative)
    ('PRD', 'api_throttle_requests_per_second', '10', 'integer', 'Maximum M3 API requests per second across all workers', 'api_throttling', '{"min": 1, "max": 100}', NOW()),
    ('PRD', 'api_throttle_burst_size', '5', 'integer', 'Burst size for token bucket rate limiter', 'api_throttling', '{"min": 1, "max": 50}', NOW()),
    ('PRD', 'bulk_operation_batch_size', '50', 'integer', 'Number of items per batch in bulk operations', 'bulk_operations', '{"min": 1, "max": 100}', NOW());
