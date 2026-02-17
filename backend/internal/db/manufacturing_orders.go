package db

import (
	"context"
	"database/sql"
	"fmt"
)

// ManufacturingOrder represents a manufacturing order record - all M3 fields as strings
type ManufacturingOrder struct {
	ID             int64
	Environment    string // M3 environment (TRN or PRD)

	// M3 Core Identifiers
	CONO           string
	DIVI           string
	FACI           string
	MFNO           string
	PRNO           string
	ITNO           string

	// M3 Status Fields
	WHST           string
	WHHS           string
	WMST           string
	MOHS           string

	// M3 Quantities (strings from Data Fabric)
	ORQT           string
	MAQT           string
	ORQA           string
	RVQT           string
	RVQA           string
	MAQA           string

	// M3 Date Fields (strings YYYYMMDD)
	STDT           string
	FIDT           string
	MSTI           string
	MFTI           string
	FSTD           string
	FFID           string
	RSDT           string
	REFD           string
	RPDT           string

	// M3 Planning
	PRIO           string
	RESP           string
	PLGR           string
	WCLN           string
	PRDY           string

	// M3 Warehouse/Location
	WHLO           string
	WHSL           string
	BANO           string
	PendingPutawayQty string // Pending putaway quantity from MPTAWY.TRQT

	// M3 Reference Orders
	RORC           string
	RORN           string
	RORL           string
	RORX           string

	// M3 Hierarchy
	PRHL           string
	MFHL           string
	PRLO           string
	MFLO           string
	LEVL           string

	// M3 Configuration
	CFIN           string
	ATNR           string

	// M3 Order Type
	ORTY           string
	GETP           string

	// M3 Material/BOM
	BDCD           string
	SCEX           string
	STRT           string
	ECVE           string

	// M3 Routing
	AOID           string
	NUOP           string
	NUFO           string

	// M3 Action/Text
	ACTP           string
	TXT1           string
	TXT2           string

	// M3 Project
	PROJ           string
	ELNO           string

	// M3 Audit
	RGDT           string
	RGTM           string
	LMDT           string
	LMTS           string
	CHNO           string
	CHID           string

	// Metadata (strings from Data Fabric)
	M3Timestamp    string

	// CO Link (from MPREAL - all strings)
	LinkedCONumber string
	LinkedCOLine   string
	LinkedCOSuffix string
	AllocatedQty   string

	// MITMAS Item Master fields
	ItemType             string
	ItemDescription      string
	ItemGroup            string
	ProductGroup         string
	ProcurementGroup     string
	GroupTechnologyClass string

	SyncTime       sql.NullTime
}

// BatchInsertManufacturingOrders inserts multiple MOs efficiently with all M3 fields
func (q *Queries) BatchInsertManufacturingOrders(ctx context.Context, orders []*ManufacturingOrder, progressCallback InsertProgressCallback) error {
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
			environment,
			cono, divi, faci, mfno, prno, itno,
			whst, whhs, wmst, mohs,
			orqt, maqt, orqa, rvqt, rvqa, maqa,
			stdt, fidt, msti, mfti, fstd, ffid, rsdt, refd, rpdt,
			prio, resp, plgr, wcln, prdy,
			whlo, whsl, bano, pending_putaway_qty,
			rorc, rorn, rorl, rorx,
			prhl, mfhl, prlo, mflo, levl,
			cfin, atnr,
			orty, getp,
			bdcd, scex, strt, ecve,
			aoid, nuop, nufo,
			actp, txt1, txt2,
			proj, elno,
			rgdt, rgtm, lmdt, lmts, chno, chid,
			m3_timestamp,
			linked_co_number, linked_co_line, linked_co_suffix, allocated_qty,
			item_type, item_description, item_group, product_group, procurement_group, group_technology_class,
			sync_timestamp
		) VALUES (
			$1,
			$2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26,
			$27, $28, $29, $30, $31,
			$32, $33, $34, $35,
			$36, $37, $38, $39,
			$40, $41, $42, $43, $44,
			$45, $46,
			$47, $48,
			$49, $50, $51, $52,
			$53, $54, $55,
			$56, $57, $58,
			$59, $60,
			$61, $62, $63, $64, $65, $66,
			$67,
			$68, $69, $70, $71,
			$72, $73, $74, $75, $76, $77,
			NOW()
		)
		ON CONFLICT (environment, faci, mfno)
		DO UPDATE SET
			whst = EXCLUDED.whst,
			whhs = EXCLUDED.whhs,
			wmst = EXCLUDED.wmst,
			mohs = EXCLUDED.mohs,
			orqt = EXCLUDED.orqt,
			maqt = EXCLUDED.maqt,
			orqa = EXCLUDED.orqa,
			rvqt = EXCLUDED.rvqt,
			rvqa = EXCLUDED.rvqa,
			maqa = EXCLUDED.maqa,
			stdt = EXCLUDED.stdt,
			fidt = EXCLUDED.fidt,
			msti = EXCLUDED.msti,
			mfti = EXCLUDED.mfti,
			fstd = EXCLUDED.fstd,
			ffid = EXCLUDED.ffid,
			rsdt = EXCLUDED.rsdt,
			refd = EXCLUDED.refd,
			rpdt = EXCLUDED.rpdt,
			prio = EXCLUDED.prio,
			resp = EXCLUDED.resp,
			plgr = EXCLUDED.plgr,
			wcln = EXCLUDED.wcln,
			prdy = EXCLUDED.prdy,
			whlo = EXCLUDED.whlo,
			whsl = EXCLUDED.whsl,
			bano = EXCLUDED.bano,
			pending_putaway_qty = EXCLUDED.pending_putaway_qty,
			rorc = EXCLUDED.rorc,
			rorn = EXCLUDED.rorn,
			rorl = EXCLUDED.rorl,
			rorx = EXCLUDED.rorx,
			prhl = EXCLUDED.prhl,
			mfhl = EXCLUDED.mfhl,
			prlo = EXCLUDED.prlo,
			mflo = EXCLUDED.mflo,
			levl = EXCLUDED.levl,
			cfin = EXCLUDED.cfin,
			atnr = EXCLUDED.atnr,
			orty = EXCLUDED.orty,
			getp = EXCLUDED.getp,
			bdcd = EXCLUDED.bdcd,
			scex = EXCLUDED.scex,
			strt = EXCLUDED.strt,
			ecve = EXCLUDED.ecve,
			aoid = EXCLUDED.aoid,
			nuop = EXCLUDED.nuop,
			nufo = EXCLUDED.nufo,
			actp = EXCLUDED.actp,
			txt1 = EXCLUDED.txt1,
			txt2 = EXCLUDED.txt2,
			proj = EXCLUDED.proj,
			elno = EXCLUDED.elno,
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
			item_type = EXCLUDED.item_type,
			item_description = EXCLUDED.item_description,
			item_group = EXCLUDED.item_group,
			product_group = EXCLUDED.product_group,
			procurement_group = EXCLUDED.procurement_group,
			group_technology_class = EXCLUDED.group_technology_class,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	const insertProgressInterval = 10000
	for i, mo := range orders {
		_, err = stmt.ExecContext(ctx,
			mo.Environment,
			mo.CONO, mo.DIVI, mo.FACI, mo.MFNO, mo.PRNO, mo.ITNO,
			mo.WHST, mo.WHHS, mo.WMST, mo.MOHS,
			mo.ORQT, mo.MAQT, mo.ORQA, mo.RVQT, mo.RVQA, mo.MAQA,
			mo.STDT, mo.FIDT, mo.MSTI, mo.MFTI, mo.FSTD, mo.FFID, mo.RSDT, mo.REFD, mo.RPDT,
			mo.PRIO, mo.RESP, mo.PLGR, mo.WCLN, mo.PRDY,
			mo.WHLO, mo.WHSL, mo.BANO, mo.PendingPutawayQty,
			mo.RORC, mo.RORN, mo.RORL, mo.RORX,
			mo.PRHL, mo.MFHL, mo.PRLO, mo.MFLO, mo.LEVL,
			mo.CFIN, mo.ATNR,
			mo.ORTY, mo.GETP,
			mo.BDCD, mo.SCEX, mo.STRT, mo.ECVE,
			mo.AOID, mo.NUOP, mo.NUFO,
			mo.ACTP, mo.TXT1, mo.TXT2,
			mo.PROJ, mo.ELNO,
			mo.RGDT, mo.RGTM, mo.LMDT, mo.LMTS, mo.CHNO, mo.CHID,
			mo.M3Timestamp,
			mo.LinkedCONumber, mo.LinkedCOLine, mo.LinkedCOSuffix, mo.AllocatedQty,
			mo.ItemType, mo.ItemDescription, mo.ItemGroup, mo.ProductGroup, mo.ProcurementGroup, mo.GroupTechnologyClass,
		)
		if err != nil {
			return fmt.Errorf("failed to insert MO %s: %w", mo.MFNO, err)
		}

		// Report insertion progress every N records
		if progressCallback != nil && ((i+1)%insertProgressInterval == 0 || (i+1) == len(orders)) {
			progressCallback(i+1, len(orders))
		}
	}

	return tx.Commit()
}

// UpdateProductionOrdersFromMOs updates the production_orders unified view from MOs
func (q *Queries) UpdateProductionOrdersFromMOs(ctx context.Context) error {
	query := `
		INSERT INTO production_orders (
			environment, order_type, order_number,
			cono, divi, faci,
			prno, itno,
			ordered_quantity, manufactured_quantity,
			planned_start_date, planned_finish_date,
			actual_start_date, actual_finish_date,
			release_date, material_start_date, material_finish_date,
			status, proposal_status,
			priority, responsible, planner_group, production_line,
			warehouse, location, batch_number, pending_putaway_qty,
			rorc, rorn, rorl, rorx,
			config_number, attribute_number,
			project_number, element_number,
			lmdt, lmts,
			linked_co_number, linked_co_line, linked_co_suffix, allocated_qty,
			orty,
			mo_id, sync_timestamp, deleted_remotely
		)
		SELECT DISTINCT ON (mo.environment, mo.mfno)
			mo.environment,
			'MO',
			mo.mfno,
			mo.cono, mo.divi, mo.faci,
			mo.prno, mo.itno,
			mo.orqt, mo.maqt,
			mo.stdt, mo.fidt,
			mo.fstd, mo.ffid,
			'', mo.msti, mo.mfti,
			mo.whst, '',
			mo.prio, mo.resp, mo.plgr, mo.wcln,
			mo.whlo, mo.whsl, mo.bano, mo.pending_putaway_qty,
			mo.rorc, mo.rorn, mo.rorl, mo.rorx,
			mo.cfin, mo.atnr,
			mo.proj, mo.elno,
			mo.lmdt, mo.lmts,
			mo.linked_co_number, mo.linked_co_line, mo.linked_co_suffix, mo.allocated_qty,
			mo.orty,
			mo.id, NOW(), mo.deleted_remotely
		FROM manufacturing_orders mo
		ORDER BY mo.environment, mo.mfno,
		         CASE WHEN mo.lmdt = '' THEN '99999999' ELSE mo.lmdt END DESC,
		         mo.id DESC
		ON CONFLICT (environment, order_number, order_type)
		DO UPDATE SET
			status = EXCLUDED.status,
			ordered_quantity = EXCLUDED.ordered_quantity,
			manufactured_quantity = EXCLUDED.manufactured_quantity,
			planned_start_date = EXCLUDED.planned_start_date,
			planned_finish_date = EXCLUDED.planned_finish_date,
			actual_start_date = EXCLUDED.actual_start_date,
			actual_finish_date = EXCLUDED.actual_finish_date,
			material_start_date = EXCLUDED.material_start_date,
			material_finish_date = EXCLUDED.material_finish_date,
			priority = EXCLUDED.priority,
			responsible = EXCLUDED.responsible,
			planner_group = EXCLUDED.planner_group,
			production_line = EXCLUDED.production_line,
			warehouse = EXCLUDED.warehouse,
			location = EXCLUDED.location,
			batch_number = EXCLUDED.batch_number,
			pending_putaway_qty = EXCLUDED.pending_putaway_qty,
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
			orty = EXCLUDED.orty,
			deleted_remotely = EXCLUDED.deleted_remotely,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`

	_, err := q.db.ExecContext(ctx, query)
	return err
}

// MarkMOAsDeletedRemotely marks an MO as deleted/closed from M3
func (q *Queries) MarkMOAsDeletedRemotely(ctx context.Context, mfno string, facility string) error {
	query := `
		UPDATE manufacturing_orders
		SET deleted_remotely = true
		WHERE mfno = $1 AND faci = $2
	`
	_, err := q.db.ExecContext(ctx, query, mfno, facility)
	return err
}
