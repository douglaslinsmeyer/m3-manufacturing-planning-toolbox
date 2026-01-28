-- Add check constraint to allow 'cancelled' status for refresh_jobs
ALTER TABLE refresh_jobs
    ADD CONSTRAINT refresh_jobs_status_check
    CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'));
