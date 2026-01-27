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
	Environment           string // M3 environment (TRN or PRD)
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
	IsIgnored             bool
	MOTypeDescription     sql.NullString
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

// GetIssuesByDetectorType gets issues filtered by detector type for a specific environment
func (q *Queries) GetIssuesByDetectorType(ctx context.Context, environment, detectorType string, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE environment = $1
		AND job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND detector_type = $2
		ORDER BY detected_at DESC
		LIMIT $3
	`

	rows, err := q.db.QueryContext(ctx, query, environment, detectorType, limit)
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

// GetIssuesByFacility gets issues filtered by facility for a specific environment
func (q *Queries) GetIssuesByFacility(ctx context.Context, environment, facility string, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE environment = $1
		AND job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND facility = $2
		ORDER BY detected_at DESC
		LIMIT $3
	`

	rows, err := q.db.QueryContext(ctx, query, environment, facility, limit)
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

// GetIssuesFiltered gets issues with optional filters for a specific environment
func (q *Queries) GetIssuesFiltered(ctx context.Context, environment, detectorType, facility, warehouse string, includeIgnored bool, limit, offset int) ([]*DetectedIssue, error) {
	query := `
		SELECT di.id, di.job_id, di.detector_type, di.detected_at, di.facility, di.warehouse,
			   di.issue_key, di.production_order_number, di.production_order_type,
			   di.co_number, di.co_line, di.co_suffix, di.issue_data, di.created_at,
			   ig.id IS NOT NULL as is_ignored,
			   mot.order_type_description as mo_type_description
		FROM detected_issues di
		LEFT JOIN ignored_issues ig
			ON di.environment = ig.environment
			AND di.facility = ig.facility
			AND di.detector_type = ig.detector_type
			AND di.issue_key = ig.issue_key
			AND di.production_order_number = ig.production_order_number
		LEFT JOIN m3_manufacturing_order_types mot
			ON mot.environment = di.environment
			AND mot.order_type = di.issue_data->>'mo_type'
			AND mot.company_number = di.issue_data->>'company'
		LEFT JOIN planned_manufacturing_orders mop
			ON di.environment = mop.environment
			AND di.production_order_type = 'MOP'
			AND mop.plpn = di.production_order_number
			AND mop.faci = di.facility
		LEFT JOIN manufacturing_orders mo
			ON di.environment = mo.environment
			AND di.production_order_type = 'MO'
			AND mo.mfno = di.production_order_number
			AND mo.faci = di.facility
		WHERE di.environment = $1
		AND di.job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND COALESCE(mop.deleted_remotely, mo.deleted_remotely, false) = false
	`
	args := make([]interface{}, 0)
	args = append(args, environment)
	argNum := 2

	if detectorType != "" {
		query += fmt.Sprintf(" AND di.detector_type = $%d", argNum)
		args = append(args, detectorType)
		argNum++
	}

	if facility != "" {
		query += fmt.Sprintf(" AND di.facility = $%d", argNum)
		args = append(args, facility)
		argNum++
	}

	if warehouse != "" {
		query += fmt.Sprintf(" AND di.warehouse = $%d", argNum)
		args = append(args, warehouse)
		argNum++
	}

	if !includeIgnored {
		query += " AND ig.id IS NULL"
	}

	query += fmt.Sprintf(" ORDER BY di.detected_at DESC OFFSET $%d LIMIT $%d", argNum, argNum+1)
	args = append(args, offset, limit)

	rows, err := q.db.QueryContext(ctx, query, args...)
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
			&issue.IsIgnored,
			&issue.MOTypeDescription,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

// GetRecentIssues gets recent issues (no filter) for a specific environment
func (q *Queries) GetRecentIssues(ctx context.Context, environment string, limit int) ([]*DetectedIssue, error) {
	query := `
		SELECT id, job_id, detector_type, detected_at, facility, warehouse,
			   issue_key, production_order_number, production_order_type,
			   co_number, co_line, co_suffix, issue_data, created_at
		FROM detected_issues
		WHERE environment = $1
		AND job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY detected_at DESC
		LIMIT $2
	`

	rows, err := q.db.QueryContext(ctx, query, environment, limit)
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

// GetIssueCountForLatestJob gets the total issue count for the latest refresh job in a specific environment
func (q *Queries) GetIssueCountForLatestJob(ctx context.Context, environment string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM detected_issues
		WHERE environment = $1
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
		return 0, nil // No issues found
	}
	return count, err
}

// GetIssuesFilteredCount gets the total count of issues matching the filters for a specific environment
func (q *Queries) GetIssuesFilteredCount(ctx context.Context, environment, detectorType, facility, warehouse string, includeIgnored bool) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM detected_issues di
		LEFT JOIN ignored_issues ig
			ON di.environment = ig.environment
			AND di.facility = ig.facility
			AND di.detector_type = ig.detector_type
			AND di.issue_key = ig.issue_key
			AND di.production_order_number = ig.production_order_number
		LEFT JOIN planned_manufacturing_orders mop
			ON di.environment = mop.environment
			AND di.production_order_type = 'MOP'
			AND mop.plpn = di.production_order_number
			AND mop.faci = di.facility
		LEFT JOIN manufacturing_orders mo
			ON di.environment = mo.environment
			AND di.production_order_type = 'MO'
			AND mo.mfno = di.production_order_number
			AND mo.faci = di.facility
		WHERE di.environment = $1
		AND di.job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND COALESCE(mop.deleted_remotely, mo.deleted_remotely, false) = false
	`
	args := make([]interface{}, 0)
	args = append(args, environment)
	argNum := 2

	if detectorType != "" {
		query += fmt.Sprintf(" AND di.detector_type = $%d", argNum)
		args = append(args, detectorType)
		argNum++
	}

	if facility != "" {
		query += fmt.Sprintf(" AND di.facility = $%d", argNum)
		args = append(args, facility)
		argNum++
	}

	if warehouse != "" {
		query += fmt.Sprintf(" AND di.warehouse = $%d", argNum)
		args = append(args, warehouse)
		argNum++
	}

	if !includeIgnored {
		query += " AND ig.id IS NULL"
	}

	var count int
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// GetIssueSummary gets aggregated issue counts with warehouse grouping and nested hierarchy for a specific environment
func (q *Queries) GetIssueSummary(ctx context.Context, environment string, includeIgnored bool) (map[string]interface{}, error) {
	query := `
		SELECT
			di.detector_type,
			di.facility,
			COALESCE(di.warehouse, '') as warehouse,
			COUNT(*) as issue_count
		FROM detected_issues di
		LEFT JOIN ignored_issues ig
			ON di.environment = ig.environment
			AND di.facility = ig.facility
			AND di.detector_type = ig.detector_type
			AND di.issue_key = ig.issue_key
			AND di.production_order_number = ig.production_order_number
		LEFT JOIN planned_manufacturing_orders mop
			ON di.environment = mop.environment
			AND di.production_order_type = 'MOP'
			AND mop.plpn = di.production_order_number
			AND mop.faci = di.facility
		LEFT JOIN manufacturing_orders mo
			ON di.environment = mo.environment
			AND di.production_order_type = 'MO'
			AND mo.mfno = di.production_order_number
			AND mo.faci = di.facility
		WHERE di.environment = $1
		AND di.job_id = (
			SELECT id FROM refresh_jobs
			WHERE environment = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND COALESCE(mop.deleted_remotely, mo.deleted_remotely, false) = false
	`

	if !includeIgnored {
		query += " AND ig.id IS NULL"
	}

	query += `
		GROUP BY di.detector_type, di.facility, di.warehouse
		ORDER BY di.facility, di.warehouse, di.detector_type
	`

	rows, err := q.db.QueryContext(ctx, query, environment)
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

// IgnoreIssue marks an issue as ignored
func (q *Queries) IgnoreIssue(ctx context.Context, params IgnoreIssueParams) error {
	query := `
		INSERT INTO ignored_issues (
			environment, facility, detector_type, issue_key,
			production_order_number, production_order_type,
			co_number, co_line, notes, ignored_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (environment, facility, detector_type, issue_key, production_order_number)
		DO UPDATE SET
			ignored_at = CURRENT_TIMESTAMP,
			notes = EXCLUDED.notes,
			ignored_by = EXCLUDED.ignored_by
	`
	_, err := q.db.ExecContext(ctx, query,
		params.Environment,
		params.Facility,
		params.DetectorType,
		params.IssueKey,
		params.ProductionOrderNumber,
		params.ProductionOrderType,
		params.CONumber,
		params.COLine,
		params.Notes,
		params.IgnoredBy,
	)
	return err
}

// UnignoreIssue removes an issue from ignored list
func (q *Queries) UnignoreIssue(ctx context.Context, params UnignoreIssueParams) error {
	query := `
		DELETE FROM ignored_issues
		WHERE environment = $1
		  AND facility = $2
		  AND detector_type = $3
		  AND issue_key = $4
		  AND production_order_number = $5
	`
	_, err := q.db.ExecContext(ctx, query,
		params.Environment,
		params.Facility,
		params.DetectorType,
		params.IssueKey,
		params.ProductionOrderNumber,
	)
	return err
}

// IsIssueIgnored checks if an issue is ignored
func (q *Queries) IsIssueIgnored(ctx context.Context, params CheckIgnoredParams) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM ignored_issues
			WHERE environment = $1
			  AND facility = $2
			  AND detector_type = $3
			  AND issue_key = $4
			  AND production_order_number = $5
		)
	`
	var exists bool
	err := q.db.QueryRowContext(ctx, query,
		params.Environment,
		params.Facility,
		params.DetectorType,
		params.IssueKey,
		params.ProductionOrderNumber,
	).Scan(&exists)
	return exists, err
}
