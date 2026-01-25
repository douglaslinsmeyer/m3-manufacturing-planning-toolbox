package detectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// StartDateMismatchDetector finds MO/MOP groups for same CO Line with different start dates
type StartDateMismatchDetector struct{}

func (d *StartDateMismatchDetector) Name() string {
	return "start_date_mismatch"
}

func (d *StartDateMismatchDetector) Description() string {
	return "Detects production orders linked to the same customer order line with different start dates"
}

func (d *StartDateMismatchDetector) Detect(ctx context.Context, queries *db.Queries, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for facility %s", d.Name(), facility)

	// Find groups with multiple start dates
	query := `
		WITH combined_orders AS (
			SELECT
				linked_co_number,
				linked_co_line,
				linked_co_suffix,
				stdt,
				mfno as order_number,
				'MO' as order_type,
				faci,
				whlo,
				itno,
				cono,
				prno,
				orty
			FROM manufacturing_orders
			WHERE cono = $1
			  AND faci = $2
			  AND linked_co_number IS NOT NULL
			  AND linked_co_number != ''
			  AND stdt IS NOT NULL
			  AND stdt != ''
			  AND deleted_remotely = false

			UNION ALL

			SELECT
				linked_co_number,
				linked_co_line,
				linked_co_suffix,
				stdt,
				CAST(plpn AS VARCHAR) as order_number,
				'MOP' as order_type,
				faci,
				whlo,
				itno,
				cono,
				prno,
				orty
			FROM planned_manufacturing_orders
			WHERE cono = $1
			  AND faci = $2
			  AND linked_co_number IS NOT NULL
			  AND linked_co_number != ''
			  AND stdt IS NOT NULL
			  AND stdt != ''
			  AND deleted_remotely = false
		),
		mismatched_groups AS (
			SELECT
				linked_co_number,
				linked_co_line,
				linked_co_suffix,
				faci,
				whlo,
				itno,
				cono,
				array_agg(DISTINCT stdt ORDER BY stdt) as dates,
				array_agg(json_build_object(
					'number', order_number,
					'type', order_type,
					'date', stdt,
					'product_number', prno,
					'mo_type', orty
				) ORDER BY order_type, order_number) as orders
			FROM combined_orders
			GROUP BY linked_co_number, linked_co_line, linked_co_suffix, faci, whlo, itno, cono
			HAVING COUNT(DISTINCT stdt) > 1
		)
		SELECT
			linked_co_number,
			linked_co_line,
			linked_co_suffix,
			faci,
			whlo,
			itno,
			cono,
			dates,
			orders
		FROM mismatched_groups
	`

	rows, err := queries.DB().QueryContext(ctx, query, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query start date mismatches: %w", err)
	}
	defer rows.Close()

	issuesFound := 0

	for rows.Next() {
		var coNumber, coLine, coSuffix, faci, whlo, itno, cono string
		var datesJSON, ordersJSON []byte

		if err := rows.Scan(&coNumber, &coLine, &coSuffix, &faci, &whlo, &itno, &cono, &datesJSON, &ordersJSON); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse arrays
		var dates []string
		var orders []map[string]interface{}

		if err := json.Unmarshal(datesJSON, &dates); err != nil {
			log.Printf("Error unmarshaling dates: %v", err)
			continue
		}

		if err := json.Unmarshal(ordersJSON, &orders); err != nil {
			log.Printf("Error unmarshaling orders: %v", err)
			continue
		}

		// Build issue data
		issueData := map[string]interface{}{
			"dates":       dates,
			"orders":      orders,
			"item_number": itno,
			"warehouse":   whlo,
			"company":     cono,
		}

		if err := d.insertIssue(ctx, queries, coNumber, coLine, coSuffix, faci, whlo, issueData, orders); err != nil {
			log.Printf("Error inserting issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d date mismatch groups", d.Name(), issuesFound)
	return issuesFound, nil
}

func (d *StartDateMismatchDetector) insertIssue(ctx context.Context, queries *db.Queries, coNumber, coLine, coSuffix, facility, warehouse string, issueData map[string]interface{}, orders []map[string]interface{}) error {
	issueDataJSON, _ := json.Marshal(issueData)

	// Insert one issue per production order in the group
	for _, order := range orders {
		orderNumber, _ := order["number"].(string)
		orderType, _ := order["type"].(string)

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

		issueKey := fmt.Sprintf("%s-%s-%s", coNumber, coLine, coSuffix)

		_, err := queries.DB().ExecContext(ctx, query,
			d.Name(), facility, warehouse,
			issueKey, orderNumber, orderType,
			coNumber, coLine, coSuffix,
			issueDataJSON,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
