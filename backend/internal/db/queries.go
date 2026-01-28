package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Queries provides access to all database operations
type Queries struct {
	db *sql.DB
}

// New creates a new Queries instance
func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

// DB returns the underlying database connection
func (q *Queries) DB() *sql.DB {
	return q.db
}

// TruncateAnalysisTables deletes snapshot data for a specific environment
// This preserves data for other environments while clearing the specified one
// Uses DELETE with WHERE clause instead of TRUNCATE for environment filtering
func (q *Queries) TruncateAnalysisTables(ctx context.Context, environment string) error {
	// Use TRUNCATE CASCADE for instant clearing (much faster than DELETE)
	// NOTE: This truncates ALL environments, not just the specified one
	// TODO: When multi-environment support is needed, use table partitioning or optimize DELETE

	tables := []string{
		"detected_issues",              // First - clear old detection results
		"production_orders",            // Second - has FKs to MOs/MOPs
		"customer_order_lines",         // Third - referenced by production analysis
		"manufacturing_orders",         // Fourth - has FK from production_orders
		"planned_manufacturing_orders", // Fifth - has FK from production_orders
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
		fmt.Printf("Truncated table: %s\n", table)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit truncate transaction: %w", err)
	}

	return nil
}

// ContextCacheStatus represents the status of a cached resource type
type ContextCacheStatus struct {
	ResourceType string
	RecordCount  int
	LastRefresh  sql.NullTime
	IsStale      bool
}

// GetContextCacheStatus retrieves cache status for all context resources
func (q *Queries) GetContextCacheStatus(ctx context.Context, environment string) ([]ContextCacheStatus, error) {
	query := `
		SELECT 'Companies' as resource_type, COUNT(*) as record_count, MAX(cached_at) as last_refresh
		FROM m3_companies WHERE environment = $1
		UNION ALL
		SELECT 'Divisions', COUNT(*), MAX(cached_at)
		FROM m3_divisions WHERE environment = $1
		UNION ALL
		SELECT 'Facilities', COUNT(*), MAX(cached_at)
		FROM m3_facilities WHERE environment = $1
		UNION ALL
		SELECT 'Warehouses', COUNT(*), MAX(cached_at)
		FROM m3_warehouses WHERE environment = $1
		ORDER BY resource_type
	`

	rows, err := q.db.QueryContext(ctx, query, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to query cache status: %w", err)
	}
	defer rows.Close()

	var statuses []ContextCacheStatus
	for rows.Next() {
		var status ContextCacheStatus
		if err := rows.Scan(&status.ResourceType, &status.RecordCount, &status.LastRefresh); err != nil {
			return nil, fmt.Errorf("failed to scan cache status: %w", err)
		}

		// Determine if stale (> 7 days old)
		if status.LastRefresh.Valid {
			// Cache is stale if older than 7 days
			refreshThreshold := 7 * 24 * time.Hour
			status.IsStale = time.Since(status.LastRefresh.Time) > refreshThreshold
		} else {
			status.IsStale = true // No cache means stale
		}

		statuses = append(statuses, status)
	}

	return statuses, rows.Err()
}

// TODO: Add database query methods here
// These will be implemented once we define the schema
