package detectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// DLIXDateMismatchDetector finds CO lines within same DLIX group with misaligned MO/MOP start dates
type DLIXDateMismatchDetector struct {
	configService ConfigService
}

// NewDLIXDateMismatchDetector creates a new detector with config service
func NewDLIXDateMismatchDetector(configService ConfigService) *DLIXDateMismatchDetector {
	return &DLIXDateMismatchDetector{configService: configService}
}

func (d *DLIXDateMismatchDetector) Name() string {
	return "dlix_date_mismatch"
}

func (d *DLIXDateMismatchDetector) Label() string {
	return "Delivery Date Mismatches"
}

func (d *DLIXDateMismatchDetector) Description() string {
	return "Detects production orders within same delivery (DLIX) with misaligned start dates"
}

func (d *DLIXDateMismatchDetector) Detect(ctx context.Context, queries *db.Queries, refreshJobID, environment, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for environment %s, facility %s, refresh job %s", d.Name(), environment, facility, refreshJobID)

	// Resolve tolerance_days threshold (use facility scope, no warehouse/MO type)
	toleranceDaysRaw, foundTolerance, err := d.configService.ResolveThreshold(
		ctx, environment, d.Name(), "tolerance_days", nil, &facility, nil)
	if err != nil || !foundTolerance {
		log.Printf("[%s] Warning: failed to resolve tolerance_days: %v (using default 0)", d.Name(), err)
		toleranceDaysRaw = float64(0)
	}

	toleranceDays := int(toleranceDaysRaw.(float64))
	log.Printf("[%s] Using tolerance_days = %d for facility %s", d.Name(), toleranceDays, facility)

	// Note: Filters and status exclusions could be added to production_orders view query if needed
	// Currently using all production orders that are already filtered by the view definition

	// Find DLIX groups with start dates beyond tolerance
	// Group by linked_co_number + dlix to analyze delivery groups within same customer order
	query := fmt.Sprintf(`
		WITH dlix_production_orders AS (
			SELECT
				po.linked_co_number as co_number,
				col.dlix,
				po.linked_co_line as co_line,
				po.linked_co_suffix as co_suffix,
				po.order_number as production_order_number,
				po.order_type as production_order_type,
				po.planned_start_date,
				po.planned_finish_date,
				po.faci as facility,
				po.warehouse,
				po.itno as item_number,
				po.prno as product_number,
				po.orty as mo_type,
				po.ordered_quantity as planned_quantity,
				po.cono,
				col.codt as confirmed_delivery_date,
				col.dwdt as requested_delivery_date,
				col.cuno as customer_number,
				col.ortp as co_type_number,
				col.customer_name,
				col.co_type_description,
				col.delivery_method
			FROM production_orders po
			INNER JOIN customer_order_lines col
				ON po.linked_co_number = col.orno
				AND po.linked_co_line = col.ponr
				AND po.linked_co_suffix = col.posx
				AND po.environment = col.environment
			WHERE po.environment = $1
			  AND po.cono = $2
			  AND po.faci = $3
			  AND col.dlix IS NOT NULL
			  AND col.dlix != ''
			  AND po.planned_start_date IS NOT NULL
			  AND po.planned_start_date != ''
			  AND po.deleted_remotely = false
			  AND col.orst >= '20'
			  AND col.orst < '30'
		),
		mismatched_dlix_groups AS (
			SELECT
				co_number,
				dlix,
				facility,
				warehouse,
				item_number,
				cono,
				MIN(CAST(planned_start_date AS INTEGER)) as min_date,
				MAX(CAST(planned_start_date AS INTEGER)) as max_date,
				COUNT(DISTINCT co_line || '-' || co_suffix) as num_co_lines,
				COUNT(*) as num_production_orders,
				json_agg(DISTINCT planned_start_date ORDER BY planned_start_date) as dates,
				json_agg(json_build_object(
					'number', production_order_number,
					'type', production_order_type,
					'date', planned_start_date,
					'co_line', co_line || '-' || co_suffix,
					'product_number', product_number,
					'mo_type', mo_type,
					'quantity', planned_quantity,
					'confirmed_delivery_date', confirmed_delivery_date,
					'requested_delivery_date', requested_delivery_date,
					'customer_number', customer_number,
					'customer_name', customer_name,
					'co_type_number', co_type_number,
					'co_type_description', co_type_description,
					'delivery_method', delivery_method
				) ORDER BY production_order_type, production_order_number) as orders
			FROM dlix_production_orders
			GROUP BY co_number, dlix, facility, warehouse, item_number, cono
			-- Check if date variance exceeds tolerance (dates are YYYYMMDD strings)
			HAVING (
				-- TO_DATE subtraction returns integer days directly, no EXTRACT needed
				(TO_DATE(MAX(planned_start_date), 'YYYYMMDD') -
				 TO_DATE(MIN(planned_start_date), 'YYYYMMDD')) > %d
			)
		)
		SELECT
			co_number,
			dlix,
			facility,
			warehouse,
			item_number,
			cono,
			min_date,
			max_date,
			num_co_lines,
			num_production_orders,
			dates,
			orders
		FROM mismatched_dlix_groups
	`, toleranceDays)

	rows, err := queries.DB().QueryContext(ctx, query, environment, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query delivery date mismatches: %w", err)
	}
	defer rows.Close()

	issuesFound := 0

	for rows.Next() {
		var coNumber, dlix, faci, whlo, itno, cono string
		var minDate, maxDate int
		var numCOLines, numProdOrders int
		var datesJSON, ordersJSON []byte

		if err := rows.Scan(&coNumber, &dlix, &faci, &whlo, &itno, &cono, &minDate, &maxDate, &numCOLines, &numProdOrders, &datesJSON, &ordersJSON); err != nil {
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

		// Extract customer, CO type, and delivery info from first order (all orders share same CO)
		var customerName, customerNumber, coTypeNumber, coTypeDescription, deliveryMethod string
		if len(orders) > 0 {
			if val, ok := orders[0]["customer_name"].(string); ok {
				customerName = val
			}
			if val, ok := orders[0]["customer_number"].(string); ok {
				customerNumber = val
			}
			if val, ok := orders[0]["co_type_number"].(string); ok {
				coTypeNumber = val
			}
			if val, ok := orders[0]["co_type_description"].(string); ok {
				coTypeDescription = val
			}
			if val, ok := orders[0]["delivery_method"].(string); ok {
				deliveryMethod = val
			}
		}

		// Build issue data
		issueData := map[string]interface{}{
			"dlix":                    dlix,
			"dates":                   dates,
			"min_date":                minDate,
			"max_date":                maxDate,
			"num_co_lines":            numCOLines,
			"num_production_orders":   numProdOrders,
			"orders":                  orders,
			"tolerance_days":          toleranceDays,
			"item_number":             itno,
			"warehouse":               whlo,
			"company":                 cono,
			"customer_name":           customerName,
			"customer_number":         customerNumber,
			"co_type_number":          coTypeNumber,
			"co_type_description":     coTypeDescription,
			"delivery_method":         deliveryMethod,
		}

		if err := d.insertIssue(ctx, queries, refreshJobID, environment, coNumber, dlix, faci, whlo, issueData, orders); err != nil {
			log.Printf("Error inserting issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d delivery groups with date mismatches", d.Name(), issuesFound)
	return issuesFound, nil
}

func (d *DLIXDateMismatchDetector) insertIssue(ctx context.Context, queries *db.Queries, refreshJobID, environment, coNumber, dlix, facility, warehouse string, issueData map[string]interface{}, orders []map[string]interface{}) error {
	issueDataJSON, _ := json.Marshal(issueData)

	// Insert ONE issue per DLIX group (not per production order)
	// This avoids cluttering the UI with duplicate rows for the same DLIX problem

	// Handle edge case: empty orders array
	if len(orders) == 0 {
		log.Printf("Warning: DLIX group %s-%s has no orders, skipping", coNumber, dlix)
		return nil
	}

	// Use first order for representative top-level fields
	firstOrder := orders[0]
	orderNumber, _ := firstOrder["number"].(string)
	orderType, _ := firstOrder["type"].(string)
	coLineKey, _ := firstOrder["co_line"].(string) // Already formatted as "line-suffix"

	query := `
		INSERT INTO detected_issues (
			environment, job_id, detector_type, facility, warehouse,
			issue_key, production_order_number, production_order_type,
			co_number, co_line, co_suffix,
			issue_data
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12
		)
	`

	// Parse co_line and co_suffix from coLineKey (format: "line-suffix")
	var coLine, coSuffix string
	fmt.Sscanf(coLineKey, "%[^-]-%s", &coLine, &coSuffix)

	// Issue key is co_number + dlix to group all orders in same DLIX group
	issueKey := fmt.Sprintf("%s-DLIX-%s", coNumber, dlix)

	_, err := queries.DB().ExecContext(ctx, query,
		environment, refreshJobID, d.Name(), facility, warehouse,
		issueKey, orderNumber, orderType,
		coNumber, coLine, coSuffix,
		issueDataJSON,
	)

	return err
}
