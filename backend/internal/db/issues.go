package db

import (
	"context"
	"database/sql"
	"fmt"
)

// IssueDetectionJob represents an issue detection job
type IssueDetectionJob struct {
	ID                 int64
	JobID              string
	Status             string
	TotalDetectors     int
	CompletedDetectors int
	FailedDetectors    int
	TotalIssuesFound   int
	IssuesByType       sql.NullString // JSONB
	StartedAt          sql.NullTime
	CompletedAt        sql.NullTime
	DurationSeconds    sql.NullInt32
	ErrorMessage       sql.NullString
	CreatedAt          sql.NullTime
	UpdatedAt          sql.NullTime
}

// DetectedIssue represents a detected issue
type DetectedIssue struct {
	ID                    int64
	JobID                 string
	DetectorType          string
	DetectedAt            sql.NullTime
	Facility              string
	Warehouse             sql.NullString
	IssueKey              string
	ProductionOrderNumber sql.NullString
	ProductionOrderType   sql.NullString
	CONumber              sql.NullString
	COLine                sql.NullString
	COSuffix              sql.NullString
	IssueData             string // JSONB
	CreatedAt             sql.NullTime
}

// CreateIssueDetectionJob creates a new detection job
func (q *Queries) CreateIssueDetectionJob(ctx context.Context, jobID string, totalDetectors int) error {
	query := `
		INSERT INTO issue_detection_jobs (job_id, status, total_detectors, started_at)
		VALUES ($1, 'running', $2, NOW())
	`
	_, err := q.db.ExecContext(ctx, query, jobID, totalDetectors)
	return err
}

// UpdateDetectionProgress updates detector completion progress
func (q *Queries) UpdateDetectionProgress(ctx context.Context, jobID string, completed, total int) error {
	query := `
		UPDATE issue_detection_jobs
		SET completed_detectors = $1,
			total_detectors = $2,
			updated_at = NOW()
		WHERE job_id = $3
	`
	_, err := q.db.ExecContext(ctx, query, completed, total, jobID)
	return err
}

// IncrementFailedDetectors increments the failed detector count
func (q *Queries) IncrementFailedDetectors(ctx context.Context, jobID string) error {
	query := `
		UPDATE issue_detection_jobs
		SET failed_detectors = failed_detectors + 1,
			updated_at = NOW()
		WHERE job_id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// CompleteDetectionJob marks detection job as complete
func (q *Queries) CompleteDetectionJob(ctx context.Context, jobID string, totalIssues int, issuesByType string) error {
	query := `
		UPDATE issue_detection_jobs
		SET status = 'completed',
			total_issues_found = $1,
			issues_by_type = $2,
			completed_at = NOW(),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
			updated_at = NOW()
		WHERE job_id = $3
	`
	_, err := q.db.ExecContext(ctx, query, totalIssues, issuesByType, jobID)
	return err
}

// FailDetectionJob marks detection job as failed
func (q *Queries) FailDetectionJob(ctx context.Context, jobID string, errorMessage string) error {
	query := `
		UPDATE issue_detection_jobs
		SET status = 'failed',
			error_message = $1,
			completed_at = NOW(),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
			updated_at = NOW()
		WHERE job_id = $2
	`
	_, err := q.db.ExecContext(ctx, query, errorMessage, jobID)
	return err
}

// ClearIssuesForJob removes previous issues for a job
func (q *Queries) ClearIssuesForJob(ctx context.Context, jobID string) error {
	query := `DELETE FROM detected_issues WHERE job_id = $1`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// GetIssuesByDetectorType gets issues filtered by detector type
func (q *Queries) GetIssuesByDetectorType(ctx context.Context, detectorType string, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE detector_type = $1
		ORDER BY detected_at DESC
		LIMIT $2
	`

	rows, err := q.db.QueryContext(ctx, query, detectorType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*DetectedIssue, 0)
	for rows.Next() {
		issue := &DetectedIssue{}
		err := rows.Scan(
			&issue.ID, &issue.JobID, &issue.DetectorType, &issue.DetectedAt,
			&issue.Facility, &issue.Warehouse, &issue.IssueKey,
			&issue.ProductionOrderNumber, &issue.ProductionOrderType,
			&issue.CONumber, &issue.COLine, &issue.COSuffix,
			&issue.IssueData, &issue.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// GetIssuesByFacility gets issues filtered by facility
func (q *Queries) GetIssuesByFacility(ctx context.Context, facility string, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE facility = $1
		ORDER BY detected_at DESC
		LIMIT $2
	`

	rows, err := q.db.QueryContext(ctx, query, facility, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*DetectedIssue, 0)
	for rows.Next() {
		issue := &DetectedIssue{}
		err := rows.Scan(
			&issue.ID, &issue.JobID, &issue.DetectorType, &issue.DetectedAt,
			&issue.Facility, &issue.Warehouse, &issue.IssueKey,
			&issue.ProductionOrderNumber, &issue.ProductionOrderType,
			&issue.CONumber, &issue.COLine, &issue.COSuffix,
			&issue.IssueData, &issue.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// GetRecentIssues gets recent issues (no filter)
func (q *Queries) GetRecentIssues(ctx context.Context, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		ORDER BY detected_at DESC
		LIMIT $1
	`

	rows, err := q.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*DetectedIssue, 0)
	for rows.Next() {
		issue := &DetectedIssue{}
		err := rows.Scan(
			&issue.ID, &issue.JobID, &issue.DetectorType, &issue.DetectedAt,
			&issue.Facility, &issue.Warehouse, &issue.IssueKey,
			&issue.ProductionOrderNumber, &issue.ProductionOrderType,
			&issue.CONumber, &issue.COLine, &issue.COSuffix,
			&issue.IssueData, &issue.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// GetIssueCountForLatestJob gets the total issue count for the latest refresh job
func (q *Queries) GetIssueCountForLatestJob(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM detected_issues
		WHERE job_id = (
			SELECT id FROM refresh_jobs
			ORDER BY created_at DESC
			LIMIT 1
		)
	`

	var count int
	err := q.db.QueryRowContext(ctx, query).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil // No issues found
	}
	return count, err
}

// GetIssueSummary gets aggregated issue counts with warehouse grouping and nested hierarchy
func (q *Queries) GetIssueSummary(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			detector_type,
			facility,
			COALESCE(warehouse, '') as warehouse,
			COUNT(*) as issue_count
		FROM detected_issues
		WHERE job_id = (
			SELECT id FROM refresh_jobs
			ORDER BY created_at DESC
			LIMIT 1
		)
		GROUP BY detector_type, facility, warehouse
		ORDER BY facility, warehouse, detector_type
	`

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]interface{})
	byDetector := make(map[string]int)
	byFacility := make(map[string]int)
	byWarehouse := make(map[string]int)
	// Nested: facility -> warehouse -> detector -> count
	byFacilityWarehouseDetector := make(map[string]map[string]map[string]int)
	total := 0

	for rows.Next() {
		var detectorType, facility, warehouse string
		var count int

		if err := rows.Scan(&detectorType, &facility, &warehouse, &count); err != nil {
			return nil, err
		}

		// Aggregate by detector type
		byDetector[detectorType] += count

		// Aggregate by facility
		byFacility[facility] += count

		// Aggregate by warehouse
		if warehouse != "" {
			byWarehouse[warehouse] += count
		}

		// Build nested hierarchy: facility -> warehouse -> detector
		if _, ok := byFacilityWarehouseDetector[facility]; !ok {
			byFacilityWarehouseDetector[facility] = make(map[string]map[string]int)
		}
		if _, ok := byFacilityWarehouseDetector[facility][warehouse]; !ok {
			byFacilityWarehouseDetector[facility][warehouse] = make(map[string]int)
		}
		byFacilityWarehouseDetector[facility][warehouse][detectorType] = count

		total += count
	}

	summary["total"] = total
	summary["by_detector"] = byDetector
	summary["by_facility"] = byFacility
	summary["by_warehouse"] = byWarehouse
	summary["by_facility_warehouse_detector"] = byFacilityWarehouseDetector

	return summary, nil
}

// GetIssueByID gets a specific issue by ID
func (q *Queries) GetIssueByID(ctx context.Context, id int64) (*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE id = $1
	`

	issue := &DetectedIssue{}
	err := q.db.QueryRowContext(ctx, query, id).Scan(
		&issue.ID, &issue.JobID, &issue.DetectorType, &issue.DetectedAt,
		&issue.Facility, &issue.Warehouse, &issue.IssueKey,
		&issue.ProductionOrderNumber, &issue.ProductionOrderType,
		&issue.CONumber, &issue.COLine, &issue.COSuffix,
		&issue.IssueData, &issue.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue not found")
	}

	return issue, err
}

// GetLatestRefreshJobID gets the most recent refresh job ID
func (q *Queries) GetLatestRefreshJobID(ctx context.Context) (string, error) {
	query := `SELECT id FROM refresh_jobs ORDER BY created_at DESC LIMIT 1`

	var jobID string
	err := q.db.QueryRowContext(ctx, query).Scan(&jobID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no refresh jobs found")
	}

	return jobID, err
}
