-- Rollback: Recreate bulk operation tables
-- NOTE: Data will NOT be restored, only schema

-- Recreate bulk_operation_jobs table (from 050_add_bulk_operation_jobs.up.sql)
CREATE TABLE IF NOT EXISTS bulk_operation_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment VARCHAR(10) NOT NULL,
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_count INTEGER NOT NULL DEFAULT 0,
    successful_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    user_id TEXT,
    issue_ids BIGINT[],
    params JSONB,
    progress_percent INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_bulk_operation_jobs_env_status ON bulk_operation_jobs(environment, status);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_jobs_user_id ON bulk_operation_jobs(user_id);

-- Recreate bulk_operation_batch_results table
CREATE TABLE IF NOT EXISTS bulk_operation_batch_results (
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES bulk_operation_jobs(id) ON DELETE CASCADE,
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bulk_operation_batch_results_job_id ON bulk_operation_batch_results(job_id);

-- Recreate bulk_operation_issue_results table (from 054_add_bulk_operation_issue_results.up.sql)
CREATE TABLE IF NOT EXISTS bulk_operation_issue_results (
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES bulk_operation_jobs(id) ON DELETE CASCADE,
    issue_id BIGINT NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    is_duplicate BOOLEAN NOT NULL DEFAULT FALSE,
    primary_issue_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bulk_operation_issue_results_job_id ON bulk_operation_issue_results(job_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operation_issue_results_issue_id ON bulk_operation_issue_results(issue_id);
