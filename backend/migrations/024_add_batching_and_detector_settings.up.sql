-- ========================================
-- Add Parallel Batching and Detector Settings
-- ========================================
-- Adds system settings for Spark-optimized parallel batching
-- and detector enable/disable toggles

INSERT INTO system_settings (setting_key, setting_value, setting_type, description, category, constraints) VALUES
    -- Data Fabric Parallel Batching Settings
    ('compass_batch_size', '50000', 'integer',
     'Target records per batch for parallel processing (Spark-optimized ID range batching)',
     'performance',
     '{"min": 10000, "max": 100000, "unit": "records"}'::jsonb),

    ('compass_page_size', '10000', 'integer',
     'Page size for Data Fabric result fetching per batch (max: 10000 per API limits)',
     'performance',
     '{"min": 1000, "max": 10000, "unit": "records"}'::jsonb),

    ('compass_over_partition_factor', '1.5', 'float',
     'Over-partitioning multiplier to handle ID gaps from deleted records (e.g., 1.5 = create 50% more batches)',
     'performance',
     '{"min": 1.0, "max": 3.0}'::jsonb),

    -- Issue Detector Toggles
    ('detector_unlinked_production_orders_enabled', 'true', 'boolean',
     'Enable detection of production orders (MO/MOP) without customer order links',
     'detection', '{}'::jsonb),

    ('detector_start_date_mismatch_enabled', 'true', 'boolean',
     'Enable detection of production orders linked to same CO line with different start dates',
     'detection', '{}'::jsonb),

    ('detector_production_timing_enabled', 'true', 'boolean',
     'Enable detection of production orders with timing issues (start date too early or too late)',
     'detection', '{}'::jsonb)

ON CONFLICT (setting_key) DO NOTHING;

-- ========================================
-- Comments
-- ========================================
COMMENT ON TABLE system_settings IS 'System-wide configuration settings with batching and detector controls';
