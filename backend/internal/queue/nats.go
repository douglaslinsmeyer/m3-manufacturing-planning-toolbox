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

	// Phase distribution subjects (for parallel execution)
	SubjectSnapshotPhaseTRN      = "snapshot.phase.TRN.>"      // Wildcard for all TRN phase jobs
	SubjectSnapshotPhasePRD      = "snapshot.phase.PRD.>"      // Wildcard for all PRD phase jobs
	SubjectPhaseComplete         = "snapshot.phase.complete.%s" // snapshot.phase.complete.{parentJobId}
	SubjectSnapshotFinalize      = "snapshot.finalize.%s"       // snapshot.finalize.{jobId}

	// Batch distribution subjects (for ID range-based parallel batching)
	SubjectSnapshotBatchTRN      = "snapshot.batch.TRN.>"      // Wildcard for all TRN batch jobs
	SubjectSnapshotBatchPRD      = "snapshot.batch.PRD.>"      // Wildcard for all PRD batch jobs
	SubjectBatchComplete         = "snapshot.batch.complete.%s" // snapshot.batch.complete.{parentJobId}

	// Analysis subjects
	SubjectAnalysisRun           = "analysis.run"
	SubjectAnalysisProgress      = "analysis.progress.%s"      // analysis.progress.{jobID}
	SubjectAnalysisComplete      = "analysis.complete.%s"      // analysis.complete.{jobID}

	// Queue groups (for load balancing)
	QueueGroupSnapshot           = "snapshot-workers"
	QueueGroupPhaseWorkers       = "phase-workers"
	QueueGroupBatchWorkers       = "batch-workers" // NEW: For parallel batch processing
	QueueGroupAnalysis           = "analysis-workers"
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

// GetPhaseSubject returns subject for a specific phase job
func GetPhaseSubject(environment, jobID, phase string) string {
	return fmt.Sprintf("snapshot.phase.%s.%s.%s", environment, jobID, phase)
}

// GetPhaseCompleteSubject returns the phase completion subject
func GetPhaseCompleteSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectPhaseComplete, parentJobID)
}

// GetBatchSubject returns the subject for a specific batch type
// Example: GetBatchSubject("TRN", "mops") → "snapshot.batch.TRN.mops"
func GetBatchSubject(environment, phase string) string {
	return fmt.Sprintf("snapshot.batch.%s.%s", environment, phase)
}

// GetBatchCompleteSubject returns the subject for batch completion events
// All batch completions for a job go to the same subject
func GetBatchCompleteSubject(parentJobID string) string {
	return fmt.Sprintf(SubjectBatchComplete, parentJobID)
}
