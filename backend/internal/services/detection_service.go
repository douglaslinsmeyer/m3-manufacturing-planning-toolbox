package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/services/detectors"
)

// DetectionService manages issue detection
type DetectionService struct {
	db               *db.Queries
	registry         *detectors.DetectorRegistry
	configService    *DetectorConfigService
	progressCallback ProgressCallback
}

// NewDetectionService creates a new detection service
func NewDetectionService(database *db.Queries, configService *DetectorConfigService) *DetectionService {
	// Initialize detector registry with config service
	registry := detectors.NewDetectorRegistry()
	registry.Register(detectors.NewUnlinkedProductionOrdersDetector(configService))
	registry.Register(detectors.NewJointDeliveryDateMismatchDetector(configService))

	return &DetectionService{
		db:            database,
		registry:      registry,
		configService: configService,
	}
}

// SetProgressCallback sets the callback function for progress updates
func (s *DetectionService) SetProgressCallback(callback ProgressCallback) {
	s.progressCallback = callback
}

// reportProgress calls the progress callback if set
func (s *DetectionService) reportProgress(phase string, stepNum, totalSteps int, message string) {
	if s.progressCallback != nil {
		s.progressCallback(phase, stepNum, totalSteps, message, 0, 0, 0)
	}
}

// RunAllDetectors executes all registered detectors (respects enabled/disabled settings)
func (s *DetectionService) RunAllDetectors(ctx context.Context, jobID, environment, company, facility string) error {
	log.Printf("Starting issue detection for job %s (environment: %s, company: %s, facility: %s)", jobID, environment, company, facility)

	allDetectors := s.registry.GetAll()

	// Load detector enable/disable settings from system_settings for this environment
	enabledDetectors := s.loadEnabledDetectors(ctx, environment)

	// Filter to only enabled detectors
	activeDetectors := make([]detectors.IssueDetector, 0)
	for _, detector := range allDetectors {
		settingKey := fmt.Sprintf("detector_%s_enabled", detector.Name())

		// Check if detector is enabled (default: true if setting doesn't exist)
		if enabled, exists := enabledDetectors[settingKey]; exists && !enabled {
			log.Printf("Detector '%s' is disabled, skipping", detector.Name())
			continue
		}

		activeDetectors = append(activeDetectors, detector)
	}

	totalDetectors := len(activeDetectors)

	if totalDetectors == 0 {
		log.Println("No detectors enabled, skipping detection phase")
		return nil
	}

	log.Printf("Running %d enabled detectors (total available: %d)", totalDetectors, len(allDetectors))

	// Create detection job record
	if err := s.db.CreateIssueDetectionJob(ctx, jobID, environment, totalDetectors); err != nil {
		return fmt.Errorf("failed to create detection job: %w", err)
	}

	// Clear previous issues for this job
	if err := s.db.ClearIssuesForJob(ctx, jobID); err != nil {
		log.Printf("Warning: failed to clear previous issues: %v", err)
	}

	s.reportProgress("detection", 0, totalDetectors, "Starting issue detection")

	issuesByType := make(map[string]int)
	totalIssues := 0
	completedDetectors := 0

	for i, detector := range activeDetectors {
		log.Printf("Running detector %d/%d: %s", i+1, totalDetectors, detector.Name())
		s.reportProgress("detection", i, totalDetectors, fmt.Sprintf("Running %s detector", detector.Description()))

		issuesFound, err := detector.Detect(ctx, s.db, jobID, environment, company, facility)
		if err != nil {
			log.Printf("Detector %s failed: %v", detector.Name(), err)
			s.db.IncrementFailedDetectors(ctx, jobID)
			continue
		}

		issuesByType[detector.Name()] = issuesFound
		totalIssues += issuesFound
		completedDetectors++

		// Update progress
		if err := s.db.UpdateDetectionProgress(ctx, jobID, completedDetectors, totalDetectors); err != nil {
			log.Printf("Warning: failed to update detection progress: %v", err)
		}

		log.Printf("Detector %s found %d issues", detector.Name(), issuesFound)
	}

	// Update final results
	issuesByTypeJSON, _ := json.Marshal(issuesByType)
	if err := s.db.CompleteDetectionJob(ctx, jobID, totalIssues, string(issuesByTypeJSON)); err != nil {
		return fmt.Errorf("failed to complete detection job: %w", err)
	}

	s.reportProgress("detection", totalDetectors, totalDetectors, fmt.Sprintf("Detection complete - %d issues found", totalIssues))

	log.Printf("Issue detection completed - %d total issues found across %d enabled detectors", totalIssues, completedDetectors)
	return nil
}

// loadEnabledDetectors loads detector enable/disable settings from system_settings for a specific environment
// Returns map of setting_key → enabled (true/false)
// Default: all detectors enabled if setting doesn't exist
func (s *DetectionService) loadEnabledDetectors(ctx context.Context, environment string) map[string]bool {
	// Load settings for this environment
	settings, err := s.db.GetSystemSettings(ctx, environment)
	if err != nil {
		log.Printf("Warning: Failed to load detector settings: %v (all detectors will run)", err)
		return make(map[string]bool) // Empty map = all enabled by default
	}

	enabled := make(map[string]bool)
	for _, setting := range settings {
		if strings.HasPrefix(setting.SettingKey, "detector_") &&
			strings.HasSuffix(setting.SettingKey, "_enabled") {
			enabled[setting.SettingKey] = setting.SettingValue == "true"
		}
	}

	log.Printf("Loaded %d detector enable/disable settings for environment '%s'", len(enabled), environment)
	return enabled
}

// GetDetectorByName retrieves detector by name for async execution
func (s *DetectionService) GetDetectorByName(name string) detectors.IssueDetector {
	return s.registry.GetByName(name)
}

// GetAllDetectorNames returns all registered detector names
func (s *DetectionService) GetAllDetectorNames() []string {
	all := s.registry.GetAll()
	names := make([]string, len(all))
	for i, d := range all {
		names[i] = d.Name()
	}
	return names
}

// IsDetectorEnabled checks if detector is enabled for environment
func (s *DetectionService) IsDetectorEnabled(ctx context.Context, environment, detectorName string) (bool, error) {
	enabledMap := s.loadEnabledDetectors(ctx, environment)
	settingKey := fmt.Sprintf("detector_%s_enabled", detectorName)

	if enabled, exists := enabledMap[settingKey]; exists {
		return enabled, nil
	}

	return true, nil // Default: enabled
}
