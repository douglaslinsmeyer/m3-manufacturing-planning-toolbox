-- Add anomaly detection support columns to detected_issues table
ALTER TABLE detected_issues
  ADD COLUMN IF NOT EXISTS severity VARCHAR(20) DEFAULT 'warning',
  ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS entity_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS affected_count INTEGER,
  ADD COLUMN IF NOT EXISTS threshold_value DECIMAL(15,6),
  ADD COLUMN IF NOT EXISTS actual_value DECIMAL(15,6);

-- Create indexes for anomaly queries
CREATE INDEX IF NOT EXISTS idx_detected_issues_severity ON detected_issues(severity);
CREATE INDEX IF NOT EXISTS idx_detected_issues_entity_type ON detected_issues(entity_type);
CREATE INDEX IF NOT EXISTS idx_detected_issues_entity_id ON detected_issues(entity_id);

-- Add comments to explain new columns
COMMENT ON COLUMN detected_issues.severity IS 'Severity level: info, warning, or critical';
COMMENT ON COLUMN detected_issues.entity_type IS 'Type of entity affected: product, warehouse, system, etc.';
COMMENT ON COLUMN detected_issues.entity_id IS 'Identifier for the affected entity (e.g., product number, warehouse code)';
COMMENT ON COLUMN detected_issues.affected_count IS 'Number of records affected by this anomaly';
COMMENT ON COLUMN detected_issues.threshold_value IS 'Threshold that was breached (for anomalies)';
COMMENT ON COLUMN detected_issues.actual_value IS 'Actual measured value (for anomalies)';
