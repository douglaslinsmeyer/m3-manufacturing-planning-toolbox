package api

import (
	"encoding/json"
	"net/http"

	"github.com/pinggolf/m3-planning-tools/internal/compass"
)

// handleTestCompassQuery allows testing Compass queries directly
func (s *Server) handleTestCompassQuery(w http.ResponseWriter, r *http.Request) {
	// Get Compass client for current user session
	compassClient, err := s.getCompassClient(r)
	if err != nil {
		http.Error(w, "Failed to initialize Compass client", http.StatusInternalServerError)
		return
	}

	// Simple test query
	testQuery := `
SELECT
  mop.PLPN, mop.PLPS, mop.FACI, mop.ITNO, mop.PSTS, mop.WHST,
  mpreal.DRDN as linked_co_number,
  mpreal.DRDL as linked_co_line,
  mpreal.PQTY as allocated_qty
FROM MMOPLP mop
LEFT JOIN MPREAL mpreal
  ON mpreal.AOCA = '5'
  AND CAST(mpreal.ARDN AS BIGINT) = mop.PLPN
  AND mpreal.DOCA = '3'
  AND mpreal.deleted = 'false'
WHERE mop.deleted = 'false'
  AND mop.PSTS IN ('10', '20')
LIMIT 5
`

	// Execute query
	results, err := compassClient.ExecuteQuery(r.Context(), testQuery, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse results
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return raw results for inspection
	response := map[string]interface{}{
		"recordCount": len(resultSet.Records),
		"columns":     resultSet.Columns,
		"records":     resultSet.Records,
	}

	if len(resultSet.Records) > 0 {
		// Show field names from first record
		fieldNames := make([]string, 0)
		for key := range resultSet.Records[0] {
			fieldNames = append(fieldNames, key)
		}
		response["fieldNamesFromFirstRecord"] = fieldNames
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
