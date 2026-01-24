-- Add extended progress tracking fields for real-time SSE updates

ALTER TABLE refresh_jobs ADD COLUMN IF NOT EXISTS records_per_second REAL;
ALTER TABLE refresh_jobs ADD COLUMN IF NOT EXISTS estimated_seconds_remaining INTEGER;
ALTER TABLE refresh_jobs ADD COLUMN IF NOT EXISTS current_operation VARCHAR(200);
ALTER TABLE refresh_jobs ADD COLUMN IF NOT EXISTS current_batch INTEGER;
ALTER TABLE refresh_jobs ADD COLUMN IF NOT EXISTS total_batches INTEGER;
