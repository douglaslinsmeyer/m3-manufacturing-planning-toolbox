-- Create bulk_operation_jobs table
CREATE TABLE bulk_operation_jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment VARCHAR(10) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- 'delete', 'close', 'reschedule'
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Job parameters
    issue_ids BIGINT[] NOT NULL,
    params JSONB, -- For reschedule date, etc.

    -- Progress tracking
    total_items INTEGER NOT NULL DEFAULT 0,
    successful_items INTEGER NOT NULL DEFAULT 0,
    failed_items INTEGER NOT NULL DEFAULT 0,
    current_phase VARCHAR(100),
    progress_percentage INTEGER NOT NULL DEFAULT 0,

    -- Timing
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    -- Error tracking
    error_message TEXT
);

-- Create indexes for bulk_operation_jobs
CREATE INDEX idx_bulkop_jobs_env_status ON bulk_operation_jobs(environment, status);
CREATE INDEX idx_bulkop_jobs_user ON bulk_operation_jobs(user_id, created_at DESC);

-- Create bulk_operation_batch_results table
CREATE TABLE bulk_operation_batch_results (
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES bulk_operation_jobs(job_id) ON DELETE CASCADE,
    batch_number INTEGER NOT NULL,

    production_order_id BIGINT NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(10) NOT NULL, -- 'MOP' or 'MO'

    success BOOLEAN NOT NULL,
    error_message TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for bulk_operation_batch_results
CREATE INDEX idx_batch_results_job ON bulk_operation_batch_results(job_id);
CREATE INDEX idx_batch_results_order ON bulk_operation_batch_results(production_order_id);
