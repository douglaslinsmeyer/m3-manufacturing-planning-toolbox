package db

import (
	"context"
	"database/sql"
	"fmt"
)

// AnomalyAlert represents an anomaly alert from the anomaly_alerts table
type AnomalyAlert struct {
	ID             int64
	Environment    string
	JobID          string
	DetectorType   string
	Severity       string
	EntityType     sql.NullString
	EntityID       sql.NullString
	Message        sql.NullString
	Metrics        string // JSONB
	AffectedCount  sql.NullInt32
	ThresholdValue sql.NullFloat64
	ActualValue    sql.NullFloat64
	Status         string
	DetectedAt     sql.NullTime
	AcknowledgedAt sql.NullTime
	AcknowledgedBy sql.NullString
	ResolvedAt     sql.NullTime
	ResolvedBy     sql.NullString
	Notes          sql.NullString
	CreatedAt      sql.NullTime
	UpdatedAt      sql.NullTime
}

// InsertAnomalyAlertParams holds parameters for inserting anomaly alerts
type InsertAnomalyAlertParams struct {
	Environment    string
	JobID          string
	DetectorType   string
	Severity       string
	EntityType     sql.NullString
	EntityID       sql.NullString
	Message        sql.NullString
	Metrics        string
	AffectedCount  sql.NullInt32
	ThresholdValue sql.NullFloat64
	ActualValue    sql.NullFloat64
}

// InsertAnomalyAlert inserts a new anomaly alert
func (q *Queries) InsertAnomalyAlert(ctx context.Context, params InsertAnomalyAlertParams) error {
	query := `
		INSERT INTO anomaly_alerts (
			environment, job_id, detector_type, severity, entity_type,
			entity_id, message, metrics, affected_count, threshold_value,
			actual_value, status, detected_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active', NOW())
	`
	_, err := q.db.ExecContext(ctx, query,
		params.Environment,
		params.JobID,
		params.DetectorType,
		params.Severity,
		params.EntityType,
		params.EntityID,
		params.Message,
		params.Metrics,
		params.AffectedCount,
		params.ThresholdValue,
		params.ActualValue,
	)
	return err
}

// GetAnomaliesFiltered retrieves anomaly alerts with optional filters
func (q *Queries) GetAnomaliesFiltered(ctx context.Context, environment, severity, detectorType string, limit, offset int) ([]*AnomalyAlert, error) {
	query := `
		SELECT id, environment, job_id, detector_type, severity, entity_type, entity_id,
		       message, metrics, affected_count, threshold_value, actual_value, status,
		       detected_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by,
		       notes, created_at, updated_at
		FROM anomaly_alerts
		WHERE environment = $1
		  AND job_id = (
		      SELECT id FROM refresh_jobs
		      WHERE environment = $1
		      ORDER BY created_at DESC
		      LIMIT 1
		  )
	`
	args := []interface{}{environment}
	argNum := 2

	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argNum)
		args = append(args, severity)
		argNum++
	}

	if detectorType != "" {
		query += fmt.Sprintf(" AND detector_type = $%d", argNum)
		args = append(args, detectorType)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY severity DESC, actual_value DESC, detected_at DESC OFFSET $%d LIMIT $%d", argNum, argNum+1)
	args = append(args, offset, limit)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	anomalies := make([]*AnomalyAlert, 0)
	for rows.Next() {
		anomaly := &AnomalyAlert{}
		err := rows.Scan(
			&anomaly.ID, &anomaly.Environment, &anomaly.JobID, &anomaly.DetectorType,
			&anomaly.Severity, &anomaly.EntityType, &anomaly.EntityID,
			&anomaly.Message, &anomaly.Metrics, &anomaly.AffectedCount,
			&anomaly.ThresholdValue, &anomaly.ActualValue, &anomaly.Status,
			&anomaly.DetectedAt, &anomaly.AcknowledgedAt, &anomaly.AcknowledgedBy,
			&anomaly.ResolvedAt, &anomaly.ResolvedBy, &anomaly.Notes,
			&anomaly.CreatedAt, &anomaly.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, rows.Err()
}

// GetAnomaliesFilteredCount gets count of anomalies matching filters
func (q *Queries) GetAnomaliesFilteredCount(ctx context.Context, environment, severity, detectorType string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM anomaly_alerts
		WHERE environment = $1
		  AND job_id = (
		      SELECT id FROM refresh_jobs
		      WHERE environment = $1
		      ORDER BY created_at DESC
		      LIMIT 1
		  )
	`
	args := []interface{}{environment}
	argNum := 2

	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argNum)
		args = append(args, severity)
		argNum++
	}

	if detectorType != "" {
		query += fmt.Sprintf(" AND detector_type = $%d", argNum)
		args = append(args, detectorType)
		argNum++
	}

	var count int
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// GetAnomalySummary gets aggregated anomaly counts by severity and detector type
func (q *Queries) GetAnomalySummary(ctx context.Context, environment string) (map[string]interface{}, error) {
	query := `
		SELECT
			severity,
			detector_type,
			COUNT(*) as count
		FROM anomaly_alerts
		WHERE environment = $1
		  AND job_id = (
		      SELECT id FROM refresh_jobs
		      WHERE environment = $1
		      ORDER BY created_at DESC
		      LIMIT 1
		  )
		GROUP BY severity, detector_type
		ORDER BY
		  CASE severity
		    WHEN 'critical' THEN 1
		    WHEN 'warning' THEN 2
		    WHEN 'info' THEN 3
		  END,
		  count DESC
	`

	rows, err := q.db.QueryContext(ctx, query, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]interface{})
	bySeverity := make(map[string]int)
	byDetector := make(map[string]int)
	total := 0

	for rows.Next() {
		var severity, detectorType string
		var count int

		if err := rows.Scan(&severity, &detectorType, &count); err != nil {
			return nil, err
		}

		bySeverity[severity] += count
		byDetector[detectorType] += count
		total += count
	}

	summary["total"] = total
	summary["by_severity"] = bySeverity
	summary["by_detector"] = byDetector

	return summary, nil
}

// GetAnomalyByID retrieves a specific anomaly by ID
func (q *Queries) GetAnomalyByID(ctx context.Context, id int64) (*AnomalyAlert, error) {
	query := `
		SELECT id, environment, job_id, detector_type, severity, entity_type, entity_id,
		       message, metrics, affected_count, threshold_value, actual_value, status,
		       detected_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by,
		       notes, created_at, updated_at
		FROM anomaly_alerts
		WHERE id = $1
	`

	anomaly := &AnomalyAlert{}
	err := q.db.QueryRowContext(ctx, query, id).Scan(
		&anomaly.ID, &anomaly.Environment, &anomaly.JobID, &anomaly.DetectorType,
		&anomaly.Severity, &anomaly.EntityType, &anomaly.EntityID,
		&anomaly.Message, &anomaly.Metrics, &anomaly.AffectedCount,
		&anomaly.ThresholdValue, &anomaly.ActualValue, &anomaly.Status,
		&anomaly.DetectedAt, &anomaly.AcknowledgedAt, &anomaly.AcknowledgedBy,
		&anomaly.ResolvedAt, &anomaly.ResolvedBy, &anomaly.Notes,
		&anomaly.CreatedAt, &anomaly.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("anomaly not found")
	}

	return anomaly, err
}

// AcknowledgeAnomaly marks an anomaly as acknowledged
func (q *Queries) AcknowledgeAnomaly(ctx context.Context, id int64, acknowledgedBy string, notes sql.NullString) error {
	query := `
		UPDATE anomaly_alerts
		SET status = 'acknowledged',
		    acknowledged_at = NOW(),
		    acknowledged_by = $2,
		    notes = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND status = 'active'
	`
	result, err := q.db.ExecContext(ctx, query, id, acknowledgedBy, notes)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("anomaly not found or already acknowledged/resolved")
	}

	return nil
}

// ResolveAnomaly marks an anomaly as resolved
func (q *Queries) ResolveAnomaly(ctx context.Context, id int64, resolvedBy string, notes sql.NullString) error {
	query := `
		UPDATE anomaly_alerts
		SET status = 'resolved',
		    resolved_at = NOW(),
		    resolved_by = $2,
		    notes = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND status IN ('active', 'acknowledged')
	`
	result, err := q.db.ExecContext(ctx, query, id, resolvedBy, notes)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("anomaly not found or already resolved")
	}

	return nil
}

// DeleteAnomaly deletes an anomaly (admin only)
func (q *Queries) DeleteAnomaly(ctx context.Context, id int64) error {
	query := `DELETE FROM anomaly_alerts WHERE id = $1`
	result, err := q.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("anomaly not found")
	}

	return nil
}

// ClearAnomaliesForJob removes all anomalies for a specific job
func (q *Queries) ClearAnomaliesForJob(ctx context.Context, jobID string) error {
	query := `DELETE FROM anomaly_alerts WHERE job_id = $1`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// GetActiveAnomalyCount gets count of active anomalies for the latest job
func (q *Queries) GetActiveAnomalyCount(ctx context.Context, environment string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM anomaly_alerts
		WHERE environment = $1
		  AND status = 'active'
		  AND job_id = (
		      SELECT id FROM refresh_jobs
		      WHERE environment = $1
		      ORDER BY created_at DESC
		      LIMIT 1
		  )
	`
	var count int
	err := q.db.QueryRowContext(ctx, query, environment).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}
