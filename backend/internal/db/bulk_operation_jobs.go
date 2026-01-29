package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// BulkOperationJob represents a bulk operation job
type BulkOperationJob struct {
	JobID              string
	Environment        string
	UserID             string
	OperationType      string // 'delete', 'close', 'reschedule'
	Status             string
	IssueIDs           []int64
	Params             map[string]interface{}
	TotalItems         int
	SuccessfulItems    int
	FailedItems        int
	CurrentPhase       sql.NullString
	ProgressPercentage int
	CreatedAt          time.Time
	StartedAt          sql.NullTime
	CompletedAt        sql.NullTime
	ErrorMessage       sql.NullString
}

// BulkOperationBatchResult represents a single result in a batch
type BulkOperationBatchResult struct {
	ID                int64
	JobID             string
	BatchNumber       int
	ProductionOrderID int64
	OrderNumber       string
	OrderType         string // 'MOP' or 'MO'
	Success           bool
	ErrorMessage      sql.NullString
	CreatedAt         time.Time
}

// CreateBulkOperationJobParams contains parameters for creating a bulk operation job
type CreateBulkOperationJobParams struct {
	Environment   string
	UserID        string
	OperationType string
	IssueIDs      []int64
	Params        map[string]interface{}
}

// CreateBulkOperationJob creates a new bulk operation job
func (q *Queries) CreateBulkOperationJob(ctx context.Context, params CreateBulkOperationJobParams) (string, error) {
	jobID := uuid.New().String()

	// Marshal params to JSONB
	var paramsJSON []byte
	var err error
	if params.Params != nil {
		paramsJSON, err = json.Marshal(params.Params)
		if err != nil {
			return "", fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	query := `
		INSERT INTO bulk_operation_jobs (
			job_id, environment, user_id, operation_type, status, issue_ids, params
		) VALUES ($1, $2, $3, $4, 'pending', $5, $6)
	`

	_, err = q.db.ExecContext(ctx, query,
		jobID,
		params.Environment,
		params.UserID,
		params.OperationType,
		pq.Array(params.IssueIDs),
		paramsJSON,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create bulk operation job: %w", err)
	}

	return jobID, nil
}

// GetBulkOperationJob retrieves a bulk operation job by ID
func (q *Queries) GetBulkOperationJob(ctx context.Context, jobID string) (*BulkOperationJob, error) {
	query := `
		SELECT
			job_id, environment, user_id, operation_type, status,
			issue_ids, params,
			total_items, successful_items, failed_items,
			current_phase, progress_percentage,
			created_at, started_at, completed_at, error_message
		FROM bulk_operation_jobs
		WHERE job_id = $1
	`

	job := &BulkOperationJob{}
	var issueIDs pq.Int64Array
	var paramsJSON []byte

	err := q.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.JobID, &job.Environment, &job.UserID, &job.OperationType, &job.Status,
		&issueIDs, &paramsJSON,
		&job.TotalItems, &job.SuccessfulItems, &job.FailedItems,
		&job.CurrentPhase, &job.ProgressPercentage,
		&job.CreatedAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMessage,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bulk operation job not found: %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk operation job: %w", err)
	}

	// Convert arrays and JSON
	job.IssueIDs = []int64(issueIDs)
	if len(paramsJSON) > 0 {
		if err := json.Unmarshal(paramsJSON, &job.Params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}
	}

	return job, nil
}

// UpdateBulkOperationJobProgress updates the progress of a bulk operation job
func (q *Queries) UpdateBulkOperationJobProgress(ctx context.Context, jobID, currentPhase string, totalItems, successfulItems, failedItems, progressPct int) error {
	query := `
		UPDATE bulk_operation_jobs
		SET current_phase = $2,
		    total_items = $3,
		    successful_items = $4,
		    failed_items = $5,
		    progress_percentage = $6
		WHERE job_id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID, currentPhase, totalItems, successfulItems, failedItems, progressPct)
	return err
}

// StartBulkOperationJob marks a job as started
func (q *Queries) StartBulkOperationJob(ctx context.Context, jobID string) error {
	query := `
		UPDATE bulk_operation_jobs
		SET status = 'running',
		    started_at = NOW()
		WHERE job_id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// CompleteBulkOperationJob marks a job as completed
func (q *Queries) CompleteBulkOperationJob(ctx context.Context, jobID string) error {
	query := `
		UPDATE bulk_operation_jobs
		SET status = 'completed',
		    completed_at = NOW(),
		    progress_percentage = 100
		WHERE job_id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// FailBulkOperationJob marks a job as failed with an error message
func (q *Queries) FailBulkOperationJob(ctx context.Context, jobID, errorMsg string) error {
	query := `
		UPDATE bulk_operation_jobs
		SET status = 'failed',
		    error_message = $2,
		    completed_at = NOW()
		WHERE job_id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID, errorMsg)
	return err
}

// CancelBulkOperationJob marks a job as cancelled
func (q *Queries) CancelBulkOperationJob(ctx context.Context, jobID string) error {
	query := `
		UPDATE bulk_operation_jobs
		SET status = 'cancelled',
		    completed_at = NOW()
		WHERE job_id = $1 AND status IN ('pending', 'running')
	`
	result, err := q.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("job not found or not in cancellable state")
	}

	return nil
}

// InsertBulkOperationBatchResults inserts batch results for a job
func (q *Queries) InsertBulkOperationBatchResults(ctx context.Context, jobID string, batchNumber int, results []BulkOperationBatchResult) error {
	if len(results) == 0 {
		return nil
	}

	// Build bulk insert query
	query := `
		INSERT INTO bulk_operation_batch_results (
			job_id, batch_number, production_order_id, order_number, order_type, success, error_message
		) VALUES
	`

	values := []interface{}{}
	placeholders := []string{}
	idx := 1

	for _, result := range results {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", idx, idx+1, idx+2, idx+3, idx+4, idx+5, idx+6))
		values = append(values, jobID, batchNumber, result.ProductionOrderID, result.OrderNumber, result.OrderType, result.Success, result.ErrorMessage)
		idx += 7
	}

	query += " " + fmt.Sprintf("%s", placeholders[0])
	for i := 1; i < len(placeholders); i++ {
		query += ", " + placeholders[i]
	}

	_, err := q.db.ExecContext(ctx, query, values...)
	return err
}

// GetBulkOperationJobResults retrieves all batch results for a job
func (q *Queries) GetBulkOperationJobResults(ctx context.Context, jobID string) ([]BulkOperationBatchResult, error) {
	query := `
		SELECT
			id, job_id, batch_number, production_order_id, order_number,
			order_type, success, error_message, created_at
		FROM bulk_operation_batch_results
		WHERE job_id = $1
		ORDER BY batch_number, id
	`

	rows, err := q.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch results: %w", err)
	}
	defer rows.Close()

	var results []BulkOperationBatchResult
	for rows.Next() {
		var result BulkOperationBatchResult
		err := rows.Scan(
			&result.ID, &result.JobID, &result.BatchNumber, &result.ProductionOrderID,
			&result.OrderNumber, &result.OrderType, &result.Success, &result.ErrorMessage,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch result: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// ListBulkOperationJobs lists bulk operation jobs for a user
func (q *Queries) ListBulkOperationJobs(ctx context.Context, environment, userID string, limit int) ([]BulkOperationJob, error) {
	query := `
		SELECT
			job_id, environment, user_id, operation_type, status,
			issue_ids, params,
			total_items, successful_items, failed_items,
			current_phase, progress_percentage,
			created_at, started_at, completed_at, error_message
		FROM bulk_operation_jobs
		WHERE environment = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := q.db.QueryContext(ctx, query, environment, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query bulk operation jobs: %w", err)
	}
	defer rows.Close()

	var jobs []BulkOperationJob
	for rows.Next() {
		var job BulkOperationJob
		var issueIDs pq.Int64Array
		var paramsJSON []byte

		err := rows.Scan(
			&job.JobID, &job.Environment, &job.UserID, &job.OperationType, &job.Status,
			&issueIDs, &paramsJSON,
			&job.TotalItems, &job.SuccessfulItems, &job.FailedItems,
			&job.CurrentPhase, &job.ProgressPercentage,
			&job.CreatedAt, &job.StartedAt, &job.CompletedAt, &job.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bulk operation job: %w", err)
		}

		// Convert arrays and JSON
		job.IssueIDs = []int64(issueIDs)
		if len(paramsJSON) > 0 {
			if err := json.Unmarshal(paramsJSON, &job.Params); err != nil {
				return nil, fmt.Errorf("failed to unmarshal params: %w", err)
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}
