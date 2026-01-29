-- Issue-level results table (one row per selected issue)
-- This table maps bulk operation results back to individual issues,
-- solving the duplicate production order problem where multiple issues
-- reference the same production order.
CREATE TABLE IF NOT EXISTS bulk_operation_issue_results (
    id BIGSERIAL PRIMARY KEY,
    job_id UUID NOT NULL,
    issue_id BIGINT NOT NULL,
    production_order_id BIGINT NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    is_duplicate BOOLEAN NOT NULL DEFAULT FALSE,
    primary_issue_id BIGINT,  -- For "same order as Issue #X" display
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_issue_results_job
        FOREIGN KEY (job_id)
        REFERENCES bulk_operation_jobs(job_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_issue_results_issue
        FOREIGN KEY (issue_id)
        REFERENCES detected_issues(id)
        ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_issue_results_job_id
    ON bulk_operation_issue_results(job_id);
CREATE INDEX IF NOT EXISTS idx_issue_results_issue_id
    ON bulk_operation_issue_results(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_results_job_issue
    ON bulk_operation_issue_results(job_id, issue_id);
