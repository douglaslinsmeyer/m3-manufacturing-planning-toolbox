package detectors

import (
	"context"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// IssueDetector interface - all detectors must implement this
type IssueDetector interface {
	// Name returns the unique detector type identifier
	Name() string

	// Description returns a human-readable description
	Description() string

	// Detect runs the detection logic and returns issues found
	// Returns: issue count, error
	Detect(ctx context.Context, queries *db.Queries, company, facility string) (int, error)
}

// DetectorRegistry manages all registered detectors
type DetectorRegistry struct {
	detectors []IssueDetector
}

// NewDetectorRegistry creates a new detector registry
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make([]IssueDetector, 0),
	}
}

// Register adds a detector to the registry
func (r *DetectorRegistry) Register(detector IssueDetector) {
	r.detectors = append(r.detectors, detector)
}

// GetAll returns all registered detectors
func (r *DetectorRegistry) GetAll() []IssueDetector {
	return r.detectors
}

// GetByName retrieves a detector by its name
func (r *DetectorRegistry) GetByName(name string) IssueDetector {
	for _, d := range r.detectors {
		if d.Name() == name {
			return d
		}
	}
	return nil
}

// InitializeDetectors creates registry with default detectors
func InitializeDetectors() *DetectorRegistry {
	registry := NewDetectorRegistry()

	// Register all built-in detectors
	registry.Register(&UnlinkedProductionOrdersDetector{})
	registry.Register(&StartDateMismatchDetector{})
	registry.Register(&ProductionTimingDetector{})

	return registry
}
