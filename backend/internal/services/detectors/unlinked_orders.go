package detectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// UnlinkedProductionOrdersDetector finds MO/MOP without CO links (with configurable filters)
type UnlinkedProductionOrdersDetector struct {
	configService ConfigService
}

// NewUnlinkedProductionOrdersDetector creates a new detector with config service
func NewUnlinkedProductionOrdersDetector(configService ConfigService) *UnlinkedProductionOrdersDetector {
	return &UnlinkedProductionOrdersDetector{configService: configService}
}

func (d *UnlinkedProductionOrdersDetector) Name() string {
	return "unlinked_production_orders"
}

func (d *UnlinkedProductionOrdersDetector) Label() string {
	return "Unlinked Production Orders"
}

func (d *UnlinkedProductionOrdersDetector) Description() string {
	return "Detects manufacturing orders and planned orders without customer order links (with configurable filters)"
}

func (d *UnlinkedProductionOrdersDetector) Detect(ctx context.Context, queries *db.Queries, refreshJobID, environment, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for environment %s, facility %s, refresh job %s", d.Name(), environment, facility, refreshJobID)

	// Load global filters for this environment
	filters, err := d.configService.LoadFilters(ctx, environment, d.Name())
	if err != nil {
		log.Printf("[%s] Warning: failed to load filters: %v (using defaults)", d.Name(), err)
		filters = DetectorFilters{}
	}

	// Build filter clauses
	moStatusClause := buildStatusExclusionSQL("whst", filters.ExcludeMOStatuses)
	mopStatusClause := buildStatusExclusionSQL("psts", filters.ExcludeMOPStatuses)
	facilityClause := buildFacilityExclusionSQL(filters.ExcludeFacilities)

	// Build age filter (only flag orders older than min_order_age_days)
	ageClause := ""
	if filters.MinOrderAgeDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -filters.MinOrderAgeDays)
		cutoffDateInt := cutoffDate.Year()*10000 + int(cutoffDate.Month())*100 + cutoffDate.Day()
		ageClause = fmt.Sprintf("AND CAST(stdt AS INTEGER) < %d", cutoffDateInt)
	}

	// Build quantity filter
	quantityClause := ""
	if filters.MinQuantityThreshold > 0 {
		quantityClause = fmt.Sprintf("AND CAST(orqt AS DECIMAL) >= %.6f", filters.MinQuantityThreshold)
	}

	issuesFound := 0

	// Find unlinked MOs with filters
	moQuery := fmt.Sprintf(`
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
			orty,
			whst
		FROM manufacturing_orders
		WHERE environment = $1
		  AND cono = $2
		  AND faci = $3
		  AND (linked_co_number IS NULL OR linked_co_number = '')
		  AND deleted_remotely = false
		  %s
		  %s
		  %s
		  %s
	`, moStatusClause, facilityClause, ageClause, quantityClause)

	moRows, err := queries.DB().QueryContext(ctx, moQuery, environment, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query unlinked MOs: %w", err)
	}
	defer moRows.Close()

	for moRows.Next() {
		var orderNumber, orderType, faci, whlo, itno, orderedQty, stdt, fidt, prno, cono string
		var orty, whst sql.NullString

		if err := moRows.Scan(&orderNumber, &orderType, &faci, &whlo, &itno, &orderedQty, &stdt, &fidt, &prno, &cono, &orty, &whst); err != nil {
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
		if whst.Valid {
			issueData["status"] = whst.String
		}

		if err := d.insertIssue(ctx, queries, refreshJobID, environment, orderNumber, orderType, faci, whlo, issueData); err != nil {
			log.Printf("Error inserting MO issue: %v", err)
			continue
		}

		issuesFound++
	}

	// Build quantity filter for MOPs
	quantityClauseMOP := ""
	if filters.MinQuantityThreshold > 0 {
		quantityClauseMOP = fmt.Sprintf("AND CAST(ppqt AS DECIMAL) >= %.6f", filters.MinQuantityThreshold)
	}

	// Find unlinked MOPs with filters
	mopQuery := fmt.Sprintf(`
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
			prno,
			psts
		FROM planned_manufacturing_orders
		WHERE environment = $1
		  AND cono = $2
		  AND faci = $3
		  AND (linked_co_number IS NULL OR linked_co_number = '')
		  AND deleted_remotely = false
		  %s
		  %s
		  %s
		  %s
	`, mopStatusClause, facilityClause, ageClause, quantityClauseMOP)

	mopRows, err := queries.DB().QueryContext(ctx, mopQuery, environment, company, facility)
	if err != nil {
		return issuesFound, fmt.Errorf("failed to query unlinked MOPs: %w", err)
	}
	defer mopRows.Close()

	for mopRows.Next() {
		var orderNumber, orderType, faci, whlo, itno, orderedQty, stdt, fidt, cono string
		var orty, prno, psts sql.NullString

		if err := mopRows.Scan(&orderNumber, &orderType, &faci, &whlo, &itno, &orderedQty, &stdt, &fidt, &cono, &orty, &prno, &psts); err != nil {
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
		if psts.Valid {
			issueData["status"] = psts.String
		}

		if err := d.insertIssue(ctx, queries, refreshJobID, environment, orderNumber, orderType, faci, whlo, issueData); err != nil {
			log.Printf("Error inserting MOP issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d unlinked orders (filters applied: mo_statuses=%v, mop_statuses=%v, min_age_days=%d, facilities=%v, min_qty=%.2f)",
		d.Name(), issuesFound,
		filters.ExcludeMOStatuses, filters.ExcludeMOPStatuses,
		filters.MinOrderAgeDays, filters.ExcludeFacilities, filters.MinQuantityThreshold)
	return issuesFound, nil
}

func (d *UnlinkedProductionOrdersDetector) insertIssue(ctx context.Context, queries *db.Queries, refreshJobID, environment, orderNumber, orderType, facility, warehouse string, issueData map[string]interface{}) error {
	issueDataJSON, _ := json.Marshal(issueData)

	query := `
		INSERT INTO detected_issues (
			environment, job_id, detector_type, facility, warehouse,
			issue_key, production_order_number, production_order_type,
			issue_data
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9
		)
	`

	issueKey := orderNumber // Group by order number

	_, err := queries.DB().ExecContext(ctx, query,
		environment, refreshJobID, d.Name(), facility, warehouse,
		issueKey, orderNumber, orderType,
		issueDataJSON,
	)

	return err
}
