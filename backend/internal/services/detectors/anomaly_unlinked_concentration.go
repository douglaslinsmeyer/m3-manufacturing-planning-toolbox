package detectors

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// UnlinkedConcentrationDetector detects when a single product or CFIN accounts for
// an excessive percentage of unlinked MOPs, indicating a potential runaway
// planning issue for that specific product or configuration.
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
	var alerts []*AnomalyAlert

	// Get total unlinked count once for both queries
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
		return nil, fmt.Errorf("failed to get total unlinked count: %w", err)
	}

	// Check product concentration
	productAlerts, err := d.detectProductConcentration(ctx, env, totalUnlinked)
	if err != nil {
		return nil, err
	}
	alerts = append(alerts, productAlerts...)

	// Check CFIN concentration
	cfinAlerts, err := d.detectCFINConcentration(ctx, env, totalUnlinked)
	if err != nil {
		return nil, err
	}
	alerts = append(alerts, cfinAlerts...)

	return alerts, nil
}

// detectProductConcentration detects product-based concentration anomalies
func (d *UnlinkedConcentrationDetector) detectProductConcentration(ctx context.Context, env string, totalUnlinked int) ([]*AnomalyAlert, error) {
	query := `
		WITH total_unlinked AS (
			SELECT $4::integer as total
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

	rows, err := d.DB.QueryContext(ctx, query, env, d.minAffectedCount, d.warningThreshold, totalUnlinked)
	if err != nil {
		return nil, fmt.Errorf("failed to query product concentration: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var product, warehouse string
		var unlinkedCount int
		var concentrationPct float64

		if err := rows.Scan(&product, &warehouse, &unlinkedCount, &concentrationPct); err != nil {
			log.Printf("Failed to scan product concentration row: %v", err)
			continue
		}

		// Determine severity
		severity := SeverityWarning
		threshold := d.warningThreshold
		if concentrationPct >= d.criticalThreshold {
			severity = SeverityCritical
			threshold = d.criticalThreshold
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
				"concentration_type": "product",
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product concentration rows: %w", err)
	}

	return alerts, nil
}

// detectCFINConcentration detects CFIN-based concentration anomalies
func (d *UnlinkedConcentrationDetector) detectCFINConcentration(ctx context.Context, env string, totalUnlinked int) ([]*AnomalyAlert, error) {
	query := `
		WITH total_unlinked AS (
			SELECT $4::integer as total
		)
		SELECT
			COALESCE(cfin, 'UNKNOWN') as cfin,
			whlo as warehouse,
			COUNT(*) as unlinked_count,
			ROUND((COUNT(*) * 100.0 / NULLIF((SELECT total FROM total_unlinked), 0))::numeric, 2) as concentration_pct
		FROM planned_manufacturing_orders
		WHERE environment = $1
			AND (linked_co_number IS NULL OR linked_co_number = '')
			AND deleted_remotely = false
			AND psts = '20'
			AND cfin IS NOT NULL
			AND cfin != ''
		GROUP BY cfin, whlo
		HAVING COUNT(*) > $2
			AND (COUNT(*) * 100.0 / NULLIF((SELECT total FROM total_unlinked), 0)) > $3
		ORDER BY concentration_pct DESC
		LIMIT 20
	`

	rows, err := d.DB.QueryContext(ctx, query, env, d.minAffectedCount, d.warningThreshold, totalUnlinked)
	if err != nil {
		return nil, fmt.Errorf("failed to query CFIN concentration: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var cfin, warehouse string
		var unlinkedCount int
		var concentrationPct float64

		if err := rows.Scan(&cfin, &warehouse, &unlinkedCount, &concentrationPct); err != nil {
			log.Printf("Failed to scan CFIN concentration row: %v", err)
			continue
		}

		// Determine severity
		severity := SeverityWarning
		threshold := d.warningThreshold
		if concentrationPct >= d.criticalThreshold {
			severity = SeverityCritical
			threshold = d.criticalThreshold
		}

		alert := &AnomalyAlert{
			DetectorType:  d.Name(),
			Severity:      severity,
			EntityType:    "configuration", // New entity type for CFIN
			EntityID:      cfin,
			AffectedCount: unlinkedCount,
			Threshold:     threshold,
			ActualValue:   concentrationPct,
			Message: fmt.Sprintf(
				"Configuration (CFIN %s) accounts for %.2f%% of unlinked MOPs (%d of %d) in warehouse %s",
				cfin, concentrationPct, unlinkedCount, totalUnlinked, warehouse,
			),
			Metrics: map[string]interface{}{
				"cfin":               cfin,
				"warehouse":          warehouse,
				"unlinked_count":     unlinkedCount,
				"total_unlinked":     totalUnlinked,
				"concentration_pct":  concentrationPct,
				"threshold_exceeded": concentrationPct >= threshold,
				"concentration_type": "cfin",
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating CFIN concentration rows: %w", err)
	}

	return alerts, nil
}
