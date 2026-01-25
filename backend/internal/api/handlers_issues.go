package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// handleListIssues lists detected issues with filtering
func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	detectorType := r.URL.Query().Get("detector_type")
	facility := r.URL.Query().Get("facility")
	warehouse := r.URL.Query().Get("warehouse")
	limitStr := r.URL.Query().Get("limit")
	includeIgnored := r.URL.Query().Get("include_ignored") == "true"

	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Use the new filtered query that supports multiple filters
	var issues []*db.DetectedIssue
	issues, err := s.db.GetIssuesFiltered(ctx, detectorType, facility, warehouse, includeIgnored, limit)
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
			"isIgnored":    issue.IsIgnored,
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

		if issue.MOTypeDescription.Valid {
			item["moTypeDescription"] = issue.MOTypeDescription.String
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

	// Parse query parameter
	includeIgnored := r.URL.Query().Get("include_ignored") == "true"

	summary, err := s.db.GetIssueSummary(ctx, includeIgnored)
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

// handleIgnoreIssue marks an issue as ignored
func (s *Server) handleIgnoreIssue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Optional: Parse request body for notes
	var requestBody struct {
		Notes string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&requestBody)

	// Get issue details from detected_issues
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Call queries.IgnoreIssue() with extracted identifiers
	err = s.db.IgnoreIssue(ctx, db.IgnoreIssueParams{
		Facility:              issue.Facility,
		DetectorType:          issue.DetectorType,
		IssueKey:              issue.IssueKey,
		ProductionOrderNumber: issue.ProductionOrderNumber.String,
		ProductionOrderType:   issue.ProductionOrderType.String,
		CONumber:              issue.CONumber.String,
		COLine:                issue.COLine.String,
		Notes:                 requestBody.Notes,
		// TODO: Add ignored_by from user context when auth is implemented
		IgnoredBy: "",
	})
	if err != nil {
		http.Error(w, "Failed to ignore issue", http.StatusInternalServerError)
		return
	}

	// Create audit log entry
	err = s.auditService.Log(ctx, services.AuditParams{
		EntityType: "issue",
		EntityID:   fmt.Sprintf("%d", issueID),
		Operation:  "ignore",
		// TODO: Add UserID and UserName from auth context
		Facility: issue.Facility,
		Metadata: map[string]interface{}{
			"detector_type":           issue.DetectorType,
			"production_order_number": issue.ProductionOrderNumber.String,
			"production_order_type":   issue.ProductionOrderType.String,
			"co_number":               issue.CONumber.String,
			"co_line":                 issue.COLine.String,
			"issue_key":               issue.IssueKey,
			"notes":                   requestBody.Notes,
		},
		IPAddress: getIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleUnignoreIssue removes an issue from ignored list
func (s *Server) handleUnignoreIssue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Get issue details from detected_issues
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Call queries.UnignoreIssue() with extracted identifiers
	err = s.db.UnignoreIssue(ctx, db.UnignoreIssueParams{
		Facility:              issue.Facility,
		DetectorType:          issue.DetectorType,
		IssueKey:              issue.IssueKey,
		ProductionOrderNumber: issue.ProductionOrderNumber.String,
	})
	if err != nil {
		http.Error(w, "Failed to unignore issue", http.StatusInternalServerError)
		return
	}

	// Create audit log entry
	err = s.auditService.Log(ctx, services.AuditParams{
		EntityType: "issue",
		EntityID:   fmt.Sprintf("%d", issueID),
		Operation:  "unignore",
		// TODO: Add UserID and UserName from auth context
		Facility: issue.Facility,
		Metadata: map[string]interface{}{
			"detector_type":           issue.DetectorType,
			"production_order_number": issue.ProductionOrderNumber.String,
			"issue_key":               issue.IssueKey,
		},
		IPAddress: getIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleDeletePlannedMO deletes a Manufacturing Order Proposal (MOP) via M3 API
func (s *Server) handleDeletePlannedMO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Get issue details to extract MOP number
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Verify this is a MOP
	if issue.ProductionOrderType.String != "MOP" {
		http.Error(w, "This operation is only valid for MOPs", http.StatusBadRequest)
		return
	}

	// Get M3 API client
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		http.Error(w, "Failed to get M3 API client", http.StatusInternalServerError)
		return
	}

	// Parse PLPN (Planned Order Number) from production order number
	plpn, err := strconv.ParseInt(issue.ProductionOrderNumber.String, 10, 64)
	if err != nil {
		http.Error(w, "Invalid planned order number format", http.StatusBadRequest)
		return
	}

	// Parse issue data JSON
	var issueData map[string]interface{}
	if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err != nil {
		log.Printf("Failed to parse issue data: %v", err)
	}

	// Call M3 API to delete the MOP
	params := map[string]string{
		"PLPN": fmt.Sprintf("%d", plpn),
	}

	// Add company if available from issue data
	if issueData != nil {
		if companyStr, ok := issueData["company"].(string); ok {
			params["CONO"] = companyStr
		}
	}

	// Execute M3 API call
	response, err := m3Client.Execute(ctx, "PMS170MI", "DelPlannedMO", params)
	if err != nil {
		log.Printf("Failed to delete MOP %s: %v", issue.ProductionOrderNumber.String, err)
		http.Error(w, fmt.Sprintf("Failed to delete MOP: %v", err), http.StatusInternalServerError)
		return
	}

	// Mark the MOP as deleted in our database
	err = s.db.MarkMOPAsDeletedRemotely(ctx, plpn, issue.Facility)
	if err != nil {
		log.Printf("Failed to mark MOP %d as deleted: %v", plpn, err)
		// Continue anyway - M3 delete succeeded
	}

	// Create audit log entry
	err = s.auditService.Log(ctx, services.AuditParams{
		EntityType: "issue",
		EntityID:   fmt.Sprintf("%d", issueID),
		Operation:  "delete_mop",
		Facility:   issue.Facility,
		Metadata: map[string]interface{}{
			"detector_type":           issue.DetectorType,
			"production_order_number": issue.ProductionOrderNumber.String,
			"production_order_type":   issue.ProductionOrderType.String,
			"m3_response":             response,
		},
		IPAddress: getIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	// Return success with M3 response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"m3_response": response,
	})
}

// getIPAddress extracts the client IP address from the request
func getIPAddress(r *http.Request) string {
	// Check for X-Forwarded-For header (proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
