package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// ManufacturingOrder represents a manufacturing order record
type ManufacturingOrder struct {
	ID             int64
	CONO           int
	DIVI           string
	Facility       string
	MONumber       string
	ProductNumber  string
	ItemNumber     string
	Status         string
	WHHS           string
	WMST           string
	MOHS           string
	OrderedQty     float64
	ManufacturedQty float64
	STDT           sql.NullInt32
	FIDT           sql.NullInt32
	RSDT           sql.NullInt32
	REFD           sql.NullInt32
	RORC           int
	RORN           string
	RORL           int
	RORX           int
	PRHL           string
	MFHL           string
	LEVL           int
	Attributes     json.RawMessage
	LMDT           sql.NullInt32
	SyncTime       sql.NullTime
}

// BatchInsertManufacturingOrders inserts multiple MOs efficiently
func (q *Queries) BatchInsertManufacturingOrders(ctx context.Context, orders []*ManufacturingOrder) error {
	if len(orders) == 0 {
		return nil
	}

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO manufacturing_orders (
			cono, divi, facility, mo_number, product_number, item_number,
			status, whhs, wmst, mohs,
			ordered_quantity, manufactured_quantity,
			stdt, fidt, rsdt, refd,
			rorc, rorn, rorl, rorx,
			prhl, mfhl, levl,
			attributes, lmdt, sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20,
			$21, $22, $23,
			$24, $25, NOW()
		)
		ON CONFLICT (facility, mo_number)
		DO UPDATE SET
			status = EXCLUDED.status,
			whhs = EXCLUDED.whhs,
			wmst = EXCLUDED.wmst,
			mohs = EXCLUDED.mohs,
			ordered_quantity = EXCLUDED.ordered_quantity,
			manufactured_quantity = EXCLUDED.manufactured_quantity,
			stdt = EXCLUDED.stdt,
			fidt = EXCLUDED.fidt,
			rsdt = EXCLUDED.rsdt,
			refd = EXCLUDED.refd,
			attributes = EXCLUDED.attributes,
			lmdt = EXCLUDED.lmdt,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, mo := range orders {
		var stdt, fidt, rsdt, refd interface{}
		if mo.STDT.Valid && mo.STDT.Int32 != 0 {
			stdt = mo.STDT.Int32
		}
		if mo.FIDT.Valid && mo.FIDT.Int32 != 0 {
			fidt = mo.FIDT.Int32
		}
		if mo.RSDT.Valid && mo.RSDT.Int32 != 0 {
			rsdt = mo.RSDT.Int32
		}
		if mo.REFD.Valid && mo.REFD.Int32 != 0 {
			refd = mo.REFD.Int32
		}

		_, err = stmt.ExecContext(ctx,
			mo.CONO, mo.DIVI, mo.Facility, mo.MONumber, mo.ProductNumber, mo.ItemNumber,
			mo.Status, mo.WHHS, mo.WMST, mo.MOHS,
			mo.OrderedQty, mo.ManufacturedQty,
			stdt, fidt, rsdt, refd,
			mo.RORC, mo.RORN, mo.RORL, mo.RORX,
			mo.PRHL, mo.MFHL, mo.LEVL,
			mo.Attributes, mo.LMDT,
		)
		if err != nil {
			return fmt.Errorf("failed to insert MO %s: %w", mo.MONumber, err)
		}
	}

	return tx.Commit()
}

// UpdateProductionOrdersFromMOs updates the production_orders unified view from MOs
func (q *Queries) UpdateProductionOrdersFromMOs(ctx context.Context) error {
	query := `
		INSERT INTO production_orders (
			order_number, order_type,
			item_number, item_description,
			facility, warehouse,
			planned_start_date, planned_finish_date,
			ordered_quantity, status,
			mo_id, cono, divi,
			rorc, rorn, rorl, rorx,
			lmdt, sync_timestamp
		)
		SELECT DISTINCT ON (mo.mo_number)
			mo.mo_number,
			'MO',
			mo.item_number,
			'',  -- Description comes from item master
			mo.facility,
			mo.warehouse,
			TO_DATE(NULLIF(mo.stdt, 0)::text, 'YYYYMMDD'),
			TO_DATE(NULLIF(mo.fidt, 0)::text, 'YYYYMMDD'),
			mo.ordered_quantity,
			mo.status,
			mo.id,
			mo.cono,
			mo.divi,
			mo.rorc,
			mo.rorn,
			mo.rorl,
			mo.rorx,
			mo.lmdt,  -- lmdt is INTEGER (YYYYMMDD format), no conversion needed
			NOW()
		FROM manufacturing_orders mo
		WHERE mo.stdt IS NOT NULL
		  AND mo.fidt IS NOT NULL
		ORDER BY mo.mo_number, mo.lmdt DESC NULLS LAST, mo.id DESC
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
