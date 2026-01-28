-- Remove check constraint for cancelled status
ALTER TABLE refresh_jobs
    DROP CONSTRAINT IF EXISTS refresh_jobs_status_check;
