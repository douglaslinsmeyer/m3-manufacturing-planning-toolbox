package detectors

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// UnlinkedConcentrationDetector detects when a single product accounts for
// an excessive percentage of unlinked MOPs, indicating a potential runaway
// planning issue for that specific product.
type UnlinkedConcentrationDetector struct {
	*BaseAnomalyDetector
	warningThreshold  float64 // Default: 10%
	criticalThreshold float64 // Default: 50%
	minAffectedCount  int     // Default: 100
}

// NewUnlinkedConcentrationDetector creates a new unlinked concentration detector
func NewUnlinkedConcentrationDetector(db *sql.DB, enabled bool, warningThreshold, criticalThreshold float64, minAffectedCount int) *UnlinkedConcentrationDetector {
	return &UnlinkedConcentrationDetector{
		BaseAnomalyDetector: NewBaseAnomalyDetector(db, enabled),
		warningThreshold:    warningThreshold,
		criticalThreshold:   criticalThreshold,
		minAffectedCount:    minAffectedCount,
	}
}

// Name returns the detector name
func (d *UnlinkedConcentrationDetector) Name() string {
	return "anomaly_unlinked_concentration"
}

// Detect performs the anomaly detection
func (d *UnlinkedConcentrationDetector) Detect(ctx context.Context, env string) ([]*AnomalyAlert, error) {
	query := `
		WITH total_unlinked AS (
			SELECT COUNT(*) as total
			FROM planned_manufacturing_orders
			WHERE environment = $1
			  AND (linked_co_number IS NULL OR linked_co_number = '')
			  AND deleted_remotely = false
			  AND psts = '20'
		)
		SELECT
			COALESCE(prno, 'UNKNOWN') as product,
			whlo as warehouse,
			COUNT(*) as unlinked_count,
			ROUND((COUNT(*) * 100.0 / NULLIF((SELECT total FROM total_unlinked), 0))::numeric, 2) as concentration_pct
		FROM planned_manufacturing_orders
		WHERE environment = $1
			AND (linked_co_number IS NULL OR linked_co_number = '')
			AND deleted_remotely = false
			AND psts = '20'
		GROUP BY prno, whlo
		HAVING COUNT(*) > $2
			AND (COUNT(*) * 100.0 / NULLIF((SELECT total FROM total_unlinked), 0)) > $3
		ORDER BY concentration_pct DESC
		LIMIT 20
	`

	rows, err := d.DB.QueryContext(ctx, query, env, d.minAffectedCount, d.warningThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query unlinked concentration: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var product, warehouse string
		var unlinkedCount int
		var concentrationPct float64

		if err := rows.Scan(&product, &warehouse, &unlinkedCount, &concentrationPct); err != nil {
			log.Printf("Failed to scan unlinked concentration row: %v", err)
			continue
		}

		// Determine severity
		severity := SeverityWarning
		threshold := d.warningThreshold
		if concentrationPct >= d.criticalThreshold {
			severity = SeverityCritical
			threshold = d.criticalThreshold
		}

		// Get total unlinked count for context
		var totalUnlinked int
		if err := d.DB.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM planned_manufacturing_orders
			WHERE environment = $1
			  AND (linked_co_number IS NULL OR linked_co_number = '')
			  AND deleted_remotely = false
			  AND psts = '20'
		`, env).Scan(&totalUnlinked); err != nil {
			log.Printf("Failed to get total unlinked count: %v", err)
			totalUnlinked = 0
		}

		alert := &AnomalyAlert{
			DetectorType:  d.Name(),
			Severity:      severity,
			EntityType:    EntityTypeProduct,
			EntityID:      product,
			AffectedCount: unlinkedCount,
			Threshold:     threshold,
			ActualValue:   concentrationPct,
			Message: fmt.Sprintf(
				"Product %s accounts for %.2f%% of unlinked MOPs (%d of %d) in warehouse %s",
				product, concentrationPct, unlinkedCount, totalUnlinked, warehouse,
			),
			Metrics: map[string]interface{}{
				"product":            product,
				"warehouse":          warehouse,
				"unlinked_count":     unlinkedCount,
				"total_unlinked":     totalUnlinked,
				"concentration_pct":  concentrationPct,
				"threshold_exceeded": concentrationPct >= threshold,
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unlinked concentration rows: %w", err)
	}

	return alerts, nil
}
