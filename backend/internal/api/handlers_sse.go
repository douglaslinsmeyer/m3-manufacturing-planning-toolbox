package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
)

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	JobID                     string  `json:"jobId"`
	Status                    string  `json:"status"`
	Progress                  int     `json:"progress"`
	CurrentStep               string  `json:"currentStep,omitempty"`
	CompletedSteps            int     `json:"completedSteps,omitempty"`
	TotalSteps                int     `json:"totalSteps,omitempty"`
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

// handleSnapshotProgressSSE streams real-time progress updates via Server-Sent Events
func (s *Server) handleSnapshotProgressSSE(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get ResponseController to extend write deadlines for long-lived SSE connections
	rc := http.NewResponseController(w)

	// Context for managing subscriptions
	ctx := r.Context()

	// Send initial connection event
	rc.SetWriteDeadline(time.Now().Add(30 * time.Second))
	fmt.Fprintf(w, "event: connected\ndata: {\"message\": \"Connected to progress stream\"}\n\n")
	flusher.Flush()

	// Get current job status from database and send it immediately
	job, err := s.db.GetRefreshJob(ctx, jobID)
	if err != nil {
		log.Printf("Failed to get job %s: %v", jobID, err)
	} else if job == nil {
		log.Printf("Job %s not found in database", jobID)
	} else {
		log.Printf("Sending initial status for job %s (status: %s, progress: %d%%)", jobID, job.Status, job.ProgressPct)
		initialUpdate := jobToProgressUpdate(job)
		sendSSEEvent(w, flusher, rc, "progress", initialUpdate)
	}

	// Subscribe to NATS progress topics
	progressSubject := queue.GetProgressSubject(jobID)
	completeSubject := queue.GetCompleteSubject(jobID)
	errorSubject := queue.GetErrorSubject(jobID)

	// Channel to receive messages from all subscriptions
	msgChan := make(chan *nats.Msg, 10)

	// Subscribe to progress updates
	progressSub, err := s.natsManager.Subscribe(progressSubject, func(msg *nats.Msg) {
		select {
		case msgChan <- msg:
		case <-ctx.Done():
			// Context cancelled, don't block
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to progress: %v", err)
		sendSSEEvent(w, flusher, rc, "error", map[string]string{"error": "Failed to subscribe to progress updates"})
		return
	}
	defer progressSub.Unsubscribe()

	// Subscribe to completion events
	completeSub, err := s.natsManager.Subscribe(completeSubject, func(msg *nats.Msg) {
		select {
		case msgChan <- msg:
		case <-ctx.Done():
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to completion: %v", err)
		sendSSEEvent(w, flusher, rc, "error", map[string]string{"error": "Failed to subscribe to completion events"})
		return
	}
	defer completeSub.Unsubscribe()

	// Subscribe to error events
	errorSub, err := s.natsManager.Subscribe(errorSubject, func(msg *nats.Msg) {
		select {
		case msgChan <- msg:
		case <-ctx.Done():
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to errors: %v", err)
		sendSSEEvent(w, flusher, rc, "error", map[string]string{"error": "Failed to subscribe to error events"})
		return
	}
	defer errorSub.Unsubscribe()

	// Heartbeat ticker - send every 5 seconds to stay well below 15s WriteTimeout
	heartbeat := time.NewTicker(5 * time.Second)
	defer heartbeat.Stop()

	log.Printf("SSE connection established for job %s", jobID)

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			log.Printf("SSE connection closed for job %s", jobID)
			return

		case msg := <-msgChan:
			// Parse and send progress update
			var update ProgressUpdate
			if err := json.Unmarshal(msg.Data, &update); err != nil {
				log.Printf("Failed to parse progress update: %v", err)
				continue
			}

			// Determine event type based on status
			eventType := "progress"
			if update.Status == "completed" {
				eventType = "complete"
			} else if update.Status == "failed" {
				eventType = "error"
			}

			sendSSEEvent(w, flusher, rc, eventType, update)

			// If job is completed or failed, close the connection after a brief delay
			if update.Status == "completed" || update.Status == "failed" {
				time.Sleep(500 * time.Millisecond)
				return
			}

		case <-heartbeat.C:
			// Send heartbeat to keep connection alive
			rc.SetWriteDeadline(time.Now().Add(30 * time.Second))
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// sendSSEEvent sends a Server-Sent Event and extends write deadline
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, rc *http.ResponseController, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal SSE data: %v", err)
		return
	}

	// Extend write deadline before writing to prevent timeout
	rc.SetWriteDeadline(time.Now().Add(30 * time.Second))

	log.Printf("Sending SSE event - type: %s, data: %s", eventType, string(jsonData))
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, jsonData)
	flusher.Flush()
}

// jobToProgressUpdate converts a RefreshJob to a ProgressUpdate
func jobToProgressUpdate(job *db.RefreshJob) ProgressUpdate {
	update := ProgressUpdate{
		JobID:            job.ID,
		Status:           job.Status,
		Progress:         job.ProgressPct,
		CompletedSteps:   job.CompletedSteps,
		TotalSteps:       job.TotalSteps,
		COLinesProcessed: job.COLinesProcessed,
		MOsProcessed:     job.MOsProcessed,
		MOPsProcessed:    job.MOPsProcessed,
	}

	if job.CurrentStep.Valid {
		update.CurrentStep = job.CurrentStep.String
	}
	if job.RecordsPerSecond.Valid {
		update.RecordsPerSecond = job.RecordsPerSecond.Float64
	}
	if job.EstimatedSecondsRemaining.Valid {
		update.EstimatedSecondsRemaining = int(job.EstimatedSecondsRemaining.Int32)
	}
	if job.CurrentOperation.Valid {
		update.CurrentOperation = job.CurrentOperation.String
	}
	if job.CurrentBatch.Valid {
		update.CurrentBatch = int(job.CurrentBatch.Int32)
	}
	if job.TotalBatches.Valid {
		update.TotalBatches = int(job.TotalBatches.Int32)
	}
	if job.ErrorMessage.Valid {
		update.Error = job.ErrorMessage.String
	}

	return update
}
