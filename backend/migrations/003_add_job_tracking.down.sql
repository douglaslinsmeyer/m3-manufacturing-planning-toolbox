-- Rollback job tracking

DROP TRIGGER IF EXISTS update_refresh_jobs_updated_at ON refresh_jobs;

DROP INDEX IF EXISTS idx_snapshot_metadata_job_id;
ALTER TABLE snapshot_metadata DROP COLUMN IF EXISTS job_id;

DROP INDEX IF EXISTS idx_refresh_jobs_created_at;
DROP INDEX IF EXISTS idx_refresh_jobs_environment;
DROP INDEX IF EXISTS idx_refresh_jobs_status;

DROP TABLE IF EXISTS refresh_jobs;
