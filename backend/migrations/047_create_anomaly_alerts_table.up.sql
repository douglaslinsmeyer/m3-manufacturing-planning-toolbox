-- Create dedicated anomaly_alerts table separate from detected_issues
CREATE TABLE anomaly_alerts (
  id SERIAL PRIMARY KEY,
  environment VARCHAR(10) NOT NULL,
  job_id VARCHAR(36) NOT NULL,
  detector_type VARCHAR(100) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  entity_type VARCHAR(50),
  entity_id VARCHAR(100),
  message TEXT,
  metrics JSONB,
  affected_count INTEGER,
  threshold_value DECIMAL(15,6),
  actual_value DECIMAL(15,6),
  status VARCHAR(20) DEFAULT 'active',
  detected_at TIMESTAMP DEFAULT NOW(),
  acknowledged_at TIMESTAMP,
  acknowledged_by VARCHAR(100),
  resolved_at TIMESTAMP,
  resolved_by VARCHAR(100),
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_anomaly_alerts_environment ON anomaly_alerts(environment);
CREATE INDEX idx_anomaly_alerts_job_id ON anomaly_alerts(job_id);
CREATE INDEX idx_anomaly_alerts_severity ON anomaly_alerts(severity);
CREATE INDEX idx_anomaly_alerts_status ON anomaly_alerts(status);
CREATE INDEX idx_anomaly_alerts_entity ON anomaly_alerts(entity_type, entity_id);
CREATE INDEX idx_anomaly_alerts_detector_type ON anomaly_alerts(detector_type);
CREATE INDEX idx_anomaly_alerts_detected_at ON anomaly_alerts(detected_at DESC);

-- Foreign key to refresh_jobs
ALTER TABLE anomaly_alerts
  ADD CONSTRAINT fk_anomaly_alerts_job
  FOREIGN KEY (job_id) REFERENCES refresh_jobs(id)
  ON DELETE CASCADE;

-- Check constraint for status
ALTER TABLE anomaly_alerts
  ADD CONSTRAINT chk_anomaly_status
  CHECK (status IN ('active', 'acknowledged', 'resolved'));

-- Check constraint for severity
ALTER TABLE anomaly_alerts
  ADD CONSTRAINT chk_anomaly_severity
  CHECK (severity IN ('info', 'warning', 'critical'));

-- Comments for documentation
COMMENT ON TABLE anomaly_alerts IS 'Statistical anomalies detected across aggregate data patterns, separate from individual record issues';
COMMENT ON COLUMN anomaly_alerts.severity IS 'Severity level: info, warning, or critical';
COMMENT ON COLUMN anomaly_alerts.entity_type IS 'Type of entity affected: product, warehouse, or system';
COMMENT ON COLUMN anomaly_alerts.entity_id IS 'Identifier for the affected entity (e.g., product number, warehouse code)';
COMMENT ON COLUMN anomaly_alerts.affected_count IS 'Number of records affected by this anomaly';
COMMENT ON COLUMN anomaly_alerts.threshold_value IS 'Threshold that was breached';
COMMENT ON COLUMN anomaly_alerts.actual_value IS 'Actual measured value that triggered the alert';
COMMENT ON COLUMN anomaly_alerts.metrics IS 'Full statistical data and context for the anomaly';
