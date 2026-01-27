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


// DataBatchJobMessage represents work for loading one data type (MOPs, MOs, or COs)
type DataBatchJobMessage struct {
	JobID       string `json:"jobId"`
	ParentJobID string `json:"parentJobId"`
	DataType    string `json:"dataType"`    // "mops", "mos", "cos"
	Environment string `json:"environment"`
	AccessToken string `json:"accessToken"`
	Company     string `json:"company"`
	Facility    string `json:"facility"`
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

// DetectorJobMessage represents work for running one detector
type DetectorJobMessage struct {
	JobID        string `json:"jobId"`        // "abc123-unlinked"
	ParentJobID  string `json:"parentJobId"`  // "abc123"
	DetectorName string `json:"detectorName"` // "unlinked_production_orders"
	Environment  string `json:"environment"`  // "TRN" or "PRD"
	Company      string `json:"company"`      // "100"
	Facility     string `json:"facility"`     // "AZ1"
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
	// Subscribe to TRN batch work distribution (parallel data loading)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotBatchTRN,
		queue.QueueGroupBatchWorkers,
		w.handleBatchJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to TRN batch jobs: %w", err)
	}

	// Subscribe to PRD batch work distribution (batch worker)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotBatchPRD,
		queue.QueueGroupBatchWorkers,
		w.handleBatchJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to PRD batch jobs: %w", err)
	}

	// Subscribe to TRN detector jobs (detector worker)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotDetectorTRN,
		queue.QueueGroupBatchWorkers,
		w.handleDetectorJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to TRN detector jobs: %w", err)
	}

	// Subscribe to PRD detector jobs (detector worker)
	_, err = w.nats.QueueSubscribe(
		queue.SubjectSnapshotDetectorPRD,
		queue.QueueGroupBatchWorkers,
		w.handleDetectorJob,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to PRD detector jobs: %w", err)
	}

	// IMPORTANT: Limit concurrent message processing by handling synchronously
	// Each worker will process one message at a time, allowing NATS to distribute
	// the remaining messages to other available workers

	log.Println("Snapshot worker started and listening for jobs, phase work, and batch work")
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

// processRefresh coordinates parallel data refresh using NATS batch distribution with ID range partitioning
// Optimized for Apache Spark: Uses predicate pushdown (WHERE ID >= X AND ID < Y) instead of OFFSET/LIMIT
func (w *SnapshotWorker) processRefresh(req SnapshotRefreshMessage) error {
	ctx := context.Background()
	log.Printf("Coordinating refresh job %s with parallel ID range batching", req.JobID)

	// Start the job
	if err := w.db.StartJob(ctx, req.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Phase 0: Truncate database (must complete first)
	log.Printf("Phase 0: Truncating database for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Preparing database", "Truncating tables",
		0, 6, 0, 0, 0, 0, 0, 0, 0, 0)

	if err := w.db.TruncateAnalysisTables(ctx, req.Environment); err != nil {
		w.publishError(req.JobID, fmt.Sprintf("Truncate failed: %v", err))
		w.db.FailJob(ctx, req.JobID, err.Error())
		return fmt.Errorf("truncate failed: %w", err)
	}

	log.Printf("Phase 0 complete: Database truncated")
	w.publishDetailedProgress(req.JobID, "running", "Database prepared", "Publishing data jobs",
		1, 4, 20, 0, 0, 0, 0, 0, 0, 0)

	// Phase 1: Publish 3 data jobs to NATS (one per data type) and wait for completion
	log.Printf("Phase 1: Publishing 3 data jobs (MOPs, MOs, COs) to NATS...")
	return w.publishDataJobs(req)
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

	ctx := context.Background()

	// Check if parent job has been cancelled
	parentJob, err := w.db.GetRefreshJob(ctx, job.ParentJobID)
	if err != nil {
		log.Printf("Failed to check parent job status: %v", err)
		// Continue processing if we can't check status (assume not cancelled)
	} else if parentJob != nil && parentJob.Status == "failed" && parentJob.ErrorMessage.Valid && parentJob.ErrorMessage.String == "Cancelled by user" {
		log.Printf("Parent job %s was cancelled, skipping %s batch", job.ParentJobID, job.DataType)
		return
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

	var recordCount int
	var fetchErr error

	// Execute full query based on data type (no ID range filtering)
	switch job.DataType {
	case "mops":
		recordCount, fetchErr = snapshotService.RefreshPlannedOrders(ctx, job.Environment, job.Company, job.Facility)

	case "mos":
		recordCount, fetchErr = snapshotService.RefreshManufacturingOrders(ctx, job.Environment, job.Company, job.Facility)

	case "cos":
		recordCount, fetchErr = snapshotService.RefreshOpenCustomerOrderLines(ctx, job.Environment, job.Company, job.Facility)

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
	parentJob, err := w.db.GetRefreshJob(ctx, job.ParentJobID)
	if err != nil {
		log.Printf("Failed to check parent job status: %v", err)
		// Continue processing if we can't check status (assume not cancelled)
	} else if parentJob != nil && parentJob.Status == "failed" && parentJob.ErrorMessage.Valid && parentJob.ErrorMessage.String == "Cancelled by user" {
		log.Printf("Parent job %s was cancelled, skipping detector %s", job.ParentJobID, job.DetectorName)
		return
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
	issuesFound, err := detector.Detect(ctx, w.db, job.Environment, job.Company, job.Facility)

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
	w.publishDetailedProgress(req.JobID, "running", "Publishing data jobs", "Distributing work to workers",
		1, 4, 25, 0, 0, 0, 0, 0, 0, 3)

	// Publish 3 jobs (one per data type)
	dataTypes := []string{"mops", "mos", "cos"}

	for _, dataType := range dataTypes {
		job := DataBatchJobMessage{
			JobID:       fmt.Sprintf("%s-%s", req.JobID, dataType),
			ParentJobID: req.JobID,
			DataType:    dataType,
			Environment: req.Environment,
			AccessToken: req.AccessToken,
			Company:     req.Company,
			Facility:    req.Facility,
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
	var mu sync.Mutex

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
			w.publishError(req.JobID, errMsg)
			w.db.FailJob(ctx, req.JobID, completion.Error)
			cancel() // Cancel context to abort
			return
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

		w.publishDetailedProgress(req.JobID, "running", "Loading data",
			fmt.Sprintf("Loaded %s", completion.DataType),
			2, 4, progress,
			totalCos, totalMos, totalMops,
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
		4, 4, 85,
		totalCos, totalMos, totalMops,
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
			4, 4, 100, totalCos, totalMos, totalMops, 0, 0, 0, 0)
		w.publishComplete(req.JobID)
		return nil
	}

	log.Printf("Found %d enabled detectors to run", totalDetectors)

	// Create detection job record
	if err := w.db.CreateIssueDetectionJob(ctx, req.JobID, totalDetectors); err != nil {
		return fmt.Errorf("failed to create detection job: %w", err)
	}

	// Clear previous issues for this job
	if err := w.db.ClearIssuesForJob(ctx, req.JobID); err != nil {
		log.Printf("Warning: failed to clear previous issues: %v", err)
	}

	// Publish detector jobs
	for _, detectorName := range enabledDetectors {
		job := DetectorJobMessage{
			JobID:        fmt.Sprintf("%s-%s", req.JobID, detectorName),
			ParentJobID:  req.JobID,
			DetectorName: detectorName,
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

	// Track completions
	completedDetectors := 0
	failedDetectors := 0
	issuesByDetector := make(map[string]int)
	var mu sync.Mutex

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

		if !completion.Success {
			errMsg := fmt.Sprintf("Detector %s failed: %s", completion.DetectorName, completion.Error)
			log.Printf(errMsg)
			failedDetectors++

			// Update failed detector count in DB
			dbCtx := context.Background()
			w.db.IncrementFailedDetectors(dbCtx, req.JobID)

			// Continue processing other detectors (don't abort)
		} else {
			issuesByDetector[completion.DetectorName] = completion.IssuesFound
			log.Printf("Detector %s found %d issues (duration: %dms)",
				completion.DetectorName, completion.IssuesFound, completion.DurationMs)
		}

		completedDetectors++

		// Calculate progress (85% base + 15% for detection)
		progress := 85 + (15 * completedDetectors / totalDetectors)

		// Update detection progress in DB
		dbCtx := context.Background()
		w.db.UpdateDetectionProgress(dbCtx, req.JobID, completedDetectors, totalDetectors)

		log.Printf("Detection progress: %d/%d detectors completed", completedDetectors, totalDetectors)

		w.publishDetailedProgress(req.JobID, "running", "Running issue detection",
			fmt.Sprintf("Completed %s detector", completion.DetectorName),
			4, 4, progress,
			totalCos, totalMos, totalMops,
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

				// Mark refresh job complete
				w.db.CompleteJob(dbCtx, req.JobID)

				statusMsg := "Data refresh completed"
				if failed > 0 {
					statusMsg = fmt.Sprintf("Data refresh completed (%d detectors failed)", failed)
				}

				w.publishDetailedProgress(req.JobID, "completed", statusMsg,
					fmt.Sprintf("All data loaded and analyzed successfully (%d issues found)", totalIssues),
					4, 4, 100, totalCos, totalMos, totalMops,
					0, 0, totalDetectors, totalDetectors)
				w.publishComplete(req.JobID)

				log.Printf("Snapshot refresh job %s completed successfully", req.JobID)
				return nil
			}
		}
	}
}
