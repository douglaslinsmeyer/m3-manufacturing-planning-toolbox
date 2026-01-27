package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
	"github.com/pinggolf/m3-planning-tools/internal/workers"
)

// DetectorInfo represents a detector with its metadata
type DetectorInfo struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// TriggerDetectionRequest represents a request to trigger specific detectors
type TriggerDetectionRequest struct {
	Environment   string   `json:"environment"`   // "TRN" or "PRD"
	DetectorNames []string `json:"detectorNames"` // List of detector names to run
}

// TriggerDetectionResponse represents the response from triggering detection
type TriggerDetectionResponse struct {
	JobID       string   `json:"jobId"`
	Environment string   `json:"environment"`
	Detectors   []string `json:"detectors"`
	Status      string   `json:"status"`
	Message     string   `json:"message"`
}

// HandleListDetectors returns all available detectors with their enabled status
func HandleListDetectors(database *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environment := r.URL.Query().Get("environment")
		if environment == "" {
			environment = "TRN" // Default
		}

		ctx := r.Context()

		// Initialize detection service to get detector list
		detectorConfigService := services.NewDetectorConfigService(database)
		detectionService := services.NewDetectionService(database, detectorConfigService)

		// Get all registered detectors
		allDetectors := detectionService.GetAllDetectorNames()

		// Build response with enabled status
		detectorInfos := make([]DetectorInfo, 0, len(allDetectors))
		for _, name := range allDetectors {
			// Get detector instance for description
			detector := detectionService.GetDetectorByName(name)
			if detector == nil {
				continue
			}

			// Check if enabled
			enabled, err := detectionService.IsDetectorEnabled(ctx, environment, name)
			if err != nil {
				log.Printf("Failed to check enabled status for %s: %v", name, err)
				enabled = true // Default to enabled if check fails
			}

			detectorInfos = append(detectorInfos, DetectorInfo{
			Name:        detector.Name(),
			Label:       detector.Label(),
			Description: detector.Description(),
			Enabled:     enabled,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detectorInfos)
}
}

// HandleTriggerDetection triggers specific detectors without a full refresh
func HandleTriggerDetection(natsManager *queue.Manager, database *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TriggerDetectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		// Validate environment
		if req.Environment != "TRN" && req.Environment != "PRD" {
			http.Error(w, "environment must be TRN or PRD", http.StatusBadRequest)
			return
		}

		// Validate detector names
		if len(req.DetectorNames) == 0 {
			http.Error(w, "detectorNames cannot be empty", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Get latest completed refresh job for this environment
		latestRefreshJob, err := database.GetLatestRefreshJobByEnvironment(ctx, req.Environment)
		if err != nil {
			http.Error(w, fmt.Sprintf("No completed refresh job found for environment %s", req.Environment), http.StatusNotFound)
			return
		}

		// Get company and facility from latest refresh job data
		company, facility, err := database.GetRefreshJobContext(ctx, latestRefreshJob.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get refresh job context: %v", err), http.StatusInternalServerError)
			return
		}

		// Create new detection job ID (format: "det-{timestamp}" to stay under 36 char limit)
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		detectionJobID := fmt.Sprintf("det-%d", timestamp)

		log.Printf("Triggering detection job %s for environment %s with detectors: %v",
			detectionJobID, req.Environment, req.DetectorNames)

		// Initialize detection service
		detectorConfigService := services.NewDetectorConfigService(database)
		detectionService := services.NewDetectionService(database, detectorConfigService)

		// Validate all detector names exist
		for _, name := range req.DetectorNames {
			detector := detectionService.GetDetectorByName(name)
			if detector == nil {
				http.Error(w, fmt.Sprintf("Unknown detector: %s", name), http.StatusBadRequest)
				return
			}
		}

		// Create detection job record
		if err := database.CreateIssueDetectionJob(ctx, detectionJobID, req.Environment, len(req.DetectorNames)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create detection job: %v", err), http.StatusInternalServerError)
			return
		}

		// Clear previous issues for selected detectors only
		for _, detectorName := range req.DetectorNames {
			if err := database.ClearIssuesForDetector(ctx, latestRefreshJob.ID, detectorName); err != nil {
				log.Printf("Warning: failed to clear issues for detector %s: %v", detectorName, err)
			}
		}

		// Publish detector jobs to NATS
		for _, detectorName := range req.DetectorNames {
			job := workers.DetectorJobMessage{
				JobID:        fmt.Sprintf("%s-%s", detectionJobID, detectorName),
				ParentJobID:  detectionJobID,
				DetectorName: detectorName,
				Environment:  req.Environment,
				Company:      company,
				Facility:     facility,
			}

			data, _ := json.Marshal(job)
			subject := queue.GetDetectorSubject(req.Environment, detectorName)
			if err := natsManager.Publish(subject, data); err != nil {
				log.Printf("Failed to publish detector job %s: %v", detectorName, err)
				database.FailDetectionJob(ctx, detectionJobID, fmt.Sprintf("Failed to publish detector job: %v", err))
				http.Error(w, fmt.Sprintf("Failed to publish detector job: %v", err), http.StatusInternalServerError)
				return
			}

			log.Printf("Published detector job: %s to subject: %s", detectorName, subject)
		}

		// Return success response
		response := TriggerDetectionResponse{
			JobID:       detectionJobID,
			Environment: req.Environment,
			Detectors:   req.DetectorNames,
			Status:      "running",
			Message:     fmt.Sprintf("Triggered %d detector(s) for environment %s", len(req.DetectorNames), req.Environment),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
