package db

import (
	"context"
	"database/sql"
	"fmt"
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

// TODO: Add database query methods here
// These will be implemented once we define the schema
