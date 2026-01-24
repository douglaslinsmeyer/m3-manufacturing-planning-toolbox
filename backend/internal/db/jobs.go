package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RefreshJob represents a data refresh job
type RefreshJob struct {
	ID                        string
	Environment               string
	UserID                    sql.NullString
	Status                    string
	CurrentStep               sql.NullString
	TotalSteps                int
	CompletedSteps            int
	ProgressPct               int
	COLinesProcessed          int
	MOsProcessed              int
	MOPsProcessed             int
	RecordsPerSecond          sql.NullFloat64
	EstimatedSecondsRemaining sql.NullInt32
	CurrentOperation          sql.NullString
	CurrentBatch              sql.NullInt32
	TotalBatches              sql.NullInt32
	StartedAt                 sql.NullTime
	CompletedAt               sql.NullTime
	DurationSeconds           sql.NullInt32
	ErrorMessage              sql.NullString
	RetryCount                int
	MaxRetries                int
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// CreateRefreshJob creates a new refresh job
func (q *Queries) CreateRefreshJob(ctx context.Context, jobID, environment, userID string) error {
	query := `
		INSERT INTO refresh_jobs (
			id, environment, user_id, status, total_steps, max_retries
		) VALUES ($1, $2, $3, 'pending', 3, 3)
	`
	_, err := q.db.ExecContext(ctx, query, jobID, environment, userID)
	return err
}

// UpdateJobStatus updates the status of a refresh job
func (q *Queries) UpdateJobStatus(ctx context.Context, jobID, status string) error {
	query := `
		UPDATE refresh_jobs
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := q.db.ExecContext(ctx, query, status, jobID)
	return err
}

// UpdateJobProgress updates the progress of a refresh job
func (q *Queries) UpdateJobProgress(ctx context.Context, jobID, currentStep string, completedSteps, totalSteps int) error {
	progressPct := 0
	if totalSteps > 0 {
		progressPct = (completedSteps * 100) / totalSteps
	}

	query := `
		UPDATE refresh_jobs
		SET current_step = $1,
		    completed_steps = $2,
		    total_steps = $3,
		    progress_percentage = $4,
		    updated_at = NOW()
		WHERE id = $5
	`
	_, err := q.db.ExecContext(ctx, query, currentStep, completedSteps, totalSteps, progressPct, jobID)
	return err
}

// UpdateJobRecordCounts updates the processed record counts
func (q *Queries) UpdateJobRecordCounts(ctx context.Context, jobID string, coLines, mos, mops int) error {
	query := `
		UPDATE refresh_jobs
		SET co_lines_processed = $1,
		    mos_processed = $2,
		    mops_processed = $3,
		    updated_at = NOW()
		WHERE id = $4
	`
	_, err := q.db.ExecContext(ctx, query, coLines, mos, mops, jobID)
	return err
}

// UpdateJobExtendedProgress updates extended progress information
func (q *Queries) UpdateJobExtendedProgress(ctx context.Context, jobID, currentOperation string, recordsPerSecond float64, estimatedSecondsRemaining, currentBatch, totalBatches int) error {
	query := `
		UPDATE refresh_jobs
		SET current_operation = $1,
		    records_per_second = $2,
		    estimated_seconds_remaining = $3,
		    current_batch = $4,
		    total_batches = $5,
		    updated_at = NOW()
		WHERE id = $6
	`
	_, err := q.db.ExecContext(ctx, query, currentOperation, recordsPerSecond, estimatedSecondsRemaining, currentBatch, totalBatches, jobID)
	return err
}

// StartJob marks a job as started
func (q *Queries) StartJob(ctx context.Context, jobID string) error {
	query := `
		UPDATE refresh_jobs
		SET status = 'running',
		    started_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// CompleteJob marks a job as completed
func (q *Queries) CompleteJob(ctx context.Context, jobID string) error {
	query := `
		UPDATE refresh_jobs
		SET status = 'completed',
		    completed_at = NOW(),
		    progress_percentage = 100,
		    duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// FailJob marks a job as failed with an error message
func (q *Queries) FailJob(ctx context.Context, jobID, errorMsg string) error {
	query := `
		UPDATE refresh_jobs
		SET status = 'failed',
		    error_message = $1,
		    completed_at = NOW(),
		    duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err := q.db.ExecContext(ctx, query, errorMsg, jobID)
	return err
}

// IncrementRetryCount increments the retry count for a job
func (q *Queries) IncrementRetryCount(ctx context.Context, jobID string) error {
	query := `
		UPDATE refresh_jobs
		SET retry_count = retry_count + 1,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.ExecContext(ctx, query, jobID)
	return err
}

// GetRefreshJob gets a refresh job by ID
func (q *Queries) GetRefreshJob(ctx context.Context, jobID string) (*RefreshJob, error) {
	query := `
		SELECT
			id, environment, user_id, status,
			current_step, total_steps, completed_steps, progress_percentage,
			co_lines_processed, mos_processed, mops_processed,
			records_per_second, estimated_seconds_remaining,
			current_operation, current_batch, total_batches,
			started_at, completed_at, duration_seconds,
			error_message, retry_count, max_retries,
			created_at, updated_at
		FROM refresh_jobs
		WHERE id = $1
	`

	job := &RefreshJob{}
	err := q.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.Environment, &job.UserID, &job.Status,
		&job.CurrentStep, &job.TotalSteps, &job.CompletedSteps, &job.ProgressPct,
		&job.COLinesProcessed, &job.MOsProcessed, &job.MOPsProcessed,
		&job.RecordsPerSecond, &job.EstimatedSecondsRemaining,
		&job.CurrentOperation, &job.CurrentBatch, &job.TotalBatches,
		&job.StartedAt, &job.CompletedAt, &job.DurationSeconds,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// GetLatestRefreshJob gets the most recent refresh job for an environment
func (q *Queries) GetLatestRefreshJob(ctx context.Context, environment string) (*RefreshJob, error) {
	query := `
		SELECT
			id, environment, user_id, status,
			current_step, total_steps, completed_steps, progress_percentage,
			co_lines_processed, mos_processed, mops_processed,
			records_per_second, estimated_seconds_remaining,
			current_operation, current_batch, total_batches,
			started_at, completed_at, duration_seconds,
			error_message, retry_count, max_retries,
			created_at, updated_at
		FROM refresh_jobs
		WHERE environment = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	job := &RefreshJob{}
	err := q.db.QueryRowContext(ctx, query, environment).Scan(
		&job.ID, &job.Environment, &job.UserID, &job.Status,
		&job.CurrentStep, &job.TotalSteps, &job.CompletedSteps, &job.ProgressPct,
		&job.COLinesProcessed, &job.MOsProcessed, &job.MOPsProcessed,
		&job.RecordsPerSecond, &job.EstimatedSecondsRemaining,
		&job.CurrentOperation, &job.CurrentBatch, &job.TotalBatches,
		&job.StartedAt, &job.CompletedAt, &job.DurationSeconds,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No jobs yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest job: %w", err)
	}

	return job, nil
}

// GetActiveRefreshJob gets the currently running or pending refresh job for an environment
// Returns nil if no active job exists
func (q *Queries) GetActiveRefreshJob(ctx context.Context, environment string) (*RefreshJob, error) {
	query := `
		SELECT
			id, environment, user_id, status,
			current_step, total_steps, completed_steps, progress_percentage,
			co_lines_processed, mos_processed, mops_processed,
			records_per_second, estimated_seconds_remaining,
			current_operation, current_batch, total_batches,
			started_at, completed_at, duration_seconds,
			error_message, retry_count, max_retries,
			created_at, updated_at
		FROM refresh_jobs
		WHERE environment = $1
		  AND status IN ('pending', 'running')
		ORDER BY created_at DESC
		LIMIT 1
	`

	job := &RefreshJob{}
	err := q.db.QueryRowContext(ctx, query, environment).Scan(
		&job.ID, &job.Environment, &job.UserID, &job.Status,
		&job.CurrentStep, &job.TotalSteps, &job.CompletedSteps, &job.ProgressPct,
		&job.COLinesProcessed, &job.MOsProcessed, &job.MOPsProcessed,
		&job.RecordsPerSecond, &job.EstimatedSecondsRemaining,
		&job.CurrentOperation, &job.CurrentBatch, &job.TotalBatches,
		&job.StartedAt, &job.CompletedAt, &job.DurationSeconds,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No active job
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active job: %w", err)
	}

	return job, nil
}
