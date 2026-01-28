package detectors

import (
	"context"
	"database/sql"
)

// AnomalyDetector is the interface for anomaly detection implementations.
// Unlike IssueDetector which analyzes individual records, AnomalyDetector
// analyzes aggregate patterns and statistical anomalies across datasets.
type AnomalyDetector interface {
	// Name returns the unique identifier for this detector
	Name() string

	// Detect performs anomaly detection and returns alerts
	Detect(ctx context.Context, env string) ([]*AnomalyAlert, error)

	// Enabled returns whether this detector is currently enabled
	Enabled() bool
}

// AnomalyAlert represents a detected anomaly
type AnomalyAlert struct {
	// DetectorType is the unique identifier for the detector (e.g., "anomaly_unlinked_concentration")
	DetectorType string

	// Severity indicates the severity level: "info", "warning", or "critical"
	Severity string

	// EntityType describes what kind of entity is affected: "product", "warehouse", "system", etc.
	EntityType string

	// EntityID is the identifier for the affected entity (e.g., product number, warehouse code)
	EntityID string

	// Message is a human-readable description of the anomaly
	Message string

	// Metrics contains additional statistical data about the anomaly
	Metrics map[string]interface{}

	// AffectedCount is the number of records affected by this anomaly
	AffectedCount int

	// Threshold is the threshold value that was breached (if applicable)
	Threshold float64

	// ActualValue is the actual measured value that triggered the alert
	ActualValue float64
}

// Severity constants
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Entity type constants
const (
	EntityTypeProduct   = "product"
	EntityTypeWarehouse = "warehouse"
	EntityTypeSystem    = "system"
)

// BaseAnomalyDetector provides common functionality for anomaly detectors
type BaseAnomalyDetector struct {
	DB      *sql.DB
	enabled bool
}

// NewBaseAnomalyDetector creates a new base detector
func NewBaseAnomalyDetector(db *sql.DB, enabled bool) *BaseAnomalyDetector {
	return &BaseAnomalyDetector{
		DB:      db,
		enabled: enabled,
	}
}

// Enabled returns whether the detector is enabled
func (b *BaseAnomalyDetector) Enabled() bool {
	return b.enabled
}
