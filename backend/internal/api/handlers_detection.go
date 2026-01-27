package api

import (
	"net/http"

	"github.com/pinggolf/m3-planning-tools/internal/handlers"
)

// handleListDetectors lists all available detectors with their enabled status
func (s *Server) handleListDetectors(w http.ResponseWriter, r *http.Request) {
	handlers.HandleListDetectors(s.db)(w, r)
}

// handleTriggerDetection triggers specific detectors without a full refresh
func (s *Server) handleTriggerDetection(w http.ResponseWriter, r *http.Request) {
	handlers.HandleTriggerDetection(s.natsManager, s.db)(w, r)
}
