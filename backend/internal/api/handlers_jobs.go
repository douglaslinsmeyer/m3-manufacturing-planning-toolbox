package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

// handleGetBulkOperationJob returns the status of a bulk operation job
func (s *Server) handleGetBulkOperationJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Get job from database
	job, err := s.db.GetBulkOperationJob(ctx, jobID)
	if err != nil {
		log.Printf("Failed to get bulk operation job %s: %v", jobID, err)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Return job status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":              job.JobID,
		"environment":         job.Environment,
		"operation_type":      job.OperationType,
		"status":              job.Status,
		"total_items":         job.TotalItems,
		"successful_items":    job.SuccessfulItems,
		"failed_items":        job.FailedItems,
		"current_phase":       job.CurrentPhase.String,
		"progress_percentage": job.ProgressPercentage,
		"created_at":          job.CreatedAt,
		"started_at":          job.StartedAt.Time,
		"completed_at":        job.CompletedAt.Time,
		"error_message":       job.ErrorMessage.String,
	})
}

// handleGetBulkOperationJobProgress streams progress updates via SSE
func (s *Server) handleGetBulkOperationJobProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Subscribe to progress updates
	progressSubject := fmt.Sprintf("bulkop.progress.%s", jobID)
	sub, err := s.natsManager.Subscribe(progressSubject, func(msg *nats.Msg) {
		// Forward progress message to SSE stream
		fmt.Fprintf(w, "data: %s\n\n", string(msg.Data))
		flusher.Flush()
	})
	if err != nil {
		log.Printf("Failed to subscribe to progress updates for job %s: %v", jobID, err)
		http.Error(w, "Failed to subscribe to progress updates", http.StatusInternalServerError)
		return
	}
	defer sub.Unsubscribe()

	// Keep connection alive until client disconnects or job completes
	// We'll check job status periodically
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			log.Printf("Client disconnected from job %s progress stream", jobID)
			return

		case <-ticker.C:
			// Check job status
			job, err := s.db.GetBulkOperationJob(r.Context(), jobID)
			if err != nil {
				log.Printf("Failed to get job status for %s: %v", jobID, err)
				return
			}

			// If job is completed or failed, close the stream
			if job.Status == "completed" || job.Status == "failed" || job.Status == "cancelled" {
				log.Printf("Job %s finished with status: %s", jobID, job.Status)
				return
			}
		}
	}
}

// handleCancelBulkOperation cancels a running bulk operation job
func (s *Server) handleCancelBulkOperation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Update job status in database
	if err := s.db.CancelBulkOperationJob(ctx, jobID); err != nil {
		log.Printf("Failed to cancel job %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish cancellation to NATS
	cancelSubject := fmt.Sprintf("bulkop.cancel.%s", jobID)
	s.natsManager.Publish(cancelSubject, []byte(jobID))

	log.Printf("Job %s cancellation requested", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Job cancellation requested",
		"job_id":  jobID,
	})
}

// handleGetBulkOperationIssueResults retrieves issue-level results for a job
func (s *Server) handleGetBulkOperationIssueResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Verify job exists and user has access
	_, err := s.db.GetBulkOperationJob(ctx, jobID)
	if err != nil {
		log.Printf("Failed to get bulk operation job %s: %v", jobID, err)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// TODO: Add authorization check (verify user owns this job)
	// Use job variable once auth is implemented

	// Fetch issue results
	results, err := s.db.GetBulkOperationIssueResults(ctx, jobID)
	if err != nil {
		log.Printf("Failed to get issue results for job %s: %v", jobID, err)
		http.Error(w, "Failed to get issue results", http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":  jobID,
		"results": results,
		"total":   len(results),
	})
}

// handleListBulkOperationJobs lists bulk operation jobs for the current user
func (s *Server) handleListBulkOperationJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user context
	session, _ := s.sessionStore.Get(r, "m3-session")
	userID, _ := session.Values["user_id"].(string)
	environment, _ := session.Values["environment"].(string)

	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Parse limit from query parameters
	limit := 20 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get jobs from database
	jobs, err := s.db.ListBulkOperationJobs(ctx, environment, userID, limit)
	if err != nil {
		log.Printf("Failed to list bulk operation jobs: %v", err)
		http.Error(w, "Failed to list jobs", http.StatusInternalServerError)
		return
	}

	// Transform to API response
	jobsResponse := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		jobsResponse = append(jobsResponse, map[string]interface{}{
			"job_id":              job.JobID,
			"operation_type":      job.OperationType,
			"status":              job.Status,
			"total_items":         job.TotalItems,
			"successful_items":    job.SuccessfulItems,
			"failed_items":        job.FailedItems,
			"progress_percentage": job.ProgressPercentage,
			"created_at":          job.CreatedAt,
			"started_at":          job.StartedAt.Time,
			"completed_at":        job.CompletedAt.Time,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs": jobsResponse,
	})
}

// handleGetBulkOperationJobResults returns detailed results for a job
func (s *Server) handleGetBulkOperationJobResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Get job results from database
	results, err := s.db.GetBulkOperationJobResults(ctx, jobID)
	if err != nil {
		log.Printf("Failed to get results for job %s: %v", jobID, err)
		http.Error(w, "Failed to get job results", http.StatusInternalServerError)
		return
	}

	// Transform to API response
	resultsResponse := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		resultsResponse = append(resultsResponse, map[string]interface{}{
			"batch_number":        result.BatchNumber,
			"production_order_id": result.ProductionOrderID,
			"order_number":        result.OrderNumber,
			"order_type":          result.OrderType,
			"success":             result.Success,
			"error_message":       result.ErrorMessage.String,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":  jobID,
		"results": resultsResponse,
	})
}
