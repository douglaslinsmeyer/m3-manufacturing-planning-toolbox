package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/services/detectors"
)

// DetectorConfigService manages detector configuration loading and resolution
// Implements detectors.ConfigService interface
type DetectorConfigService struct {
	queries *db.Queries
}

// NewDetectorConfigService creates a new detector config service
func NewDetectorConfigService(queries *db.Queries) *DetectorConfigService {
	return &DetectorConfigService{queries: queries}
}

// HierarchicalThreshold represents JSONB structure for hierarchical thresholds
type HierarchicalThreshold struct {
	Global    interface{}         `json:"global"`
	Overrides []ThresholdOverride `json:"overrides"`
}

// ThresholdOverride represents a scope-specific override
type ThresholdOverride struct {
	Warehouse *string     `json:"warehouse,omitempty"`
	Facility  *string     `json:"facility,omitempty"`
	MOType    *string     `json:"moType,omitempty"`
	Value     interface{} `json:"value"`
}

// ResolveThreshold resolves hierarchical JSONB threshold with scope precedence
// Precedence: Warehouse+Facility+MOType > Warehouse+MOType > Warehouse+Facility > etc.
// Returns: value (int or float64), found (bool), error
func (s *DetectorConfigService) ResolveThreshold(
	ctx context.Context,
	environment string,
	detectorName, parameterName string,
	warehouse, facility, moType *string,
) (interface{}, bool, error) {

	// Load system settings for this environment
	settings, err := s.queries.GetSystemSettings(ctx, environment)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load settings: %w", err)
	}

	// Find the hierarchical threshold setting
	settingKey := fmt.Sprintf("detector_%s_%s", detectorName, parameterName)
	var thresholdJSON string
	for _, setting := range settings {
		if setting.SettingKey == settingKey {
			thresholdJSON = setting.SettingValue
			break
		}
	}

	if thresholdJSON == "" {
		return nil, false, nil // Setting not found
	}

	// Parse JSONB
	var hierarchical HierarchicalThreshold
	if err := json.Unmarshal([]byte(thresholdJSON), &hierarchical); err != nil {
		return nil, false, fmt.Errorf("invalid JSONB for %s: %w", settingKey, err)
	}

	// Find best matching override (highest specificity score)
	var bestMatch *ThresholdOverride
	var bestScore int

	for i := range hierarchical.Overrides {
		override := &hierarchical.Overrides[i]
		score := 0
		matches := true

		// Check warehouse match
		if override.Warehouse != nil {
			if warehouse == nil || *warehouse != *override.Warehouse {
				continue // No match
			}
			score += 4
		}

		// Check facility match
		if override.Facility != nil {
			if facility == nil || *facility != *override.Facility {
				continue // No match
			}
			score += 2
		}

		// Check MO type match
		if override.MOType != nil {
			if moType == nil || *moType != *override.MOType {
				continue // No match
			}
			score += 1
		}

		// Update best match if this is more specific
		if matches && score > bestScore {
			bestMatch = override
			bestScore = score
		}
	}

	// Return best match or global default
	if bestMatch != nil {
		log.Printf("[DetectorConfig] %s.%s = %v (override, score=%d)", detectorName, parameterName, bestMatch.Value, bestScore)
		return bestMatch.Value, true, nil
	}

	log.Printf("[DetectorConfig] %s.%s = %v (global default)", detectorName, parameterName, hierarchical.Global)
	return hierarchical.Global, true, nil
}

// LoadFilters loads global filter settings for a detector for a specific environment
// Implements detectors.ConfigService interface
func (s *DetectorConfigService) LoadFilters(ctx context.Context, environment, detectorName string) (detectors.DetectorFilters, error) {
	settings, err := s.queries.GetSystemSettings(ctx, environment)
	if err != nil {
		return detectors.DetectorFilters{}, fmt.Errorf("failed to load settings: %w", err)
	}

	filters := detectors.DetectorFilters{
		ExcludeMOStatuses:    []string{},
		ExcludeMOPStatuses:   []string{},
		MinOrderAgeDays:      0,
		ExcludeFacilities:    []string{},
		MinQuantityThreshold: 0.0,
	}

	prefix := fmt.Sprintf("detector_%s_", detectorName)

	for _, setting := range settings {
		key := setting.SettingKey
		if !strings.HasPrefix(key, prefix) || strings.HasSuffix(key, "_enabled") {
			continue
		}

		// Skip hierarchical threshold settings (those have "hierarchical": true in constraints)
		var constraints map[string]interface{}
		if len(setting.Constraints) > 0 {
			json.Unmarshal(setting.Constraints, &constraints)
			if constraints["hierarchical"] == true {
				continue
			}
		}

		switch {
		case strings.HasSuffix(key, "_exclude_mo_statuses"):
			json.Unmarshal([]byte(setting.SettingValue), &filters.ExcludeMOStatuses)
		case strings.HasSuffix(key, "_exclude_mop_statuses"):
			json.Unmarshal([]byte(setting.SettingValue), &filters.ExcludeMOPStatuses)
		case strings.HasSuffix(key, "_min_order_age_days"):
			filters.MinOrderAgeDays, _ = strconv.Atoi(setting.SettingValue)
		case strings.HasSuffix(key, "_exclude_facilities"):
			json.Unmarshal([]byte(setting.SettingValue), &filters.ExcludeFacilities)
		case strings.HasSuffix(key, "_min_quantity_threshold"):
			filters.MinQuantityThreshold, _ = strconv.ParseFloat(setting.SettingValue, 64)
		}
	}

	return filters, nil
}
