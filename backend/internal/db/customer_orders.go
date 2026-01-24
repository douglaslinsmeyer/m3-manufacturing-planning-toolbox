package db

import (
	"context"
	"database/sql"
	"fmt"
)

// CustomerOrderLine represents a customer order line - all M3 fields as strings
type CustomerOrderLine struct {
	ID   int64
	COId sql.NullInt64

	// M3 Core Identifiers
	CONO, DIVI, ORNO, PONR, POSX string

	// M3 Item Information
	ITNO, ITDS, TEDS, REPI string

	// M3 Status/Type
	ORST, ORTY string

	// M3 Facility/Warehouse
	FACI, WHLO string

	// M3 Quantities - Basic U/M
	ORQT, RNQT, ALQT, DLQT, IVQT string

	// M3 Quantities - Alternate U/M
	ORQA, RNQA, ALQA, DLQA, IVQA string

	// M3 Units
	ALUN, COFA, SPUN string

	// M3 Delivery Dates
	DWDT, DWHM, CODT, COHM, PLDT, FDED, LDED string

	// M3 Pricing
	SAPR, NEPR, LNAM, CUCD string

	// M3 Discounts
	DIP1, DIP2, DIP3, DIP4, DIP5, DIP6 string
	DIA1, DIA2, DIA3, DIA4, DIA5, DIA6 string

	// M3 Reference Orders
	RORC, RORN, RORL, RORX string

	// M3 Customer References
	CUNO, CUOR, CUPO, CUSX string

	// M3 Product/Model
	PRNO, HDPR, POPN, ALWT, ALWQ string

	// M3 Delivery/Route
	ADID, ROUT, RODN, DSDT, DSHM, MODL, TEDL, TEL2 string

	// M3 Packaging
	TEPA, PACT, CUPA string

	// M3 Partner/EDI
	E0PA, DSGP, PUSN, PUTP string

	// M3 Attributes (ATV1-ATV0)
	ATV1, ATV2, ATV3, ATV4, ATV5 string
	ATV6, ATV7, ATV8, ATV9, ATV0 string

	// M3 User-Defined Alpha (UCA1-UCA0)
	UCA1, UCA2, UCA3, UCA4, UCA5 string
	UCA6, UCA7, UCA8, UCA9, UCA0 string

	// M3 User-Defined Numeric (UDN1-UDN6)
	UDN1, UDN2, UDN3, UDN4, UDN5, UDN6 string

	// M3 User-Defined Date (UID1-UID3)
	UID1, UID2, UID3 string

	// M3 User-Defined Text
	UCT1 string

	// M3 Configuration
	ATNR, ATMO, ATPR, CFIN string

	// M3 Project
	PROJ, ELNO string

	// M3 Audit
	RGDT, RGTM, LMDT, CHNO, CHID, LMTS string

	// Metadata
	M3Timestamp string
	SyncTime    sql.NullTime
}

// Note: InsertCustomerOrderLine removed - use BatchInsertCustomerOrderLines instead

// BatchInsertCustomerOrderLines inserts multiple CO lines with all M3 fields as strings
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
			cono, divi, orno, ponr, posx,
			itno, itds, teds, repi,
			orst, orty,
			faci, whlo,
			orqt, rnqt, alqt, dlqt, ivqt,
			orqa, rnqa, alqa, dlqa, ivqa,
			alun, cofa, spun,
			dwdt, dwhm, codt, cohm, pldt, fded, lded,
			sapr, nepr, lnam, cucd,
			dip1, dip2, dip3, dip4, dip5, dip6,
			dia1, dia2, dia3, dia4, dia5, dia6,
			rorc, rorn, rorl, rorx,
			cuno, cuor, cupo, cusx,
			prno, hdpr, popn, alwt, alwq,
			adid, rout, rodn, dsdt, dshm, modl, tedl, tel2,
			tepa, pact, cupa,
			e0pa, dsgp, pusn, putp,
			atv1, atv2, atv3, atv4, atv5, atv6, atv7, atv8, atv9, atv0,
			uca1, uca2, uca3, uca4, uca5, uca6, uca7, uca8, uca9, uca0,
			udn1, udn2, udn3, udn4, udn5, udn6,
			uid1, uid2, uid3,
			uct1,
			atnr, atmo, atpr, cfin,
			proj, elno,
			rgdt, rgtm, lmdt, chno, chid, lmts,
			m3_timestamp,
			sync_timestamp
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13,
			$14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23,
			$24, $25, $26,
			$27, $28, $29, $30, $31, $32, $33,
			$34, $35, $36, $37,
			$38, $39, $40, $41, $42, $43,
			$44, $45, $46, $47, $48, $49,
			$50, $51, $52, $53,
			$54, $55, $56, $57,
			$58, $59, $60, $61, $62,
			$63, $64, $65, $66, $67, $68, $69, $70,
			$71, $72, $73,
			$74, $75, $76, $77,
			$78, $79, $80, $81, $82, $83, $84, $85, $86, $87,
			$88, $89, $90, $91, $92, $93, $94, $95, $96, $97,
			$98, $99, $100, $101, $102, $103,
			$104, $105, $106,
			$107,
			$108, $109, $110, $111,
			$112, $113,
			$114, $115, $116, $117, $118, $119,
			$120,
			NOW()
		)
		ON CONFLICT (orno, ponr, posx)
		DO UPDATE SET
			itds = EXCLUDED.itds,
			teds = EXCLUDED.teds,
			orst = EXCLUDED.orst,
			orqt = EXCLUDED.orqt,
			rnqt = EXCLUDED.rnqt,
			alqt = EXCLUDED.alqt,
			dlqt = EXCLUDED.dlqt,
			ivqt = EXCLUDED.ivqt,
			dwdt = EXCLUDED.dwdt,
			codt = EXCLUDED.codt,
			pldt = EXCLUDED.pldt,
			lmdt = EXCLUDED.lmdt,
			lmts = EXCLUDED.lmts,
			m3_timestamp = EXCLUDED.m3_timestamp,
			sync_timestamp = NOW(),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, line := range lines {
		_, err = stmt.ExecContext(ctx,
			line.CONO, line.DIVI, line.ORNO, line.PONR, line.POSX,
			line.ITNO, line.ITDS, line.TEDS, line.REPI,
			line.ORST, line.ORTY,
			line.FACI, line.WHLO,
			line.ORQT, line.RNQT, line.ALQT, line.DLQT, line.IVQT,
			line.ORQA, line.RNQA, line.ALQA, line.DLQA, line.IVQA,
			line.ALUN, line.COFA, line.SPUN,
			line.DWDT, line.DWHM, line.CODT, line.COHM, line.PLDT, line.FDED, line.LDED,
			line.SAPR, line.NEPR, line.LNAM, line.CUCD,
			line.DIP1, line.DIP2, line.DIP3, line.DIP4, line.DIP5, line.DIP6,
			line.DIA1, line.DIA2, line.DIA3, line.DIA4, line.DIA5, line.DIA6,
			line.RORC, line.RORN, line.RORL, line.RORX,
			line.CUNO, line.CUOR, line.CUPO, line.CUSX,
			line.PRNO, line.HDPR, line.POPN, line.ALWT, line.ALWQ,
			line.ADID, line.ROUT, line.RODN, line.DSDT, line.DSHM, line.MODL, line.TEDL, line.TEL2,
			line.TEPA, line.PACT, line.CUPA,
			line.E0PA, line.DSGP, line.PUSN, line.PUTP,
			line.ATV1, line.ATV2, line.ATV3, line.ATV4, line.ATV5, line.ATV6, line.ATV7, line.ATV8, line.ATV9, line.ATV0,
			line.UCA1, line.UCA2, line.UCA3, line.UCA4, line.UCA5, line.UCA6, line.UCA7, line.UCA8, line.UCA9, line.UCA0,
			line.UDN1, line.UDN2, line.UDN3, line.UDN4, line.UDN5, line.UDN6,
			line.UID1, line.UID2, line.UID3,
			line.UCT1,
			line.ATNR, line.ATMO, line.ATPR, line.CFIN,
			line.PROJ, line.ELNO,
			line.RGDT, line.RGTM, line.LMDT, line.CHNO, line.CHID, line.LMTS,
			line.M3Timestamp,
		)
		if err != nil {
			return fmt.Errorf("failed to insert line %s-%s: %w", line.ORNO, line.PONR, err)
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
