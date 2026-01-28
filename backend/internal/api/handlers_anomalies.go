package api

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// handleListAnomalies lists detected anomalies from anomaly_alerts table with filtering
func (s *Server) handleListAnomalies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	severity := r.URL.Query().Get("severity")
	detectorType := r.URL.Query().Get("detector_type")

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage >= 1 {
			page = parsedPage
		}
	}

	pageSize := 50 // default
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if parsedSize, err := strconv.Atoi(pageSizeStr); err == nil {
			switch parsedSize {
			case 25, 50, 100, 200:
				pageSize = parsedSize
			default:
				pageSize = 50
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count for pagination metadata
	totalCount, err := s.db.GetAnomaliesFilteredCount(ctx, environment, severity, detectorType)
	if err != nil {
		http.Error(w, "Failed to count anomalies", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Get filtered anomalies with pagination
	anomalies, err := s.db.GetAnomaliesFiltered(ctx, environment, severity, detectorType, pageSize, offset)
	if err != nil {
		http.Error(w, "Failed to fetch anomalies", http.StatusInternalServerError)
		return
	}

	// Transform to API response
	response := make([]map[string]interface{}, 0, len(anomalies))
	for _, anomaly := range anomalies {
		// Parse JSONB metrics
		var metrics map[string]interface{}
		if err := json.Unmarshal([]byte(anomaly.Metrics), &metrics); err != nil {
			metrics = make(map[string]interface{})
		}

		anomalyMap := map[string]interface{}{
			"id":           anomaly.ID,
			"detectorType": anomaly.DetectorType,
			"severity":     anomaly.Severity,
			"status":       anomaly.Status,
			"metrics":      metrics,
		}

		// Handle nullable time fields
		if anomaly.DetectedAt.Valid {
			anomalyMap["detectedAt"] = anomaly.DetectedAt.Time
		}
		if anomaly.CreatedAt.Valid {
			anomalyMap["createdAt"] = anomaly.CreatedAt.Time
		}
		if anomaly.UpdatedAt.Valid {
			anomalyMap["updatedAt"] = anomaly.UpdatedAt.Time
		}

		// Extract warehouse from metrics if present
		if warehouse, ok := metrics["warehouse"].(string); ok {
			anomalyMap["warehouse"] = warehouse
		}

		if anomaly.EntityType.Valid {
			anomalyMap["entityType"] = anomaly.EntityType.String
		}
		if anomaly.EntityID.Valid {
			anomalyMap["entityId"] = anomaly.EntityID.String
		}
		if anomaly.Message.Valid {
			anomalyMap["message"] = anomaly.Message.String
		}
		if anomaly.AffectedCount.Valid {
			anomalyMap["affectedCount"] = anomaly.AffectedCount.Int32
		}
		if anomaly.ThresholdValue.Valid {
			anomalyMap["thresholdValue"] = anomaly.ThresholdValue.Float64
		}
		if anomaly.ActualValue.Valid {
			anomalyMap["actualValue"] = anomaly.ActualValue.Float64
		}
		if anomaly.AcknowledgedAt.Valid {
			anomalyMap["acknowledgedAt"] = anomaly.AcknowledgedAt.Time
		}
		if anomaly.AcknowledgedBy.Valid {
			anomalyMap["acknowledgedBy"] = anomaly.AcknowledgedBy.String
		}
		if anomaly.ResolvedAt.Valid {
			anomalyMap["resolvedAt"] = anomaly.ResolvedAt.Time
		}
		if anomaly.ResolvedBy.Valid {
			anomalyMap["resolvedBy"] = anomaly.ResolvedBy.String
		}
		if anomaly.Notes.Valid {
			anomalyMap["notes"] = anomaly.Notes.String
		}

		response = append(response, anomalyMap)
	}

	// Build response with pagination
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": response,
		"pagination": PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	})
}

// handleGetAnomalySummary returns aggregated anomaly statistics
func (s *Server) handleGetAnomalySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	summary, err := s.db.GetAnomalySummary(ctx, environment)
	if err != nil {
		http.Error(w, "Failed to fetch anomaly summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// handleGetAnomalyCount returns count of active anomalies
func (s *Server) handleGetAnomalyCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	count, err := s.db.GetActiveAnomalyCount(ctx, environment)
	if err != nil {
		http.Error(w, "Failed to fetch anomaly count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": count,
	})
}

// handleAcknowledgeAnomaly acknowledges an anomaly
func (s *Server) handleAcknowledgeAnomaly(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	// Get anomaly ID from URL
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid anomaly ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	username, _ := session.Values["username"].(string)
	if username == "" {
		username = "unknown"
	}

	// Acknowledge anomaly
	notes := sql.NullString{String: req.Notes, Valid: req.Notes != ""}
	if err := s.db.AcknowledgeAnomaly(ctx, id, username, notes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Anomaly acknowledged",
	})
}

// handleResolveAnomaly marks an anomaly as resolved
func (s *Server) handleResolveAnomaly(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	// Get anomaly ID from URL
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid anomaly ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	username, _ := session.Values["username"].(string)
	if username == "" {
		username = "unknown"
	}

	// Resolve anomaly
	notes := sql.NullString{String: req.Notes, Valid: req.Notes != ""}
	if err := s.db.ResolveAnomaly(ctx, id, username, notes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Anomaly resolved",
	})
}
