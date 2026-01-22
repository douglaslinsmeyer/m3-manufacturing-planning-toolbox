-- Add job tracking for async operations

CREATE TABLE IF NOT EXISTS refresh_jobs (
    id VARCHAR(36) PRIMARY KEY,  -- UUID
    environment VARCHAR(10) NOT NULL,
    user_id VARCHAR(100),

    -- Job status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Progress tracking
    current_step VARCHAR(100),
    total_steps INTEGER DEFAULT 3,
    completed_steps INTEGER DEFAULT 0,
    progress_percentage INTEGER DEFAULT 0,

    -- Record counts
    co_lines_processed INTEGER DEFAULT 0,
    mos_processed INTEGER DEFAULT 0,
    mops_processed INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,

    -- Error tracking
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_jobs_status ON refresh_jobs(status);
CREATE INDEX idx_refresh_jobs_environment ON refresh_jobs(environment);
CREATE INDEX idx_refresh_jobs_created_at ON refresh_jobs(created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_refresh_jobs_updated_at
    BEFORE UPDATE ON refresh_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Update snapshot_metadata table to reference jobs
ALTER TABLE snapshot_metadata ADD COLUMN IF NOT EXISTS job_id VARCHAR(36);
CREATE INDEX IF NOT EXISTS idx_snapshot_metadata_job_id ON snapshot_metadata(job_id);
