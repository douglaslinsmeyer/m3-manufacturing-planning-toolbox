-- Migration 017: Add Issue Detection System
-- Creates tables for storing detected data quality issues and tracking detection jobs

-- Detected issues table
CREATE TABLE detected_issues (
    id BIGSERIAL PRIMARY KEY,

    -- Detection metadata
    job_id VARCHAR(36) NOT NULL,  -- Links to refresh_jobs.id
    detector_type VARCHAR(50) NOT NULL,
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Context filters
    facility VARCHAR(10) NOT NULL,
    warehouse VARCHAR(10),

    -- Issue grouping key (for aggregation)
    issue_key VARCHAR(200) NOT NULL,

    -- Affected records
    production_order_number VARCHAR(50),
    production_order_type VARCHAR(10),  -- 'MO' or 'MOP'
    co_number VARCHAR(50),
    co_line VARCHAR(50),
    co_suffix VARCHAR(50),

    -- Issue details (JSONB for flexibility)
    issue_data JSONB NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_order_type CHECK (production_order_type IN ('MO', 'MOP') OR production_order_type IS NULL)
);

-- Indexes for efficient querying
CREATE INDEX idx_detected_issues_job_id ON detected_issues(job_id);
CREATE INDEX idx_detected_issues_detector_type ON detected_issues(detector_type);
CREATE INDEX idx_detected_issues_facility ON detected_issues(facility);
CREATE INDEX idx_detected_issues_detected_at ON detected_issues(detected_at DESC);
CREATE INDEX idx_detected_issues_issue_key ON detected_issues(issue_key);
CREATE INDEX idx_detected_issues_po_number ON detected_issues(production_order_number);
CREATE INDEX idx_detected_issues_co ON detected_issues(co_number, co_line);

-- Add comments
COMMENT ON TABLE detected_issues IS 'Stores detected data quality and planning issues from issue detectors';
COMMENT ON COLUMN detected_issues.issue_key IS 'Grouping key for related issues (e.g., CO number-line for date mismatches)';
COMMENT ON COLUMN detected_issues.issue_data IS 'JSONB field storing detector-specific issue details';

-- Issue detection jobs tracking table
CREATE TABLE issue_detection_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_id VARCHAR(36) NOT NULL UNIQUE,  -- Same as refresh_jobs.id

    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Detector execution tracking
    total_detectors INTEGER NOT NULL,
    completed_detectors INTEGER DEFAULT 0,
    failed_detectors INTEGER DEFAULT 0,

    -- Results summary
    total_issues_found INTEGER DEFAULT 0,
    issues_by_type JSONB,  -- {"unlinked_orders": 5, "date_mismatch": 12, ...}

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,

    error_message TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issue_detection_jobs_job_id ON issue_detection_jobs(job_id);
CREATE INDEX idx_issue_detection_jobs_status ON issue_detection_jobs(status);

-- Add comments
COMMENT ON TABLE issue_detection_jobs IS 'Tracks execution of issue detection runs linked to snapshot refresh jobs';
COMMENT ON COLUMN issue_detection_jobs.issues_by_type IS 'JSONB summary of issue counts by detector type';
