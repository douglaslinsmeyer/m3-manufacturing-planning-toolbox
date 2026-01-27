package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// AuditService provides audit logging functionality
type AuditService struct {
	queries *db.Queries
}

// NewAuditService creates a new audit service
func NewAuditService(queries *db.Queries) *AuditService {
	return &AuditService{queries: queries}
}

// AuditParams contains all fields for an audit log entry
type AuditParams struct {
	// Required fields
	EntityType string
	Operation  string

	// Optional identification
	EntityID string
	UserID   string
	UserName string

	// Optional context
	Environment string
	Company     string
	Facility    string
	Warehouse   string

	// Flexible metadata
	Metadata map[string]interface{}

	// Optional HTTP context
	IPAddress string
	UserAgent string
}

// Log creates an audit log entry
func (s *AuditService) Log(ctx context.Context, params AuditParams) error {
	// Marshal metadata to JSONB
	var metadataJSON []byte
	var err error
	if params.Metadata != nil {
		metadataJSON, err = json.Marshal(params.Metadata)
		if err != nil {
			return err
		}
	}

	// Insert audit log record
	return s.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		Environment: sql.NullString{String: params.Environment, Valid: params.Environment != ""},
		EntityType:  params.EntityType,
		EntityID:    sql.NullString{String: params.EntityID, Valid: params.EntityID != ""},
		Operation:   params.Operation,
		UserID:      sql.NullString{String: params.UserID, Valid: params.UserID != ""},
		UserName:    sql.NullString{String: params.UserName, Valid: params.UserName != ""},
		Company:     sql.NullString{String: params.Company, Valid: params.Company != ""},
		Facility:    sql.NullString{String: params.Facility, Valid: params.Facility != ""},
		Warehouse:   sql.NullString{String: params.Warehouse, Valid: params.Warehouse != ""},
		Metadata:    metadataJSON,
		IPAddress:   sql.NullString{String: params.IPAddress, Valid: params.IPAddress != ""},
		UserAgent:   sql.NullString{String: params.UserAgent, Valid: params.UserAgent != ""},
	})
}

// QueryAuditLog retrieves audit logs with flexible filtering
func (s *AuditService) QueryAuditLog(
	ctx context.Context,
	entityType, operation, userID string,
	startTime, endTime time.Time,
	limit int,
) ([]db.AuditLog, error) {
	return s.queries.GetAuditLogs(ctx, db.GetAuditLogsParams{
		EntityType: sql.NullString{String: entityType, Valid: entityType != ""},
		Operation:  sql.NullString{String: operation, Valid: operation != ""},
		UserID:     sql.NullString{String: userID, Valid: userID != ""},
		StartTime:  sql.NullTime{Time: startTime, Valid: !startTime.IsZero()},
		EndTime:    sql.NullTime{Time: endTime, Valid: !endTime.IsZero()},
		Limit:      int32(limit),
	})
}
