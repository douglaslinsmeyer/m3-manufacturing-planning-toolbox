package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/services/detectors"
)

// DetectionService manages issue detection
type DetectionService struct {
	db               *db.Queries
	registry         *detectors.DetectorRegistry
	progressCallback ProgressCallback
}

// NewDetectionService creates a new detection service
func NewDetectionService(database *db.Queries) *DetectionService {
	return &DetectionService{
		db:       database,
		registry: detectors.InitializeDetectors(),
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

// RunAllDetectors executes all registered detectors
func (s *DetectionService) RunAllDetectors(ctx context.Context, jobID, company, facility string) error {
	log.Printf("Starting issue detection for job %s (company: %s, facility: %s)", jobID, company, facility)

	allDetectors := s.registry.GetAll()
	totalDetectors := len(allDetectors)

	// Create detection job record
	if err := s.db.CreateIssueDetectionJob(ctx, jobID, totalDetectors); err != nil {
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

	for i, detector := range allDetectors {
		log.Printf("Running detector %d/%d: %s", i+1, totalDetectors, detector.Name())
		s.reportProgress("detection", i, totalDetectors, fmt.Sprintf("Running %s detector", detector.Description()))

		issuesFound, err := detector.Detect(ctx, s.db, company, facility)
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

	log.Printf("Issue detection completed - %d total issues found across %d detectors", totalIssues, completedDetectors)
	return nil
}
