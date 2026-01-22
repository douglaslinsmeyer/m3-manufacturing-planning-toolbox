package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// PlannedManufacturingOrder represents a planned manufacturing order record
type PlannedManufacturingOrder struct {
	ID            int64
	CONO          int
	DIVI          string
	MOPNumber     string
	PLPS          int
	Facility      string
	ItemNumber    string
	Status        string
	PSTS          string
	WHST          string
	PlannedQty    float64
	STDT          sql.NullInt32
	FIDT          sql.NullInt32
	PLDT          sql.NullInt32
	RORC          int
	RORN          string
	RORL          int
	RORX          int
	Messages      json.RawMessage
	Attributes    json.RawMessage
	LMDT          sql.NullInt32
	SyncTime      sql.NullTime
}

// BatchInsertPlannedOrders inserts multiple MOPs efficiently
func (q *Queries) BatchInsertPlannedOrders(ctx context.Context, orders []*PlannedManufacturingOrder) error {
	if len(orders) == 0 {
		return nil
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO planned_manufacturing_orders (
			cono, divi, mop_number, plps, facility, item_number,
			status, psts, whst,
			planned_quantity,
			stdt, fidt, pldt,
			rorc, rorn, rorl, rorx,
			messages, attributes, lmdt, sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10,
			$11, $12, $13,
			$14, $15, $16, $17,
			$18, $19, $20, NOW()
		)
		ON CONFLICT (mop_number)
		DO UPDATE SET
			status = EXCLUDED.status,
			psts = EXCLUDED.psts,
			whst = EXCLUDED.whst,
			planned_quantity = EXCLUDED.planned_quantity,
			stdt = EXCLUDED.stdt,
			fidt = EXCLUDED.fidt,
			pldt = EXCLUDED.pldt,
			messages = EXCLUDED.messages,
			attributes = EXCLUDED.attributes,
			lmdt = EXCLUDED.lmdt,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, mop := range orders {
		var stdt, fidt, pldt interface{}
		if mop.STDT.Valid && mop.STDT.Int32 != 0 {
			stdt = mop.STDT.Int32
		}
		if mop.FIDT.Valid && mop.FIDT.Int32 != 0 {
			fidt = mop.FIDT.Int32
		}
		if mop.PLDT.Valid && mop.PLDT.Int32 != 0 {
			pldt = mop.PLDT.Int32
		}

		_, err = stmt.ExecContext(ctx,
			mop.CONO, mop.DIVI, mop.MOPNumber, mop.PLPS, mop.Facility, mop.ItemNumber,
			mop.Status, mop.PSTS, mop.WHST,
			mop.PlannedQty,
			stdt, fidt, pldt,
			mop.RORC, mop.RORN, mop.RORL, mop.RORX,
			mop.Messages, mop.Attributes, mop.LMDT,
		)
		if err != nil {
			return fmt.Errorf("failed to insert MOP %s: %w", mop.MOPNumber, err)
		}
	}

	return tx.Commit()
}

// UpdateProductionOrdersFromMOPs updates the production_orders unified view from MOPs
func (q *Queries) UpdateProductionOrdersFromMOPs(ctx context.Context) error {
	query := `
		INSERT INTO production_orders (
			order_number, order_type,
			item_number, item_description,
			facility, warehouse,
			planned_start_date, planned_finish_date,
			ordered_quantity, status,
			mop_id, cono, divi,
			rorc, rorn, rorl, rorx,
			lmdt, sync_timestamp
		)
		SELECT DISTINCT ON (mop.mop_number)
			mop.mop_number,
			'MOP',
			mop.item_number,
			'',  -- Description comes from item master
			mop.facility,
			mop.warehouse,
			TO_DATE(NULLIF(mop.stdt, 0)::text, 'YYYYMMDD'),
			TO_DATE(NULLIF(mop.fidt, 0)::text, 'YYYYMMDD'),
			mop.planned_quantity,
			mop.status,
			mop.id,
			mop.cono,
			mop.divi,
			mop.rorc,
			mop.rorn,
			mop.rorl,
			mop.rorx,
			mop.lmdt,  -- lmdt is INTEGER (YYYYMMDD format), no conversion needed
			NOW()
		FROM planned_manufacturing_orders mop
		WHERE mop.stdt IS NOT NULL
		  AND mop.fidt IS NOT NULL
		ORDER BY mop.mop_number, mop.lmdt DESC NULLS LAST, mop.id DESC
		ON CONFLICT (order_number, order_type)
		DO UPDATE SET
			status = EXCLUDED.status,
			ordered_quantity = EXCLUDED.ordered_quantity,
			planned_start_date = EXCLUDED.planned_start_date,
			planned_finish_date = EXCLUDED.planned_finish_date,
			lmdt = EXCLUDED.lmdt,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`

	_, err := q.db.ExecContext(ctx, query)
	return err
}
