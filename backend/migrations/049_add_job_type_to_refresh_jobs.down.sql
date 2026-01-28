DROP INDEX IF EXISTS idx_refresh_jobs_job_type;
ALTER TABLE refresh_jobs DROP CONSTRAINT IF EXISTS chk_refresh_jobs_job_type;
ALTER TABLE refresh_jobs DROP COLUMN IF EXISTS job_type;
