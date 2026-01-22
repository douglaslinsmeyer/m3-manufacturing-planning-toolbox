package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
}

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	JobID          string `json:"jobId"`
	Status         string `json:"status"`
	CurrentStep    string `json:"currentStep"`
	CompletedSteps int    `json:"completedSteps"`
	TotalSteps     int    `json:"totalSteps"`
	ProgressPct    int    `json:"progressPercentage"`
	COLines        int    `json:"coLinesProcessed,omitempty"`
	MOs            int    `json:"mosProcessed,omitempty"`
	MOPs           int    `json:"mopsProcessed,omitempty"`
	Error          string `json:"error,omitempty"`
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

	// Start the job
	if err := w.db.StartJob(ctx, req.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Publish initial progress
	w.publishProgress(req.JobID, "running", "Starting data refresh", 0, 4, 0, 0, 0, 0)

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

	// Create snapshot service
	snapshotService := services.NewSnapshotService(compassClient, w.db)

	// Execute the full refresh using the service (it handles all orchestration)
	w.publishProgress(req.JobID, "running", "Loading production orders", 0, 3, 0, 0, 0, 0)

	if err := snapshotService.RefreshAll(ctx); err != nil {
		w.db.FailJob(ctx, req.JobID, fmt.Sprintf("Refresh failed: %v", err))
		w.publishError(req.JobID, err.Error())
		return err
	}

	// Complete the job
	w.db.CompleteJob(ctx, req.JobID)
	w.publishProgress(req.JobID, "completed", "Data refresh completed", 3, 3, 100, 0, 0, 0)
	w.publishComplete(req.JobID)

	log.Printf("Refresh job %s completed successfully", req.JobID)
	return nil
}

// publishProgress publishes a progress update to NATS
func (w *SnapshotWorker) publishProgress(jobID, status, currentStep string, completedSteps, totalSteps, progressPct, coLines, mos, mops int) {
	update := ProgressUpdate{
		JobID:          jobID,
		Status:         status,
		CurrentStep:    currentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     totalSteps,
		ProgressPct:    progressPct,
		COLines:        coLines,
		MOs:            mos,
		MOPs:           mops,
	}

	data, _ := json.Marshal(update)
	subject := queue.GetProgressSubject(jobID)

	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish progress: %v", err)
	}
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
