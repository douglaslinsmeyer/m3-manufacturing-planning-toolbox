package detectors

import (
	"context"
	"fmt"
	"strings"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// ConfigService interface to avoid circular dependency
// Services package will implement this interface
type ConfigService interface {
	// ResolveThreshold resolves a hierarchical threshold value for a specific environment
	ResolveThreshold(ctx context.Context, environment, detectorName, parameterName string, warehouse, facility, moType *string) (interface{}, bool, error)

	// LoadFilters loads global filter settings for a detector for a specific environment
	LoadFilters(ctx context.Context, environment, detectorName string) (DetectorFilters, error)
}

// DetectorFilters contains global filter settings
type DetectorFilters struct {
	ExcludeMOStatuses    []string
	ExcludeMOPStatuses   []string
	MinOrderAgeDays      int
	ExcludeFacilities    []string
	MinQuantityThreshold float64
}

// Helper functions shared by all detectors

// buildStatusExclusionSQL builds a WHERE clause to exclude specific statuses
func buildStatusExclusionSQL(columnName string, statuses []string) string {
	if len(statuses) == 0 {
		return ""
	}
	quoted := make([]string, len(statuses))
	for i, s := range statuses {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}
	return fmt.Sprintf("AND %s NOT IN (%s)", columnName, strings.Join(quoted, ","))
}

// buildFacilityExclusionSQL builds a WHERE clause to exclude specific facilities
func buildFacilityExclusionSQL(facilities []string) string {
	if len(facilities) == 0 {
		return ""
	}
	quoted := make([]string, len(facilities))
	for i, f := range facilities {
		quoted[i] = fmt.Sprintf("'%s'", f)
	}
	return fmt.Sprintf("AND faci NOT IN (%s)", strings.Join(quoted, ","))
}

// IssueDetector interface - all detectors must implement this
type IssueDetector interface {
	// Name returns the unique detector type identifier
	Name() string

	// Label returns a short, user-friendly display name
	Label() string

	// Description returns a human-readable description
	Description() string

	// Detect runs the detection logic and returns issues found
	// Returns: issue count, error
	Detect(ctx context.Context, queries *db.Queries, refreshJobID, environment, company, facility string) (int, error)
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

