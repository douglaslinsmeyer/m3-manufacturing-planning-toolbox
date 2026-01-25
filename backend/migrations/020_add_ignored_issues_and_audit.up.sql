-- ========================================
-- AUDIT LOG TABLE (Flexible, Reusable)
-- ========================================
-- Generic audit log for tracking all user actions across the system
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,

    -- Core audit fields
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(100),           -- Optional: user identifier (from auth system)
    user_name VARCHAR(200),         -- Optional: human-readable user name

    -- Operation metadata
    entity_type VARCHAR(50) NOT NULL,     -- e.g., 'issue', 'order', 'setting', 'bulk_operation'
    entity_id VARCHAR(200),              -- Optional: specific entity identifier
    operation VARCHAR(50) NOT NULL,      -- e.g., 'ignore', 'unignore', 'create', 'update', 'delete'

    -- Context fields (optional, for filtering/reporting)
    company VARCHAR(10),
    facility VARCHAR(10),
    warehouse VARCHAR(10),

    -- Flexible metadata storage
    metadata JSONB,                      -- Operation-specific data (what changed, why, etc.)

    -- HTTP context (for debugging)
    ip_address INET,
    user_agent TEXT,

    -- Indexing for common queries
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common audit queries
CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp DESC);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_operation ON audit_log(operation);
CREATE INDEX idx_audit_log_facility ON audit_log(facility) WHERE facility IS NOT NULL;
CREATE INDEX idx_audit_log_metadata ON audit_log USING gin(metadata);  -- JSONB indexing

-- ========================================
-- IGNORED ISSUES TABLE (Business Logic)
-- ========================================
-- Table to track ignored issues (persists across refreshes)
CREATE TABLE ignored_issues (
    id BIGSERIAL PRIMARY KEY,
    facility VARCHAR(10) NOT NULL,
    detector_type VARCHAR(50) NOT NULL,
    issue_key VARCHAR(200) NOT NULL,
    production_order_number VARCHAR(50),
    production_order_type VARCHAR(10),
    co_number VARCHAR(50),
    co_line VARCHAR(50),
    ignored_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ignored_by VARCHAR(100),  -- User who ignored it (matches audit_log.user_id)
    notes TEXT,  -- Optional: reason for ignoring
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint: same issue can only be ignored once
    UNIQUE(facility, detector_type, issue_key, production_order_number)
);

-- Index for fast lookups when joining with detected_issues
CREATE INDEX idx_ignored_issues_lookup
    ON ignored_issues(facility, detector_type, issue_key, production_order_number);

-- Index for querying by user
CREATE INDEX idx_ignored_issues_by_user ON ignored_issues(ignored_by) WHERE ignored_by IS NOT NULL;
