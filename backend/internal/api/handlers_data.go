package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Placeholder handlers for data endpoints
// These will be implemented once we have the full database schema and M3 client

// RefreshRequest represents a refresh request
type RefreshRequest struct {
	JobID       string `json:"jobId"`
	Environment string `json:"environment"`
	AccessToken string `json:"accessToken"`
	Company     string `json:"company"`
	Facility    string `json:"facility"`
}

// handleSnapshotRefresh initiates a data refresh from M3 via NATS
func (s *Server) handleSnapshotRefresh(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment and access token
	environment, _ := session.Values["environment"].(string)
	accessToken, err := s.authManager.GetAccessToken(session)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusUnauthorized)
		return
	}

	// Get effective context (respects temporary overrides)
	effectiveContext := s.contextService.GetEffectiveContext(session)

	// Validate that company and facility are set
	if effectiveContext.Company == "" {
		http.Error(w, "Company context is not set. Please select a company before refreshing data.", http.StatusBadRequest)
		return
	}
	if effectiveContext.Facility == "" {
		http.Error(w, "Facility context is not set. Please select a facility before refreshing data.", http.StatusBadRequest)
		return
	}

	// DEBUG: Print token and context for manual testing
	log.Printf("=== SNAPSHOT REFRESH REQUEST ===")
	log.Printf("Token: %s", accessToken)
	log.Printf("Environment: %s", environment)
	log.Printf("Company: %s", effectiveContext.Company)
	log.Printf("Facility: %s", effectiveContext.Facility)
	log.Printf("=======================================")

	// Generate job ID
	jobID := generateJobID()

	// Create job record in database
	ctx := r.Context()
	userID := session.Values["user_id"]
	if userID == nil {
		userID = "anonymous"
	}

	if err := s.db.CreateRefreshJob(ctx, jobID, environment, userID.(string)); err != nil {
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	// Publish refresh request to NATS
	refreshMsg := RefreshRequest{
		JobID:       jobID,
		Environment: environment,
		AccessToken: accessToken,
		Company:     effectiveContext.Company,
		Facility:    effectiveContext.Facility,
	}

	msgData, _ := json.Marshal(refreshMsg)
	subject := getRefreshSubject(environment)

	if err := s.natsManager.Publish(subject, msgData); err != nil {
		s.db.FailJob(ctx, jobID, "Failed to publish job to queue")
		http.Error(w, "Failed to queue refresh job", http.StatusInternalServerError)
		return
	}

	log.Printf("Snapshot refresh job %s queued for environment %s", jobID, environment)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "queued",
		"jobId":   jobID,
		"message": "Snapshot refresh job queued",
	})
}

// getRefreshSubject returns the NATS subject for refresh based on environment
func getRefreshSubject(environment string) string {
	if environment == "TRN" {
		return "snapshot.refresh.TRN"
	}
	return "snapshot.refresh.PRD"
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}

// handleCancelRefresh cancels a running snapshot refresh job
func (s *Server) handleCancelRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get job to verify it exists and is cancellable
	job, err := s.db.GetRefreshJob(ctx, jobID)
	if err != nil {
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}

	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Check if job is in a cancellable state
	if job.Status != "pending" && job.Status != "running" {
		http.Error(w, fmt.Sprintf("Job cannot be cancelled (status: %s)", job.Status), http.StatusBadRequest)
		return
	}

	// Mark job as failed with cancellation message
	if err := s.db.FailJob(ctx, jobID, "Cancelled by user"); err != nil {
		http.Error(w, "Failed to cancel job", http.StatusInternalServerError)
		return
	}

	log.Printf("Snapshot refresh job %s cancelled by user", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "cancelled",
		"jobId":   jobID,
		"message": "Snapshot refresh job cancelled",
	})
}

// handleSnapshotStatus returns the status of the current snapshot refresh
func (s *Server) handleSnapshotStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	ctx := r.Context()

	// Get latest job for this environment
	job, err := s.db.GetLatestRefreshJob(ctx, environment)
	if err != nil {
		http.Error(w, "Failed to get job status", http.StatusInternalServerError)
		return
	}

	// If no job exists, return idle status
	if job == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "idle",
			"progress": 0,
		})
		return
	}

	// Return job status
	response := map[string]interface{}{
		"jobId":             job.ID,
		"status":            job.Status,
		"progress":          job.ProgressPct,
		"completedSteps":    job.CompletedSteps,
		"totalSteps":        job.TotalSteps,
		"coLinesProcessed":  job.COLinesProcessed,
		"mosProcessed":      job.MOsProcessed,
		"mopsProcessed":     job.MOPsProcessed,
	}

	if job.CurrentStep.Valid {
		response["currentStep"] = job.CurrentStep.String
	}
	if job.StartedAt.Valid {
		response["startedAt"] = job.StartedAt.Time
	}
	if job.CompletedAt.Valid {
		response["completedAt"] = job.CompletedAt.Time
	}
	if job.DurationSeconds.Valid {
		response["durationSeconds"] = job.DurationSeconds.Int32
	}
	if job.ErrorMessage.Valid && job.ErrorMessage.String != "" {
		response["error"] = job.ErrorMessage.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetActiveJob returns the currently active refresh job if one exists
func (s *Server) handleGetActiveJob(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	ctx := r.Context()

	// Get active job for this environment
	job, err := s.db.GetActiveRefreshJob(ctx, environment)
	if err != nil {
		http.Error(w, "Failed to get active job", http.StatusInternalServerError)
		return
	}

	// If no active job exists, return null
	if job == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jobId": nil,
		})
		return
	}

	// Return job ID and basic status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobId":  job.ID,
		"status": job.Status,
	})
}

// handleSnapshotSummary returns summary statistics of the current snapshot
func (s *Server) handleSnapshotSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session first
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Get counts from database for this environment
	var productionOrdersCount, moCount, mopCount, coLinesCount int

	s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM production_orders WHERE environment = $1", environment).Scan(&productionOrdersCount)
	s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM manufacturing_orders WHERE environment = $1", environment).Scan(&moCount)
	s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM planned_manufacturing_orders WHERE environment = $1", environment).Scan(&mopCount)
	s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM customer_order_lines WHERE environment = $1", environment).Scan(&coLinesCount)

	// Get last refresh time from most recent completed job

	job, _ := s.db.GetLatestRefreshJob(ctx, environment)

	var lastRefresh interface{}
	if job != nil && job.CompletedAt.Valid {
		lastRefresh = job.CompletedAt.Time
	}

	// Get inconsistency count from latest detection job
	issueCount, err := s.db.GetIssueCountForLatestJob(ctx, environment)
	if err != nil {
		log.Printf("Warning: failed to get issue count: %v", err)
		issueCount = 0 // Graceful fallback
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalProductionOrders":    productionOrdersCount,
		"totalManufacturingOrders": moCount,
		"totalPlannedOrders":       mopCount,
		"totalCustomerOrderLines":  coLinesCount,
		"lastRefresh":              lastRefresh,
		"inconsistenciesCount":     issueCount,
	})
}

// handleListProductionOrders lists all production orders (unified MO/MOP view)
func (s *Server) handleListProductionOrders(w http.ResponseWriter, r *http.Request) {
	// TODO: Query production_orders table
	// Support filtering by date range, facility, status, etc.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// handleGetProductionOrder gets a single production order
func (s *Server) handleGetProductionOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Query production_orders table by ID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id": id,
	})
}

// handleGetManufacturingOrder gets full MO details
func (s *Server) handleGetManufacturingOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Query manufacturing_orders table with all details
	// Include operations, materials, actuals, etc.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id": id,
	})
}

// handleGetPlannedOrder gets full MOP details
func (s *Server) handleGetPlannedOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Query planned_manufacturing_orders table with all details
	// Include planning parameters, demand references, etc.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id": id,
	})
}

// handleListInconsistencies lists detected planning inconsistencies
func (s *Server) handleListInconsistencies(w http.ResponseWriter, r *http.Request) {
	// TODO: Run analysis logic to detect inconsistencies
	// Compare production order dates with CO delivery dates

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// handleGetTimeline returns timeline view of production orders
func (s *Server) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	// TODO: Generate timeline data for visualization

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": []interface{}{},
	})
}
