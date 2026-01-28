package detectors

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// MOPDemandRatioDetector detects when the ratio of unlinked MOPs to actual
// customer order demand is excessive, indicating potential over-planning.
type MOPDemandRatioDetector struct {
	*BaseAnomalyDetector
	warningMOPsPerCOLine      float64 // Default: 10
	criticalMOPsPerCOLine     float64 // Default: 50
	criticalMOPsPerUnitDemand float64 // Default: 5
}

// NewMOPDemandRatioDetector creates a new MOP-to-demand ratio detector
func NewMOPDemandRatioDetector(db *sql.DB, enabled bool, warningMOPsPerCOLine, criticalMOPsPerCOLine, criticalMOPsPerUnitDemand float64) *MOPDemandRatioDetector {
	return &MOPDemandRatioDetector{
		BaseAnomalyDetector:       NewBaseAnomalyDetector(db, enabled),
		warningMOPsPerCOLine:      warningMOPsPerCOLine,
		criticalMOPsPerCOLine:     criticalMOPsPerCOLine,
		criticalMOPsPerUnitDemand: criticalMOPsPerUnitDemand,
	}
}

// Name returns the detector name
func (d *MOPDemandRatioDetector) Name() string {
	return "anomaly_mop_demand_ratio"
}

// Detect performs the anomaly detection
func (d *MOPDemandRatioDetector) Detect(ctx context.Context, env string) ([]*AnomalyAlert, error) {
	query := `
		WITH product_demand AS (
			SELECT
				itno as product,
				whlo as warehouse,
				COUNT(*) as co_line_count,
				SUM(CAST(orqt AS DECIMAL)) as total_demand_qty
			FROM customer_order_lines
			WHERE environment = $1
			  AND orst >= '20' AND orst < '66'
			  AND deleted_remotely = false
			GROUP BY itno, whlo
		),
		unlinked_mops AS (
			SELECT
				prno as product,
				whlo as warehouse,
				COUNT(*) as unlinked_mop_count
			FROM planned_manufacturing_orders
			WHERE environment = $1
			  AND (linked_co_number IS NULL OR linked_co_number = '')
			  AND deleted_remotely = false
			  AND psts = '20'
			GROUP BY prno, whlo
		)
		SELECT
			u.product,
			u.warehouse,
			u.unlinked_mop_count,
			d.co_line_count,
			d.total_demand_qty,
			ROUND((u.unlinked_mop_count::float / NULLIF(d.co_line_count, 0))::numeric, 2) as mops_per_co_line,
			ROUND((u.unlinked_mop_count::float / NULLIF(d.total_demand_qty, 0))::numeric, 2) as mops_per_unit_demand
		FROM unlinked_mops u
		INNER JOIN product_demand d ON u.product = d.product AND u.warehouse = d.warehouse
		WHERE (u.unlinked_mop_count::float / NULLIF(d.co_line_count, 0)) > $2
		   OR (u.unlinked_mop_count::float / NULLIF(d.total_demand_qty, 0)) > $3
		ORDER BY mops_per_co_line DESC
		LIMIT 20
	`

	rows, err := d.DB.QueryContext(ctx, query, env, d.warningMOPsPerCOLine, d.criticalMOPsPerUnitDemand)
	if err != nil {
		return nil, fmt.Errorf("failed to query MOP-to-demand ratio: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var product, warehouse string
		var unlinkedMOPCount, coLineCount int
		var totalDemandQty, mopsPerCOLine, mopsPerUnitDemand float64

		if err := rows.Scan(&product, &warehouse, &unlinkedMOPCount, &coLineCount, &totalDemandQty, &mopsPerCOLine, &mopsPerUnitDemand); err != nil {
			log.Printf("Failed to scan MOP-to-demand ratio row: %v", err)
			continue
		}

		// Determine severity based on both ratios
		severity := SeverityWarning
		threshold := d.warningMOPsPerCOLine
		actualValue := mopsPerCOLine

		if mopsPerCOLine >= d.criticalMOPsPerCOLine || mopsPerUnitDemand >= d.criticalMOPsPerUnitDemand {
			severity = SeverityCritical
			if mopsPerCOLine >= d.criticalMOPsPerCOLine {
				threshold = d.criticalMOPsPerCOLine
			} else {
				threshold = d.criticalMOPsPerUnitDemand
				actualValue = mopsPerUnitDemand
			}
		}

		alert := &AnomalyAlert{
			DetectorType:  d.Name(),
			Severity:      severity,
			EntityType:    EntityTypeProduct,
			EntityID:      product,
			AffectedCount: unlinkedMOPCount,
			Threshold:     threshold,
			ActualValue:   actualValue,
			Message: fmt.Sprintf(
				"Product %s has excessive unlinked MOPs relative to demand in warehouse %s: %d unlinked MOPs for %d CO lines (%.0f units demand). Ratios: %.2f MOPs/CO line, %.2f MOPs/unit",
				product, warehouse, unlinkedMOPCount, coLineCount, totalDemandQty, mopsPerCOLine, mopsPerUnitDemand,
			),
			Metrics: map[string]interface{}{
				"product":                 product,
				"warehouse":               warehouse,
				"unlinked_mop_count":      unlinkedMOPCount,
				"co_line_count":           coLineCount,
				"total_demand_qty":        totalDemandQty,
				"mops_per_co_line":        mopsPerCOLine,
				"mops_per_unit_demand":    mopsPerUnitDemand,
				"threshold_mops_per_line": d.criticalMOPsPerCOLine,
				"threshold_mops_per_unit": d.criticalMOPsPerUnitDemand,
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating MOP-to-demand ratio rows: %w", err)
	}

	return alerts, nil
}
