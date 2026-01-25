package detectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// ProductionTimingDetector finds production orders where start date is too early (>= 3 days before delivery) or too late (after delivery)
type ProductionTimingDetector struct{}

func (d *ProductionTimingDetector) Name() string {
	return "production_timing"
}

func (d *ProductionTimingDetector) Description() string {
	return "Detects production orders with start dates too early (>=3 days before delivery) or too late (after delivery)"
}

func (d *ProductionTimingDetector) Detect(ctx context.Context, queries *db.Queries, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for facility %s", d.Name(), facility)

	query := `
		SELECT
			po.order_number,
			po.order_type,
			po.linked_co_number,
			po.linked_co_line,
			po.linked_co_suffix,
			po.planned_start_date as start_date,
			col.codt as confirmed_delivery_date,
			-- Calculate days difference (CODT - planned_start_date)
			(TO_DATE(NULLIF(col.codt, ''), 'YYYYMMDD') - TO_DATE(NULLIF(po.planned_start_date, ''), 'YYYYMMDD')) as days_until_delivery,
			po.faci,
			po.warehouse,
			po.itno,
			po.cono,
			po.prno,
			po.orty
		FROM production_orders po
		INNER JOIN customer_order_lines col
			ON col.orno = po.linked_co_number
			AND col.ponr::VARCHAR = po.linked_co_line
		WHERE po.cono = $1
		  AND po.faci = $2
		  AND po.linked_co_number IS NOT NULL
		  AND po.linked_co_number != ''
		  AND po.planned_start_date IS NOT NULL AND po.planned_start_date != ''
		  AND col.codt IS NOT NULL AND col.codt != ''
		  AND po.deleted_remotely = false
		  AND (
			  -- Too early: starts 3+ days before delivery
			  (TO_DATE(NULLIF(col.codt, ''), 'YYYYMMDD') - TO_DATE(NULLIF(po.planned_start_date, ''), 'YYYYMMDD')) >= 3
			  OR
			  -- Too late: starts after delivery date
			  (TO_DATE(NULLIF(col.codt, ''), 'YYYYMMDD') - TO_DATE(NULLIF(po.planned_start_date, ''), 'YYYYMMDD')) < 0
		  )
	`

	rows, err := queries.DB().QueryContext(ctx, query, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query production timing issues: %w", err)
	}
	defer rows.Close()

	issuesFound := 0

	for rows.Next() {
		var orderNumber, orderType, coNumber, coLine, coSuffix, startDate, deliveryDate, faci, whlo, itno, cono, prno string
		var orty sql.NullString
		var daysDifference int

		if err := rows.Scan(&orderNumber, &orderType, &coNumber, &coLine, &coSuffix, &startDate, &deliveryDate, &daysDifference, &faci, &whlo, &itno, &cono, &prno, &orty); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Determine timing issue type
		var timingIssue string
		if daysDifference < 0 {
			timingIssue = "too_late"
		} else {
			timingIssue = "too_early"
		}

		// Build issue data
		issueData := map[string]interface{}{
			"start_date":       startDate,
			"delivery_date":    deliveryDate,
			"days_difference":  daysDifference,
			"timing_issue":     timingIssue,
			"item_number":      itno,
			"warehouse":        whlo,
			"co_number":        coNumber,
			"co_line":          coLine,
			"company":          cono,
			"product_number":   prno,
		}
		if orty.Valid {
			issueData["mo_type"] = orty.String
		}

		if err := d.insertIssue(ctx, queries, orderNumber, orderType, coNumber, coLine, coSuffix, faci, whlo, issueData); err != nil {
			log.Printf("Error inserting issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d timing issues", d.Name(), issuesFound)
	return issuesFound, nil
}

func (d *ProductionTimingDetector) insertIssue(ctx context.Context, queries *db.Queries, orderNumber, orderType, coNumber, coLine, coSuffix, facility, warehouse string, issueData map[string]interface{}) error {
	issueDataJSON, _ := json.Marshal(issueData)

	query := `
		INSERT INTO detected_issues (
			job_id, detector_type, facility, warehouse,
			issue_key, production_order_number, production_order_type,
			co_number, co_line, co_suffix,
			issue_data
		)
		SELECT
			id, $1, $2, $3,
			$4, $5, $6,
			$7, $8, $9,
			$10
		FROM refresh_jobs
		ORDER BY created_at DESC
		LIMIT 1
	`

	issueKey := orderNumber // Individual issue per production order

	_, err := queries.DB().ExecContext(ctx, query,
		d.Name(), facility, warehouse,
		issueKey, orderNumber, orderType,
		coNumber, coLine, coSuffix,
		issueDataJSON,
	)

	return err
}
