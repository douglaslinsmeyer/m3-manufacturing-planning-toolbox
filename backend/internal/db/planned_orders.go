package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// PlannedManufacturingOrder represents a planned manufacturing order record - all M3 fields as strings
type PlannedManufacturingOrder struct {
	ID              int64

	// M3 Core Identifiers
	CONO            string
	DIVI            string
	FACI            string
	PLPN            string
	PLPS            string
	PRNO            string
	ITNO            string

	// M3 Status
	PSTS            string
	WHST            string
	ACTP            string

	// M3 Order Type
	ORTY            string
	GETY            string

	// M3 Quantities (strings from Data Fabric)
	PPQT            string
	ORQA            string

	// M3 Dates (strings YYYYMMDD)
	RELD            string
	STDT            string
	FIDT            string
	MSTI            string
	MFTI            string
	PLDT            string

	// M3 Planning
	RESP            string
	PRIP            string
	PLGR            string
	WCLN            string
	PRDY            string

	// M3 Warehouse
	WHLO            string

	// M3 Reference Orders
	RORC            string
	RORN            string
	RORL            string
	RORX            string
	RORH            string

	// M3 Hierarchy
	PLLO            string
	PLHL            string

	// M3 Configuration
	ATNR            string
	CFIN            string

	// M3 Project
	PROJ            string
	ELNO            string

	// M3 Messages (JSONB)
	Messages        json.RawMessage

	// M3 Planning Parameters
	NUAU            string
	ORDP            string

	// M3 Audit
	RGDT            string
	RGTM            string
	LMDT            string
	LMTS            string
	CHNO            string
	CHID            string

	// Metadata (strings from Data Fabric)
	M3Timestamp     string

	// CO Link (all strings)
	LinkedCONumber  string
	LinkedCOLine    string
	LinkedCOSuffix  string
	AllocatedQty    string

	SyncTime        sql.NullTime
}

// BatchInsertPlannedOrders inserts multiple MOPs efficiently with all M3 fields
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
			cono, divi, faci, plpn, plps, prno, itno,
			psts, whst, actp,
			orty, gety,
			ppqt, orqa,
			reld, stdt, fidt, msti, mfti, pldt,
			resp, prip, plgr, wcln, prdy,
			whlo,
			rorc, rorn, rorl, rorx, rorh,
			pllo, plhl,
			atnr, cfin,
			proj, elno,
			messages,
			nuau, ordp,
			rgdt, rgtm, lmdt, lmts, chno, chid,
			m3_timestamp,
			linked_co_number, linked_co_line, linked_co_suffix, allocated_qty,
			sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12,
			$13, $14,
			$15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26,
			$27, $28, $29, $30, $31,
			$32, $33,
			$34, $35,
			$36, $37,
			$38,
			$39, $40,
			$41, $42, $43, $44, $45, $46,
			$47,
			$48, $49, $50, $51,
			NOW()
		)
		ON CONFLICT (plpn)
		DO UPDATE SET
			psts = EXCLUDED.psts,
			whst = EXCLUDED.whst,
			actp = EXCLUDED.actp,
			orty = EXCLUDED.orty,
			gety = EXCLUDED.gety,
			ppqt = EXCLUDED.ppqt,
			orqa = EXCLUDED.orqa,
			reld = EXCLUDED.reld,
			stdt = EXCLUDED.stdt,
			fidt = EXCLUDED.fidt,
			msti = EXCLUDED.msti,
			mfti = EXCLUDED.mfti,
			pldt = EXCLUDED.pldt,
			resp = EXCLUDED.resp,
			prip = EXCLUDED.prip,
			plgr = EXCLUDED.plgr,
			wcln = EXCLUDED.wcln,
			prdy = EXCLUDED.prdy,
			whlo = EXCLUDED.whlo,
			rorc = EXCLUDED.rorc,
			rorn = EXCLUDED.rorn,
			rorl = EXCLUDED.rorl,
			rorx = EXCLUDED.rorx,
			rorh = EXCLUDED.rorh,
			pllo = EXCLUDED.pllo,
			plhl = EXCLUDED.plhl,
			atnr = EXCLUDED.atnr,
			cfin = EXCLUDED.cfin,
			proj = EXCLUDED.proj,
			elno = EXCLUDED.elno,
			messages = EXCLUDED.messages,
			nuau = EXCLUDED.nuau,
			ordp = EXCLUDED.ordp,
			rgdt = EXCLUDED.rgdt,
			rgtm = EXCLUDED.rgtm,
			lmdt = EXCLUDED.lmdt,
			lmts = EXCLUDED.lmts,
			chno = EXCLUDED.chno,
			chid = EXCLUDED.chid,
			m3_timestamp = EXCLUDED.m3_timestamp,
			linked_co_number = EXCLUDED.linked_co_number,
			linked_co_line = EXCLUDED.linked_co_line,
			linked_co_suffix = EXCLUDED.linked_co_suffix,
			allocated_qty = EXCLUDED.allocated_qty,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, mop := range orders {
		_, err = stmt.ExecContext(ctx,
			mop.CONO, mop.DIVI, mop.FACI, mop.PLPN, mop.PLPS, mop.PRNO, mop.ITNO,
			mop.PSTS, mop.WHST, mop.ACTP,
			mop.ORTY, mop.GETY,
			mop.PPQT, mop.ORQA,
			mop.RELD, mop.STDT, mop.FIDT, mop.MSTI, mop.MFTI, mop.PLDT,
			mop.RESP, mop.PRIP, mop.PLGR, mop.WCLN, mop.PRDY,
			mop.WHLO,
			mop.RORC, mop.RORN, mop.RORL, mop.RORX, mop.RORH,
			mop.PLLO, mop.PLHL,
			mop.ATNR, mop.CFIN,
			mop.PROJ, mop.ELNO,
			mop.Messages,
			mop.NUAU, mop.ORDP,
			mop.RGDT, mop.RGTM, mop.LMDT, mop.LMTS, mop.CHNO, mop.CHID,
			mop.M3Timestamp,
			mop.LinkedCONumber, mop.LinkedCOLine, mop.LinkedCOSuffix, mop.AllocatedQty,
		)
		if err != nil {
			return fmt.Errorf("failed to insert MOP %s: %w", mop.PLPN, err)
		}
	}

	return tx.Commit()
}

// UpdateProductionOrdersFromMOPs updates the production_orders unified view from MOPs
func (q *Queries) UpdateProductionOrdersFromMOPs(ctx context.Context) error {
	query := `
		INSERT INTO production_orders (
			order_type, order_number,
			cono, divi, faci,
			prno, itno,
			ordered_quantity, manufactured_quantity,
			planned_start_date, planned_finish_date,
			actual_start_date, actual_finish_date,
			release_date, material_start_date, material_finish_date,
			status, proposal_status,
			priority, responsible, planner_group, production_line,
			warehouse, location, batch_number,
			rorc, rorn, rorl, rorx,
			config_number, attribute_number,
			project_number, element_number,
			lmdt, lmts,
			linked_co_number, linked_co_line, linked_co_suffix, allocated_qty,
			mop_id, sync_timestamp
		)
		SELECT DISTINCT ON (mop.plpn)
			'MOP',
			mop.plpn,
			mop.cono, mop.divi, mop.faci,
			mop.prno, mop.itno,
			mop.ppqt, '',
			mop.stdt, mop.fidt,
			'', '',
			mop.reld, mop.msti, mop.mfti,
			mop.whst, mop.psts,
			mop.prip, mop.resp, mop.plgr, mop.wcln,
			mop.whlo, '', '',
			mop.rorc, mop.rorn, mop.rorl, mop.rorx,
			mop.cfin, mop.atnr,
			mop.proj, mop.elno,
			mop.lmdt, mop.lmts,
			mop.linked_co_number, mop.linked_co_line, mop.linked_co_suffix, mop.allocated_qty,
			mop.id, NOW()
		FROM planned_manufacturing_orders mop
		ORDER BY mop.plpn,
		         CASE WHEN mop.lmdt = '' THEN '99999999' ELSE mop.lmdt END DESC,
		         mop.id DESC
		ON CONFLICT (order_number)
		DO UPDATE SET
			status = EXCLUDED.status,
			proposal_status = EXCLUDED.proposal_status,
			ordered_quantity = EXCLUDED.ordered_quantity,
			planned_start_date = EXCLUDED.planned_start_date,
			planned_finish_date = EXCLUDED.planned_finish_date,
			release_date = EXCLUDED.release_date,
			material_start_date = EXCLUDED.material_start_date,
			material_finish_date = EXCLUDED.material_finish_date,
			priority = EXCLUDED.priority,
			responsible = EXCLUDED.responsible,
			planner_group = EXCLUDED.planner_group,
			production_line = EXCLUDED.production_line,
			warehouse = EXCLUDED.warehouse,
			config_number = EXCLUDED.config_number,
			attribute_number = EXCLUDED.attribute_number,
			project_number = EXCLUDED.project_number,
			element_number = EXCLUDED.element_number,
			lmdt = EXCLUDED.lmdt,
			lmts = EXCLUDED.lmts,
			linked_co_number = EXCLUDED.linked_co_number,
			linked_co_line = EXCLUDED.linked_co_line,
			linked_co_suffix = EXCLUDED.linked_co_suffix,
			allocated_qty = EXCLUDED.allocated_qty,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`

	_, err := q.db.ExecContext(ctx, query)
	return err
}
