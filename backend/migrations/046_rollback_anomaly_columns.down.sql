-- Re-add anomaly columns to detected_issues if rolling back this migration
ALTER TABLE detected_issues
  ADD COLUMN IF NOT EXISTS severity VARCHAR(20) DEFAULT 'warning',
  ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS entity_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS affected_count INTEGER,
  ADD COLUMN IF NOT EXISTS threshold_value DECIMAL(15,6),
  ADD COLUMN IF NOT EXISTS actual_value DECIMAL(15,6);

CREATE INDEX IF NOT EXISTS idx_detected_issues_severity ON detected_issues(severity);
CREATE INDEX IF NOT EXISTS idx_detected_issues_entity_type ON detected_issues(entity_type);
CREATE INDEX IF NOT EXISTS idx_detected_issues_entity_id ON detected_issues(entity_id);
