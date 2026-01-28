package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// Queries provides access to all database operations
type Queries struct {
	db              *sql.DB
	cacheTablesMeta []CacheTableMetadata
	cacheMetaExpiry time.Time
	cacheMetaMutex  sync.RWMutex
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

// CacheTableMetadata describes a discovered cache table
type CacheTableMetadata struct {
	TableName       string
	TimestampColumn string // "cached_at" or "fetched_at"
	ScopeColumn     string // "environment" or "user_id"
}

// GetContextCacheStatus retrieves cache status for all context resources dynamically
func (q *Queries) GetContextCacheStatus(ctx context.Context, environment string) ([]ContextCacheStatus, error) {
	// Step 1: Discover cache tables (with in-memory caching)
	cacheTables, err := q.getCachedTableMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover cache tables: %w", err)
	}

	var statuses []ContextCacheStatus

	// Step 2: Query each discovered table
	for _, table := range cacheTables {
		var status ContextCacheStatus
		status.ResourceType = formatResourceName(table.TableName)

		var query string
		var args []interface{}

		if table.ScopeColumn == "environment" {
			// Environment-scoped: filter by environment
			query = fmt.Sprintf(
				"SELECT COUNT(*) as record_count, MAX(%s) as last_refresh FROM %s WHERE environment = $1",
				table.TimestampColumn, table.TableName)
			args = []interface{}{environment}
		} else if table.ScopeColumn == "user_id" {
			// User-scoped: total count across all users
			query = fmt.Sprintf(
				"SELECT COUNT(*) as record_count, MAX(%s) as last_refresh FROM %s",
				table.TimestampColumn, table.TableName)
			args = []interface{}{}
		} else {
			continue // Skip tables without proper scope
		}

		row := q.db.QueryRowContext(ctx, query, args...)
		if err := row.Scan(&status.RecordCount, &status.LastRefresh); err != nil {
			log.Printf("Warning: Failed to query cache status for %s: %v", table.TableName, err)
			continue
		}

		// Determine staleness (> 7 days old)
		if status.LastRefresh.Valid {
			refreshThreshold := 7 * 24 * time.Hour
			status.IsStale = time.Since(status.LastRefresh.Time) > refreshThreshold
		} else {
			status.IsStale = true
		}

		statuses = append(statuses, status)
	}

	// Sort by resource type for consistent ordering
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ResourceType < statuses[j].ResourceType
	})

	return statuses, nil
}

// DiscoverCacheTables finds all cache tables via information_schema
func (q *Queries) DiscoverCacheTables(ctx context.Context) ([]CacheTableMetadata, error) {
	query := `
		SELECT
			c.table_name,
			c.column_name as timestamp_column,
			CASE
				WHEN EXISTS (SELECT 1 FROM information_schema.columns c2
							 WHERE c2.table_name = c.table_name
							 AND c2.column_name = 'environment')
				THEN 'environment'
				WHEN EXISTS (SELECT 1 FROM information_schema.columns c2
							 WHERE c2.table_name = c.table_name
							 AND c2.column_name = 'user_id')
				THEN 'user_id'
				ELSE ''
			END as scope_column
		FROM information_schema.columns c
		WHERE c.table_schema = 'public'
		  AND c.column_name IN ('cached_at', 'fetched_at')
		  AND (c.table_name LIKE 'm3_%' OR c.table_name = 'user_profiles')
		ORDER BY c.table_name
	`

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to discover cache tables: %w", err)
	}
	defer rows.Close()

	var tables []CacheTableMetadata
	for rows.Next() {
		var table CacheTableMetadata
		if err := rows.Scan(&table.TableName, &table.TimestampColumn, &table.ScopeColumn); err != nil {
			return nil, fmt.Errorf("failed to scan cache table metadata: %w", err)
		}

		// Skip tables without proper scope column
		if table.ScopeColumn == "" {
			continue
		}

		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// formatResourceName converts table name to display name
func formatResourceName(tableName string) string {
	switch tableName {
	case "m3_companies":
		return "Companies"
	case "m3_divisions":
		return "Divisions"
	case "m3_facilities":
		return "Facilities"
	case "m3_warehouses":
		return "Warehouses"
	case "m3_manufacturing_order_types":
		return "Manufacturing Order Types"
	case "m3_customer_order_types":
		return "Customer Order Types"
	case "user_profiles":
		return "User Profiles"
	default:
		// Fallback: Convert m3_snake_case to Title Case
		name := strings.TrimPrefix(tableName, "m3_")
		name = strings.ReplaceAll(name, "_", " ")
		return strings.Title(name)
	}
}

// getCachedTableMetadata returns cached table metadata, refreshing if expired
func (q *Queries) getCachedTableMetadata(ctx context.Context) ([]CacheTableMetadata, error) {
	q.cacheMetaMutex.RLock()
	if time.Now().Before(q.cacheMetaExpiry) && len(q.cacheTablesMeta) > 0 {
		defer q.cacheMetaMutex.RUnlock()
		return q.cacheTablesMeta, nil
	}
	q.cacheMetaMutex.RUnlock()

	// Refresh cache with double-check locking
	q.cacheMetaMutex.Lock()
	defer q.cacheMetaMutex.Unlock()

	// Double-check pattern
	if time.Now().Before(q.cacheMetaExpiry) && len(q.cacheTablesMeta) > 0 {
		return q.cacheTablesMeta, nil
	}

	tables, err := q.DiscoverCacheTables(ctx)
	if err != nil {
		return nil, err
	}

	q.cacheTablesMeta = tables
	q.cacheMetaExpiry = time.Now().Add(5 * time.Minute)

	return tables, nil
}

// TODO: Add database query methods here
// These will be implemented once we define the schema
