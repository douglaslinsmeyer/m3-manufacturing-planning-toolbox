package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// handleListIssues lists detected issues with filtering
func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	detectorType := r.URL.Query().Get("detector_type")
	facility := r.URL.Query().Get("facility")
	warehouse := r.URL.Query().Get("warehouse")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Use the new filtered query that supports multiple filters
	var issues []*db.DetectedIssue
	issues, err := s.db.GetIssuesFiltered(ctx, detectorType, facility, warehouse, limit)
	if err != nil {
		http.Error(w, "Failed to fetch issues", http.StatusInternalServerError)
		return
	}

	// Transform to API response
	response := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		item := map[string]interface{}{
			"id":           issue.ID,
			"detectorType": issue.DetectorType,
			"facility":     issue.Facility,
			"issueKey":     issue.IssueKey,
		}

		if issue.DetectedAt.Valid {
			item["detectedAt"] = issue.DetectedAt.Time
		}

		if issue.Warehouse.Valid {
			item["warehouse"] = issue.Warehouse.String
		}

		if issue.ProductionOrderNumber.Valid {
			item["productionOrderNumber"] = issue.ProductionOrderNumber.String
		}

		if issue.ProductionOrderType.Valid {
			item["productionOrderType"] = issue.ProductionOrderType.String
		}

		if issue.CONumber.Valid {
			item["coNumber"] = issue.CONumber.String
		}

		if issue.COLine.Valid {
			item["coLine"] = issue.COLine.String
		}

		if issue.COSuffix.Valid {
			item["coSuffix"] = issue.COSuffix.String
		}

		// Parse issue data JSON
		var issueData map[string]interface{}
		if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err == nil {
			item["issueData"] = issueData
		}

		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetIssueSummary returns aggregated issue statistics
func (s *Server) handleGetIssueSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary, err := s.db.GetIssueSummary(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch issue summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// handleGetIssueDetail gets a specific issue with full details
func (s *Server) handleGetIssueDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	issue, err := s.db.GetIssueByID(ctx, id)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Build detailed response
	response := map[string]interface{}{
		"id":           issue.ID,
		"detectorType": issue.DetectorType,
		"facility":     issue.Facility,
		"issueKey":     issue.IssueKey,
	}

	if issue.DetectedAt.Valid {
		response["detectedAt"] = issue.DetectedAt.Time
	}

	if issue.Warehouse.Valid {
		response["warehouse"] = issue.Warehouse.String
	}

	if issue.ProductionOrderNumber.Valid {
		response["productionOrderNumber"] = issue.ProductionOrderNumber.String
	}

	if issue.ProductionOrderType.Valid {
		response["productionOrderType"] = issue.ProductionOrderType.String
	}

	if issue.CONumber.Valid {
		response["coNumber"] = issue.CONumber.String
	}

	if issue.COLine.Valid {
		response["coLine"] = issue.COLine.String
	}

	if issue.COSuffix.Valid {
		response["coSuffix"] = issue.COSuffix.String
	}

	// Parse issue data JSON
	var issueData map[string]interface{}
	if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err == nil {
		response["issueData"] = issueData
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
