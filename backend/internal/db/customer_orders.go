package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// CustomerOrderLine represents a customer order line record
type CustomerOrderLine struct {
	ID           int64
	CONO         int
	DIVI         string
	OrderNumber  string
	LineNumber   string
	LineSuffix   string
	ItemNumber   string
	ItemDesc     string
	Status       string
	RORC         int
	RORN         string
	RORL         int
	RORX         int
	OrderedQty   float64
	DeliveredQty float64
	DWDT         sql.NullInt32
	CODT         sql.NullInt32
	PLDT         sql.NullInt32
	Attributes   json.RawMessage
	LMDT         sql.NullInt32
	SyncTime     sql.NullTime
}

// InsertCustomerOrderLine inserts a new customer order line
func (q *Queries) InsertCustomerOrderLine(ctx context.Context, line *CustomerOrderLine) error {
	query := `
		INSERT INTO customer_order_lines (
			cono, divi, order_number, line_number, line_suffix,
			item_number, item_description, status,
			rorc, rorn, rorl, rorx,
			ordered_quantity, delivered_quantity,
			dwdt, codt, pldt,
			attributes, lmdt, sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12,
			$13, $14,
			$15, $16, $17,
			$18, $19, NOW()
		)
		ON CONFLICT (order_number, line_number, line_suffix)
		DO UPDATE SET
			item_description = EXCLUDED.item_description,
			status = EXCLUDED.status,
			ordered_quantity = EXCLUDED.ordered_quantity,
			delivered_quantity = EXCLUDED.delivered_quantity,
			dwdt = EXCLUDED.dwdt,
			codt = EXCLUDED.codt,
			pldt = EXCLUDED.pldt,
			attributes = EXCLUDED.attributes,
			lmdt = EXCLUDED.lmdt,
			sync_timestamp = NOW(),
			updated_at = NOW()
		RETURNING id
	`

	var dwdt, codt, pldt interface{}
	if line.DWDT.Valid && line.DWDT.Int32 != 0 {
		dwdt = line.DWDT.Int32
	}
	if line.CODT.Valid && line.CODT.Int32 != 0 {
		codt = line.CODT.Int32
	}
	if line.PLDT.Valid && line.PLDT.Int32 != 0 {
		pldt = line.PLDT.Int32
	}

	err := q.db.QueryRowContext(ctx, query,
		line.CONO, line.DIVI, line.OrderNumber, line.LineNumber, line.LineSuffix,
		line.ItemNumber, line.ItemDesc, line.Status,
		line.RORC, line.RORN, line.RORL, line.RORX,
		line.OrderedQty, line.DeliveredQty,
		dwdt, codt, pldt,
		line.Attributes, line.LMDT,
	).Scan(&line.ID)

	if err != nil {
		return fmt.Errorf("failed to insert customer order line: %w", err)
	}

	return nil
}

// BatchInsertCustomerOrderLines inserts multiple CO lines efficiently
func (q *Queries) BatchInsertCustomerOrderLines(ctx context.Context, lines []*CustomerOrderLine) error {
	if len(lines) == 0 {
		return nil
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO customer_order_lines (
			cono, divi, order_number, line_number, line_suffix,
			item_number, item_description, status,
			rorc, rorn, rorl, rorx,
			ordered_quantity, delivered_quantity,
			dwdt, codt, pldt,
			attributes, lmdt, sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12,
			$13, $14,
			$15, $16, $17,
			$18, $19, NOW()
		)
		ON CONFLICT (order_number, line_number, line_suffix)
		DO UPDATE SET
			item_description = EXCLUDED.item_description,
			status = EXCLUDED.status,
			ordered_quantity = EXCLUDED.ordered_quantity,
			delivered_quantity = EXCLUDED.delivered_quantity,
			dwdt = EXCLUDED.dwdt,
			codt = EXCLUDED.codt,
			pldt = EXCLUDED.pldt,
			attributes = EXCLUDED.attributes,
			lmdt = EXCLUDED.lmdt,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, line := range lines {
		var dwdt, codt, pldt interface{}
		if line.DWDT.Valid && line.DWDT.Int32 != 0 {
			dwdt = line.DWDT.Int32
		}
		if line.CODT.Valid && line.CODT.Int32 != 0 {
			codt = line.CODT.Int32
		}
		if line.PLDT.Valid && line.PLDT.Int32 != 0 {
			pldt = line.PLDT.Int32
		}

		_, err = stmt.ExecContext(ctx,
			line.CONO, line.DIVI, line.OrderNumber, line.LineNumber, line.LineSuffix,
			line.ItemNumber, line.ItemDesc, line.Status,
			line.RORC, line.RORN, line.RORL, line.RORX,
			line.OrderedQty, line.DeliveredQty,
			dwdt, codt, pldt,
			line.Attributes, line.LMDT,
		)
		if err != nil {
			return fmt.Errorf("failed to insert line %s-%s: %w", line.OrderNumber, line.LineNumber, err)
		}
	}

	return tx.Commit()
}

// GetLastSyncDate gets the last LMDT value for incremental loading
// DEPRECATED: System now uses full refresh with table truncation.
// Kept for backward compatibility or potential future incremental mode.
func (q *Queries) GetLastSyncDate(ctx context.Context, tableName string) (int, error) {
	var lmdt sql.NullInt32

	query := ""
	switch tableName {
	case "customer_order_lines":
		query = "SELECT MAX(lmdt) FROM customer_order_lines WHERE lmdt IS NOT NULL"
	case "manufacturing_orders":
		query = "SELECT MAX(lmdt) FROM manufacturing_orders WHERE lmdt IS NOT NULL"
	case "planned_manufacturing_orders":
		query = "SELECT MAX(lmdt) FROM planned_manufacturing_orders WHERE lmdt IS NOT NULL"
	default:
		return 0, fmt.Errorf("unknown table: %s", tableName)
	}

	err := q.db.QueryRowContext(ctx, query).Scan(&lmdt)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if lmdt.Valid {
		return int(lmdt.Int32), nil
	}

	// No previous sync, return a far past date for full load
	return 20200101, nil // January 1, 2020
}
