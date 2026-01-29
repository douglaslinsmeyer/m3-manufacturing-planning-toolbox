package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// BulkOperationWorker handles async bulk operation jobs
type BulkOperationWorker struct {
	nats           *queue.Manager
	db             *db.Queries
	m3Client       *m3api.Client
	rateLimiter    *services.RateLimiterService
	jobContexts    map[string]context.CancelFunc // Track job cancellation contexts
	jobContextsMux sync.RWMutex                  // Protect concurrent access
}

// NewBulkOperationWorker creates a new bulk operation worker
func NewBulkOperationWorker(
	natsManager *queue.Manager,
	database *db.Queries,
	m3Client *m3api.Client,
	rateLimiter *services.RateLimiterService,
) *BulkOperationWorker {
	return &BulkOperationWorker{
		nats:        natsManager,
		db:          database,
		m3Client:    m3Client,
		rateLimiter: rateLimiter,
		jobContexts: make(map[string]context.CancelFunc),
	}
}

// Message Types

// BulkOperationJobMessage represents a bulk operation coordinator job
type BulkOperationJobMessage struct {
	JobID       string                 `json:"job_id"`
	Environment string                 `json:"environment"`
	UserID      string                 `json:"user_id"`
	Operation   string                 `json:"operation"` // "delete", "close", "reschedule"
	IssueIDs    []int64                `json:"issue_ids"`
	Params      map[string]interface{} `json:"params,omitempty"` // For reschedule date, etc.
}

// BulkOperationBatchMessage represents work for processing one batch
type BulkOperationBatchMessage struct {
	JobID            string                   `json:"job_id"`
	BatchNumber      int                      `json:"batch_number"`
	Environment      string                   `json:"environment"`
	Operation        string                   `json:"operation"`
	ProductionOrders []ProductionOrderBatch   `json:"production_orders"`
	Params           map[string]interface{}   `json:"params,omitempty"`
}

// ProductionOrderBatch contains details for a single production order in a batch
type ProductionOrderBatch struct {
	ID          int64  `json:"id"`           // Database ID
	OrderNumber string `json:"order_number"` // MOP PLPN or MO MFNO
	OrderType   string `json:"order_type"`   // "MOP" or "MO"
	Facility    string `json:"facility"`     // Required for MOs
	Company     string `json:"company"`      // Required for API calls
}

// IssueToOrderMapping tracks which issues reference which production orders
type IssueToOrderMapping struct {
	IssueID           int64
	ProductionOrderID int64
	OrderNumber       string
	OrderType         string
	Facility          string
	Company           string
}

// DuplicateInfo tracks duplicate detection metadata
type DuplicateInfo struct {
	OrderToIssues  map[int64][]int64 // production_order_id -> []issue_id
	PrimaryIssueID map[int64]int64   // production_order_id -> first issue_id
}

// BulkOpBatchStartMessage signals that a worker has picked up a batch job
type BulkOpBatchStartMessage struct {
	JobID       string    `json:"job_id"`
	BatchNumber int       `json:"batch_number"`
	StartTime   time.Time `json:"start_time"`
}

// BulkOpBatchCompletionMessage signals batch completion
type BulkOpBatchCompletionMessage struct {
	JobID       string `json:"job_id"`
	BatchNumber int    `json:"batch_number"`
	Successful  int    `json:"successful"`
	Failed      int    `json:"failed"`
	Error       string `json:"error,omitempty"`
}

// BulkOpProgressMessage represents progress update
type BulkOpProgressMessage struct {
	JobID              string `json:"job_id"`
	Status             string `json:"status"`
	Phase              string `json:"phase"`
	ProgressPercentage int    `json:"progress_percentage"`
	TotalItems         int    `json:"total_items"`
	SuccessfulItems    int    `json:"successful_items"`
	FailedItems        int    `json:"failed_items"`
}

// Start starts the bulk operation worker and subscribes to NATS subjects
func (w *BulkOperationWorker) Start(ctx context.Context) error {
	log.Println("Starting bulk operation worker...")

	// Subscribe to coordinator jobs (TRN and PRD)
	if _, err := w.nats.QueueSubscribe(
		queue.SubjectBulkOpRequestTRN,
		queue.QueueGroupBulkOpCoordinator,
		w.handleCoordinatorJob,
	); err != nil {
		return fmt.Errorf("failed to subscribe to TRN bulk operation requests: %w", err)
	}

	if _, err := w.nats.QueueSubscribe(
		queue.SubjectBulkOpRequestPRD,
		queue.QueueGroupBulkOpCoordinator,
		w.handleCoordinatorJob,
	); err != nil {
		return fmt.Errorf("failed to subscribe to PRD bulk operation requests: %w", err)
	}

	// Subscribe to batch jobs (wildcard for all environments and job IDs)
	if _, err := w.nats.QueueSubscribe(
		"bulkop.batch.>",
		queue.QueueGroupBulkOpWorkers,
		w.handleBatchJob,
	); err != nil {
		return fmt.Errorf("failed to subscribe to bulk operation batch jobs: %w", err)
	}

	// Subscribe to cancellation requests (broadcast, not queue)
	if _, err := w.nats.Subscribe(
		"bulkop.cancel.*",
		w.handleCancellation,
	); err != nil {
		return fmt.Errorf("failed to subscribe to bulk operation cancellations: %w", err)
	}

	log.Println("Bulk operation worker started successfully")
	return nil
}

// handleCoordinatorJob handles bulk operation coordinator jobs
func (w *BulkOperationWorker) handleCoordinatorJob(msg *nats.Msg) {
	var jobMsg BulkOperationJobMessage
	if err := json.Unmarshal(msg.Data, &jobMsg); err != nil {
		log.Printf("Failed to unmarshal bulk operation job message: %v", err)
		return
	}

	log.Printf("Coordinator picked up bulk operation job %s (operation: %s, issues: %d)",
		jobMsg.JobID, jobMsg.Operation, len(jobMsg.IssueIDs))

	// Create cancellable context for this job
	ctx, cancel := context.WithCancel(context.Background())
	w.registerJobContext(jobMsg.JobID, cancel)
	defer w.unregisterJobContext(jobMsg.JobID)

	// Execute coordinator logic
	if err := w.runCoordinator(ctx, jobMsg); err != nil {
		log.Printf("Coordinator job %s failed: %v", jobMsg.JobID, err)
		w.db.FailBulkOperationJob(context.Background(), jobMsg.JobID, err.Error())
		w.publishProgress(jobMsg.JobID, "failed", "Error", 0, 0, 0, 0)
	}
}

// runCoordinator orchestrates the bulk operation
func (w *BulkOperationWorker) runCoordinator(ctx context.Context, jobMsg BulkOperationJobMessage) error {
	// Mark job as started
	if err := w.db.StartBulkOperationJob(ctx, jobMsg.JobID); err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}

	w.publishProgress(jobMsg.JobID, "running", "Fetching issue mappings", 5, 0, 0, 0)

	// Step 1: Fetch ALL issue-to-order mappings (no deduplication yet)
	mappings, err := w.fetchIssueToOrderMappings(ctx, jobMsg.Environment, jobMsg.IssueIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch issue mappings: %w", err)
	}

	if len(mappings) == 0 {
		// No orders to process - mark as completed
		w.db.CompleteBulkOperationJob(ctx, jobMsg.JobID)
		w.publishProgress(jobMsg.JobID, "completed", "No orders to process", 100, 0, 0, 0)
		return nil
	}

	// Step 2: Detect duplicates for progress reporting
	duplicateInfo := w.detectDuplicates(mappings)
	uniqueOrderCount := len(duplicateInfo.OrderToIssues)

	log.Printf("Job %s: Found %d issues referencing %d unique production orders",
		jobMsg.JobID, len(mappings), uniqueOrderCount)

	// Step 3: Deduplicate for M3 execution (only process each order once)
	orders := w.deduplicateForExecution(mappings)

	w.publishProgress(jobMsg.JobID, "running",
		fmt.Sprintf("Found %d unique production orders from %d issues", uniqueOrderCount, len(mappings)),
		10, uniqueOrderCount, 0, 0)

	// Step 2: Get batch size from settings
	batchSize, err := w.getBatchSize(ctx, jobMsg.Environment)
	if err != nil {
		return fmt.Errorf("failed to get batch size: %w", err)
	}

	// Step 3: Split into batches and publish to NATS
	batches := w.createBatches(orders, batchSize)
	totalBatches := len(batches)

	log.Printf("Job %s: Split %d orders into %d batches", jobMsg.JobID, len(orders), totalBatches)

	// Subscribe to batch completions
	completionSubject := queue.GetBulkOpBatchCompleteSubject(jobMsg.JobID)
	completionChan := make(chan *BulkOpBatchCompletionMessage, totalBatches)

	sub, err := w.nats.Subscribe(completionSubject, func(msg *nats.Msg) {
		var completion BulkOpBatchCompletionMessage
		if err := json.Unmarshal(msg.Data, &completion); err != nil {
			log.Printf("Failed to unmarshal batch completion: %v", err)
			return
		}
		completionChan <- &completion
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to batch completions: %w", err)
	}
	defer sub.Unsubscribe()

	// Publish all batch jobs to NATS
	for i, batch := range batches {
		batchMsg := BulkOperationBatchMessage{
			JobID:            jobMsg.JobID,
			BatchNumber:      i + 1,
			Environment:      jobMsg.Environment,
			Operation:        jobMsg.Operation,
			ProductionOrders: batch,
			Params:           jobMsg.Params,
		}

		batchData, _ := json.Marshal(batchMsg)
		batchSubject := queue.GetBulkOpBatchSubject(jobMsg.Environment, jobMsg.JobID)

		if err := w.nats.Publish(batchSubject, batchData); err != nil {
			log.Printf("Failed to publish batch %d: %v", i+1, err)
		}
	}

	w.publishProgress(jobMsg.JobID, "running", fmt.Sprintf("Processing %d batches", totalBatches), 20, len(orders), 0, 0)

	// Step 4: Wait for all batch completions
	successfulItems := 0
	failedItems := 0
	completedBatches := 0

	timeout := time.After(30 * time.Minute) // 30 minute timeout

	for completedBatches < totalBatches {
		select {
		case completion := <-completionChan:
			completedBatches++
			successfulItems += completion.Successful
			failedItems += completion.Failed

			// Calculate progress: 20% to 90% during batch processing
			progress := 20 + (completedBatches * 70 / totalBatches)

			w.publishProgress(jobMsg.JobID, "running",
				fmt.Sprintf("Completed batch %d/%d", completedBatches, totalBatches),
				progress, len(orders), successfulItems, failedItems)

			log.Printf("Job %s: Batch %d/%d completed (success: %d, failed: %d)",
				jobMsg.JobID, completedBatches, totalBatches, completion.Successful, completion.Failed)

		case <-timeout:
			return fmt.Errorf("timeout waiting for batch completions")

		case <-ctx.Done():
			return fmt.Errorf("job cancelled")
		}
	}

	// Update final counts
	w.db.UpdateBulkOperationJobProgress(ctx, jobMsg.JobID, "Finalizing", len(orders), successfulItems, failedItems, 90)

	// Step 5: Expand order-level results to issue-level results
	w.publishProgress(jobMsg.JobID, "running", "Mapping results to issues", 95, len(orders), successfulItems, failedItems)

	if err := w.expandOrderResultsToIssues(ctx, jobMsg.JobID, mappings, duplicateInfo); err != nil {
		log.Printf("Warning: Failed to expand results to issues: %v", err)
		// Don't fail the job - batch results are still available
	}

	// Step 6: Complete job
	if err := w.db.CompleteBulkOperationJob(ctx, jobMsg.JobID); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	w.publishProgress(jobMsg.JobID, "completed", "Completed", 100, len(orders), successfulItems, failedItems)

	log.Printf("Job %s completed: %d successful, %d failed", jobMsg.JobID, successfulItems, failedItems)

	return nil
}

// handleBatchJob handles bulk operation batch processing
func (w *BulkOperationWorker) handleBatchJob(msg *nats.Msg) {
	var batchMsg BulkOperationBatchMessage
	if err := json.Unmarshal(msg.Data, &batchMsg); err != nil {
		log.Printf("Failed to unmarshal batch message: %v", err)
		return
	}

	log.Printf("Worker picked up batch job %s #%d (operation: %s, orders: %d)",
		batchMsg.JobID, batchMsg.BatchNumber, batchMsg.Operation, len(batchMsg.ProductionOrders))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Check if job is cancelled
	if w.isJobCancelled(ctx, batchMsg.JobID) {
		log.Printf("Batch %d of job %s skipped (job cancelled)", batchMsg.BatchNumber, batchMsg.JobID)
		return
	}

	// Apply rate limiting
	if err := w.rateLimiter.Wait(ctx, batchMsg.Environment); err != nil {
		log.Printf("Rate limiter error for batch %d: %v", batchMsg.BatchNumber, err)
		return
	}

	// Execute batch operation
	results, err := w.executeBatch(ctx, batchMsg)

	// Count successes and failures
	successful := 0
	failed := 0
	for _, result := range results {
		if result.Success {
			successful++
		} else {
			failed++
		}
	}

	// Store batch results in database
	if len(results) > 0 {
		if err := w.db.InsertBulkOperationBatchResults(ctx, batchMsg.JobID, batchMsg.BatchNumber, results); err != nil {
			log.Printf("Failed to store batch results: %v", err)
		}
	}

	// Publish completion message
	completion := BulkOpBatchCompletionMessage{
		JobID:       batchMsg.JobID,
		BatchNumber: batchMsg.BatchNumber,
		Successful:  successful,
		Failed:      failed,
	}
	if err != nil {
		completion.Error = err.Error()
	}

	completionData, _ := json.Marshal(completion)
	completionSubject := queue.GetBulkOpBatchCompleteSubject(batchMsg.JobID)
	w.nats.Publish(completionSubject, completionData)

	log.Printf("Batch %d of job %s completed: %d successful, %d failed",
		batchMsg.BatchNumber, batchMsg.JobID, successful, failed)
}

// executeBatch executes the M3 API operation for a batch
func (w *BulkOperationWorker) executeBatch(ctx context.Context, batchMsg BulkOperationBatchMessage) ([]db.BulkOperationBatchResult, error) {
	switch batchMsg.Operation {
	case "delete":
		return w.executeDeleteBatch(ctx, batchMsg)
	case "close":
		return w.executeCloseBatch(ctx, batchMsg)
	case "reschedule":
		return w.executeRescheduleBatch(ctx, batchMsg)
	default:
		return nil, fmt.Errorf("unknown operation: %s", batchMsg.Operation)
	}
}

// executeDeleteBatch executes delete operation for a batch
func (w *BulkOperationWorker) executeDeleteBatch(ctx context.Context, batchMsg BulkOperationBatchMessage) ([]db.BulkOperationBatchResult, error) {
	// Group by order type (MOP vs MO)
	mops := []ProductionOrderBatch{}
	mos := []ProductionOrderBatch{}

	for _, order := range batchMsg.ProductionOrders {
		if order.OrderType == "MOP" {
			mops = append(mops, order)
		} else if order.OrderType == "MO" {
			mos = append(mos, order)
		}
	}

	results := []db.BulkOperationBatchResult{}

	// Delete MOPs
	if len(mops) > 0 {
		mopRequests := []m3api.BulkRequestItem{}
		for _, mop := range mops {
			mopRequests = append(mopRequests, m3api.BulkRequestItem{
				Program:     "MMS100MI",
				Transaction: "DltPlannedMO",
				Record: map[string]string{
					"PLPN": mop.OrderNumber,
					"CONO": mop.Company,
				},
			})
		}

		response, err := w.m3Client.ExecuteBulk(ctx, mopRequests)
		if err != nil {
			// Mark all as failed
			for _, mop := range mops {
				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mop.ID,
					OrderNumber:       mop.OrderNumber,
					OrderType:         "MOP",
					Success:           false,
					ErrorMessage:      sql.NullString{String: err.Error(), Valid: true},
				})
			}
		} else {
			// Process results
			for i, mop := range mops {
				success := i < len(response.Results) && response.Results[i].IsSuccess()
				errorMsg := ""
				if !success && i < len(response.Results) {
					errorMsg = response.Results[i].ErrorMessage
				}

				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mop.ID,
					OrderNumber:       mop.OrderNumber,
					OrderType:         "MOP",
					Success:           success,
					ErrorMessage:      sql.NullString{String: errorMsg, Valid: errorMsg != ""},
				})
			}
		}
	}

	// Delete MOs
	if len(mos) > 0 {
		moRequests := []m3api.BulkRequestItem{}
		for _, mo := range mos {
			moRequests = append(moRequests, m3api.BulkRequestItem{
				Program:     "MMS002MI",
				Transaction: "DltManOrd",
				Record: map[string]string{
					"MFNO": mo.OrderNumber,
					"FACI": mo.Facility,
					"CONO": mo.Company,
				},
			})
		}

		response, err := w.m3Client.ExecuteBulk(ctx, moRequests)
		if err != nil {
			// Mark all as failed
			for _, mo := range mos {
				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mo.ID,
					OrderNumber:       mo.OrderNumber,
					OrderType:         "MO",
					Success:           false,
					ErrorMessage:      sql.NullString{String: err.Error(), Valid: true},
				})
			}
		} else {
			// Process results
			for i, mo := range mos {
				success := i < len(response.Results) && response.Results[i].IsSuccess()
				errorMsg := ""
				if !success && i < len(response.Results) {
					errorMsg = response.Results[i].ErrorMessage
				}

				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mo.ID,
					OrderNumber:       mo.OrderNumber,
					OrderType:         "MO",
					Success:           success,
					ErrorMessage:      sql.NullString{String: errorMsg, Valid: errorMsg != ""},
				})
			}
		}
	}

	return results, nil
}

// executeCloseBatch executes close operation for a batch (MOs only)
func (w *BulkOperationWorker) executeCloseBatch(ctx context.Context, batchMsg BulkOperationBatchMessage) ([]db.BulkOperationBatchResult, error) {
	requests := []m3api.BulkRequestItem{}

	for _, order := range batchMsg.ProductionOrders {
		if order.OrderType == "MO" {
			requests = append(requests, m3api.BulkRequestItem{
				Program:     "MOS100MI",
				Transaction: "Close",
				Record: map[string]string{
					"MFNO": order.OrderNumber,
					"FACI": order.Facility,
					"CONO": order.Company,
				},
			})
		}
	}

	results := []db.BulkOperationBatchResult{}

	if len(requests) == 0 {
		return results, nil
	}

	response, err := w.m3Client.ExecuteBulk(ctx, requests)
	if err != nil {
		// Mark all as failed
		for _, order := range batchMsg.ProductionOrders {
			results = append(results, db.BulkOperationBatchResult{
				JobID:             batchMsg.JobID,
				BatchNumber:       batchMsg.BatchNumber,
				ProductionOrderID: order.ID,
				OrderNumber:       order.OrderNumber,
				OrderType:         order.OrderType,
				Success:           false,
				ErrorMessage:      sql.NullString{String: err.Error(), Valid: true},
			})
		}
		return results, err
	}

	// Process results
	for i, order := range batchMsg.ProductionOrders {
		if order.OrderType != "MO" {
			continue
		}

		success := i < len(response.Results) && response.Results[i].IsSuccess()
		errorMsg := ""
		if !success && i < len(response.Results) {
			errorMsg = response.Results[i].ErrorMessage
		}

		results = append(results, db.BulkOperationBatchResult{
			JobID:             batchMsg.JobID,
			BatchNumber:       batchMsg.BatchNumber,
			ProductionOrderID: order.ID,
			OrderNumber:       order.OrderNumber,
			OrderType:         order.OrderType,
			Success:           success,
			ErrorMessage:      sql.NullString{String: errorMsg, Valid: errorMsg != ""},
		})
	}

	return results, nil
}

// executeRescheduleBatch executes reschedule operation for a batch
func (w *BulkOperationWorker) executeRescheduleBatch(ctx context.Context, batchMsg BulkOperationBatchMessage) ([]db.BulkOperationBatchResult, error) {
	// Extract new_date from params
	newDateStr, ok := batchMsg.Params["new_date"].(string)
	if !ok {
		return nil, fmt.Errorf("new_date parameter missing or invalid")
	}

	// Parse date (expected format: YYYYMMDD)
	_, err := strconv.Atoi(newDateStr)
	if err != nil || len(newDateStr) != 8 {
		return nil, fmt.Errorf("invalid date format: %s (expected YYYYMMDD)", newDateStr)
	}

	// Group by order type
	mops := []ProductionOrderBatch{}
	mos := []ProductionOrderBatch{}

	for _, order := range batchMsg.ProductionOrders {
		if order.OrderType == "MOP" {
			mops = append(mops, order)
		} else if order.OrderType == "MO" {
			mos = append(mos, order)
		}
	}

	results := []db.BulkOperationBatchResult{}

	// Reschedule MOPs
	if len(mops) > 0 {
		mopRequests := []m3api.BulkRequestItem{}
		for _, mop := range mops {
			mopRequests = append(mopRequests, m3api.BulkRequestItem{
				Program:     "MMS100MI",
				Transaction: "ChgPlannedMO",
				Record: map[string]string{
					"PLPN": mop.OrderNumber,
					"CONO": mop.Company,
					"PLDT": newDateStr, // Planned order date
					"RELD": newDateStr, // Release date
				},
			})
		}

		response, err := w.m3Client.ExecuteBulk(ctx, mopRequests)
		if err != nil {
			for _, mop := range mops {
				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mop.ID,
					OrderNumber:       mop.OrderNumber,
					OrderType:         "MOP",
					Success:           false,
					ErrorMessage:      sql.NullString{String: err.Error(), Valid: true},
				})
			}
		} else {
			for i, mop := range mops {
				success := i < len(response.Results) && response.Results[i].IsSuccess()
				errorMsg := ""
				if !success && i < len(response.Results) {
					errorMsg = response.Results[i].ErrorMessage
				}

				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mop.ID,
					OrderNumber:       mop.OrderNumber,
					OrderType:         "MOP",
					Success:           success,
					ErrorMessage:      sql.NullString{String: errorMsg, Valid: errorMsg != ""},
				})
			}
		}
	}

	// Reschedule MOs
	if len(mos) > 0 {
		moRequests := []m3api.BulkRequestItem{}
		for _, mo := range mos {
			moRequests = append(moRequests, m3api.BulkRequestItem{
				Program:     "MMS100MI",
				Transaction: "ChgMO",
				Record: map[string]string{
					"MFNO": mo.OrderNumber,
					"FACI": mo.Facility,
					"CONO": mo.Company,
					"STDT": newDateStr, // Start date
					"FIDT": newDateStr, // Finish date
				},
			})
		}

		response, err := w.m3Client.ExecuteBulk(ctx, moRequests)
		if err != nil {
			for _, mo := range mos {
				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mo.ID,
					OrderNumber:       mo.OrderNumber,
					OrderType:         "MO",
					Success:           false,
					ErrorMessage:      sql.NullString{String: err.Error(), Valid: true},
				})
			}
		} else {
			for i, mo := range mos {
				success := i < len(response.Results) && response.Results[i].IsSuccess()
				errorMsg := ""
				if !success && i < len(response.Results) {
					errorMsg = response.Results[i].ErrorMessage
				}

				results = append(results, db.BulkOperationBatchResult{
					JobID:             batchMsg.JobID,
					BatchNumber:       batchMsg.BatchNumber,
					ProductionOrderID: mo.ID,
					OrderNumber:       mo.OrderNumber,
					OrderType:         "MO",
					Success:           success,
					ErrorMessage:      sql.NullString{String: errorMsg, Valid: errorMsg != ""},
				})
			}
		}
	}

	return results, nil
}

// handleCancellation handles job cancellation requests
func (w *BulkOperationWorker) handleCancellation(msg *nats.Msg) {
	jobID := string(msg.Data)

	log.Printf("Cancellation requested for job %s", jobID)

	// Cancel the job context
	w.jobContextsMux.RLock()
	cancelFunc, exists := w.jobContexts[jobID]
	w.jobContextsMux.RUnlock()

	if exists && cancelFunc != nil {
		cancelFunc()
		log.Printf("Job %s cancelled", jobID)
	}
}

// Helper Methods

// fetchProductionOrders fetches all production orders for the given issue IDs
// fetchIssueToOrderMappings fetches ALL issue-to-order mappings (no DISTINCT!)
// This preserves the many-to-one relationship between issues and production orders
func (w *BulkOperationWorker) fetchIssueToOrderMappings(
	ctx context.Context,
	environment string,
	issueIDs []int64,
) ([]IssueToOrderMapping, error) {
	// Query WITHOUT DISTINCT to preserve all issue→order relationships
	query := `
		SELECT
			di.id AS issue_id,
			po.id AS production_order_id,
			po.order_number,
			po.order_type,
			po.faci AS facility,
			po.cono
		FROM detected_issues di
		JOIN production_orders po
			ON po.order_number = di.production_order_number
			AND po.order_type = di.production_order_type
			AND po.environment = di.environment
		WHERE di.id = ANY($1)
		  AND di.environment = $2
		ORDER BY po.id, di.id
	`

	rows, err := w.db.DB().QueryContext(ctx, query, pq.Array(issueIDs), environment)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue mappings: %w", err)
	}
	defer rows.Close()

	mappings := []IssueToOrderMapping{}
	for rows.Next() {
		var m IssueToOrderMapping
		var facility sql.NullString
		var company sql.NullString

		if err := rows.Scan(
			&m.IssueID,
			&m.ProductionOrderID,
			&m.OrderNumber,
			&m.OrderType,
			&facility,
			&company,
		); err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if facility.Valid {
			m.Facility = facility.String
		}
		if company.Valid {
			m.Company = company.String
		}

		mappings = append(mappings, m)
	}

	return mappings, rows.Err()
}

// detectDuplicates analyzes mappings to find duplicate order references
func (w *BulkOperationWorker) detectDuplicates(
	mappings []IssueToOrderMapping,
) DuplicateInfo {
	info := DuplicateInfo{
		OrderToIssues:  make(map[int64][]int64),
		PrimaryIssueID: make(map[int64]int64),
	}

	for _, m := range mappings {
		info.OrderToIssues[m.ProductionOrderID] = append(
			info.OrderToIssues[m.ProductionOrderID],
			m.IssueID,
		)

		// First issue for this order becomes the primary
		if _, exists := info.PrimaryIssueID[m.ProductionOrderID]; !exists {
			info.PrimaryIssueID[m.ProductionOrderID] = m.IssueID
		}
	}

	return info
}

// deduplicateForExecution creates unique production order batches for M3 API
func (w *BulkOperationWorker) deduplicateForExecution(
	mappings []IssueToOrderMapping,
) []ProductionOrderBatch {
	seen := make(map[int64]bool)
	orders := []ProductionOrderBatch{}

	for _, m := range mappings {
		if seen[m.ProductionOrderID] {
			continue
		}
		seen[m.ProductionOrderID] = true

		orders = append(orders, ProductionOrderBatch{
			ID:          m.ProductionOrderID,
			OrderNumber: m.OrderNumber,
			OrderType:   m.OrderType,
			Facility:    m.Facility,
			Company:     m.Company,
		})
	}

	return orders
}

// expandOrderResultsToIssues maps order-level M3 results to all affected issues
func (w *BulkOperationWorker) expandOrderResultsToIssues(
	ctx context.Context,
	jobID string,
	mappings []IssueToOrderMapping,
	duplicateInfo DuplicateInfo,
) error {
	// Fetch batch results (order-level results from M3 API)
	batchResults, err := w.db.GetBulkOperationJobResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to fetch batch results: %w", err)
	}

	// Build map: production_order_id -> batch result
	orderResults := make(map[int64]*db.BulkOperationBatchResult)
	for i := range batchResults {
		orderResults[batchResults[i].ProductionOrderID] = &batchResults[i]
	}

	// Create issue result for each mapping
	issueResults := []db.BulkOperationIssueResult{}
	for _, mapping := range mappings {
		batchResult := orderResults[mapping.ProductionOrderID]
		if batchResult == nil {
			// Order wasn't processed (shouldn't happen, but handle gracefully)
			issueResults = append(issueResults, db.BulkOperationIssueResult{
				JobID:             jobID,
				IssueID:           mapping.IssueID,
				ProductionOrderID: mapping.ProductionOrderID,
				OrderNumber:       mapping.OrderNumber,
				OrderType:         mapping.OrderType,
				Success:           false,
				ErrorMessage:      "Order was not processed",
				IsDuplicate:       false,
			})
			continue
		}

		// Check if this is a duplicate
		issueIDs := duplicateInfo.OrderToIssues[mapping.ProductionOrderID]
		isDuplicate := len(issueIDs) > 1
		var primaryIssueID *int64
		if isDuplicate {
			primary := duplicateInfo.PrimaryIssueID[mapping.ProductionOrderID]
			if mapping.IssueID != primary {
				primaryIssueID = &primary
			}
		}

		// Extract error message from nullable
		var errorMsg string
		if batchResult.ErrorMessage.Valid {
			errorMsg = batchResult.ErrorMessage.String
		}

		issueResults = append(issueResults, db.BulkOperationIssueResult{
			JobID:             jobID,
			IssueID:           mapping.IssueID,
			ProductionOrderID: mapping.ProductionOrderID,
			OrderNumber:       mapping.OrderNumber,
			OrderType:         mapping.OrderType,
			Success:           batchResult.Success,
			ErrorMessage:      errorMsg,
			IsDuplicate:       isDuplicate,
			PrimaryIssueID:    primaryIssueID,
		})
	}

	// Bulk insert all issue results
	if err := w.db.InsertBulkOperationIssueResults(ctx, issueResults); err != nil {
		return fmt.Errorf("failed to insert issue results: %w", err)
	}

	log.Printf("Job %s: Expanded %d order results to %d issue results",
		jobID, len(orderResults), len(issueResults))

	return nil
}

// getBatchSize retrieves the batch size setting from database
func (w *BulkOperationWorker) getBatchSize(ctx context.Context, environment string) (int, error) {
	settings, err := w.db.GetSystemSettings(ctx, environment)
	if err != nil {
		return 50, nil // Default to 50 if settings not found
	}

	for _, setting := range settings {
		if setting.SettingKey == "bulk_operation_batch_size" {
			if val, err := strconv.Atoi(setting.SettingValue); err == nil {
				return val, nil
			}
		}
	}

	return 50, nil // Default
}

// createBatches splits orders into batches
func (w *BulkOperationWorker) createBatches(orders []ProductionOrderBatch, batchSize int) [][]ProductionOrderBatch {
	batches := [][]ProductionOrderBatch{}

	for i := 0; i < len(orders); i += batchSize {
		end := i + batchSize
		if end > len(orders) {
			end = len(orders)
		}
		batches = append(batches, orders[i:end])
	}

	return batches
}

// publishProgress publishes progress update to NATS
func (w *BulkOperationWorker) publishProgress(jobID, status, phase string, progressPct, totalItems, successfulItems, failedItems int) {
	progress := BulkOpProgressMessage{
		JobID:              jobID,
		Status:             status,
		Phase:              phase,
		ProgressPercentage: progressPct,
		TotalItems:         totalItems,
		SuccessfulItems:    successfulItems,
		FailedItems:        failedItems,
	}

	data, _ := json.Marshal(progress)
	subject := queue.GetBulkOpProgressSubject(jobID)
	w.nats.Publish(subject, data)
}

// isJobCancelled checks if a job has been cancelled
func (w *BulkOperationWorker) isJobCancelled(ctx context.Context, jobID string) bool {
	job, err := w.db.GetBulkOperationJob(ctx, jobID)
	if err != nil {
		return false
	}
	return job.Status == "cancelled"
}

// registerJobContext registers a cancellation context for a job
func (w *BulkOperationWorker) registerJobContext(jobID string, cancel context.CancelFunc) {
	w.jobContextsMux.Lock()
	defer w.jobContextsMux.Unlock()
	w.jobContexts[jobID] = cancel
}

// unregisterJobContext unregisters a cancellation context for a job
func (w *BulkOperationWorker) unregisterJobContext(jobID string) {
	w.jobContextsMux.Lock()
	defer w.jobContextsMux.Unlock()
	delete(w.jobContexts, jobID)
}
