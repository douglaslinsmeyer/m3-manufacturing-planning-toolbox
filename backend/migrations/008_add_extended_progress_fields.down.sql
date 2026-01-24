-- Remove extended progress tracking fields

ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS records_per_second;
ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS estimated_seconds_remaining;
ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS current_operation;
ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS current_batch;
ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS total_batches;
