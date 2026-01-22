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

// TruncateAnalysisTables truncates all M3 snapshot tables for full refresh
// This clears all data from production orders, MOs, MOPs, COs, and CO lines
// while preserving permanent data like jobs and metadata
func (q *Queries) TruncateAnalysisTables(ctx context.Context) error {
	// Order is critical to respect foreign key constraints
	// Only parent tables listed - CASCADE handles children
	tables := []string{
		"production_orders",            // First - references mo_id/mop_id
		"customer_orders",              // Second - cascades to CO lines, deliveries
		"manufacturing_orders",         // Third - cascades to mo_operations, mo_materials
		"planned_manufacturing_orders", // Fourth - no dependencies
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit truncation transaction: %w", err)
	}

	return nil
}

// TODO: Add database query methods here
// These will be implemented once we define the schema
