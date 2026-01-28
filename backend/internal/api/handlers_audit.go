package api

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// AuditLogListResponse wraps audit log data with pagination metadata
type AuditLogListResponse struct {
	Data       []map[string]interface{} `json:"data"`
	Pagination PaginationMeta           `json:"pagination"`
}

// handleListAuditLogs lists audit logs with filtering and pagination
func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get environment from session
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		http.Error(w, "Environment not set in session", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	entityType := r.URL.Query().Get("entity_type")
	operation := r.URL.Query().Get("operation")
	userID := r.URL.Query().Get("user_id")
	facility := r.URL.Query().Get("facility")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

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

	// Build query parameters
	params := db.GetAuditLogsParams{
		Environment: sql.NullString{String: environment, Valid: true},
		Limit:       int32(pageSize),
		Offset:      int32(offset),
	}

	if entityType != "" {
		params.EntityType = sql.NullString{String: entityType, Valid: true}
	}

	if operation != "" {
		params.Operation = sql.NullString{String: operation, Valid: true}
	}

	if userID != "" {
		params.UserID = sql.NullString{String: userID, Valid: true}
	}

	if facility != "" {
		params.Facility = sql.NullString{String: facility, Valid: true}
	}

	// Parse time filters (RFC3339 format)
	if startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			params.StartTime = sql.NullTime{Time: startTime, Valid: true}
		}
	}

	if endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			params.EndTime = sql.NullTime{Time: endTime, Valid: true}
		}
	}

	// Get total count for pagination metadata
	totalCount, err := s.db.GetAuditLogsCount(ctx, params)
	if err != nil {
		http.Error(w, "Failed to count audit logs", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1 // At least 1 page even with 0 results
	}

	// Get paginated audit logs
	logs, err := s.db.GetAuditLogs(ctx, params)
	if err != nil {
		http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	// Transform to API response
	response := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		item := map[string]interface{}{
			"id":         log.ID,
			"timestamp":  log.Timestamp,
			"entityType": log.EntityType,
			"operation":  log.Operation,
			"createdAt":  log.CreatedAt,
		}

		// Handle optional fields
		if log.UserID.Valid {
			item["userId"] = log.UserID.String
		}

		if log.UserName.Valid {
			item["userName"] = log.UserName.String
		}

		if log.EntityID.Valid {
			item["entityId"] = log.EntityID.String
		}

		if log.Company.Valid {
			item["company"] = log.Company.String
		}

		if log.Facility.Valid {
			item["facility"] = log.Facility.String
		}

		if log.Warehouse.Valid {
			item["warehouse"] = log.Warehouse.String
		}

		if log.IPAddress.Valid {
			item["ipAddress"] = log.IPAddress.String
		}

		if log.UserAgent.Valid {
			item["userAgent"] = log.UserAgent.String
		}

		// Parse metadata JSONB
		if log.Metadata != nil && len(log.Metadata) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(log.Metadata, &metadata); err == nil {
				item["metadata"] = metadata
			}
		}

		response = append(response, item)
	}

	// Return response with pagination
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuditLogListResponse{
		Data: response,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	})
}
