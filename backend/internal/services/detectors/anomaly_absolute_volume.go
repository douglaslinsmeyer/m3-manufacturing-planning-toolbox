package detectors

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// AbsoluteVolumeDetector detects when a product/warehouse combination has
// an excessive absolute count of unlinked MOPs, regardless of percentages
// or ratios. This catches situations where the raw volume is concerning.
type AbsoluteVolumeDetector struct {
	*BaseAnomalyDetector
	warningThreshold  int // Default: 1000
	criticalThreshold int // Default: 10000
}

// NewAbsoluteVolumeDetector creates a new absolute volume detector
func NewAbsoluteVolumeDetector(db *sql.DB, enabled bool, warningThreshold, criticalThreshold int) *AbsoluteVolumeDetector {
	return &AbsoluteVolumeDetector{
		BaseAnomalyDetector: NewBaseAnomalyDetector(db, enabled),
		warningThreshold:    warningThreshold,
		criticalThreshold:   criticalThreshold,
	}
}

// Name returns the detector name
func (d *AbsoluteVolumeDetector) Name() string {
	return "anomaly_absolute_volume"
}

// Detect performs the anomaly detection
func (d *AbsoluteVolumeDetector) Detect(ctx context.Context, env string) ([]*AnomalyAlert, error) {
	query := `
		SELECT
			COALESCE(prno, 'UNKNOWN') as product,
			whlo as warehouse,
			COUNT(*) as unlinked_count
		FROM planned_manufacturing_orders
		WHERE environment = $1
		  AND (linked_co_number IS NULL OR linked_co_number = '')
		  AND deleted_remotely = false
		  AND psts = '20'
		GROUP BY prno, whlo
		HAVING COUNT(*) > $2
		ORDER BY unlinked_count DESC
		LIMIT 20
	`

	rows, err := d.DB.QueryContext(ctx, query, env, d.warningThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query absolute volume: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var product, warehouse string
		var unlinkedCount int

		if err := rows.Scan(&product, &warehouse, &unlinkedCount); err != nil {
			log.Printf("Failed to scan absolute volume row: %v", err)
			continue
		}

		// Determine severity
		severity := SeverityWarning
		threshold := float64(d.warningThreshold)
		if unlinkedCount >= d.criticalThreshold {
			severity = SeverityCritical
			threshold = float64(d.criticalThreshold)
		}

		alert := &AnomalyAlert{
			DetectorType:  d.Name(),
			Severity:      severity,
			EntityType:    EntityTypeProduct,
			EntityID:      product,
			AffectedCount: unlinkedCount,
			Threshold:     threshold,
			ActualValue:   float64(unlinkedCount),
			Message: fmt.Sprintf(
				"Product %s in warehouse %s has %d unlinked MOPs (threshold: %.0f)",
				product, warehouse, unlinkedCount, threshold,
			),
			Metrics: map[string]interface{}{
				"product":        product,
				"warehouse":      warehouse,
				"unlinked_count": unlinkedCount,
				"threshold":      threshold,
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating absolute volume rows: %w", err)
	}

	return alerts, nil
}
