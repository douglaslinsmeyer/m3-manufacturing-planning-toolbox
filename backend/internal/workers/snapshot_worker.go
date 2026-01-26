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

// DataBatchJobMessage represents work for a data batch (ID range)
type DataBatchJobMessage struct {
	JobID        string      `json:"jobId"`
	ParentJobID  string      `json:"parentJobId"`
	Phase        string      `json:"phase"`        // "mops", "mos", "cos"
	BatchNumber  int         `json:"batchNumber"`  // 1-based
	TotalBatches int         `json:"totalBatches"`
	MinID        interface{} `json:"minId"`        // int64 for PLPN, string for MFNO/ORNO
	MaxID        interface{} `json:"maxId"`        // int64 for PLPN, string for MFNO/ORNO
	Environment  string      `json:"environment"`
	AccessToken  string      `json:"accessToken"`
	Company      string      `json:"company"`
	Facility     string      `json:"facility"`
}

// BatchCompletionMessage signals batch completion
type BatchCompletionMessage struct {
	JobID        string `json:"jobId"`
	ParentJobID  string `json:"parentJobId"`
	Phase        string `json:"phase"`
	BatchNumber  int    `json:"batchNumber"`
	RecordCount  int    `json:"recordCount"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
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

	// Subscribe to TRN batch work distribution (batch worker for ID range-based batching)
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
	w.publishDetailedProgress(req.JobID, "running", "Database prepared", "Analyzing dataset ranges",
		1, 6, 10, 0, 0, 0, 0, 0, 0, 0)

	// Phase 1: Query MIN/MAX/COUNT for all 3 entity types (parallel)
	log.Printf("Phase 1: Querying data ranges (MIN/MAX/COUNT)...")

	type rangeResult struct {
		phase string
		meta  *services.RangeMetadata
		err   error
	}

	rangeChan := make(chan rangeResult, 3)

	// Query ranges in parallel
	go func() {
		meta, err := snapshotService.QueryMOPRange(ctx, req.Company, req.Facility)
		rangeChan <- rangeResult{"mops", meta, err}
	}()
	go func() {
		meta, err := snapshotService.QueryMORange(ctx, req.Company, req.Facility)
		rangeChan <- rangeResult{"mos", meta, err}
	}()
	go func() {
		meta, err := snapshotService.QueryCORange(ctx, req.Company, req.Facility)
		rangeChan <- rangeResult{"cos", meta, err}
	}()

	// Collect range results
	ranges := make(map[string]*services.RangeMetadata)
	for i := 0; i < 3; i++ {
		result := <-rangeChan
		if result.err != nil {
			errMsg := fmt.Sprintf("Range query failed for %s: %v", result.phase, result.err)
			w.publishError(req.JobID, errMsg)
			w.db.FailJob(ctx, req.JobID, result.err.Error())
			return fmt.Errorf(errMsg)
		}
		ranges[result.phase] = result.meta
	}

	log.Printf("Data ranges - MOPs: %d records, MOs: %d records, COs: %d records",
		ranges["mops"].TotalRecords,
		ranges["mos"].TotalRecords,
		ranges["cos"].TotalRecords)

	w.publishDetailedProgress(req.JobID, "running", "Ranges queried", "Calculating batch partitions",
		2, 6, 20,
		0, 0, 0,
		0, 0, 0, 0)

	// Phase 2: Calculate batch ranges
	log.Printf("Phase 2: Calculating batch partitions...")

	// Load batching settings
	batchSize := services.LoadSystemSettingInt(w.db, "compass_batch_size", 50000)
	overPartitionFactor := services.LoadSystemSettingFloat(w.db, "compass_over_partition_factor", 1.5)

	log.Printf("Batching settings: batch_size=%d, over_partition_factor=%.1f", batchSize, overPartitionFactor)

	mopBatches := services.CalculateBatchRanges(*ranges["mops"], batchSize, overPartitionFactor)
	moBatches := services.CalculateBatchRanges(*ranges["mos"], batchSize, overPartitionFactor)
	coBatches := services.CalculateBatchRanges(*ranges["cos"], batchSize, overPartitionFactor)

	totalBatches := len(mopBatches) + len(moBatches) + len(coBatches)

	// Count single batches for small datasets (nil = no partitioning, use full query)
	if len(mopBatches) == 0 && ranges["mops"].TotalRecords > 0 {
		totalBatches++
	}
	if len(moBatches) == 0 && ranges["mos"].TotalRecords > 0 {
		totalBatches++
	}
	if len(coBatches) == 0 && ranges["cos"].TotalRecords > 0 {
		totalBatches++
	}

	log.Printf("Batch plan: %d MOP batches, %d MO batches, %d CO batches (total: %d)",
		len(mopBatches), len(moBatches), len(coBatches), totalBatches)

	if totalBatches == 0 {
		log.Printf("No data to load, skipping to finalize")
		w.publishDetailedProgress(req.JobID, "running", "No data found", "Running finalize",
			3, 6, 80, 0, 0, 0, 0, 0, 0, 0)
		return w.runFinalizeForBatches(req, map[string]int{"mops": 0, "mos": 0, "cos": 0})
	}

	// Phase 3: Publish batch jobs to NATS and wait for completion
	return w.publishAndWaitForBatches(req, mopBatches, moBatches, coBatches, ranges, totalBatches)
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

	log.Printf("Processing batch %d/%d for %s (range: %v-%v)",
		job.BatchNumber, job.TotalBatches, job.Phase, job.MinID, job.MaxID)

	ctx := context.Background()

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

	// Execute batch query based on phase and ID type
	switch job.Phase {
	case "mops":
		// Numeric ID (PLPN)
		minID, okMin := job.MinID.(float64) // JSON unmarshals numbers as float64
		maxID, okMax := job.MaxID.(float64)
		if !okMin || !okMax {
			log.Printf("Invalid ID types for MOP batch: minID=%T, maxID=%T", job.MinID, job.MaxID)
			w.publishBatchCompletion(job, 0, fmt.Errorf("invalid ID types"))
			return
		}
		recordCount, fetchErr = snapshotService.RefreshMOPBatch(ctx, job.Company, job.Facility,
			int64(minID), int64(maxID))

	case "mos":
		// String ID (MFNO)
		minID, okMin := job.MinID.(string)
		maxID, okMax := job.MaxID.(string)
		if !okMin || !okMax {
			log.Printf("Invalid ID types for MO batch: minID=%T, maxID=%T", job.MinID, job.MaxID)
			w.publishBatchCompletion(job, 0, fmt.Errorf("invalid ID types"))
			return
		}
		recordCount, fetchErr = snapshotService.RefreshMOBatch(ctx, job.Company, job.Facility, minID, maxID)

	case "cos":
		// String ID (ORNO)
		minID, okMin := job.MinID.(string)
		maxID, okMax := job.MaxID.(string)
		if !okMin || !okMax {
			log.Printf("Invalid ID types for CO batch: minID=%T, maxID=%T", job.MinID, job.MaxID)
			w.publishBatchCompletion(job, 0, fmt.Errorf("invalid ID types"))
			return
		}
		recordCount, fetchErr = snapshotService.RefreshCOBatch(ctx, job.Company, job.Facility, minID, maxID)

	default:
		log.Printf("Unknown phase: %s", job.Phase)
		w.publishBatchCompletion(job, 0, fmt.Errorf("unknown phase: %s", job.Phase))
		return
	}

	// Publish batch completion
	w.publishBatchCompletion(job, recordCount, fetchErr)
}

// publishBatchCompletion publishes a batch completion message
func (w *SnapshotWorker) publishBatchCompletion(job DataBatchJobMessage, recordCount int, err error) {
	completion := BatchCompletionMessage{
		JobID:       job.JobID,
		ParentJobID: job.ParentJobID,
		Phase:       job.Phase,
		BatchNumber: job.BatchNumber,
		RecordCount: recordCount,
		Success:     err == nil,
	}
	if err != nil {
		completion.Error = err.Error()
		log.Printf("Batch %d/%d for %s failed: %v", job.BatchNumber, job.TotalBatches, job.Phase, err)
	} else {
		log.Printf("Batch %d/%d for %s completed: %d records", job.BatchNumber, job.TotalBatches, job.Phase, recordCount)
	}

	data, _ := json.Marshal(completion)
	completeSubject := queue.GetBatchCompleteSubject(job.ParentJobID)
	if err := w.nats.Publish(completeSubject, data); err != nil {
		log.Printf("Failed to publish batch completion: %v", err)
	}
}

// publishAndWaitForBatches publishes batch jobs to NATS and waits for all to complete
func (w *SnapshotWorker) publishAndWaitForBatches(
	req SnapshotRefreshMessage,
	mopBatches, moBatches, coBatches []services.BatchRange,
	ranges map[string]*services.RangeMetadata,
	totalBatches int,
) error {
	ctx := context.Background()

	log.Printf("Phase 3: Publishing %d batch jobs to NATS queue...", totalBatches)
	w.publishDetailedProgress(req.JobID, "running", "Publishing batch jobs", "Distributing work to workers",
		3, 6, 30, 0, 0, 0, 0, 0, 0, totalBatches)

	batchCount := 0

	// Publish MOP batches
	if len(mopBatches) > 0 {
		for _, batch := range mopBatches {
			batchJob := DataBatchJobMessage{
				JobID:        fmt.Sprintf("%s-mop-b%d", req.JobID, batch.BatchNumber),
				ParentJobID:  req.JobID,
				Phase:        "mops",
				BatchNumber:  batch.BatchNumber,
				TotalBatches: len(mopBatches),
				MinID:        batch.MinID,
				MaxID:        batch.MaxID,
				Environment:  req.Environment,
				AccessToken:  req.AccessToken,
				Company:      req.Company,
				Facility:     req.Facility,
			}

			data, _ := json.Marshal(batchJob)
			subject := queue.GetBatchSubject(req.Environment, "mops")
			if err := w.nats.Publish(subject, data); err != nil {
				log.Printf("Failed to publish MOP batch: %v", err)
				w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
				w.db.FailJob(ctx, req.JobID, err.Error())
				return err
			}
			batchCount++
		}
		log.Printf("Published %d MOP batch jobs", len(mopBatches))
	} else if ranges["mops"].TotalRecords > 0 {
		// Single batch (no partitioning) - use full MIN/MAX range
		batchJob := DataBatchJobMessage{
			JobID:        fmt.Sprintf("%s-mop-b1", req.JobID),
			ParentJobID:  req.JobID,
			Phase:        "mops",
			BatchNumber:  1,
			TotalBatches: 1,
			MinID:        ranges["mops"].MinID,
			MaxID:        ranges["mops"].MaxID,
			Environment:  req.Environment,
			AccessToken:  req.AccessToken,
			Company:      req.Company,
			Facility:     req.Facility,
		}

		data, _ := json.Marshal(batchJob)
		subject := queue.GetBatchSubject(req.Environment, "mops")
		if err := w.nats.Publish(subject, data); err != nil {
			log.Printf("Failed to publish MOP batch: %v", err)
			w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
			w.db.FailJob(ctx, req.JobID, err.Error())
			return err
		}
		batchCount++
		log.Printf("Published 1 MOP batch job (no partitioning)")
	}

	// Publish MO batches
	if len(moBatches) > 0 {
		for _, batch := range moBatches {
			batchJob := DataBatchJobMessage{
				JobID:        fmt.Sprintf("%s-mo-b%d", req.JobID, batch.BatchNumber),
				ParentJobID:  req.JobID,
				Phase:        "mos",
				BatchNumber:  batch.BatchNumber,
				TotalBatches: len(moBatches),
				MinID:        batch.MinID,
				MaxID:        batch.MaxID,
				Environment:  req.Environment,
				AccessToken:  req.AccessToken,
				Company:      req.Company,
				Facility:     req.Facility,
			}

			data, _ := json.Marshal(batchJob)
			subject := queue.GetBatchSubject(req.Environment, "mos")
			if err := w.nats.Publish(subject, data); err != nil {
				log.Printf("Failed to publish MO batch: %v", err)
				w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
				w.db.FailJob(ctx, req.JobID, err.Error())
				return err
			}
			batchCount++
		}
		log.Printf("Published %d MO batch jobs", len(moBatches))
	} else if ranges["mos"].TotalRecords > 0 {
		batchJob := DataBatchJobMessage{
			JobID:        fmt.Sprintf("%s-mo-b1", req.JobID),
			ParentJobID:  req.JobID,
			Phase:        "mos",
			BatchNumber:  1,
			TotalBatches: 1,
			MinID:        ranges["mos"].MinID,
			MaxID:        ranges["mos"].MaxID,
			Environment:  req.Environment,
			AccessToken:  req.AccessToken,
			Company:      req.Company,
			Facility:     req.Facility,
		}

		data, _ := json.Marshal(batchJob)
		subject := queue.GetBatchSubject(req.Environment, "mos")
		if err := w.nats.Publish(subject, data); err != nil {
			log.Printf("Failed to publish MO batch: %v", err)
			w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
			w.db.FailJob(ctx, req.JobID, err.Error())
			return err
		}
		batchCount++
		log.Printf("Published 1 MO batch job (no partitioning)")
	}

	// Publish CO batches
	if len(coBatches) > 0 {
		for _, batch := range coBatches {
			batchJob := DataBatchJobMessage{
				JobID:        fmt.Sprintf("%s-co-b%d", req.JobID, batch.BatchNumber),
				ParentJobID:  req.JobID,
				Phase:        "cos",
				BatchNumber:  batch.BatchNumber,
				TotalBatches: len(coBatches),
				MinID:        batch.MinID,
				MaxID:        batch.MaxID,
				Environment:  req.Environment,
				AccessToken:  req.AccessToken,
				Company:      req.Company,
				Facility:     req.Facility,
			}

			data, _ := json.Marshal(batchJob)
			subject := queue.GetBatchSubject(req.Environment, "cos")
			if err := w.nats.Publish(subject, data); err != nil {
				log.Printf("Failed to publish CO batch: %v", err)
				w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
				w.db.FailJob(ctx, req.JobID, err.Error())
				return err
			}
			batchCount++
		}
		log.Printf("Published %d CO batch jobs", len(coBatches))
	} else if ranges["cos"].TotalRecords > 0 {
		batchJob := DataBatchJobMessage{
			JobID:        fmt.Sprintf("%s-co-b1", req.JobID),
			ParentJobID:  req.JobID,
			Phase:        "cos",
			BatchNumber:  1,
			TotalBatches: 1,
			MinID:        ranges["cos"].MinID,
			MaxID:        ranges["cos"].MaxID,
			Environment:  req.Environment,
			AccessToken:  req.AccessToken,
			Company:      req.Company,
			Facility:     req.Facility,
		}

		data, _ := json.Marshal(batchJob)
		subject := queue.GetBatchSubject(req.Environment, "cos")
		if err := w.nats.Publish(subject, data); err != nil {
			log.Printf("Failed to publish CO batch: %v", err)
			w.publishError(req.JobID, fmt.Sprintf("Failed to publish batch: %v", err))
			w.db.FailJob(ctx, req.JobID, err.Error())
			return err
		}
		batchCount++
		log.Printf("Published 1 CO batch job (no partitioning)")
	}

	log.Printf("Published %d total batch jobs, waiting for completion...", batchCount)

	// Wait for all batches to complete
	return w.waitForBatchesAndFinalize(req, batchCount, ranges)
}

// waitForBatchesAndFinalize waits for all batches to complete, then runs finalize and detection
func (w *SnapshotWorker) waitForBatchesAndFinalize(
	req SnapshotRefreshMessage,
	totalBatches int,
	ranges map[string]*services.RangeMetadata,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	log.Printf("Phase 4: Waiting for %d batches to complete...", totalBatches)

	// Track batch completions
	completedBatches := 0
	recordsByPhase := map[string]int{
		"mops": 0,
		"mos":  0,
		"cos":  0,
	}
	var mu sync.Mutex

	// Subscribe to batch completion events
	completeSubject := queue.GetBatchCompleteSubject(req.JobID)
	sub, err := w.nats.Subscribe(completeSubject, func(msg *nats.Msg) {
		var completion BatchCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to parse batch completion: %v", err)
			return
		}

		mu.Lock()
		completedBatches++
		recordsByPhase[completion.Phase] += completion.RecordCount

		// Calculate progress percentage (batches are 30-70% of total work)
		batchProgress := 30 + int(float64(completedBatches)/float64(totalBatches)*40)

		log.Printf("Batch %d/%d completed: %s batch %d, %d records (total %s: %d)",
			completedBatches, totalBatches,
			completion.Phase, completion.BatchNumber, completion.RecordCount,
			completion.Phase, recordsByPhase[completion.Phase])

		// Publish detailed progress update
		w.publishDetailedProgress(req.JobID, "running",
			"Loading data batches",
			fmt.Sprintf("Batch %d/%d complete", completedBatches, totalBatches),
			4, 6, batchProgress,
			recordsByPhase["cos"], recordsByPhase["mos"], recordsByPhase["mops"],
			0, 0, completedBatches, totalBatches)

		mu.Unlock()
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to subscribe to batch completions: %v", err)
		w.publishError(req.JobID, errMsg)
		w.db.FailJob(context.Background(), req.JobID, errMsg)
		return fmt.Errorf(errMsg)
	}
	defer sub.Unsubscribe()

	// Wait for all batches with timeout
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			errMsg := fmt.Sprintf("Timeout waiting for batches (%d/%d completed)", completedBatches, totalBatches)
			w.publishError(req.JobID, errMsg)
			w.db.FailJob(context.Background(), req.JobID, errMsg)
			return fmt.Errorf(errMsg)

		case <-ticker.C:
			mu.Lock()
			complete := completedBatches >= totalBatches
			mu.Unlock()

			if complete {
				log.Printf("All %d batches completed successfully", totalBatches)
				log.Printf("Total records loaded - MOPs: %d, MOs: %d, COs: %d",
					recordsByPhase["mops"], recordsByPhase["mos"], recordsByPhase["cos"])

				// Run finalize and detection
				return w.runFinalizeForBatches(req, recordsByPhase)
			}
		}
	}
}

// runFinalizeForBatches executes finalize and detection phases after batches complete
func (w *SnapshotWorker) runFinalizeForBatches(req SnapshotRefreshMessage, recordsByPhase map[string]int) error {
	ctx := context.Background()

	// Phase 5: Finalize unified views
	log.Printf("Phase 5: Running finalize for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Finalizing data", "Updating production orders view",
		5, 6, 80,
		recordsByPhase["cos"],
		recordsByPhase["mos"],
		recordsByPhase["mops"],
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

	// Phase 6: Detection
	log.Printf("Phase 6: Running detection for job %s", req.JobID)
	w.publishDetailedProgress(req.JobID, "running", "Running issue detectors", "Analyzing data",
		6, 6, 90,
		recordsByPhase["cos"],
		recordsByPhase["mos"],
		recordsByPhase["mops"],
		0, 0, 0, 0)

	detectionService := services.NewDetectionService(w.db)
	if err := detectionService.RunAllDetectors(ctx, req.JobID, req.Company, req.Facility); err != nil {
		log.Printf("Detection warning: %v", err)
		// Don't fail job on detection errors
	}

	// Mark job complete
	w.db.CompleteJob(ctx, req.JobID)
	w.publishDetailedProgress(req.JobID, "completed", "Data refresh completed", "All data loaded successfully",
		6, 6, 100,
		recordsByPhase["cos"],
		recordsByPhase["mos"],
		recordsByPhase["mops"],
		0, 0, 0, 0)
	w.publishComplete(req.JobID)

	log.Printf("Refresh job %s completed successfully - MOPs: %d, MOs: %d, COs: %d",
		req.JobID,
		recordsByPhase["mops"],
		recordsByPhase["mos"],
		recordsByPhase["cos"])

	return nil
}
