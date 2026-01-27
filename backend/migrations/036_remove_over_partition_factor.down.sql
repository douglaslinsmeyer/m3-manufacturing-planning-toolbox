-- ========================================
-- Restore Over-Partitioning Factor Setting
-- ========================================

INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints)
VALUES
    ('TRN', 'compass_over_partition_factor', '1.5', 'float',
     'Over-partitioning multiplier to handle ID gaps from deleted records (e.g., 1.5 = create 50% more batches)',
     'data_refresh', '{"min": 1.0, "max": 3.0}'::jsonb),
    ('PRD', 'compass_over_partition_factor', '1.5', 'float',
     'Over-partitioning multiplier to handle ID gaps from deleted records (e.g., 1.5 = create 50% more batches)',
     'data_refresh', '{"min": 1.0, "max": 3.0}'::jsonb)
ON CONFLICT (environment, setting_key) DO NOTHING;
