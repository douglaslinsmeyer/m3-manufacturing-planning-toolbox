package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// SnapshotWorker handles async snapshot refresh jobs
type SnapshotWorker struct {
	nats   *queue.Manager
	db     *db.Queries
	config *config.Config
}

// NewSnapshotWorker creates a new snapshot worker
func NewSnapshotWorker(nats *queue.Manager, database *db.Queries, cfg *config.Config) *SnapshotWorker {
	return &SnapshotWorker{
		nats:   nats,
		db:     database,
		config: cfg,
	}
}

// SnapshotRefreshMessage represents a snapshot refresh request
type SnapshotRefreshMessage struct {
	JobID       string `json:"jobId"`
	Environment string `json:"environment"`
	UserID      string `json:"userId,omitempty"`
	AccessToken string `json:"accessToken"`
	Company     string `json:"company"`
	Facility    string `json:"facility"`
}

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	JobID                     string  `json:"jobId"`
	Status                    string  `json:"status"`
	Progress                  int     `json:"progress"`
	CurrentStep               string  `json:"currentStep"`
	CompletedSteps            int     `json:"completedSteps"`
	TotalSteps                int     `json:"totalSteps"`
	COLinesProcessed          int     `json:"coLinesProcessed,omitempty"`
	MOsProcessed              int     `json:"mosProcessed,omitempty"`
	MOPsProcessed             int     `json:"mopsProcessed,omitempty"`
	RecordsPerSecond          float64 `json:"recordsPerSecond,omitempty"`
	EstimatedSecondsRemaining int     `json:"estimatedTimeRemaining,omitempty"`
	CurrentOperation          string  `json:"currentOperation,omitempty"`
	CurrentBatch              int     `json:"currentBatch,omitempty"`
	TotalBatches              int     `json:"totalBatches,omitempty"`
	Error                     string  `json:"error,omitempty"`
}

// Start starts the snapshot worker and subscribes to NATS subjects
func (w *SnapshotWorker) Start() error {
	log.Println("Starting snapshot worker...")

	// Subscribe to TRN refresh requests
	_, err := w.nats.QueueSubscribe(
		queue.SubjectSnapshotRefreshTRN,
		queue.QueueGroupSnapshot,
		w.handleRefreshRequest,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to TRN refresh: %w", err)
	}

	// Subscribe to PRD refresh requests
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotRefreshPRD,
		queue.QueueGroupSnapshot,
		w.handleRefreshRequest,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to PRD refresh: %w", err)
	}

	log.Println("Snapshot worker started and listening for jobs")
	return nil
}

// handleRefreshRequest handles a snapshot refresh request
func (w *SnapshotWorker) handleRefreshRequest(msg *nats.Msg) {
	log.Printf("Received refresh request on subject: %s", msg.Subject)

	// Parse message
	var req SnapshotRefreshMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("Failed to parse refresh request: %v", err)
		return
	}

	// Process the refresh with error recovery
	if err := w.processRefreshWithRetry(req); err != nil {
		log.Printf("Refresh job %s failed after retries: %v", req.JobID, err)
	}
}

// processRefreshWithRetry processes a refresh job with automatic retry on failure
func (w *SnapshotWorker) processRefreshWithRetry(req SnapshotRefreshMessage) error {
	ctx := context.Background()

	// Get job details to check retry count
	job, err := w.db.GetRefreshJob(ctx, req.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Check if max retries exceeded
	if job.RetryCount >= job.MaxRetries {
		log.Printf("Job %s exceeded max retries (%d)", req.JobID, job.MaxRetries)
		w.db.FailJob(ctx, req.JobID, fmt.Sprintf("Exceeded maximum retries (%d)", job.MaxRetries))
		w.publishError(req.JobID, fmt.Sprintf("Exceeded maximum retries (%d)", job.MaxRetries))
		return fmt.Errorf("max retries exceeded")
	}

	// Process the job
	if err := w.processRefresh(req); err != nil {
		// Increment retry count
		w.db.IncrementRetryCount(ctx, req.JobID)

		// Get updated job
		job, _ := w.db.GetRefreshJob(ctx, req.JobID)

		// Check if we should retry
		if job.RetryCount < job.MaxRetries {
			log.Printf("Job %s failed (attempt %d/%d), will retry: %v", req.JobID, job.RetryCount, job.MaxRetries, err)
			// Re-publish to NATS for retry after a delay
			// For now, just log - in production we'd use NATS delayed retry pattern
			return err
		}

		// Max retries exceeded
		w.db.FailJob(ctx, req.JobID, err.Error())
		w.publishError(req.JobID, err.Error())
		return err
	}

	return nil
}

// processRefresh performs the actual data refresh
func (w *SnapshotWorker) processRefresh(req SnapshotRefreshMessage) error {
	ctx := context.Background()
	startTime := time.Now()

	// Start the job
	if err := w.db.StartJob(ctx, req.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Progress tracking variables
	totalSteps := 4
	var finalMopCount, finalMoCount, finalCoCount int

	// Publish initial progress
	w.publishDetailedProgress(req.JobID, "running", "Starting data refresh", "Initializing refresh process",
		0, totalSteps, 0, 0, 0, 0, 0, 0, 0, 0)

	// Get environment config
	envConfig, err := w.config.GetEnvironmentConfig(req.Environment)
	if err != nil {
		w.db.FailJob(ctx, req.JobID, err.Error())
		w.publishError(req.JobID, err.Error())
		return err
	}

	// Create Compass client with the provided access token
	getToken := func() (string, error) {
		return req.AccessToken, nil
	}
	compassClient := compass.NewClient(envConfig.CompassBaseURL, getToken)

	// Create snapshot service with progress callback
	snapshotService := services.NewSnapshotService(compassClient, w.db)

	// Set up progress callback to track phases
	snapshotService.SetProgressCallback(func(phase string, stepNum, total int, message string, mopCount, moCount, coCount int) {
		// Store final counts
		finalMopCount = mopCount
		finalMoCount = moCount
		finalCoCount = coCount
		elapsed := time.Since(startTime).Seconds()

		// Map phases to operations
		var operation string
		switch phase {
		case "truncate":
			operation = "Preparing database"
		case "mops":
			operation = message
		case "mos":
			operation = message
		case "cos":
			operation = message
		case "finalize":
			operation = "Finalizing data refresh"
		case "complete":
			operation = "Data refresh completed"
		default:
			operation = message
		}

		progressPct := (stepNum * 100) / totalSteps

		// Calculate rate and ETA
		totalRecords := mopCount + moCount + coCount
		recordsPerSec := 0.0
		estimatedRemaining := 0

		if elapsed > 0 && totalRecords > 0 {
			recordsPerSec = float64(totalRecords) / elapsed
			remainingSteps := totalSteps - stepNum
			if remainingSteps > 0 {
				// Rough estimate: assume similar record counts per remaining step
				avgRecordsPerStep := float64(totalRecords) / float64(stepNum+1)
				remainingRecords := avgRecordsPerStep * float64(remainingSteps)
				estimatedRemaining = int(remainingRecords / recordsPerSec)
			}
		}

		// Publish detailed progress
		w.publishDetailedProgress(req.JobID, "running",
			operation, // Send operation description as currentStep
			operation, // Keep same for currentOperation
			stepNum, totalSteps, progressPct,
			coCount, moCount, mopCount,
			recordsPerSec, estimatedRemaining,
			0, 0) // No batch tracking yet
	})

	// Execute the full refresh using the service (it handles all orchestration)
	log.Printf("Starting refresh for company %s, facility %s", req.Company, req.Facility)

	if err := snapshotService.RefreshAll(ctx, req.Company, req.Facility); err != nil {
		w.db.FailJob(ctx, req.JobID, fmt.Sprintf("Refresh failed: %v", err))
		w.publishError(req.JobID, err.Error())
		return err
	}

	// Get final counts and calculate final metrics
	elapsed := time.Since(startTime).Seconds()
	totalRecords := finalMopCount + finalMoCount + finalCoCount
	recordsPerSec := 0.0
	if elapsed > 0 {
		recordsPerSec = float64(totalRecords) / elapsed
	}

	// Complete the job
	w.db.CompleteJob(ctx, req.JobID)
	w.publishDetailedProgress(req.JobID, "completed", "Data refresh completed", "All data successfully loaded",
		totalSteps, totalSteps, 100,
		finalCoCount, finalMoCount, finalMopCount,
		recordsPerSec, 0, 0, 0)
	w.publishComplete(req.JobID)

	log.Printf("Refresh job %s completed successfully in %.2f seconds", req.JobID, elapsed)
	return nil
}

// publishProgress publishes a progress update to NATS
func (w *SnapshotWorker) publishProgress(jobID, status, currentStep string, completedSteps, totalSteps, progressPct, coLines, mos, mops int) {
	update := ProgressUpdate{
		JobID:            jobID,
		Status:           status,
		Progress:         progressPct,
		CurrentStep:      currentStep,
		CompletedSteps:   completedSteps,
		TotalSteps:       totalSteps,
		COLinesProcessed: coLines,
		MOsProcessed:     mos,
		MOPsProcessed:    mops,
	}

	data, _ := json.Marshal(update)
	subject := queue.GetProgressSubject(jobID)

	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish progress: %v", err)
	}

	// Also update database
	ctx := context.Background()
	w.db.UpdateJobProgress(ctx, jobID, currentStep, completedSteps, totalSteps)
}

// publishDetailedProgress publishes a detailed progress update with extended metrics
func (w *SnapshotWorker) publishDetailedProgress(jobID, status, currentStep, currentOperation string, completedSteps, totalSteps, progressPct, coLines, mos, mops int, recordsPerSec float64, estimatedSecsRemaining, currentBatch, totalBatches int) {
	update := ProgressUpdate{
		JobID:                     jobID,
		Status:                    status,
		Progress:                  progressPct,
		CurrentStep:               currentStep,
		CompletedSteps:            completedSteps,
		TotalSteps:                totalSteps,
		COLinesProcessed:          coLines,
		MOsProcessed:              mos,
		MOPsProcessed:             mops,
		RecordsPerSecond:          recordsPerSec,
		EstimatedSecondsRemaining: estimatedSecsRemaining,
		CurrentOperation:          currentOperation,
		CurrentBatch:              currentBatch,
		TotalBatches:              totalBatches,
	}

	data, _ := json.Marshal(update)
	subject := queue.GetProgressSubject(jobID)

	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish progress: %v", err)
	}

	// Update database with extended progress
	ctx := context.Background()
	w.db.UpdateJobProgress(ctx, jobID, currentStep, completedSteps, totalSteps)
	w.db.UpdateJobRecordCounts(ctx, jobID, coLines, mos, mops)
	w.db.UpdateJobExtendedProgress(ctx, jobID, currentOperation, recordsPerSec, estimatedSecsRemaining, currentBatch, totalBatches)
}

// publishComplete publishes a completion message
func (w *SnapshotWorker) publishComplete(jobID string) {
	subject := queue.GetCompleteSubject(jobID)
	data := []byte(fmt.Sprintf(`{"jobId":"%s","status":"completed"}`, jobID))

	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish completion: %v", err)
	}
}

// publishError publishes an error message
func (w *SnapshotWorker) publishError(jobID, errorMsg string) {
	update := ProgressUpdate{
		JobID:  jobID,
		Status: "failed",
		Error:  errorMsg,
	}

	data, _ := json.Marshal(update)
	subject := queue.GetErrorSubject(jobID)

	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish error: %v", err)
	}
}
