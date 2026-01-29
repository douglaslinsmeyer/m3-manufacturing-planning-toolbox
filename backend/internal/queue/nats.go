package queue

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Manager handles NATS connection and messaging
type Manager struct {
	conn    *nats.Conn
	url     string
	options []nats.Option
}

// NewManager creates a new NATS manager
func NewManager(natsURL string) (*Manager, error) {
	options := []nats.Option{
		nats.Name("M3 Planning Tools"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("NATS connection closed")
		}),
	}

	// Connect to NATS
	conn, err := nats.Connect(natsURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Printf("Connected to NATS at %s", natsURL)

	return &Manager{
		conn:    conn,
		url:     natsURL,
		options: options,
	}, nil
}

// Close closes the NATS connection
func (m *Manager) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

// Conn returns the NATS connection
func (m *Manager) Conn() *nats.Conn {
	return m.conn
}

// Publish publishes a message to a subject
func (m *Manager) Publish(subject string, data []byte) error {
	return m.conn.Publish(subject, data)
}

// Subscribe subscribes to a subject with a handler
func (m *Manager) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return m.conn.Subscribe(subject, handler)
}

// QueueSubscribe creates a queue subscriber (load balanced across workers)
func (m *Manager) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return m.conn.QueueSubscribe(subject, queue, handler)
}

// Request sends a request and waits for a response
func (m *Manager) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return m.conn.Request(subject, data, timeout)
}

// NATS Subject Patterns

const (
	// Snapshot refresh subjects
	SubjectSnapshotRefresh       = "snapshot.refresh"
	SubjectSnapshotRefreshTRN    = "snapshot.refresh.TRN"
	SubjectSnapshotRefreshPRD    = "snapshot.refresh.PRD"
	SubjectSnapshotProgress      = "snapshot.progress.%s"      // snapshot.progress.{jobID}
	SubjectSnapshotComplete      = "snapshot.complete.%s"      // snapshot.complete.{jobID}
	SubjectSnapshotError         = "snapshot.error.%s"         // snapshot.error.{jobID}
	SubjectSnapshotCancel        = "snapshot.cancel.%s"        // snapshot.cancel.{jobID}

	// Batch distribution subjects (for parallel data loading)
	SubjectSnapshotBatchTRN      = "snapshot.batch.TRN.>"      // Wildcard for all TRN batch jobs
	SubjectSnapshotBatchPRD      = "snapshot.batch.PRD.>"      // Wildcard for all PRD batch jobs
	SubjectBatchStart            = "snapshot.batch.start.%s"    // snapshot.batch.start.{parentJobId}
	SubjectBatchComplete         = "snapshot.batch.complete.%s" // snapshot.batch.complete.{parentJobId}

	// Detector distribution subjects (for parallel detector execution)
	SubjectSnapshotDetectorTRN   = "snapshot.detector.TRN.>"      // Wildcard for all TRN detector jobs
	SubjectSnapshotDetectorPRD   = "snapshot.detector.PRD.>"      // Wildcard for all PRD detector jobs
	SubjectDetectorStart         = "snapshot.detector.start.%s"    // snapshot.detector.start.{parentJobId}
	SubjectDetectorComplete      = "snapshot.detector.complete.%s" // snapshot.detector.complete.{parentJobId}

	// Detector coordinator subjects (for manual detection triggers)
	SubjectDetectorCoordinateTRN = "snapshot.detector.coordinate.TRN" // Coordinator jobs for TRN
	SubjectDetectorCoordinatePRD = "snapshot.detector.coordinate.PRD" // Coordinator jobs for PRD

	// Analysis subjects
	SubjectAnalysisRun           = "analysis.run"
	SubjectAnalysisProgress      = "analysis.progress.%s"      // analysis.progress.{jobID}
	SubjectAnalysisComplete      = "analysis.complete.%s"      // analysis.complete.{jobID}

	// Bulk operation subjects
	SubjectBulkOpRequestTRN      = "bulkop.request.TRN"             // Bulk operation request for TRN
	SubjectBulkOpRequestPRD      = "bulkop.request.PRD"             // Bulk operation request for PRD
	SubjectBulkOpBatch           = "bulkop.batch.%s.%s"             // bulkop.batch.{ENV}.{jobID}
	SubjectBulkOpBatchStart      = "bulkop.batch.start.%s"          // bulkop.batch.start.{jobID}
	SubjectBulkOpBatchComplete   = "bulkop.batch.complete.%s"       // bulkop.batch.complete.{jobID}
	SubjectBulkOpProgress        = "bulkop.progress.%s"             // bulkop.progress.{jobID}
	SubjectBulkOpComplete        = "bulkop.complete.%s"             // bulkop.complete.{jobID}
	SubjectBulkOpError           = "bulkop.error.%s"                // bulkop.error.{jobID}
	SubjectBulkOpCancel          = "bulkop.cancel.%s"               // bulkop.cancel.{jobID}

	// Queue groups (for load balancing)
	QueueGroupSnapshot           = "snapshot-workers"
	QueueGroupBatchWorkers       = "batch-workers"
	QueueGroupAnalysis           = "analysis-workers"
	QueueGroupBulkOpCoordinator  = "bulkop-coordinators"
	QueueGroupBulkOpWorkers      = "bulkop-workers"
)

// GetSnapshotRefreshSubject returns the subject for snapshot refresh based on environment
func GetSnapshotRefreshSubject(environment string) string {
	switch environment {
	case "TRN":
		return SubjectSnapshotRefreshTRN
	case "PRD":
		return SubjectSnapshotRefreshPRD
	default:
		return SubjectSnapshotRefresh
	}
}

// GetProgressSubject returns the progress subject for a job
func GetProgressSubject(jobID string) string {
	return fmt.Sprintf(SubjectSnapshotProgress, jobID)
}

// GetCompleteSubject returns the completion subject for a job
func GetCompleteSubject(jobID string) string {
	return fmt.Sprintf(SubjectSnapshotComplete, jobID)
}

// GetErrorSubject returns the error subject for a job
func GetErrorSubject(jobID string) string {
	return fmt.Sprintf(SubjectSnapshotError, jobID)
}

// GetBatchSubject returns the subject for a specific batch type
// Example: GetBatchSubject("TRN", "mops") → "snapshot.batch.TRN.mops"
func GetBatchSubject(environment, phase string) string {
	return fmt.Sprintf("snapshot.batch.%s.%s", environment, phase)
}

// GetPhaseProgressSubject returns the subject for phase sub-progress updates
// Example: GetPhaseProgressSubject("abc123") → "snapshot.phase.progress.abc123"
func GetPhaseProgressSubject(jobID string) string {
	return fmt.Sprintf("snapshot.phase.progress.%s", jobID)
}

// GetBatchStartSubject returns the subject for batch start events
// All batch start notifications for a job go to the same subject
func GetBatchStartSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectBatchStart, parentJobID)
}

// GetBatchCompleteSubject returns the subject for batch completion events
// All batch completions for a job go to the same subject
func GetBatchCompleteSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectBatchComplete, parentJobID)
}

// GetDetectorSubject returns the subject for a specific detector job
// Example: GetDetectorSubject("TRN", "unlinked_production_orders") → "snapshot.detector.TRN.unlinked_production_orders"
func GetDetectorSubject(environment, detectorName string) string {
	return fmt.Sprintf("snapshot.detector.%s.%s", environment, detectorName)
}

// GetDetectorStartSubject returns the subject for detector start events
// All detector start notifications for a job go to the same subject
func GetDetectorStartSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectDetectorStart, parentJobID)
}

// GetDetectorCompleteSubject returns the subject for detector completion events
// All detector completions for a job go to the same subject
func GetDetectorCompleteSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectDetectorComplete, parentJobID)
}

// GetDetectorCoordinateSubject returns the subject for detector coordinator jobs
// Example: GetDetectorCoordinateSubject("TRN") → "snapshot.detector.coordinate.TRN"
func GetDetectorCoordinateSubject(environment string) string {
	switch environment {
	case "TRN":
		return SubjectDetectorCoordinateTRN
	case "PRD":
		return SubjectDetectorCoordinatePRD
	default:
		return fmt.Sprintf("snapshot.detector.coordinate.%s", environment)
	}
}

// GetBulkOpRequestSubject returns the subject for bulk operation requests
// Example: GetBulkOpRequestSubject("TRN") → "bulkop.request.TRN"
func GetBulkOpRequestSubject(environment string) string {
	switch environment {
	case "TRN":
		return SubjectBulkOpRequestTRN
	case "PRD":
		return SubjectBulkOpRequestPRD
	default:
		return fmt.Sprintf("bulkop.request.%s", environment)
	}
}

// GetBulkOpBatchSubject returns the subject for a specific bulk operation batch
// Example: GetBulkOpBatchSubject("TRN", "abc123") → "bulkop.batch.TRN.abc123"
func GetBulkOpBatchSubject(environment, jobID string) string {
	return fmt.Sprintf(SubjectBulkOpBatch, environment, jobID)
}

// GetBulkOpBatchStartSubject returns the subject for bulk operation batch start events
func GetBulkOpBatchStartSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpBatchStart, jobID)
}

// GetBulkOpBatchCompleteSubject returns the subject for bulk operation batch completion events
func GetBulkOpBatchCompleteSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpBatchComplete, jobID)
}

// GetBulkOpProgressSubject returns the subject for bulk operation progress updates
func GetBulkOpProgressSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpProgress, jobID)
}

// GetBulkOpCompleteSubject returns the subject for bulk operation completion
func GetBulkOpCompleteSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpComplete, jobID)
}

// GetBulkOpErrorSubject returns the subject for bulk operation errors
func GetBulkOpErrorSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpError, jobID)
}

// GetBulkOpCancelSubject returns the subject for bulk operation cancellation
func GetBulkOpCancelSubject(jobID string) string {
	return fmt.Sprintf(SubjectBulkOpCancel, jobID)
}
