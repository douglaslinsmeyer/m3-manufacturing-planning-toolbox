package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// IssuesListResponse wraps issue data with pagination metadata
type IssuesListResponse struct {
	Data       []map[string]interface{} `json:"data"`
	Pagination PaginationMeta           `json:"pagination"`
}

// PaginationMeta contains pagination information
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
}

// handleListIssues lists detected issues with filtering
func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	detectorType := r.URL.Query().Get("detector_type")
	facility := r.URL.Query().Get("facility")
	warehouse := r.URL.Query().Get("warehouse")
	includeIgnored := r.URL.Query().Get("include_ignored") == "true"

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
			// Validate page size is one of the allowed values
			switch parsedSize {
			case 25, 50, 100, 200:
				pageSize = parsedSize
			default:
				pageSize = 50 // default to 50 if invalid
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count for pagination metadata
	totalCount, err := s.db.GetIssuesFilteredCount(ctx, environment, detectorType, facility, warehouse, includeIgnored)
	if err != nil {
		http.Error(w, "Failed to count issues", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1 // At least 1 page even with 0 results
	}

	// Get filtered issues with pagination
	var issues []*db.DetectedIssue
	issues, err = s.db.GetIssuesFiltered(ctx, environment, detectorType, facility, warehouse, includeIgnored, pageSize, offset)
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

	// Wrap response with pagination metadata
	paginatedResponse := IssuesListResponse{
		Data: response,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedResponse)
}

// handleGetIssueSummary returns aggregated issue statistics
func (s *Server) handleGetIssueSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Parse query parameter
	includeIgnored := r.URL.Query().Get("include_ignored") == "true"

	summary, err := s.db.GetIssueSummary(ctx, environment, includeIgnored)
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

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

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
		Environment:           environment,
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
		Environment: environment,
		EntityType:  "issue",
		EntityID:    fmt.Sprintf("%d", issueID),
		Operation:   "ignore",
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

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

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
		Environment:           environment,
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
		Environment: environment,
		EntityType:  "issue",
		EntityID:    fmt.Sprintf("%d", issueID),
		Operation:   "unignore",
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

// handleDeleteMO deletes a Manufacturing Order (MO) via M3 API
func (s *Server) handleDeleteMO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Get issue details to extract MO number
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Verify this is an MO
	if issue.ProductionOrderType.String != "MO" {
		http.Error(w, "This operation is only valid for MOs", http.StatusBadRequest)
		return
	}

	// Parse issue data JSON
	var issueData map[string]interface{}
	if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err != nil {
		http.Error(w, "Failed to parse issue data", http.StatusInternalServerError)
		return
	}

	// Verify status is <= 22
	if statusStr, ok := issueData["status"].(string); ok {
		if status, err := strconv.Atoi(statusStr); err == nil && status > 22 {
			http.Error(w, "MO status is too advanced for deletion. Use Close instead.", http.StatusBadRequest)
			return
		}
	}

	// Get M3 API client
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		http.Error(w, "Failed to get M3 API client", http.StatusInternalServerError)
		return
	}

	// Call M3 API to delete the MO
	params := map[string]string{
		"MFNO": issue.ProductionOrderNumber.String,
	}

	// Add company if available from issue data
	if companyStr, ok := issueData["company"].(string); ok {
		params["CONO"] = companyStr
	}

	// Execute M3 API call
	response, err := m3Client.Execute(ctx, "PMS100MI", "DltMO", params)
	if err != nil {
		log.Printf("Failed to delete MO %s: %v", issue.ProductionOrderNumber.String, err)
		http.Error(w, fmt.Sprintf("Failed to delete MO: %v", err), http.StatusInternalServerError)
		return
	}

	// Mark the MO as deleted in our database
	err = s.db.MarkMOAsDeletedRemotely(ctx, issue.ProductionOrderNumber.String, issue.Facility)
	if err != nil {
		log.Printf("Failed to mark MO as deleted: %v", err)
		// Continue anyway - M3 delete succeeded
	}

	// Create audit log entry
	err = s.auditService.Log(ctx, services.AuditParams{
		EntityType: "issue",
		EntityID:   fmt.Sprintf("%d", issueID),
		Operation:  "delete_mo",
		Facility:   issue.Facility,
		Metadata: map[string]interface{}{
			"detector_type":           issue.DetectorType,
			"production_order_number": issue.ProductionOrderNumber.String,
			"production_order_type":   issue.ProductionOrderType.String,
			"status":                  issueData["status"],
			"m3_response":             response,
		},
		IPAddress: getIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Return success with M3 response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"m3_response": response,
	})
}

// handleCloseMO closes a Manufacturing Order (MO) via M3 API
func (s *Server) handleCloseMO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Get issue details
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Verify this is an MO
	if issue.ProductionOrderType.String != "MO" {
		http.Error(w, "This operation is only valid for MOs", http.StatusBadRequest)
		return
	}

	// Parse issue data JSON
	var issueData map[string]interface{}
	if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err != nil {
		http.Error(w, "Failed to parse issue data", http.StatusInternalServerError)
		return
	}

	// Verify status is > 22
	if statusStr, ok := issueData["status"].(string); ok {
		if status, err := strconv.Atoi(statusStr); err == nil && status <= 22 {
			http.Error(w, "MO status allows deletion. Use Delete instead.", http.StatusBadRequest)
			return
		}
	}

	// Get M3 API client
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		http.Error(w, "Failed to get M3 API client", http.StatusInternalServerError)
		return
	}

	// Call M3 API to close the MO
	params := map[string]string{
		"MFNO": issue.ProductionOrderNumber.String,
		"FACI": issue.Facility,
	}

	// Execute M3 API call
	response, err := m3Client.Execute(ctx, "PMS100MI", "CloseMO", params)
	if err != nil {
		log.Printf("Failed to close MO %s: %v", issue.ProductionOrderNumber.String, err)
		http.Error(w, fmt.Sprintf("Failed to close MO: %v", err), http.StatusInternalServerError)
		return
	}

	// Mark the MO as deleted (closed) in our database
	err = s.db.MarkMOAsDeletedRemotely(ctx, issue.ProductionOrderNumber.String, issue.Facility)
	if err != nil {
		log.Printf("Failed to mark MO as closed: %v", err)
		// Continue anyway - M3 close succeeded
	}

	// Create audit log entry
	err = s.auditService.Log(ctx, services.AuditParams{
		EntityType: "issue",
		EntityID:   fmt.Sprintf("%d", issueID),
		Operation:  "close_mo",
		Facility:   issue.Facility,
		Metadata: map[string]interface{}{
			"detector_type":           issue.DetectorType,
			"production_order_number": issue.ProductionOrderNumber.String,
			"production_order_type":   issue.ProductionOrderType.String,
			"status":                  issueData["status"],
			"m3_response":             response,
		},
		IPAddress: getIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Return success with M3 response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"m3_response": response,
	})
}

// getCurrentDateYYYYMMDD returns current date in YYYYMMDD format
func getCurrentDateYYYYMMDD() int {
	now := time.Now()
	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

// getNextBusinessDay returns next business day in YYYYMMDD format (skips weekends)
func getNextBusinessDay() int {
	now := time.Now()

	// Start with tomorrow
	next := now.AddDate(0, 0, 1)

	// Skip Saturday and Sunday
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.AddDate(0, 0, 1)
	}

	// Return in YYYYMMDD format
	return next.Year()*10000 + int(next.Month())*100 + next.Day()
}

// getAlignmentDate checks if date is in the past and adjusts to next business day if needed
func getAlignmentDate(minDate int) (alignmentDate int, wasAdjusted bool) {
	currentDate := getCurrentDateYYYYMMDD()

	if minDate < currentDate {
		// Earliest date is in the past, use next business day
		return getNextBusinessDay(), true
	}

	// Earliest date is today or future, use it as-is
	return minDate, false
}

// handleAlignEarliestMOs aligns all production orders in a JDCD group to the earliest date
func (s *Server) handleAlignEarliestMOs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse issue ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	issueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid issue ID", http.StatusBadRequest)
		return
	}

	// Get issue details
	issue, err := s.db.GetIssueByID(ctx, issueID)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Verify this is a joint delivery date mismatch issue
	if issue.DetectorType != "joint_delivery_date_mismatch" {
		http.Error(w, "This operation is only valid for joint delivery date mismatch issues", http.StatusBadRequest)
		return
	}

	// Parse issue data
	var issueData map[string]interface{}
	if err := json.Unmarshal([]byte(issue.IssueData), &issueData); err != nil {
		http.Error(w, "Failed to parse issue data", http.StatusInternalServerError)
		return
	}

	// Extract alignment target date and orders
	minDate := int(issueData["min_date"].(float64))
	orders, ok := issueData["orders"].([]interface{})
	if !ok || len(orders) == 0 {
		http.Error(w, "No orders found in issue data", http.StatusBadRequest)
		return
	}

	// Check if min_date is in the past and adjust to next business day if needed
	alignmentDate, dateAdjusted := getAlignmentDate(minDate)
	alignmentDateStr := fmt.Sprintf("%d", alignmentDate)

	if dateAdjusted {
		log.Printf("Earliest date %d is in the past, adjusted to next business day: %d", minDate, alignmentDate)
	}

	// Get M3 API client
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		http.Error(w, "Failed to get M3 API client", http.StatusInternalServerError)
		return
	}

	// Process each order
	alignedCount := 0
	skippedCount := 0
	failedCount := 0
	failures := []map[string]string{}

	for _, orderInterface := range orders {
		order, ok := orderInterface.(map[string]interface{})
		if !ok {
			continue
		}

		orderNumber, _ := order["number"].(string)
		orderType, _ := order["type"].(string)
		currentDate, _ := order["date"].(string)

		// Skip if already aligned to target date
		if currentDate == alignmentDateStr {
			skippedCount++
			log.Printf("Skipping %s %s - already aligned to %s", orderType, orderNumber, alignmentDateStr)
			continue
		}

		var alignErr error

		if orderType == "MO" {
			// Reschedule MO to alignment date (may be adjusted from min_date if past)
			alignErr = s.rescheduleMO(ctx, m3Client, orderNumber, issue.Facility, alignmentDateStr)
		} else if orderType == "MOP" {
			// Update MOP dates (maintaining duration)
			alignErr = s.updateMOPDates(ctx, m3Client, orderNumber, currentDate, alignmentDateStr)
		} else {
			alignErr = fmt.Errorf("unknown order type: %s", orderType)
		}

		if alignErr != nil {
			failedCount++
			failures = append(failures, map[string]string{
				"order": orderNumber,
				"type":  orderType,
				"error": alignErr.Error(),
			})
			log.Printf("Failed to align %s %s: %v", orderType, orderNumber, alignErr)
		} else {
			alignedCount++
			log.Printf("Successfully aligned %s %s to %s", orderType, orderNumber, alignmentDateStr)
		}
	}

	// Create audit log for successful alignments
	if alignedCount > 0 {
		jdcd, _ := issueData["jdcd"].(string)

		// Get environment from session
		session, _ := s.sessionStore.Get(r, "m3-session")
		environment, _ := session.Values["environment"].(string)

		err = s.auditService.Log(ctx, services.AuditParams{
			Environment: environment,
			EntityType:  "jdcd_group",
			EntityID:    jdcd,
			Operation:   "align_earliest",
			Facility:    issue.Facility,
			Metadata: map[string]interface{}{
				"aligned_count":     alignedCount,
				"failed_count":      failedCount,
				"skipped_count":     skippedCount,
				"target_date":       alignmentDateStr,
				"date_adjusted":     dateAdjusted,
				"original_min_date": minDate,
				"co_number":         issue.CONumber.String,
				"jdcd":              jdcd,
			},
			IPAddress: getIPAddress(r),
			UserAgent: r.Header.Get("User-Agent"),
		})
		if err != nil {
			log.Printf("Failed to create audit log: %v", err)
		}
	}

	// Return summary response with date adjustment info
	response := map[string]interface{}{
		"success":       failedCount == 0,
		"aligned_count": alignedCount,
		"skipped_count": skippedCount,
		"failed_count":  failedCount,
		"total_orders":  len(orders),
		"target_date":   alignmentDateStr,
		"date_adjusted": dateAdjusted,
		"failures":      failures,
	}

	if dateAdjusted {
		response["original_min_date"] = fmt.Sprintf("%d", minDate)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// rescheduleMO reschedules a manufacturing order to a new start date
func (s *Server) rescheduleMO(ctx context.Context, m3Client *m3api.Client, mfno, facility, newStartDate string) error {
	// Get MO details from production_orders view
	moQuery := `
		SELECT prno, ordered_quantity
		FROM production_orders
		WHERE order_number = $1 AND faci = $2 AND order_type = 'MO'
		LIMIT 1
	`
	var prno, orqa string
	err := s.db.DB().QueryRowContext(ctx, moQuery, mfno, facility).Scan(&prno, &orqa)
	if err != nil {
		return fmt.Errorf("failed to get MO details: %w", err)
	}

	// Call M3 API to reschedule
	params := map[string]string{
		"FACI": facility,
		"PRNO": prno,
		"MFNO": mfno,
		"ORQA": orqa,
		"WLDE": "0",           // Infinite/no bottlenecks
		"STDT": newStartDate,  // New aligned start date
		"DSP1": "1",           // Auto-approve: date earlier than today
		"DSP2": "1",           // Auto-approve: MO connected to order
		"DSP3": "1",           // Auto-approve: order contains subcontract
		"DSP4": "1",           // Auto-approve: quantity not divisible
	}

	log.Printf("Rescheduling MO %s to %s (FACI: %s, PRNO: %s, ORQA: %s)", mfno, newStartDate, facility, prno, orqa)

	response, err := m3Client.Execute(ctx, "PMS100MI", "Reschedule", params)
	if err != nil {
		return fmt.Errorf("M3 API error: %w", err)
	}

	// Log successful reschedule
	if response != nil {
		log.Printf("Successfully rescheduled MO %s to %s", mfno, newStartDate)
	}

	return nil
}

// updateMOPDates updates a MOP's start and finish dates (maintaining production duration)
func (s *Server) updateMOPDates(ctx context.Context, m3Client *m3api.Client, plpnStr, currentStartDate, newStartDate string) error {
	plpn, err := strconv.ParseInt(plpnStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid PLPN format: %w", err)
	}

	// Get MOP details for current finish date
	mopQuery := `
		SELECT stdt, fidt
		FROM planned_manufacturing_orders
		WHERE plpn = $1
		LIMIT 1
	`
	var currentStart, currentFinish string
	err = s.db.DB().QueryRowContext(ctx, mopQuery, plpn).Scan(&currentStart, &currentFinish)
	if err != nil {
		return fmt.Errorf("failed to get MOP details: %w", err)
	}

	// Validate dates
	if currentStart == "" || currentStart == "0" {
		currentStart = currentStartDate // Fall back to issue data
	}
	if currentFinish == "" || currentFinish == "0" {
		return fmt.Errorf("MOP %d has no finish date, cannot calculate new finish", plpn)
	}

	// Calculate new finish date based on desired start date and maintaining production duration
	startInt, err := strconv.Atoi(currentStart)
	if err != nil {
		return fmt.Errorf("invalid current start date: %w", err)
	}
	finishInt, err := strconv.Atoi(currentFinish)
	if err != nil {
		return fmt.Errorf("invalid current finish date: %w", err)
	}
	newStartInt, err := strconv.Atoi(newStartDate)
	if err != nil {
		return fmt.Errorf("invalid new start date: %w", err)
	}

	duration := finishInt - startInt
	if duration < 0 {
		log.Printf("Warning: MOP %d has negative duration (start %s > finish %s), using 0", plpn, currentStart, currentFinish)
		duration = 0
	}

	newFinishInt := newStartInt + duration
	newFinishDate := fmt.Sprintf("%d", newFinishInt)

	log.Printf("Updating MOP %d: finish date %s → %s (maintaining %d day duration for target start %s)",
		plpn, currentFinish, newFinishDate, duration, newStartDate)

	// Call M3 API to update MOP - NOTE: MOPs can only update finish date, not start date
	params := map[string]string{
		"PLPN": plpnStr,
		"FIDT": newFinishDate, // Only update finish date (calculated to align with desired start)
		"IGWA": "1",           // Ignore warnings
	}

	response, err := m3Client.Execute(ctx, "PMS170MI", "Updat", params)
	if err != nil {
		return fmt.Errorf("M3 API error: %w", err)
	}

	// Log successful update
	if response != nil {
		log.Printf("Successfully updated MOP %d finish date to %s", plpn, newFinishDate)
	}

	return nil
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
