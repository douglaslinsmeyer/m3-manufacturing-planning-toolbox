package detectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// COQuantityMismatchDetector finds CO lines where production order quantities don't match remaining quantity
// Implements M3 putaway logic: MOs don't reduce CO remaining qty until putaway complete
type COQuantityMismatchDetector struct {
	configService ConfigService
}

// NewCOQuantityMismatchDetector creates a new detector with config service
func NewCOQuantityMismatchDetector(configService ConfigService) *COQuantityMismatchDetector {
	return &COQuantityMismatchDetector{configService: configService}
}

func (d *COQuantityMismatchDetector) Name() string {
	return "co_quantity_mismatch"
}

func (d *COQuantityMismatchDetector) Label() string {
	return "CO Line Quantity Mismatches"
}

func (d *COQuantityMismatchDetector) Description() string {
	return "Detects customer order lines where production order quantities don't match remaining quantity (RNQA), accounting for pending putaway status"
}

func (d *COQuantityMismatchDetector) Detect(ctx context.Context, queries *db.Queries, refreshJobID, environment, company, facility string) (int, error) {
	log.Printf("[%s] Running detector for environment %s, facility %s, refresh job %s", d.Name(), environment, facility, refreshJobID)

	// Resolve tolerance threshold
	toleranceRaw, found, err := d.configService.ResolveThreshold(
		ctx, environment, d.Name(), "tolerance_threshold", nil, &facility, nil)
	if err != nil || !found {
		log.Printf("[%s] Warning: failed to resolve tolerance_threshold: %v (using default 0.01)", d.Name(), err)
		toleranceRaw = float64(0.01)
	}
	tolerance := toleranceRaw.(float64)
	log.Printf("[%s] Using tolerance_threshold = %.6f for facility %s", d.Name(), tolerance, facility)

	// Build the detection query with putaway logic
	query := fmt.Sprintf(`
WITH production_orders_countable AS (
    -- Determine which POs should count toward CO line supply
    SELECT
        linked_co_number,
        linked_co_line,
        linked_co_suffix,
        faci,
        warehouse,
        cono,
        environment,
        order_number,
        order_type,
        itno,
        bano,
        prno,
        orty,
        planned_start_date,
        planned_finish_date,
        ordered_quantity,
        manufactured_quantity,
        pending_putaway_qty,
        -- Determine if this PO should count toward supply
        CASE
            -- MOPs always count (no manufactured quantity)
            WHEN order_type = 'MOP' THEN true
            -- MOs count if: incomplete (maqt = 0/NULL) OR (completed but pending putaway)
            WHEN order_type = 'MO' THEN (
                NULLIF(manufactured_quantity, '') IS NULL OR
                CAST(NULLIF(manufactured_quantity, '') AS DECIMAL) = 0 OR
                (CAST(NULLIF(manufactured_quantity, '') AS DECIMAL) > 0 AND
                 CAST(NULLIF(pending_putaway_qty, '') AS DECIMAL) > 0)
            )
            ELSE false
        END as should_count
    FROM production_orders
    WHERE environment = $1
      AND cono = $2
      AND faci = $3
      AND linked_co_number IS NOT NULL
      AND linked_co_number != ''
      AND deleted_remotely = false
),
co_po_aggregated AS (
    -- Aggregate countable PO quantities by CO line
    SELECT
        linked_co_number,
        linked_co_line,
        linked_co_suffix,
        faci,
        warehouse,
        cono,
        environment,
        COUNT(*) FILTER (WHERE should_count) as countable_po_count,
        COUNT(*) as total_po_count,
        SUM(CAST(NULLIF(ordered_quantity, '') AS DECIMAL)) FILTER (WHERE should_count) as total_po_quantity,
        -- Aggregate all PO details including putaway status
        json_agg(json_build_object(
            'number', order_number,
            'type', order_type,
            'quantity', ordered_quantity,
            'manufactured_quantity', manufactured_quantity,
            'pending_putaway_qty', pending_putaway_qty,
            'should_count', should_count,
            'start_date', planned_start_date,
            'finish_date', planned_finish_date,
            'item_number', itno,
            'lot_number', bano,
            'product_number', prno,
            'mo_type', orty
        ) ORDER BY order_type, order_number) as production_orders
    FROM production_orders_countable
    GROUP BY linked_co_number, linked_co_line, linked_co_suffix,
             faci, warehouse, cono, environment
),
quantity_mismatches AS (
    -- Join to CO lines and find mismatches
    SELECT
        agg.linked_co_number as co_number,
        agg.linked_co_line as co_line,
        agg.linked_co_suffix as co_suffix,
        agg.faci as facility,
        agg.warehouse,
        agg.cono,
        agg.countable_po_count,
        agg.total_po_count,
        agg.total_po_quantity,
        CAST(NULLIF(col.rnqa, '') AS DECIMAL) as co_remaining_quantity,
        (agg.total_po_quantity - CAST(NULLIF(col.rnqa, '') AS DECIMAL)) as quantity_variance,
        col.itno as item_number,
        col.dwdt as requested_delivery_date,
        col.codt as confirmed_delivery_date,
        col.cuno as customer_number,
        col.customer_name,
        col.ortp as co_type_number,
        col.co_type_description,
        col.delivery_method,
        col.orst as co_status,
        agg.production_orders
    FROM co_po_aggregated agg
    INNER JOIN customer_order_lines col
        ON agg.linked_co_number = col.orno
        AND agg.linked_co_line = col.ponr
        AND agg.linked_co_suffix = col.posx
        AND agg.environment = col.environment
    WHERE col.orst >= '20' AND col.orst < '30'  -- Reserved only
      AND col.rnqa IS NOT NULL AND col.rnqa != ''
      AND ABS(agg.total_po_quantity - CAST(NULLIF(col.rnqa, '') AS DECIMAL)) > %f  -- tolerance
)
SELECT * FROM quantity_mismatches ORDER BY ABS(quantity_variance) DESC
`, tolerance)

	rows, err := queries.DB().QueryContext(ctx, query, environment, company, facility)
	if err != nil {
		return 0, fmt.Errorf("failed to query quantity mismatches: %w", err)
	}
	defer rows.Close()

	issuesFound := 0

	for rows.Next() {
		var coNumber, coLine, coSuffix, facilityCode, warehouse, cono string
		var countablePOCount, totalPOCount int
		var totalPOQty, coRemainingQty, variance sql.NullFloat64
		var itemNumber, requestedDeliveryDate, confirmedDeliveryDate string
		var customerNumber, customerName, coTypeNumber, coTypeDescription, deliveryMethod, coStatus string
		var ordersJSON []byte

		if err := rows.Scan(
			&coNumber, &coLine, &coSuffix, &facilityCode, &warehouse, &cono,
			&countablePOCount, &totalPOCount, &totalPOQty, &coRemainingQty, &variance,
			&itemNumber, &requestedDeliveryDate, &confirmedDeliveryDate,
			&customerNumber, &customerName, &coTypeNumber, &coTypeDescription, &deliveryMethod, &coStatus,
			&ordersJSON,
		); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse production orders JSON
		var orders []map[string]interface{}
		if err := json.Unmarshal(ordersJSON, &orders); err != nil {
			log.Printf("Error unmarshaling orders: %v", err)
			continue
		}

		// Build issue data
		issueData := map[string]interface{}{
			"co_line":                coLine,
			"co_suffix":              coSuffix,
			"co_remaining_quantity":  coRemainingQty.Float64,
			"total_po_quantity":      totalPOQty.Float64,
			"quantity_variance":      variance.Float64,
			"countable_po_count":     countablePOCount,
			"total_po_count":         totalPOCount,
			"item_number":            itemNumber,
			"customer_number":        customerNumber,
			"customer_name":          customerName,
			"requested_delivery_date": requestedDeliveryDate,
			"confirmed_delivery_date": confirmedDeliveryDate,
			"co_type_number":         coTypeNumber,
			"co_type_description":    coTypeDescription,
			"delivery_method":        deliveryMethod,
			"co_status":              coStatus,
			"tolerance_threshold":    tolerance,
			"production_orders":      orders,
		}

		if err := d.insertIssue(ctx, queries, refreshJobID, environment, coNumber, coLine, coSuffix, facilityCode, warehouse, issueData, orders); err != nil {
			log.Printf("Error inserting issue: %v", err)
			continue
		}

		issuesFound++
	}

	log.Printf("[%s] Found %d CO lines with quantity mismatches (countable_pos != remaining_qty)", d.Name(), issuesFound)
	return issuesFound, nil
}

func (d *COQuantityMismatchDetector) insertIssue(ctx context.Context, queries *db.Queries,
	refreshJobID, environment, coNumber, coLine, coSuffix, facility, warehouse string,
	issueData map[string]interface{}, orders []map[string]interface{}) error {

	issueDataJSON, _ := json.Marshal(issueData)

	// Handle edge case: empty orders array
	if len(orders) == 0 {
		log.Printf("Warning: CO line %s-%s-%s has no orders, skipping", coNumber, coLine, coSuffix)
		return nil
	}

	// Use first production order for representative top-level fields
	firstOrder := orders[0]
	orderNumber, _ := firstOrder["number"].(string)
	orderType, _ := firstOrder["type"].(string)

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

	// Issue key is co_number + co_line + co_suffix (one issue per CO line)
	issueKey := fmt.Sprintf("%s-%s-%s", coNumber, coLine, coSuffix)

	_, err := queries.DB().ExecContext(ctx, query,
		environment, refreshJobID, d.Name(), facility, warehouse,
		issueKey, orderNumber, orderType,
		coNumber, coLine, coSuffix,
		issueDataJSON,
	)

	return err
}
