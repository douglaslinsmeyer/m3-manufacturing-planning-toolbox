-- Migration 040: Add Refresh Job Phase Tracking
-- Adds tables to track individual phase and detector execution for crash recovery and performance analysis

-- Refresh job data loading phases (MOPs, MOs, COs)
CREATE TABLE refresh_job_phases (
    id BIGSERIAL PRIMARY KEY,
    job_id VARCHAR(36) NOT NULL REFERENCES refresh_jobs(id) ON DELETE CASCADE,

    -- Phase identification
    phase_type VARCHAR(20) NOT NULL,  -- 'mops', 'mos', 'cos'

    -- Phase status
    status VARCHAR(20) NOT NULL,  -- 'pending', 'running', 'completed', 'failed'

    -- Results
    record_count INTEGER DEFAULT 0,
    error_message TEXT,

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms BIGINT,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_phase_type CHECK (phase_type IN ('mops', 'mos', 'cos')),
    CONSTRAINT chk_phase_status CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    UNIQUE(job_id, phase_type)
);

-- Refresh job detector executions
CREATE TABLE refresh_job_detectors (
    id BIGSERIAL PRIMARY KEY,
    job_id VARCHAR(36) NOT NULL REFERENCES refresh_jobs(id) ON DELETE CASCADE,

    -- Detector identification
    detector_name VARCHAR(100) NOT NULL,
    display_label VARCHAR(200) NOT NULL,

    -- Detector status
    status VARCHAR(20) NOT NULL,  -- 'pending', 'running', 'completed', 'failed'

    -- Results
    issues_found INTEGER DEFAULT 0,
    error_message TEXT,

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms BIGINT,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_detector_status CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    UNIQUE(job_id, detector_name)
);

-- Indexes for efficient querying
CREATE INDEX idx_refresh_job_phases_job_id ON refresh_job_phases(job_id);
CREATE INDEX idx_refresh_job_phases_status ON refresh_job_phases(status);
CREATE INDEX idx_refresh_job_detectors_job_id ON refresh_job_detectors(job_id);
CREATE INDEX idx_refresh_job_detectors_status ON refresh_job_detectors(status);

-- Triggers for updated_at
CREATE TRIGGER update_refresh_job_phases_updated_at
    BEFORE UPDATE ON refresh_job_phases
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refresh_job_detectors_updated_at
    BEFORE UPDATE ON refresh_job_detectors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE refresh_job_phases IS 'Tracks individual data loading phase execution for crash recovery and performance analysis';
COMMENT ON TABLE refresh_job_detectors IS 'Tracks individual detector execution for crash recovery and performance analysis';
