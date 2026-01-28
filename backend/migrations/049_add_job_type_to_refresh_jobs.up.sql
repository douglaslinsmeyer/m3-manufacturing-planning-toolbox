-- Add job_type column with default for existing records
ALTER TABLE refresh_jobs
  ADD COLUMN job_type VARCHAR(20) NOT NULL DEFAULT 'snapshot_refresh';

-- Add check constraint for valid job types
ALTER TABLE refresh_jobs
  ADD CONSTRAINT chk_refresh_jobs_job_type
  CHECK (job_type IN ('snapshot_refresh', 'manual_detection'));

-- Add index for efficient filtering
CREATE INDEX idx_refresh_jobs_job_type ON refresh_jobs(job_type);

-- Add comment
COMMENT ON COLUMN refresh_jobs.job_type IS 'Type of job: snapshot_refresh for full data refresh, manual_detection for on-demand detector runs';
