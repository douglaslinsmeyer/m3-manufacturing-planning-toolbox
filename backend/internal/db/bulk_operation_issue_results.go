package db

import (
	"context"
	"database/sql"
	"fmt"
)

// BulkOperationIssueResult represents a result for a single issue
type BulkOperationIssueResult struct {
	ID                int64  `json:"id"`
	JobID             string `json:"job_id"`
	IssueID           int64  `json:"issue_id"`
	ProductionOrderID int64  `json:"production_order_id"`
	OrderNumber       string `json:"order_number"`
	OrderType         string `json:"order_type"`
	Success           bool   `json:"success"`
	ErrorMessage      string `json:"error_message,omitempty"`
	IsDuplicate       bool   `json:"is_duplicate"`
	PrimaryIssueID    *int64 `json:"primary_issue_id,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// InsertBulkOperationIssueResults bulk inserts issue results
func (q *Queries) InsertBulkOperationIssueResults(
	ctx context.Context,
	results []BulkOperationIssueResult,
) error {
	if len(results) == 0 {
		return nil
	}

	query := `
		INSERT INTO bulk_operation_issue_results (
			job_id, issue_id, production_order_id, order_number, order_type,
			success, error_message, is_duplicate, primary_issue_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, result := range results {
		_, err := stmt.ExecContext(ctx,
			result.JobID,
			result.IssueID,
			result.ProductionOrderID,
			result.OrderNumber,
			result.OrderType,
			result.Success,
			result.ErrorMessage,
			result.IsDuplicate,
			result.PrimaryIssueID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert issue result: %w", err)
		}
	}

	return tx.Commit()
}

// GetBulkOperationIssueResults retrieves all issue results for a job
func (q *Queries) GetBulkOperationIssueResults(
	ctx context.Context,
	jobID string,
) ([]BulkOperationIssueResult, error) {
	query := `
		SELECT
			id, job_id, issue_id, production_order_id, order_number, order_type,
			success, error_message, is_duplicate, primary_issue_id, created_at
		FROM bulk_operation_issue_results
		WHERE job_id = $1
		ORDER BY created_at, issue_id
	`

	rows, err := q.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue results: %w", err)
	}
	defer rows.Close()

	results := []BulkOperationIssueResult{}
	for rows.Next() {
		var result BulkOperationIssueResult
		var errorMsg sql.NullString
		var primaryIssueID sql.NullInt64

		err := rows.Scan(
			&result.ID,
			&result.JobID,
			&result.IssueID,
			&result.ProductionOrderID,
			&result.OrderNumber,
			&result.OrderType,
			&result.Success,
			&errorMsg,
			&result.IsDuplicate,
			&primaryIssueID,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue result: %w", err)
		}

		if errorMsg.Valid {
			result.ErrorMessage = errorMsg.String
		}
		if primaryIssueID.Valid {
			result.PrimaryIssueID = &primaryIssueID.Int64
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
