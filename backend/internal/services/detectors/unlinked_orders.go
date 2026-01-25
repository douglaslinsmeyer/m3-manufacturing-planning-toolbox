package detectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// UnlinkedProductionOrdersDetector finds MO/MOP without CO links
type UnlinkedProductionOrdersDetector struct{}

func (d *UnlinkedProductionOrdersDetector) Name() string {
	return "unlinked_production_orders"
}

func (d *UnlinkedProductionOrdersDetector) Description() string {
	return "Detects manufacturing orders and planned orders without customer order links"
}

func (d *UnlinkedProductionOrdersDetector) Detect(ctx context.Context, queries *db.Queries, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for facility %s", d.Name(), facility)

	issuesFound := 0

	// Find unlinked MOs
	moQuery := `
		SELECT
			mfno as order_number,
			'MO' as order_type,
			faci,
			whlo,
			itno,
			orqt as ordered_qty,
			stdt,
			fidt,
			prno,
			cono,
			orty
		FROM manufacturing_orders
		WHERE cono = $1
		  AND faci = $2
		  AND (linked_co_number IS NULL OR linked_co_number = '')
		  AND deleted_remotely = false
	`

	moRows, err := queries.DB().QueryContext(ctx, moQuery, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query unlinked MOs: %w", err)
	}
	defer moRows.Close()

	for moRows.Next() {
		var orderNumber, orderType, faci, whlo, itno, orderedQty, stdt, fidt, prno, cono string
		var orty sql.NullString

		if err := moRows.Scan(&orderNumber, &orderType, &faci, &whlo, &itno, &orderedQty, &stdt, &fidt, &prno, &cono, &orty); err != nil {
			log.Printf("Error scanning MO row: %v", err)
			continue
		}

		// Build issue data
		issueData := map[string]interface{}{
			"item_number":      itno,
			"ordered_quantity": orderedQty,
			"start_date":       stdt,
			"finish_date":      fidt,
			"warehouse":        whlo,
			"product_number":   prno,
			"company":          cono,
		}
		if orty.Valid {
			issueData["mo_type"] = orty.String
		}

		if err := d.insertIssue(ctx, queries, orderNumber, orderType, faci, whlo, issueData); err != nil {
			log.Printf("Error inserting MO issue: %v", err)
			continue
		}

		issuesFound++
	}

	// Find unlinked MOPs
	mopQuery := `
		SELECT
			CAST(plpn AS VARCHAR) as order_number,
			'MOP' as order_type,
			faci,
			whlo,
			itno,
			ppqt as ordered_qty,
			stdt,
			fidt,
			cono,
			orty,
			prno
		FROM planned_manufacturing_orders
		WHERE cono = $1
		  AND faci = $2
		  AND (linked_co_number IS NULL OR linked_co_number = '')
		  AND deleted_remotely = false
	`

	mopRows, err := queries.DB().QueryContext(ctx, mopQuery, company, facility)
	if err != nil {
		return issuesFound, fmt.Errorf("failed to query unlinked MOPs: %w", err)
	}
	defer mopRows.Close()

	for mopRows.Next() {
		var orderNumber, orderType, faci, whlo, itno, orderedQty, stdt, fidt, cono string
		var orty, prno sql.NullString

		if err := mopRows.Scan(&orderNumber, &orderType, &faci, &whlo, &itno, &orderedQty, &stdt, &fidt, &cono, &orty, &prno); err != nil {
			log.Printf("Error scanning MOP row: %v", err)
			continue
		}

		issueData := map[string]interface{}{
			"item_number":      itno,
			"ordered_quantity": orderedQty,
			"start_date":       stdt,
			"finish_date":      fidt,
			"warehouse":        whlo,
			"company":          cono,
		}
		if orty.Valid {
			issueData["mo_type"] = orty.String
		}
		if prno.Valid {
			issueData["product_number"] = prno.String
		}

		if err := d.insertIssue(ctx, queries, orderNumber, orderType, faci, whlo, issueData); err != nil {
			log.Printf("Error inserting MOP issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d unlinked orders", d.Name(), issuesFound)
	return issuesFound, nil
}

func (d *UnlinkedProductionOrdersDetector) insertIssue(ctx context.Context, queries *db.Queries, orderNumber, orderType, facility, warehouse string, issueData map[string]interface{}) error {
	issueDataJSON, _ := json.Marshal(issueData)

	query := `
		INSERT INTO detected_issues (
			job_id, detector_type, facility, warehouse,
			issue_key, production_order_number, production_order_type,
			issue_data
		)
		SELECT
			id, $1, $2, $3,
			$4, $5, $6,
			$7
		FROM refresh_jobs
		ORDER BY created_at DESC
		LIMIT 1
	`

	issueKey := orderNumber // Group by order number

	_, err := queries.DB().ExecContext(ctx, query,
		d.Name(), facility, warehouse,
		issueKey, orderNumber, orderType,
		issueDataJSON,
	)

	return err
}
