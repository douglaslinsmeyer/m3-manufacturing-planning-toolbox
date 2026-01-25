package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
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

// PhaseProgress represents the status of a single parallel phase
type PhaseProgress struct {
	Phase       string    `json:"phase"`             // "mops", "mos", "cos"
	Status      string    `json:"status"`            // "pending", "running", "completed", "failed"
	RecordCount int       `json:"recordCount"`       // Records processed
	StartTime   time.Time `json:"startTime,omitempty"`
	EndTime     time.Time `json:"endTime,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	JobID                     string          `json:"jobId"`
	Status                    string          `json:"status"`
	Progress                  int             `json:"progress"`
	CurrentStep               string          `json:"currentStep"`
	CompletedSteps            int             `json:"completedSteps"`
	TotalSteps                int             `json:"totalSteps"`
	ParallelPhases            []PhaseProgress `json:"parallelPhases,omitempty"` // NEW: Parallel phase tracking
	COLinesProcessed          int             `json:"coLinesProcessed,omitempty"`
	MOsProcessed              int             `json:"mosProcessed,omitempty"`
	MOPsProcessed             int             `json:"mopsProcessed,omitempty"`
	RecordsPerSecond          float64         `json:"recordsPerSecond,omitempty"`
	EstimatedSecondsRemaining int             `json:"estimatedTimeRemaining,omitempty"`
	CurrentOperation          string          `json:"currentOperation,omitempty"`
	CurrentBatch              int             `json:"currentBatch,omitempty"`
	TotalBatches              int             `json:"totalBatches,omitempty"`
	Error                     string          `json:"error,omitempty"`
}

// PhaseJobMessage represents work for a single phase
type PhaseJobMessage struct {
	JobID       string `json:"jobId"`
	ParentJobID string `json:"parentJobId"` // Links to main refresh job
	Phase       string `json:"phase"`       // "mops", "mos", "cos"
	Environment string `json:"environment"`
	AccessToken string `json:"accessToken"`
	Company     string `json:"company"`
	Facility    string `json:"facility"`
}

// PhaseCompletionMessage signals phase completion
type PhaseCompletionMessage struct {
	JobID       string `json:"jobId"`
	ParentJobID string `json:"parentJobID"`
	Phase       string `json:"phase"`
	RecordCount int    `json:"recordCount"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// Start starts the snapshot worker and subscribes to NATS subjects
func (w *SnapshotWorker) Start() error {
	log.Println("Starting snapshot worker...")

	// Subscribe to TRN refresh requests (coordinator)
	_, err := w.nats.QueueSubscribe(
		queue.SubjectSnapshotRefreshTRN,
		queue.QueueGroupSnapshot,
		w.handleRefreshRequest,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to TRN refresh: %w", err)
	}

	// Subscribe to PRD refresh requests (coordinator)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotRefreshPRD,
		queue.QueueGroupSnapshot,
		w.handleRefreshRequest,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to PRD refresh: %w", err)
	}

	// Subscribe to TRN phase work distribution (phase worker)
	// Use wildcard subscription to catch all phase types
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotPhaseTRN,
		queue.QueueGroupPhaseWorkers,
		w.handlePhaseJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to TRN phase jobs: %w", err)
	}

	// Subscribe to PRD phase work distribution (phase worker)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotPhasePRD,
		queue.QueueGroupPhaseWorkers,
		w.handlePhaseJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to PRD phase jobs: %w", err)
	}

	// IMPORTANT: Limit concurrent message processing by handling synchronously
	// Each worker will process one message at a time, allowing NATS to distribute
	// the remaining messages to other available workers

	log.Println("Snapshot worker started and listening for jobs and phase work")
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

// processRefresh coordinates the parallel data refresh using NATS phase distribution
func (w *SnapshotWorker) processRefresh(req SnapshotRefreshMessage) error {
	ctx := context.Background()
	log.Printf("Coordinating refresh job %s", req.JobID)

	// Start the job
	if err := w.db.StartJob(ctx, req.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Phase 0: Truncate database (must complete first)
	log.Printf("Phase 0: Truncating database for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Preparing database", "Truncating tables",
		0, 5, 0, 0, 0, 0, 0, 0, 0, 0)

	envConfig, err := w.config.GetEnvironmentConfig(req.Environment)
	if err != nil {
		w.db.FailJob(ctx, req.JobID, err.Error())
		w.publishError(req.JobID, err.Error())
		return err
	}

	getToken := func() (string, error) {
		return req.AccessToken, nil
	}
	compassClient := compass.NewClient(envConfig.CompassBaseURL, getToken)
	snapshotService := services.NewSnapshotService(compassClient, w.db)

	if err := w.db.TruncateAnalysisTables(ctx); err != nil {
		w.publishError(req.JobID, fmt.Sprintf("Truncate failed: %v", err))
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf("truncate failed: %w", err)
	}

	log.Printf("Phase 0 complete: Database truncated")
	w.publishDetailedProgress(req.JobID, "running", "Database prepared", "Starting parallel data load",
		1, 5, 20, 0, 0, 0, 0, 0, 0, 0)

	// Phases 1-3: Publish 3 phase jobs to NATS queue for parallel execution
	phases := []string{"mops", "mos", "cos"}
	for i, phase := range phases {
		phaseJob := PhaseJobMessage{
			JobID:       fmt.Sprintf("%s-%s", req.JobID, phase),
			ParentJobID: req.JobID,
			Phase:       phase,
			Environment: req.Environment,
			AccessToken: req.AccessToken,
			Company:     req.Company,
			Facility:    req.Facility,
		}

		data, _ := json.Marshal(phaseJob)
		subject := queue.GetPhaseSubject(req.Environment, req.JobID, phase)
		if err := w.nats.Publish(subject, data); err != nil {
			log.Printf("Failed to publish phase job %s: %v", phase, err)
			w.publishError(req.JobID, fmt.Sprintf("Failed to publish phase %s: %v", phase, err))
			w.db.FailJob(ctx, req.JobID, err.Error())
			return fmt.Errorf("failed to publish phase %s: %w", phase, err)
		}
		log.Printf("Published phase job: %s to %s", phase, subject)

		// Add small delay between publishes to allow NATS to distribute round-robin
		// This prevents one worker from grabbing all messages before others can respond
		if i < len(phases)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Wait for all phases to complete and then finalize
	return w.waitForPhasesAndFinalize(req, snapshotService)
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

// waitForPhasesAndFinalize waits for all 3 phases to complete, then finalizes
func (w *SnapshotWorker) waitForPhasesAndFinalize(req SnapshotRefreshMessage, snapshotService *services.SnapshotService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Track phase completions
	completions := make(map[string]*PhaseCompletionMessage)
	var mu sync.Mutex

	// Subscribe to phase completion events for this job
	completeSubject := queue.GetPhaseCompleteSubject(req.JobID)
	sub, err := w.nats.Subscribe(completeSubject, func(msg *nats.Msg) {
		var completion PhaseCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to parse completion: %v", err)
			return
		}

		mu.Lock()
		completions[completion.Phase] = &completion
		log.Printf("Phase %s completed: %d records", completion.Phase, completion.RecordCount)

		// Update progress
		w.publishPhaseProgress(req.JobID, completions)
		mu.Unlock()
	})
	if err != nil {
		log.Printf("Failed to subscribe to completions: %v", err)
		w.publishError(req.JobID, "Failed to subscribe to phase completions")
		w.db.FailJob(context.Background(), req.JobID, "subscription error")
		return err
	}
	defer sub.Unsubscribe()

	// Wait for all 3 phases (with timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.publishError(req.JobID, "Phase completion timeout")
			w.db.FailJob(context.Background(), req.JobID, "timeout waiting for phases")
			return fmt.Errorf("timeout waiting for phases")

		case <-ticker.C:
			mu.Lock()
			allComplete := len(completions) == 3
			anyFailed := false
			for _, c := range completions {
				if !c.Success {
					anyFailed = true
					break
				}
			}
			mu.Unlock()

			if allComplete {
				if anyFailed {
					w.publishError(req.JobID, "One or more phases failed")
					w.db.FailJob(context.Background(), req.JobID, "phase failure")
					return fmt.Errorf("one or more phases failed")
				}
				// All phases successful - run finalize
				return w.runFinalize(req, completions)
			}
		}
	}
}

// runFinalize executes finalize and detection phases
func (w *SnapshotWorker) runFinalize(req SnapshotRefreshMessage, completions map[string]*PhaseCompletionMessage) error {
	ctx := context.Background()

	// Phase 4: Finalize unified views
	log.Printf("Phase 4: Running finalize for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Finalizing data", "Updating production orders",
		4, 5, 80,
		getRecordCount(completions, "cos"),
		getRecordCount(completions, "mos"),
		getRecordCount(completions, "mops"),
		0, 0, 0, 0)

	if err := w.db.UpdateProductionOrdersFromMOPs(ctx); err != nil {
		w.publishError(req.JobID, fmt.Sprintf("Finalize MOPs failed: %v", err))
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf("finalize MOPs failed: %w", err)
	}
	if err := w.db.UpdateProductionOrdersFromMOs(ctx); err != nil {
		w.publishError(req.JobID, fmt.Sprintf("Finalize MOs failed: %v", err))
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf("finalize MOs failed: %w", err)
	}

	// Phase 5: Detection
	log.Printf("Phase 5: Running detection for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Running issue detectors", "Analyzing data",
		5, 5, 90,
		getRecordCount(completions, "cos"),
		getRecordCount(completions, "mos"),
		getRecordCount(completions, "mops"),
		0, 0, 0, 0)

	detectionService := services.NewDetectionService(w.db)
	if err := detectionService.RunAllDetectors(ctx, req.JobID, req.Company, req.Facility); err != nil {
		log.Printf("Detection warning: %v", err)
		// Don't fail job on detection errors
	}

	// Mark job complete
	w.db.CompleteJob(ctx, req.JobID)
	w.publishDetailedProgress(req.JobID, "completed", "Data refresh completed", "All data loaded successfully",
		5, 5, 100,
		getRecordCount(completions, "cos"),
		getRecordCount(completions, "mos"),
		getRecordCount(completions, "mops"),
		0, 0, 0, 0)
	w.publishComplete(req.JobID)

	log.Printf("Refresh job %s completed successfully", req.JobID)
	return nil
}

// handlePhaseJob processes a single phase (MOPs, MOs, or COs)
func (w *SnapshotWorker) handlePhaseJob(msg *nats.Msg) {
	var job PhaseJobMessage
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		log.Printf("Failed to parse phase job: %v", err)
		return
	}

	log.Printf("Processing phase %s for job %s", job.Phase, job.ParentJobID)

	ctx := context.Background()

	// Get environment config
	envConfig, err := w.config.GetEnvironmentConfig(job.Environment)
	if err != nil {
		log.Printf("Failed to get environment config: %v", err)
		w.publishPhaseCompletion(job, 0, err)
		return
	}

	// Create Compass client
	getToken := func() (string, error) {
		return job.AccessToken, nil
	}
	compassClient := compass.NewClient(envConfig.CompassBaseURL, getToken)
	snapshotService := services.NewSnapshotService(compassClient, w.db)

	var recordCount int
	var fetchErr error

	switch job.Phase {
	case "mops":
		refs, err := snapshotService.RefreshPlannedOrders(ctx, job.Company, job.Facility)
		recordCount = len(refs)
		fetchErr = err

	case "mos":
		refs, err := snapshotService.RefreshManufacturingOrders(ctx, job.Company, job.Facility)
		recordCount = len(refs)
		fetchErr = err

	case "cos":
		count, err := snapshotService.RefreshOpenCustomerOrderLines(ctx, job.Company, job.Facility)
		recordCount = count
		fetchErr = err

	default:
		log.Printf("Unknown phase: %s", job.Phase)
		return
	}

	// Publish completion
	w.publishPhaseCompletion(job, recordCount, fetchErr)
}

// publishPhaseCompletion publishes a phase completion message
func (w *SnapshotWorker) publishPhaseCompletion(job PhaseJobMessage, recordCount int, err error) {
	completion := PhaseCompletionMessage{
		JobID:       job.JobID,
		ParentJobID: job.ParentJobID,
		Phase:       job.Phase,
		RecordCount: recordCount,
		Success:     err == nil,
	}
	if err != nil {
		completion.Error = err.Error()
		log.Printf("Phase %s failed: %v", job.Phase, err)
	} else {
		log.Printf("Phase %s completed: %d records", job.Phase, recordCount)
	}

	data, _ := json.Marshal(completion)
	completeSubject := queue.GetPhaseCompleteSubject(job.ParentJobID)
	if err := w.nats.Publish(completeSubject, data); err != nil {
		log.Printf("Failed to publish phase completion: %v", err)
	}
}

// publishPhaseProgress publishes progress update with parallel phase status
func (w *SnapshotWorker) publishPhaseProgress(jobID string, completions map[string]*PhaseCompletionMessage) {
	// Convert completions map to PhaseProgress array
	phases := make([]PhaseProgress, 0, 3)

	for _, phaseName := range []string{"mops", "mos", "cos"} {
		if completion, exists := completions[phaseName]; exists {
			phases = append(phases, PhaseProgress{
				Phase:       phaseName,
				Status:      "completed",
				RecordCount: completion.RecordCount,
			})
		} else {
			phases = append(phases, PhaseProgress{
				Phase:  phaseName,
				Status: "running",
			})
		}
	}

	// Calculate overall progress (phase 0: 20%, phases 1-3: 60%, phases 4-5: 20%)
	completedPhases := len(completions)
	phaseProgress := 20 + (completedPhases * 20) // 20, 40, 60
	if completedPhases == 3 {
		phaseProgress = 60 // Cap at 60% until finalize runs
	}

	update := ProgressUpdate{
		JobID:            jobID,
		Status:           "running",
		Progress:         phaseProgress,
		CurrentStep:      fmt.Sprintf("Loading data (%d/3 phases complete)", completedPhases),
		ParallelPhases:   phases,
		COLinesProcessed: getRecordCount(completions, "cos"),
		MOsProcessed:     getRecordCount(completions, "mos"),
		MOPsProcessed:    getRecordCount(completions, "mops"),
	}

	// Debug: Log phases array
	log.Printf("DEBUG publishPhaseProgress: jobID=%s, phases count=%d, phases=%+v", jobID, len(phases), phases)

	// Publish to NATS for SSE streaming
	data, _ := json.Marshal(update)
	log.Printf("DEBUG publishPhaseProgress: JSON length=%d, JSON=%s", len(data), string(data))
	if err := w.nats.Publish(queue.GetProgressSubject(jobID), data); err != nil {
		log.Printf("Failed to publish phase progress: %v", err)
	}

	// Update database
	ctx := context.Background()
	w.db.UpdateJobRecordCounts(ctx, jobID,
		getRecordCount(completions, "cos"),
		getRecordCount(completions, "mos"),
		getRecordCount(completions, "mops"),
	)
}

// getRecordCount helper to extract record count from completions map
func getRecordCount(completions map[string]*PhaseCompletionMessage, phase string) int {
	if c, exists := completions[phase]; exists {
		return c.RecordCount
	}
	return 0
}
