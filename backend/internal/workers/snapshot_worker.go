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
	nats           *queue.Manager
	db             *db.Queries
	config         *config.Config
	jobContexts    map[string]context.CancelFunc // Track job cancellation contexts
	jobContextsMux sync.RWMutex                  // Protect concurrent access
}

// NewSnapshotWorker creates a new snapshot worker
func NewSnapshotWorker(nats *queue.Manager, database *db.Queries, cfg *config.Config) *SnapshotWorker {
	return &SnapshotWorker{
		nats:        nats,
		db:          database,
		config:      cfg,
		jobContexts: make(map[string]context.CancelFunc),
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
	Language    string `json:"language"`
}

// PhaseProgress represents the status of a single parallel phase
type PhaseProgress struct {
	Phase            string    `json:"phase"`                      // "mops", "mos", "cos"
	Status           string    `json:"status"`                     // "pending", "running", "completed", "failed"
	CurrentOperation string    `json:"currentOperation,omitempty"` // "Querying...", "Processing...", "Inserting..."
	RecordCount      int       `json:"recordCount"`                // Records processed
	StartTime        time.Time `json:"startTime,omitempty"`
	EndTime          time.Time `json:"endTime,omitempty"`
	Error            string    `json:"error,omitempty"`
}

// DetectorProgress represents the status of a single parallel detector
type DetectorProgress struct {
	DetectorName string    `json:"detectorName"`       // "unlinked_production_orders"
	DisplayLabel string    `json:"displayLabel"`       // "Unlinked Production Orders"
	Status       string    `json:"status"`             // "pending", "running", "completed", "failed"
	IssuesFound  int       `json:"issuesFound"`        // Issues detected
	DurationMs   int64     `json:"durationMs"`         // Execution time
	StartTime    time.Time `json:"startTime,omitempty"`
	EndTime      time.Time `json:"endTime,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	JobID                     string             `json:"jobId"`
	Status                    string             `json:"status"`
	Progress                  int                `json:"progress"`
	CurrentStep               string             `json:"currentStep"`
	CompletedSteps            int                `json:"completedSteps"`
	TotalSteps                int                `json:"totalSteps"`
	ParallelPhases            []PhaseProgress    `json:"parallelPhases,omitempty"`    // Parallel data loading tracking
	ParallelDetectors         []DetectorProgress `json:"parallelDetectors,omitempty"` // Parallel detector tracking
	COLinesProcessed          int                `json:"coLinesProcessed,omitempty"`
	MOsProcessed              int                `json:"mosProcessed,omitempty"`
	MOPsProcessed             int                `json:"mopsProcessed,omitempty"`
	RecordsPerSecond          float64            `json:"recordsPerSecond,omitempty"`
	EstimatedSecondsRemaining int                `json:"estimatedTimeRemaining,omitempty"`
	CurrentOperation          string             `json:"currentOperation,omitempty"`
	CurrentBatch              int                `json:"currentBatch,omitempty"`
	TotalBatches              int                `json:"totalBatches,omitempty"`
	Error                     string             `json:"error,omitempty"`
}


// DataBatchJobMessage represents work for loading one data type (MOPs, MOs, or COs)
type DataBatchJobMessage struct {
	JobID       string `json:"jobId"`
	ParentJobID string `json:"parentJobId"`
	DataType    string `json:"dataType"`    // "mops", "mos", "cos"
	Environment string `json:"environment"`
	AccessToken string `json:"accessToken"`
	Company     string `json:"company"`
	Facility    string `json:"facility"`
	Language    string `json:"language"`
}

// BatchStartMessage signals that a worker has picked up a batch job
type BatchStartMessage struct {
	JobID       string    `json:"jobId"`
	ParentJobID string    `json:"parentJobId"`
	DataType    string    `json:"dataType"` // "mops", "mos", "cos"
	StartTime   time.Time `json:"startTime"`
}

// BatchCompletionMessage signals data type loading completion
type BatchCompletionMessage struct {
	JobID       string `json:"jobId"`
	ParentJobID string `json:"parentJobId"`
	DataType    string `json:"dataType"` // "mops", "mos", "cos"
	RecordCount int    `json:"recordCount"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// PhaseSubProgressMessage signals intermediate progress within a data type loading
type PhaseSubProgressMessage struct {
	JobID            string `json:"jobId"`            // "abc123-mops"
	ParentJobID      string `json:"parentJobId"`      // "abc123"
	DataType         string `json:"dataType"`         // "mops", "mos", "cos"
	CurrentOperation string `json:"currentOperation"` // "Querying...", "Processing...", "Inserting..."
	RecordCount      int    `json:"recordCount"`      // Running count if available
}

// DetectorJobMessage represents work for running one detector
type DetectorJobMessage struct {
	JobID        string `json:"jobId"`        // "abc123-unlinked"
	ParentJobID  string `json:"parentJobId"`  // "abc123"
	DetectorName string `json:"detectorName"` // "unlinked_production_orders"
	DisplayLabel string `json:"displayLabel"` // "Unlinked Production Orders"
	Environment  string `json:"environment"`  // "TRN" or "PRD"
	Company      string `json:"company"`      // "100"
	Facility     string `json:"facility"`     // "AZ1"
}

// DetectorStartMessage signals that a worker has picked up a detector job
type DetectorStartMessage struct {
	JobID        string    `json:"jobId"`
	ParentJobID  string    `json:"parentJobId"`
	DetectorName string    `json:"detectorName"`
	DisplayLabel string    `json:"displayLabel"`
	StartTime    time.Time `json:"startTime"`
}

// DetectorCompletionMessage signals detector execution completion
type DetectorCompletionMessage struct {
	JobID        string `json:"jobId"`
	ParentJobID  string `json:"parentJobId"`
	DetectorName string `json:"detectorName"`
	IssuesFound  int    `json:"issuesFound"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"durationMs"`
}

// DetectorCoordinatorMessage triggers coordination of a manual detection job
type DetectorCoordinatorMessage struct {
	JobID          string   `json:"jobId"`          // Detection job ID (e.g., "det-123456789")
	Environment    string   `json:"environment"`    // "TRN" or "PRD"
	DetectorNames  []string `json:"detectorNames"`  // List of detectors being run
	TotalDetectors int      `json:"totalDetectors"` // Total count for progress tracking
	Company        string   `json:"company"`        // Company code
	Facility       string   `json:"facility"`       // Facility code
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

	// Subscribe to each data type individually for parallel distribution
	// IMPORTANT: Wildcard subscriptions (snapshot.batch.TRN.>) create a single FIFO queue,
	// causing sequential processing. Individual subscriptions with the same queue group
	// enable NATS to distribute messages in parallel across workers.
	dataTypes := []string{"mops", "mos", "cos"}
	environments := []string{"TRN", "PRD"}

	for _, env := range environments {
		for _, dataType := range dataTypes {
			subject := queue.GetBatchSubject(env, dataType)
			_, err := w.nats.QueueSubscribe(
				subject,
				queue.QueueGroupBatchWorkers,
				w.handleBatchJob,
			)
			if err != nil {
				return fmt.Errorf("failed to subscribe to %s %s batch jobs: %w", env, dataType, err)
			}
		}
		log.Printf("Subscribed to %d %s data batch queues for parallel processing", len(dataTypes), env)
	}

	// Subscribe to each detector individually for parallel distribution
	// IMPORTANT: This list must match detector registration in detection_service.go
	// When adding new detectors, update this list to enable parallel execution.
	detectorNames := []string{
		"unlinked_production_orders",
		"joint_delivery_date_mismatch",
		"dlix_date_mismatch",
		// DISABLED: "co_quantity_mismatch" - requires PAQT from MPTAWY table which has severe performance issues
	}

	for _, env := range environments {
		for _, detectorName := range detectorNames {
			subject := queue.GetDetectorSubject(env, detectorName)
			_, err := w.nats.QueueSubscribe(
				subject,
				queue.QueueGroupBatchWorkers,
				w.handleDetectorJob,
			)
			if err != nil {
				return fmt.Errorf("failed to subscribe to %s %s detector: %w", env, detectorName, err)
			}
		}
		log.Printf("Subscribed to %d %s detector queues for parallel processing", len(detectorNames), env)
	}

	// Subscribe to manual detection coordinator jobs
	for _, env := range environments {
		coordSubject := queue.GetDetectorCoordinateSubject(env)
		_, err := w.nats.QueueSubscribe(
			coordSubject,
			"detector-coordinators",
			w.handleManualDetectionCoordinator,
		)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s detection coordinator: %w", env, err)
		}
	}
	log.Println("Subscribed to manual detection coordinator queues")

	// Subscribe to cancellation requests (all workers should listen)
	_, err = w.nats.Subscribe("snapshot.cancel.*", w.handleCancelRequest)
	if err != nil {
		return fmt.Errorf("failed to subscribe to cancellation requests: %w", err)
	}

	// IMPORTANT: Limit concurrent message processing by handling synchronously
	// Each worker will process one message at a time, allowing NATS to distribute
	// the remaining messages to other available workers

	log.Println("Snapshot worker started and listening for jobs, phase work, batch work, and cancellation requests")
	return nil
}

// createJobContext creates and stores a cancellable context for a job
func (w *SnapshotWorker) createJobContext(jobID string) context.Context {
	w.jobContextsMux.Lock()
	defer w.jobContextsMux.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.jobContexts[jobID] = cancel
	log.Printf("Created cancellable context for job: %s", jobID)
	return ctx
}

// cancelJobContext cancels the context for a job
func (w *SnapshotWorker) cancelJobContext(jobID string) {
	w.jobContextsMux.Lock()
	defer w.jobContextsMux.Unlock()

	if cancel, exists := w.jobContexts[jobID]; exists {
		cancel()
		delete(w.jobContexts, jobID)
		log.Printf("Cancelled context for job: %s", jobID)
	}
}

// getJobContext retrieves the context for a job (returns Background if not found)
func (w *SnapshotWorker) getJobContext(jobID string) context.Context {
	w.jobContextsMux.RLock()
	defer w.jobContextsMux.RUnlock()

	// Note: In distributed environment, contexts are not shared across workers
	// Each worker maintains its own context map
	// For batch/detector jobs, we check DB status instead
	return context.Background()
}

// isJobCancelled checks if a job has been cancelled in the database
func (w *SnapshotWorker) isJobCancelled(jobID string) bool {
	ctx := context.Background()
	job, err := w.db.GetRefreshJob(ctx, jobID)
	if err != nil {
		return false
	}
	return job.Status == "cancelled"
}

// handleCancelRequest handles a cancellation request for a job
func (w *SnapshotWorker) handleCancelRequest(msg *nats.Msg) {
	// Extract jobID from subject (format: snapshot.cancel.{jobID})
	parts := len("snapshot.cancel.")
	if len(msg.Subject) <= parts {
		log.Printf("Invalid cancel subject: %s", msg.Subject)
		return
	}
	jobID := msg.Subject[parts:]

	log.Printf("Received cancellation request for job: %s", jobID)
	w.cancelJobContext(jobID)
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

// processRefresh coordinates parallel data refresh using NATS batch distribution with ID range partitioning
// Optimized for Apache Spark: Uses predicate pushdown (WHERE ID >= X AND ID < Y) instead of OFFSET/LIMIT
func (w *SnapshotWorker) processRefresh(req SnapshotRefreshMessage) error {
	// Create cancellable context for this job
	ctx := w.createJobContext(req.JobID)
	defer w.cancelJobContext(req.JobID) // Clean up context when done

	log.Printf("Coordinating refresh job %s with parallel ID range batching", req.JobID)

	// Start the job
	if err := w.db.StartJob(ctx, req.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Check for cancellation before starting
	if ctx.Err() != nil {
		log.Printf("Job %s cancelled before starting", req.JobID)
		return fmt.Errorf("job cancelled: %w", ctx.Err())
	}

	// Phase 0: Truncate database (must complete first)
	log.Printf("Phase 0: Truncating database for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Preparing database", "Truncating tables",
		0, 4, 0, 0, 0, 0, nil, nil, 0, 0, 0, 0)

	if err := w.db.TruncateAnalysisTables(ctx, req.Environment); err != nil {
		// Check if error is due to cancellation
		if ctx.Err() != nil {
			log.Printf("Job %s cancelled during truncate", req.JobID)
			return fmt.Errorf("job cancelled: %w", ctx.Err())
		}
		w.publishError(req.JobID, fmt.Sprintf("Truncate failed: %v", err))
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf("truncate failed: %w", err)
	}

	// Check for cancellation after truncate
	if ctx.Err() != nil {
		log.Printf("Job %s cancelled after truncate", req.JobID)
		return fmt.Errorf("job cancelled: %w", ctx.Err())
	}

	log.Printf("Phase 0 complete: Database truncated")
	w.publishDetailedProgress(req.JobID, "running", "Database prepared", "Publishing data jobs",
		1, 4, 20, 0, 0, 0, nil, nil, 0, 0, 0, 0)

	// Phase 1: Publish 3 data jobs to NATS (one per data type) and wait for completion
	log.Printf("Phase 1: Publishing 3 data jobs (MOPs, MOs, COs) to NATS...")
	return w.publishDataJobs(req)
}

// publishDetailedProgress publishes a detailed progress update with extended metrics
func (w *SnapshotWorker) publishDetailedProgress(jobID, status, currentStep, currentOperation string, completedSteps, totalSteps, progressPct, coLines, mos, mops int, parallelPhases []PhaseProgress, parallelDetectors []DetectorProgress, recordsPerSec float64, estimatedSecsRemaining, currentBatch, totalBatches int) {
	update := ProgressUpdate{
		JobID:                     jobID,
		Status:                    status,
		Progress:                  progressPct,
		CurrentStep:               currentStep,
		CompletedSteps:            completedSteps,
		TotalSteps:                totalSteps,
		ParallelPhases:            parallelPhases,
		ParallelDetectors:         parallelDetectors,
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

// publishPhaseSubProgress publishes intermediate progress for a data type
func (w *SnapshotWorker) publishPhaseSubProgress(parentJobID, dataType, operation string, recordCount int) {
	msg := PhaseSubProgressMessage{
		JobID:            fmt.Sprintf("%s-%s", parentJobID, dataType),
		ParentJobID:      parentJobID,
		DataType:         dataType,
		CurrentOperation: operation,
		RecordCount:      recordCount,
	}

	data, _ := json.Marshal(msg)
	subject := queue.GetPhaseProgressSubject(parentJobID)
	if err := w.nats.Publish(subject, data); err != nil {
		log.Printf("Failed to publish phase sub-progress: %v", err)
		// Non-fatal, continue
	}
}

// ========================================
// Batch Processing Handlers
// ========================================

// handleBatchJob processes a single data batch with ID range filtering
// This is the worker that executes parallel batches distributed via NATS
func (w *SnapshotWorker) handleBatchJob(msg *nats.Msg) {
	var job DataBatchJobMessage
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		log.Printf("Failed to parse batch job: %v", err)
		return
	}

	log.Printf("Processing %s data for job %s", job.DataType, job.JobID)

	// Create context with timeout for Compass SQL queries
	// 30 minutes should be sufficient for even large datasets
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Check if parent job has been cancelled
	if w.isJobCancelled(job.ParentJobID) {
		log.Printf("Parent job %s was cancelled, skipping %s batch", job.ParentJobID, job.DataType)
		return
	}

	// Publish batch start notification
	startMsg := BatchStartMessage{
		JobID:       job.JobID,
		ParentJobID: job.ParentJobID,
		DataType:    job.DataType,
		StartTime:   time.Now(),
	}
	startData, _ := json.Marshal(startMsg)
	startSubject := queue.GetBatchStartSubject(job.ParentJobID)
	if err := w.nats.Publish(startSubject, startData); err != nil {
		log.Printf("Failed to publish batch start notification: %v", err)
		// Non-fatal, continue processing
	}

	// Get environment config
	envConfig, err := w.config.GetEnvironmentConfig(job.Environment)
	if err != nil {
		log.Printf("Failed to get environment config: %v", err)
		w.publishBatchCompletion(job, 0, err)
		return
	}

	// Create Compass client
	getToken := func() (string, error) {
		return job.AccessToken, nil
	}
	compassClient := compass.NewClient(envConfig.CompassBaseURL, getToken)
	snapshotService := services.NewSnapshotService(compassClient, w.db)

	// Set progress callback to publish intermediate updates
	snapshotService.SetProgressCallback(func(phase string, stepNum, totalSteps int, message string, mopCount, moCount, coCount, currentRecordCount int) {
		// Use the currentRecordCount parameter which contains the real-time progress
		w.publishPhaseSubProgress(job.ParentJobID, job.DataType, message, currentRecordCount)
	})

	var recordCount int
	var fetchErr error

	// Execute full query based on data type (no ID range filtering)
	switch job.DataType {
	case "mops":
		recordCount, fetchErr = snapshotService.RefreshPlannedOrders(ctx, job.Environment, job.Company, job.Facility)

	case "mos":
		recordCount, fetchErr = snapshotService.RefreshManufacturingOrders(ctx, job.Environment, job.Company, job.Facility)

	case "cos":
		recordCount, fetchErr = snapshotService.RefreshOpenCustomerOrderLines(ctx, job.Environment, job.Company, job.Facility, job.Language)

	default:
		log.Printf("Unknown data type: %s", job.DataType)
		w.publishBatchCompletion(job, 0, fmt.Errorf("unknown data type: %s", job.DataType))
		return
	}

	log.Printf("Completed %s data: %d records", job.DataType, recordCount)

	// Publish completion
	w.publishBatchCompletion(job, recordCount, fetchErr)
}

// publishBatchCompletion publishes a data type loading completion message
func (w *SnapshotWorker) publishBatchCompletion(job DataBatchJobMessage, recordCount int, err error) {
	completion := BatchCompletionMessage{
		JobID:       job.JobID,
		ParentJobID: job.ParentJobID,
		DataType:    job.DataType,
		RecordCount: recordCount,
		Success:     err == nil,
	}
	if err != nil {
		completion.Error = err.Error()
		log.Printf("Loading %s failed: %v", job.DataType, err)
	} else {
		log.Printf("Loaded %s: %d records", job.DataType, recordCount)
	}

	data, _ := json.Marshal(completion)
	completeSubject := queue.GetBatchCompleteSubject(job.ParentJobID)
	if err := w.nats.Publish(completeSubject, data); err != nil {
		log.Printf("Failed to publish completion: %v", err)
	}
}

// handleDetectorJob processes a single detector execution
// This is the worker that executes individual detectors distributed via NATS
func (w *SnapshotWorker) handleDetectorJob(msg *nats.Msg) {
	var job DetectorJobMessage
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		log.Printf("Failed to parse detector job: %v", err)
		return
	}

	startTime := time.Now()
	log.Printf("Processing detector '%s' for job %s (environment: %s)",
		job.DetectorName, job.ParentJobID, job.Environment)

	ctx := context.Background()

	// Check if parent job has been cancelled
	if w.isJobCancelled(job.ParentJobID) {
		log.Printf("Parent job %s was cancelled, skipping detector %s", job.ParentJobID, job.DetectorName)
		return
	}

	// Publish detector start notification
	startMsg := DetectorStartMessage{
		JobID:        job.JobID,
		ParentJobID:  job.ParentJobID,
		DetectorName: job.DetectorName,
		DisplayLabel: job.DisplayLabel,
		StartTime:    startTime,
	}
	startData, _ := json.Marshal(startMsg)
	startSubject := queue.GetDetectorStartSubject(job.ParentJobID)
	if err := w.nats.Publish(startSubject, startData); err != nil {
		log.Printf("Failed to publish detector start notification: %v", err)
		// Non-fatal, continue processing
	}

	// Initialize detector services
	detectorConfigService := services.NewDetectorConfigService(w.db)
	detectionService := services.NewDetectionService(w.db, detectorConfigService)

	// Get detector by name
	detector := detectionService.GetDetectorByName(job.DetectorName)
	if detector == nil {
		errMsg := fmt.Sprintf("detector not found: %s", job.DetectorName)
		log.Printf("ERROR: %s", errMsg)
		w.publishDetectorCompletion(job, 0, fmt.Errorf(errMsg), startTime)
		return
	}

	// Check if enabled
	enabled, err := detectionService.IsDetectorEnabled(ctx, job.Environment, job.DetectorName)
	if err != nil {
		log.Printf("Failed to check detector enabled status: %v", err)
		w.publishDetectorCompletion(job, 0, err, startTime)
		return
	}

	if !enabled {
		log.Printf("Detector '%s' is disabled for environment %s, skipping",
			job.DetectorName, job.Environment)
		// Publish success with 0 issues (skipped but not failed)
		w.publishDetectorCompletion(job, 0, nil, startTime)
		return
	}

	// Execute detector
	issuesFound, err := detector.Detect(ctx, w.db, job.ParentJobID, job.Environment, job.Company, job.Facility)

	if err != nil {
		log.Printf("Detector '%s' failed: %v", job.DetectorName, err)
	} else {
		log.Printf("Detector '%s' completed: %d issues found (%dms)",
			job.DetectorName, issuesFound, time.Since(startTime).Milliseconds())
	}

	// Publish completion
	w.publishDetectorCompletion(job, issuesFound, err, startTime)
}

// publishDetectorCompletion publishes a detector execution completion message
func (w *SnapshotWorker) publishDetectorCompletion(job DetectorJobMessage, issuesFound int, err error, startTime time.Time) {
	completion := DetectorCompletionMessage{
		JobID:        job.JobID,
		ParentJobID:  job.ParentJobID,
		DetectorName: job.DetectorName,
		IssuesFound:  issuesFound,
		Success:      err == nil,
		DurationMs:   time.Since(startTime).Milliseconds(),
	}
	if err != nil {
		completion.Error = err.Error()
		log.Printf("Detector %s failed: %v", job.DetectorName, err)
	} else {
		log.Printf("Detector %s completed: %d issues", job.DetectorName, issuesFound)
	}

	data, _ := json.Marshal(completion)
	completeSubject := queue.GetDetectorCompleteSubject(job.ParentJobID)
	if err := w.nats.Publish(completeSubject, data); err != nil {
		log.Printf("Failed to publish detector completion: %v", err)
	}
}

// publishDataJobs publishes 3 data jobs (MOPs, MOs, COs) to NATS and waits for completion
func (w *SnapshotWorker) publishDataJobs(req SnapshotRefreshMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	log.Printf("Phase 1: Publishing 3 data jobs to NATS queue...")

	// Initialize parallel phases as pending
	initialPhases := []PhaseProgress{
		{Phase: "mops", Status: "pending"},
		{Phase: "mos", Status: "pending"},
		{Phase: "cos", Status: "pending"},
	}

	// Create phase records in database for tracking
	dataTypes := []string{"mops", "mos", "cos"}
	for _, phaseType := range dataTypes {
		if err := w.db.CreateRefreshJobPhase(ctx, req.JobID, phaseType); err != nil {
			log.Printf("Warning: failed to create phase record for %s: %v", phaseType, err)
			// Non-fatal, continue
		}
	}

	w.publishDetailedProgress(req.JobID, "running", "Publishing data jobs", "Distributing work to workers",
		1, 4, 25, 0, 0, 0, initialPhases, nil, 0, 0, 0, 3)

	// Publish 3 jobs (one per data type)

	for _, dataType := range dataTypes {
		job := DataBatchJobMessage{
			JobID:       fmt.Sprintf("%s-%s", req.JobID, dataType),
			ParentJobID: req.JobID,
			DataType:    dataType,
			Environment: req.Environment,
			AccessToken: req.AccessToken,
			Company:     req.Company,
			Facility:    req.Facility,
			Language:    req.Language,
		}

		data, _ := json.Marshal(job)
		subject := queue.GetBatchSubject(req.Environment, dataType)
		if err := w.nats.Publish(subject, data); err != nil {
			errMsg := fmt.Sprintf("Failed to publish %s job: %v", dataType, err)
			w.publishError(req.JobID, errMsg)
			w.db.FailJob(ctx, req.JobID, err.Error())
			return fmt.Errorf(errMsg)
		}
	}

	log.Printf("Published 3 data jobs, waiting for completion...")

	// Phase 2: Wait for all 3 jobs to complete
	return w.waitForDataJobs(req)
}

// waitForDataJobs waits for exactly 3 data jobs to complete, then runs finalize and detection
func (w *SnapshotWorker) waitForDataJobs(req SnapshotRefreshMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	log.Printf("Phase 2: Waiting for 3 data jobs to complete...")

	// Track completions
	completedJobs := 0
	recordsByType := map[string]int{
		"mops": 0,
		"mos":  0,
		"cos":  0,
	}

	// Track parallel phase states
	phaseStates := map[string]*PhaseProgress{
		"mops": {Phase: "mops", Status: "pending"},
		"mos":  {Phase: "mos", Status: "pending"},
		"cos":  {Phase: "cos", Status: "pending"},
	}
	var mu sync.Mutex

	// Subscribe to batch start events
	startSubject := queue.GetBatchStartSubject(req.JobID)
	startSub, err := w.nats.Subscribe(startSubject, func(msg *nats.Msg) {
		var start BatchStartMessage
		if err := json.Unmarshal(msg.Data, &start); err != nil {
			log.Printf("Failed to parse batch start message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Update phase state to running
		phaseStates[start.DataType].Status = "running"
		phaseStates[start.DataType].StartTime = start.StartTime

		// Persist phase start to database
		dbCtx := context.Background()
		if err := w.db.StartRefreshJobPhase(dbCtx, req.JobID, start.DataType); err != nil {
			log.Printf("Warning: failed to persist phase start for %s: %v", start.DataType, err)
			// Non-fatal, continue
		}

		log.Printf("Data job started: %s", start.DataType)

		// Convert phase states to slice for JSON
		parallelPhases := make([]PhaseProgress, 0, 3)
		for _, phase := range []string{"mops", "mos", "cos"} {
			parallelPhases = append(parallelPhases, *phaseStates[phase])
		}

		// Send progress update showing running status
		totalMops := recordsByType["mops"]
		totalMos := recordsByType["mos"]
		totalCos := recordsByType["cos"]
		progress := 25 + (completedJobs * 15)

		w.publishDetailedProgress(req.JobID, "running", "Loading data",
			fmt.Sprintf("Loading %s", start.DataType),
			2, 4, progress,
			totalCos, totalMos, totalMops,
			parallelPhases,
			nil,
			0, 0, completedJobs, 3)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to batch starts: %w", err)
	}
	defer startSub.Unsubscribe()

	// Subscribe to phase sub-progress events (intermediate updates)
	subProgressSubject := queue.GetPhaseProgressSubject(req.JobID)
	subProgressSub, err := w.nats.Subscribe(subProgressSubject, func(msg *nats.Msg) {
		var subProgress PhaseSubProgressMessage
		if err := json.Unmarshal(msg.Data, &subProgress); err != nil {
			log.Printf("Failed to parse phase sub-progress: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Ignore updates after phase completed (race condition protection)
		if phaseStates[subProgress.DataType].Status == "completed" ||
			phaseStates[subProgress.DataType].Status == "failed" {
			return
		}

		// Update phase state with current operation
		phaseStates[subProgress.DataType].CurrentOperation = subProgress.CurrentOperation
		if subProgress.RecordCount > 0 {
			phaseStates[subProgress.DataType].RecordCount = subProgress.RecordCount
		}

		log.Printf("Phase %s: %s", subProgress.DataType, subProgress.CurrentOperation)

		// Aggregate and publish progress update
		parallelPhases := make([]PhaseProgress, 0, 3)
		for _, phase := range []string{"mops", "mos", "cos"} {
			parallelPhases = append(parallelPhases, *phaseStates[phase])
		}

		progress := 25 + (completedJobs * 15)
		w.publishDetailedProgress(req.JobID, "running", "Loading data",
			subProgress.CurrentOperation, 2, 4, progress,
			recordsByType["cos"], recordsByType["mos"], recordsByType["mops"],
			parallelPhases, nil, 0, 0, completedJobs, 3)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to phase sub-progress: %w", err)
	}
	defer subProgressSub.Unsubscribe()

	// Subscribe to completion events
	completeSubject := queue.GetBatchCompleteSubject(req.JobID)
	subscription, err := w.nats.Subscribe(completeSubject, func(msg *nats.Msg) {
		var completion BatchCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to parse completion message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if !completion.Success {
			errMsg := fmt.Sprintf("Data job %s failed: %s", completion.DataType, completion.Error)
			log.Printf(errMsg)

			// Update phase state to failed
			phaseStates[completion.DataType].Status = "failed"
			phaseStates[completion.DataType].Error = completion.Error
			phaseStates[completion.DataType].EndTime = time.Now()

			// Persist phase failure to database
			dbCtx := context.Background()
			if err := w.db.FailRefreshJobPhase(dbCtx, req.JobID, completion.DataType, completion.Error); err != nil {
				log.Printf("Warning: failed to persist phase failure for %s: %v", completion.DataType, err)
			}

			w.publishError(req.JobID, errMsg)
			w.db.FailJob(ctx, req.JobID, completion.Error)
			cancel() // Cancel context to abort
			return
		}

		// Update phase state to completed
		phaseStates[completion.DataType].Status = "completed"
		phaseStates[completion.DataType].CurrentOperation = "" // Clear transient operation
		phaseStates[completion.DataType].RecordCount = completion.RecordCount
		phaseStates[completion.DataType].EndTime = time.Now()

		// Persist phase completion to database
		dbCtx := context.Background()
		if err := w.db.CompleteRefreshJobPhase(dbCtx, req.JobID, completion.DataType, completion.RecordCount); err != nil {
			log.Printf("Warning: failed to persist phase completion for %s: %v", completion.DataType, err)
			// Non-fatal, continue
		}

		completedJobs++
		recordsByType[completion.DataType] = completion.RecordCount

		totalMops := recordsByType["mops"]
		totalMos := recordsByType["mos"]
		totalCos := recordsByType["cos"]

		// Calculate progress
		progress := 25 + (completedJobs * 15) // 25% base + 15% per job (up to 70%)

		log.Printf("Data job completed: %s (%d records), total: %d/3 jobs",
			completion.DataType, completion.RecordCount, completedJobs)

		// Convert phase states to slice for JSON
		parallelPhases := make([]PhaseProgress, 0, 3)
		for _, phase := range []string{"mops", "mos", "cos"} {
			parallelPhases = append(parallelPhases, *phaseStates[phase])
		}

		w.publishDetailedProgress(req.JobID, "running", "Loading data",
			fmt.Sprintf("Loaded %s", completion.DataType),
			2, 4, progress,
			totalCos, totalMos, totalMops,
			parallelPhases,
			nil,
			0, 0, completedJobs, 3)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to completions: %w", err)
	}
	defer subscription.Unsubscribe()

	// Wait for all 3 jobs to complete
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			completed := completedJobs
			mu.Unlock()
			if completed < 3 {
				return fmt.Errorf("timeout waiting for data jobs (completed %d/3)", completed)
			}
		case <-ticker.C:
			mu.Lock()
			completed := completedJobs
			mu.Unlock()

			if completed >= 3 {
				log.Printf("All 3 data jobs completed")
				mu.Lock()
				totalMops := recordsByType["mops"]
				totalMos := recordsByType["mos"]
				totalCos := recordsByType["cos"]
				mu.Unlock()

				log.Printf("Total records - MOPs: %d, MOs: %d, COs: %d", totalMops, totalMos, totalCos)

				// Phase 3: Finalize and detection
				return w.runFinalize(req, totalCos, totalMos, totalMops)
			}
		}
	}
}

// runFinalize runs finalize and detection phases
func (w *SnapshotWorker) runFinalize(req SnapshotRefreshMessage, totalCos, totalMos, totalMops int) error {
	ctx := context.Background()

	// Phase 3: Finalize
	log.Printf("Phase 3: Running finalize for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Finalizing data", "Updating production orders view",
		3, 4, 75,
		totalCos, totalMos, totalMops,
		nil,
		nil,
		0, 0, 0, 0)

	if err := w.db.UpdateProductionOrdersFromMOPs(ctx); err != nil {
		errMsg := fmt.Sprintf("Finalize MOPs failed: %v", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf(errMsg)
	}

	if err := w.db.UpdateProductionOrdersFromMOs(ctx); err != nil {
		errMsg := fmt.Sprintf("Finalize MOs failed: %v", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf(errMsg)
	}

	// Phase 4: Parallel Detection via NATS
	log.Printf("Phase 4: Publishing detector jobs for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Starting issue detection", "Publishing detector jobs",
		3, 4, 85,
		totalCos, totalMos, totalMops,
		nil,
		nil,
		0, 0, 0, 0)

	return w.publishDetectorJobs(req, totalCos, totalMos, totalMops)
}

// publishDetectorJobs publishes detector jobs to NATS and waits for completion
func (w *SnapshotWorker) publishDetectorJobs(req SnapshotRefreshMessage, totalCos, totalMos, totalMops int) error {
	ctx := context.Background()

	log.Printf("Publishing detector jobs to NATS queue...")

	// Initialize detector services to get detector list
	detectorConfigService := services.NewDetectorConfigService(w.db)
	detectionService := services.NewDetectionService(w.db, detectorConfigService)

	// Get all registered detector names
	allDetectorNames := detectionService.GetAllDetectorNames()

	// Filter to only enabled detectors
	enabledDetectors := make([]string, 0)
	for _, name := range allDetectorNames {
		isEnabled, err := detectionService.IsDetectorEnabled(ctx, req.Environment, name)
		if err != nil {
			log.Printf("Warning: failed to check enabled status for %s: %v", name, err)
			continue
		}
		if isEnabled {
			enabledDetectors = append(enabledDetectors, name)
		}
	}

	totalDetectors := len(enabledDetectors)

	if totalDetectors == 0 {
		log.Println("No detectors enabled, skipping detection phase")
		// Mark job complete immediately
		w.db.CompleteJob(ctx, req.JobID)
		w.publishDetailedProgress(req.JobID, "completed", "Data refresh completed",
			"All data loaded successfully (detection skipped)",
			4, 4, 100, totalCos, totalMos, totalMops, nil, nil, 0, 0, 0, 0)
		w.publishComplete(req.JobID)
		return nil
	}

	log.Printf("Found %d enabled detectors to run", totalDetectors)

	// Create detection job record
	if err := w.db.CreateIssueDetectionJob(ctx, req.JobID, req.Environment, totalDetectors); err != nil {
		return fmt.Errorf("failed to create detection job: %w", err)
	}

	// Clear previous issues for this job
	if err := w.db.ClearIssuesForJob(ctx, req.JobID); err != nil {
		log.Printf("Warning: failed to clear previous issues: %v", err)
	}

	// Clear previous anomalies for this job
	if err := w.db.ClearAnomaliesForJob(ctx, req.JobID); err != nil {
		log.Printf("Warning: failed to clear previous anomalies: %v", err)
	}

	// Create detector records in database for tracking
	for _, detectorName := range enabledDetectors {
		detector := detectionService.GetDetectorByName(detectorName)
		displayLabel := detectorName
		if detector != nil {
			displayLabel = detector.Label()
		}
		if err := w.db.CreateRefreshJobDetector(ctx, req.JobID, detectorName, displayLabel); err != nil {
			log.Printf("Warning: failed to create detector record for %s: %v", detectorName, err)
			// Non-fatal, continue
		}
	}

	// Publish detector jobs
	for _, detectorName := range enabledDetectors {
		// Get detector label for UI display
		detector := detectionService.GetDetectorByName(detectorName)
		displayLabel := detectorName // Fallback to name
		if detector != nil {
			displayLabel = detector.Label()
		}

		job := DetectorJobMessage{
			JobID:        fmt.Sprintf("%s-%s", req.JobID, detectorName),
			ParentJobID:  req.JobID,
			DetectorName: detectorName,
			DisplayLabel: displayLabel,
			Environment:  req.Environment,
			Company:      req.Company,
			Facility:     req.Facility,
		}

		data, _ := json.Marshal(job)
		subject := queue.GetDetectorSubject(req.Environment, detectorName)
		if err := w.nats.Publish(subject, data); err != nil {
			errMsg := fmt.Sprintf("Failed to publish %s detector job: %v", detectorName, err)
			w.publishError(req.JobID, errMsg)
			w.db.FailJob(ctx, req.JobID, err.Error())
			return fmt.Errorf(errMsg)
		}
		log.Printf("Published detector job: %s", detectorName)
	}

	log.Printf("Published %d detector jobs, waiting for completion...", totalDetectors)

	// Wait for all detector jobs to complete
	return w.waitForDetectorJobs(req, totalDetectors, totalCos, totalMos, totalMops)
}

// waitForDetectorJobs waits for all detector jobs to complete, then finalizes detection job
func (w *SnapshotWorker) waitForDetectorJobs(req SnapshotRefreshMessage, totalDetectors, totalCos, totalMos, totalMops int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	log.Printf("Waiting for %d detector jobs to complete...", totalDetectors)

	// Initialize detector services to get detector info
	detectorConfigService := services.NewDetectorConfigService(w.db)
	detectionService := services.NewDetectionService(w.db, detectorConfigService)

	// Get all enabled detector names
	allDetectorNames := detectionService.GetAllDetectorNames()
	enabledDetectors := make([]string, 0)
	detectorLabels := make(map[string]string) // Map name to display label

	for _, name := range allDetectorNames {
		isEnabled, err := detectionService.IsDetectorEnabled(ctx, req.Environment, name)
		if err != nil || !isEnabled {
			continue
		}
		enabledDetectors = append(enabledDetectors, name)

		// Get display label
		detector := detectionService.GetDetectorByName(name)
		if detector != nil {
			detectorLabels[name] = detector.Label()
		} else {
			detectorLabels[name] = name
		}
	}

	// Track detector states
	detectorStates := make(map[string]*DetectorProgress)
	for _, name := range enabledDetectors {
		detectorStates[name] = &DetectorProgress{
			DetectorName: name,
			DisplayLabel: detectorLabels[name],
			Status:       "pending",
		}
	}

	// Track completions
	completedDetectors := 0
	failedDetectors := 0
	issuesByDetector := make(map[string]int)
	var mu sync.Mutex

	// Subscribe to detector start events
	startSubject := queue.GetDetectorStartSubject(req.JobID)
	startSub, err := w.nats.Subscribe(startSubject, func(msg *nats.Msg) {
		var start DetectorStartMessage
		if err := json.Unmarshal(msg.Data, &start); err != nil {
			log.Printf("Failed to parse detector start message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Update detector state to running
		if state, exists := detectorStates[start.DetectorName]; exists {
			state.Status = "running"
			state.StartTime = start.StartTime
		}

		// Persist detector start to database
		dbCtx := context.Background()
		if err := w.db.StartRefreshJobDetector(dbCtx, req.JobID, start.DetectorName); err != nil {
			log.Printf("Warning: failed to persist detector start for %s: %v", start.DetectorName, err)
			// Non-fatal, continue
		}

		log.Printf("Detector started: %s (%s)", start.DetectorName, start.DisplayLabel)

		// Convert detector states to slice for JSON
		parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
		for _, name := range enabledDetectors {
			parallelDetectors = append(parallelDetectors, *detectorStates[name])
		}

		// Calculate progress
		progress := 85 + (15 * completedDetectors / totalDetectors)

		// Send progress update showing running status
		w.publishDetailedProgress(req.JobID, "running", "Running issue detection",
			fmt.Sprintf("Running %s detector", start.DisplayLabel),
			4, 4, progress,
			totalCos, totalMos, totalMops,
			nil,
			parallelDetectors,
			0, 0, 0, 0)
	})

	if err != nil {
		errMsg := fmt.Sprintf("failed to subscribe to detector starts: %w", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailJob(ctx, req.JobID, errMsg)
		return fmt.Errorf(errMsg)
	}
	defer startSub.Unsubscribe()

	// Subscribe to completion events
	completeSubject := queue.GetDetectorCompleteSubject(req.JobID)
	subscription, err := w.nats.Subscribe(completeSubject, func(msg *nats.Msg) {
		var completion DetectorCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to parse detector completion message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Update detector state
		if state, exists := detectorStates[completion.DetectorName]; exists {
			dbCtx := context.Background()

			if !completion.Success {
				state.Status = "failed"
				state.Error = completion.Error
				state.DurationMs = completion.DurationMs
				state.EndTime = time.Now()
				failedDetectors++

				errMsg := fmt.Sprintf("Detector %s failed: %s", completion.DetectorName, completion.Error)
				log.Printf(errMsg)

				// Persist detector failure to database
				if err := w.db.FailRefreshJobDetector(dbCtx, req.JobID, completion.DetectorName, completion.Error, completion.DurationMs); err != nil {
					log.Printf("Warning: failed to persist detector failure for %s: %v", completion.DetectorName, err)
				}

				// Update failed detector count in DB
				w.db.IncrementFailedDetectors(dbCtx, req.JobID)

				// Continue processing other detectors (don't abort)
			} else {
				state.Status = "completed"
				state.IssuesFound = completion.IssuesFound
				state.DurationMs = completion.DurationMs
				state.EndTime = time.Now()

				// Persist detector completion to database
				if err := w.db.CompleteRefreshJobDetector(dbCtx, req.JobID, completion.DetectorName, completion.IssuesFound, completion.DurationMs); err != nil {
					log.Printf("Warning: failed to persist detector completion for %s: %v", completion.DetectorName, err)
					// Non-fatal, continue
				}

				issuesByDetector[completion.DetectorName] = completion.IssuesFound
				log.Printf("Detector %s found %d issues (duration: %dms)",
					completion.DetectorName, completion.IssuesFound, completion.DurationMs)
			}
		}

		completedDetectors++

		// Calculate progress (85% base + 15% for detection)
		progress := 85 + (15 * completedDetectors / totalDetectors)

		// Update detection progress in DB
		dbCtx := context.Background()
		w.db.UpdateDetectionProgress(dbCtx, req.JobID, completedDetectors, totalDetectors)

		log.Printf("Detection progress: %d/%d detectors completed", completedDetectors, totalDetectors)

		// Convert detector states to slice for JSON
		parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
		for _, name := range enabledDetectors {
			parallelDetectors = append(parallelDetectors, *detectorStates[name])
		}

		w.publishDetailedProgress(req.JobID, "running", "Running issue detection",
			fmt.Sprintf("Completed %s detector", completion.DetectorName),
			4, 4, progress,
			totalCos, totalMos, totalMops,
			nil,
			parallelDetectors,
			0, 0, 0, 0)
	})

	if err != nil {
		errMsg := fmt.Sprintf("failed to subscribe to detector completions: %w", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailJob(ctx, req.JobID, errMsg)
		return fmt.Errorf(errMsg)
	}
	defer subscription.Unsubscribe()

	// Wait for all detector jobs to complete
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			completed := completedDetectors
			mu.Unlock()
			if completed < totalDetectors {
				errMsg := fmt.Sprintf("timeout waiting for detector jobs (completed %d/%d)", completed, totalDetectors)
				w.publishError(req.JobID, errMsg)
				w.db.FailJob(ctx, req.JobID, errMsg)
				return fmt.Errorf(errMsg)
			}
		case <-ticker.C:
			mu.Lock()
			completed := completedDetectors
			failed := failedDetectors
			issues := make(map[string]int)
			for k, v := range issuesByDetector {
				issues[k] = v
			}
			mu.Unlock()

			if completed >= totalDetectors {
				log.Printf("All %d detector jobs completed (%d failed)", totalDetectors, failed)

				// Calculate total issues
				totalIssues := 0
				for _, count := range issues {
					totalIssues += count
				}

				// Finalize detection job in database
				issuesByTypeJSON, _ := json.Marshal(issues)
				dbCtx := context.Background()
				if err := w.db.CompleteDetectionJob(dbCtx, req.JobID, totalIssues, string(issuesByTypeJSON)); err != nil {
					log.Printf("Warning: failed to complete detection job: %v", err)
				}

				log.Printf("Detection complete: %d total issues found across %d detectors", totalIssues, len(issues))

				// Run anomaly detection after issue detection completes
				log.Printf("Running anomaly detection for job %s", req.JobID)
				detectorConfigService := services.NewDetectorConfigService(w.db)
				detectionService := services.NewDetectionService(w.db, detectorConfigService)
				if err := detectionService.RunAnomalyDetectors(dbCtx, req.JobID, req.Environment, req.Company, req.Facility); err != nil {
					log.Printf("Anomaly detection failed: %v", err)
					// Don't fail the job if anomaly detection fails
				} else {
					log.Printf("Anomaly detection completed for job %s", req.JobID)
				}

				// Mark refresh job complete
				w.db.CompleteJob(dbCtx, req.JobID)

				statusMsg := "Data refresh completed"
				if failed > 0 {
					statusMsg = fmt.Sprintf("Data refresh completed (%d detectors failed)", failed)
				}

				// Convert final detector states to slice for JSON
				parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
				for _, name := range enabledDetectors {
					parallelDetectors = append(parallelDetectors, *detectorStates[name])
				}

				w.publishDetailedProgress(req.JobID, "completed", statusMsg,
					fmt.Sprintf("All data loaded and analyzed successfully (%d issues found)", totalIssues),
					4, 4, 100, totalCos, totalMos, totalMops,
					nil,
					parallelDetectors,
					0, 0, totalDetectors, totalDetectors)
				w.publishComplete(req.JobID)

				log.Printf("Snapshot refresh job %s completed successfully", req.JobID)
				return nil
			}
		}
	}
}

// coordinateManualDetection coordinates a manual detection job, aggregating detector progress and publishing updates
func (w *SnapshotWorker) coordinateManualDetection(req DetectorCoordinatorMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Coordinating manual detection job %s with %d detectors", req.JobID, req.TotalDetectors)

	// Initialize detection service to get detector info
	detectorConfigService := services.NewDetectorConfigService(w.db)
	detectionService := services.NewDetectionService(w.db, detectorConfigService)

	// Build detector labels map
	detectorLabels := make(map[string]string)
	for _, name := range req.DetectorNames {
		detector := detectionService.GetDetectorByName(name)
		if detector != nil {
			detectorLabels[name] = detector.Label()
		} else {
			detectorLabels[name] = name
		}
	}

	// Track detector states
	detectorStates := make(map[string]*DetectorProgress)
	for _, name := range req.DetectorNames {
		detectorStates[name] = &DetectorProgress{
			DetectorName: name,
			DisplayLabel: detectorLabels[name],
			Status:       "pending",
		}
	}

	// Track completions
	completedDetectors := 0
	failedDetectors := 0
	issuesByDetector := make(map[string]int)
	var mu sync.Mutex

	// Publish initial progress
	initialDetectorStates := make([]DetectorProgress, 0, len(detectorStates))
	for _, name := range req.DetectorNames {
		initialDetectorStates = append(initialDetectorStates, *detectorStates[name])
	}
	w.publishDetailedProgress(req.JobID, "running", "Running issue detection", "Waiting for detectors to start",
		0, req.TotalDetectors, 0, 0, 0, 0, nil, initialDetectorStates, 0, 0, 0, 0)

	// Subscribe to detector start events
	startSubject := queue.GetDetectorStartSubject(req.JobID)
	startSub, err := w.nats.Subscribe(startSubject, func(msg *nats.Msg) {
		var start DetectorStartMessage
		if err := json.Unmarshal(msg.Data, &start); err != nil {
			log.Printf("Failed to parse detector start message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Update detector state to running
		if state, exists := detectorStates[start.DetectorName]; exists {
			state.Status = "running"
			state.StartTime = start.StartTime
		}

		log.Printf("Detector started: %s (%s)", start.DetectorName, start.DisplayLabel)

		// Convert detector states to slice for JSON
		parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
		for _, name := range req.DetectorNames {
			parallelDetectors = append(parallelDetectors, *detectorStates[name])
		}

		// Calculate progress
		progress := (completedDetectors * 100) / req.TotalDetectors

		// Send progress update showing running status
		w.publishDetailedProgress(req.JobID, "running", "Running issue detection",
			fmt.Sprintf("Running %s detector", start.DisplayLabel),
			0, req.TotalDetectors, progress, 0, 0, 0, nil, parallelDetectors, 0, 0, 0, 0)
	})

	if err != nil {
		errMsg := fmt.Sprintf("failed to subscribe to detector starts: %v", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailDetectionJob(ctx, req.JobID, errMsg)
		return fmt.Errorf(errMsg)
	}
	defer startSub.Unsubscribe()

	// Subscribe to completion events
	completeSubject := queue.GetDetectorCompleteSubject(req.JobID)
	subscription, err := w.nats.Subscribe(completeSubject, func(msg *nats.Msg) {
		var completion DetectorCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to parse detector completion message: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// Update detector state
		if state, exists := detectorStates[completion.DetectorName]; exists {
			if !completion.Success {
				state.Status = "failed"
				state.Error = completion.Error
				state.DurationMs = completion.DurationMs
				state.EndTime = time.Now()
				failedDetectors++

				log.Printf("Detector %s failed: %s", completion.DetectorName, completion.Error)
			} else {
				state.Status = "completed"
				state.IssuesFound = completion.IssuesFound
				state.DurationMs = completion.DurationMs
				state.EndTime = time.Now()

				issuesByDetector[completion.DetectorName] = completion.IssuesFound
				log.Printf("Detector %s found %d issues (duration: %dms)",
					completion.DetectorName, completion.IssuesFound, completion.DurationMs)
			}
		}

		completedDetectors++

		// Calculate progress
		progress := (completedDetectors * 100) / req.TotalDetectors

		log.Printf("Detection progress: %d/%d detectors completed", completedDetectors, req.TotalDetectors)

		// Convert detector states to slice for JSON
		parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
		for _, name := range req.DetectorNames {
			parallelDetectors = append(parallelDetectors, *detectorStates[name])
		}

		w.publishDetailedProgress(req.JobID, "running", "Running issue detection",
			fmt.Sprintf("Completed %s detector", completion.DetectorName),
			0, req.TotalDetectors, progress, 0, 0, 0, nil, parallelDetectors, 0, 0, 0, 0)
	})

	if err != nil {
		errMsg := fmt.Sprintf("failed to subscribe to detector completions: %v", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailDetectionJob(ctx, req.JobID, errMsg)
		return fmt.Errorf(errMsg)
	}
	defer subscription.Unsubscribe()

	// Wait for all detector jobs to complete
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			completed := completedDetectors
			mu.Unlock()
			if completed < req.TotalDetectors {
				errMsg := fmt.Sprintf("timeout waiting for detector jobs (completed %d/%d)", completed, req.TotalDetectors)
				w.publishError(req.JobID, errMsg)
				w.db.FailDetectionJob(ctx, req.JobID, errMsg)
				return fmt.Errorf(errMsg)
			}
		case <-ticker.C:
			mu.Lock()
			completed := completedDetectors
			failed := failedDetectors
			issues := make(map[string]int)
			for k, v := range issuesByDetector {
				issues[k] = v
			}
			mu.Unlock()

			if completed >= req.TotalDetectors {
				log.Printf("All %d detector jobs completed (%d failed)", req.TotalDetectors, failed)

				// Calculate total issues
				totalIssues := 0
				for _, count := range issues {
					totalIssues += count
				}

				// Finalize detection job in database
				issuesByTypeJSON, _ := json.Marshal(issues)
				dbCtx := context.Background()
				if err := w.db.CompleteDetectionJob(dbCtx, req.JobID, totalIssues, string(issuesByTypeJSON)); err != nil {
					log.Printf("Warning: failed to complete detection job: %v", err)
				}

				// Mark refresh job as completed
				if err := w.db.CompleteJob(dbCtx, req.JobID); err != nil {
					log.Printf("Warning: failed to complete refresh job: %v", err)
				}

				log.Printf("Detection complete: %d total issues found across %d detectors", totalIssues, len(issues))

				// Run anomaly detection after issue detection completes
				log.Printf("Running anomaly detection for job %s", req.JobID)
				if err := detectionService.RunAnomalyDetectors(dbCtx, req.JobID, req.Environment, req.Company, req.Facility); err != nil {
					log.Printf("Anomaly detection failed: %v", err)
					// Don't fail the job if anomaly detection fails
				} else {
					log.Printf("Anomaly detection completed for job %s", req.JobID)
				}

				// Convert final detector states to slice for JSON
				parallelDetectors := make([]DetectorProgress, 0, len(detectorStates))
				for _, name := range req.DetectorNames {
					parallelDetectors = append(parallelDetectors, *detectorStates[name])
				}

				// Publish final progress
				statusMsg := "Detection completed"
				if failed > 0 {
					statusMsg = fmt.Sprintf("Detection completed (%d detectors failed)", failed)
				}

				w.publishDetailedProgress(req.JobID, "completed", statusMsg,
					fmt.Sprintf("All detectors finished (%d issues found)", totalIssues),
					req.TotalDetectors, req.TotalDetectors, 100, 0, 0, 0, nil, parallelDetectors, 0, 0, 0, 0)
				w.publishComplete(req.JobID)

				log.Printf("Manual detection job %s completed successfully", req.JobID)
				return nil
			}
		}
	}
}

// handleManualDetectionCoordinator handles a manual detection coordinator request
func (w *SnapshotWorker) handleManualDetectionCoordinator(msg *nats.Msg) {
	var req DetectorCoordinatorMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("Failed to parse coordinator request: %v", err)
		return
	}

	log.Printf("Coordinating manual detection job: %s", req.JobID)

	// Run coordinator in goroutine (non-blocking)
	go func() {
		if err := w.coordinateManualDetection(req); err != nil {
			log.Printf("Manual detection coordination failed for job %s: %v", req.JobID, err)
		}
	}()
}
