-- Rollback anomaly detection support

-- Drop indexes
DROP INDEX IF EXISTS idx_detected_issues_entity_id;
DROP INDEX IF EXISTS idx_detected_issues_entity_type;
DROP INDEX IF EXISTS idx_detected_issues_severity;

-- Drop columns
ALTER TABLE detected_issues
  DROP COLUMN IF EXISTS actual_value,
  DROP COLUMN IF EXISTS threshold_value,
  DROP COLUMN IF EXISTS affected_count,
  DROP COLUMN IF EXISTS entity_id,
  DROP COLUMN IF EXISTS entity_type,
  DROP COLUMN IF EXISTS severity;
