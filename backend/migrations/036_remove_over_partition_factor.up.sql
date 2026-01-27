-- ========================================
-- Remove Over-Partitioning Factor Setting
-- ========================================
-- Remove compass_over_partition_factor as we're simplifying to 3 simple queries
-- No longer using ID range batching that required over-partitioning

DELETE FROM system_settings WHERE setting_key = 'compass_over_partition_factor';
