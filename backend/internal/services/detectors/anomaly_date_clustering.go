package detectors

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// DateClusteringDetector detects when an excessive percentage of MOPs
// for a product are scheduled on the same planned date, indicating
// potential bulk planning issues or misconfiguration.
type DateClusteringDetector struct {
	*BaseAnomalyDetector
	warningThreshold  float64 // Default: 80%
	criticalThreshold float64 // Default: 95%
	minAffectedCount  int     // Default: 100
}

// NewDateClusteringDetector creates a new date clustering detector
func NewDateClusteringDetector(db *sql.DB, enabled bool, warningThreshold, criticalThreshold float64, minAffectedCount int) *DateClusteringDetector {
	return &DateClusteringDetector{
		BaseAnomalyDetector: NewBaseAnomalyDetector(db, enabled),
		warningThreshold:    warningThreshold,
		criticalThreshold:   criticalThreshold,
		minAffectedCount:    minAffectedCount,
	}
}

// Name returns the detector name
func (d *DateClusteringDetector) Name() string {
	return "anomaly_date_clustering"
}

// Detect performs the anomaly detection
func (d *DateClusteringDetector) Detect(ctx context.Context, env string) ([]*AnomalyAlert, error) {
	query := `
		WITH product_totals AS (
			SELECT
				prno,
				COUNT(*) as total_mops
			FROM planned_manufacturing_orders
			WHERE environment = $1
			  AND (linked_co_number IS NULL OR linked_co_number = '')
			  AND deleted_remotely = false
			  AND psts = '20'
			GROUP BY prno
		)
		SELECT
			mop.pldt as planned_date,
			COALESCE(mop.prno, 'UNKNOWN') as product,
			mop.whlo as warehouse,
			COUNT(*) as mop_count,
			pt.total_mops,
			ROUND((COUNT(*) * 100.0 / NULLIF(pt.total_mops, 0))::numeric, 2) as date_concentration_pct
		FROM planned_manufacturing_orders mop
		INNER JOIN product_totals pt ON mop.prno = pt.prno
		WHERE mop.environment = $1
		  AND (mop.linked_co_number IS NULL OR mop.linked_co_number = '')
		  AND mop.deleted_remotely = false
		  AND mop.psts = '20'
		GROUP BY mop.pldt, mop.prno, mop.whlo, pt.total_mops
		HAVING COUNT(*) > $2
		  AND (COUNT(*) * 100.0 / NULLIF(pt.total_mops, 0)) > $3
		ORDER BY date_concentration_pct DESC
		LIMIT 20
	`

	rows, err := d.DB.QueryContext(ctx, query, env, d.minAffectedCount, d.warningThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query date clustering: %w", err)
	}
	defer rows.Close()

	var alerts []*AnomalyAlert

	for rows.Next() {
		var plannedDate sql.NullInt32
		var product, warehouse string
		var mopCount, totalMOPs int
		var dateConcentrationPct float64

		if err := rows.Scan(&plannedDate, &product, &warehouse, &mopCount, &totalMOPs, &dateConcentrationPct); err != nil {
			log.Printf("Failed to scan date clustering row: %v", err)
			continue
		}

		// Convert M3 date format (YYYYMMDD) to readable date
		var dateStr string
		if plannedDate.Valid && plannedDate.Int32 > 0 {
			dateInt := int(plannedDate.Int32)
			year := dateInt / 10000
			month := (dateInt / 100) % 100
			day := dateInt % 100
			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			dateStr = t.Format("2006-01-02")
		} else {
			dateStr = "UNKNOWN"
		}

		// Determine severity
		severity := SeverityWarning
		threshold := d.warningThreshold
		if dateConcentrationPct >= d.criticalThreshold {
			severity = SeverityCritical
			threshold = d.criticalThreshold
		}

		alert := &AnomalyAlert{
			DetectorType:  d.Name(),
			Severity:      severity,
			EntityType:    EntityTypeProduct,
			EntityID:      product,
			AffectedCount: mopCount,
			Threshold:     threshold,
			ActualValue:   dateConcentrationPct,
			Message: fmt.Sprintf(
				"Product %s has %.2f%% of unlinked MOPs on single date %s (%d of %d MOPs) in warehouse %s",
				product, dateConcentrationPct, dateStr, mopCount, totalMOPs, warehouse,
			),
			Metrics: map[string]interface{}{
				"product":                product,
				"warehouse":              warehouse,
				"planned_date":           dateStr,
				"planned_date_raw":       plannedDate.Int32,
				"mop_count_on_date":      mopCount,
				"total_product_mops":     totalMOPs,
				"date_concentration_pct": dateConcentrationPct,
				"threshold_exceeded":     dateConcentrationPct >= threshold,
			},
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating date clustering rows: %w", err)
	}

	return alerts, nil
}
