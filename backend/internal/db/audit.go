package db

import (
	"context"
	"fmt"
)

// CreateAuditLog inserts a new audit log entry
func (q *Queries) CreateAuditLog(ctx context.Context, params CreateAuditLogParams) error {
	query := `
		INSERT INTO audit_log (
			entity_type, entity_id, operation,
			user_id, user_name,
			company, facility, warehouse,
			metadata, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := q.db.ExecContext(ctx, query,
		params.EntityType,
		params.EntityID,
		params.Operation,
		params.UserID,
		params.UserName,
		params.Company,
		params.Facility,
		params.Warehouse,
		params.Metadata,
		params.IPAddress,
		params.UserAgent,
	)
	return err
}

// GetAuditLogs queries audit logs with filters
func (q *Queries) GetAuditLogs(ctx context.Context, params GetAuditLogsParams) ([]AuditLog, error) {
	query := `
		SELECT
			id, timestamp, user_id, user_name,
			entity_type, entity_id, operation,
			company, facility, warehouse,
			metadata, ip_address, user_agent, created_at
		FROM audit_log
		WHERE 1=1
	`

	var args []interface{}
	argNum := 1

	// Add filters dynamically
	if params.EntityType.Valid {
		query += fmt.Sprintf(" AND entity_type = $%d", argNum)
		args = append(args, params.EntityType.String)
		argNum++
	}

	if params.Operation.Valid {
		query += fmt.Sprintf(" AND operation = $%d", argNum)
		args = append(args, params.Operation.String)
		argNum++
	}

	if params.UserID.Valid {
		query += fmt.Sprintf(" AND user_id = $%d", argNum)
		args = append(args, params.UserID.String)
		argNum++
	}

	if params.StartTime.Valid {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, params.StartTime.Time)
		argNum++
	}

	if params.EndTime.Valid {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, params.EndTime.Time)
		argNum++
	}

	query += " ORDER BY timestamp DESC"

	if params.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, params.Limit)
	}

	// Execute query and return results
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.UserID, &log.UserName,
			&log.EntityType, &log.EntityID, &log.Operation,
			&log.Company, &log.Facility, &log.Warehouse,
			&log.Metadata, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetAuditLogsByEntity retrieves all audit entries for a specific entity
func (q *Queries) GetAuditLogsByEntity(ctx context.Context, entityType, entityID string, limit int) ([]AuditLog, error) {
	query := `
		SELECT
			id, timestamp, user_id, user_name,
			entity_type, entity_id, operation,
			company, facility, warehouse,
			metadata, ip_address, user_agent, created_at
		FROM audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	rows, err := q.db.QueryContext(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.UserID, &log.UserName,
			&log.EntityType, &log.EntityID, &log.Operation,
			&log.Company, &log.Facility, &log.Warehouse,
			&log.Metadata, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
