package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ========================================
// AUDIT LOG MODELS
// ========================================

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         int64           `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	UserID     sql.NullString  `json:"user_id,omitempty"`
	UserName   sql.NullString  `json:"user_name,omitempty"`
	EntityType string          `json:"entity_type"`
	EntityID   sql.NullString  `json:"entity_id,omitempty"`
	Operation  string          `json:"operation"`
	Company    sql.NullString  `json:"company,omitempty"`
	Facility   sql.NullString  `json:"facility,omitempty"`
	Warehouse  sql.NullString  `json:"warehouse,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IPAddress  sql.NullString  `json:"ip_address,omitempty"`
	UserAgent  sql.NullString  `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// CreateAuditLogParams contains parameters for creating an audit log
type CreateAuditLogParams struct {
	Environment sql.NullString
	EntityType  string
	EntityID    sql.NullString
	Operation   string
	UserID      sql.NullString
	UserName    sql.NullString
	Company     sql.NullString
	Facility    sql.NullString
	Warehouse   sql.NullString
	Metadata    json.RawMessage
	IPAddress   sql.NullString
	UserAgent   sql.NullString
}

// GetAuditLogsParams contains parameters for querying audit logs
type GetAuditLogsParams struct {
	Environment sql.NullString
	Facility    sql.NullString
	EntityType  sql.NullString
	Operation   sql.NullString
	UserID      sql.NullString
	StartTime   sql.NullTime
	EndTime     sql.NullTime
	Limit       int32
	Offset      int32
}

// ========================================
// IGNORED ISSUES MODELS
// ========================================

// IgnoreIssueParams contains parameters for ignoring an issue
type IgnoreIssueParams struct {
	Environment           string
	Facility              string
	DetectorType          string
	IssueKey              string
	ProductionOrderNumber string
	ProductionOrderType   string
	CONumber              string
	COLine                string
	Notes                 string
	IgnoredBy             string // User ID from auth context
}

// UnignoreIssueParams contains parameters for unignoring an issue
type UnignoreIssueParams struct {
	Environment           string
	Facility              string
	DetectorType          string
	IssueKey              string
	ProductionOrderNumber string
}

// CheckIgnoredParams contains parameters for checking if an issue is ignored
type CheckIgnoredParams struct {
	Environment           string
	Facility              string
	DetectorType          string
	IssueKey              string
	ProductionOrderNumber string
}
