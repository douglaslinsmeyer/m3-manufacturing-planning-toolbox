package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// CustomerOrderLineResponse represents a CO line returned from CFIN lookup
type CustomerOrderLineResponse struct {
	OrderNumber    string  `json:"order_number"`
	LineNumber     int     `json:"line_number"`
	LineSuffix     int     `json:"line_suffix"`
	CFIN           string  `json:"cfin"`
	ItemNumber     string  `json:"item_number"`
	ItemDesc       string  `json:"item_description"`
	OrderQuantity  float64 `json:"order_quantity"`
	OrderStatus    string  `json:"order_status"`
	Warehouse      string  `json:"warehouse"`
	RequestedDate  int     `json:"requested_date"`
	ConfirmedDate  int     `json:"confirmed_date"`
	PlannedDate    int     `json:"planned_date"`
	CustomerNumber string  `json:"customer_number"`
}

// CFINLookupResponse represents the response for CFIN lookup
type CFINLookupResponse struct {
	CFIN          string                        `json:"cfin"`
	Found         bool                          `json:"found"`
	CustomerLines []CustomerOrderLineResponse   `json:"customer_order_lines"`
	Message       string                        `json:"message,omitempty"`
}

// handleLookupCFIN performs on-demand lookup of customer orders by CFIN from Compass SQL
// GET /api/lookup/cfin/{cfin}
func (s *Server) handleLookupCFIN(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cfin := vars["cfin"]

	if cfin == "" {
		http.Error(w, "CFIN parameter is required", http.StatusBadRequest)
		return
	}

	// Get Compass client with user's session credentials
	compassClient, err := s.getCompassClient(r)
	if err != nil {
		log.Printf("Failed to get Compass client: %v", err)
		http.Error(w, "Failed to authenticate with M3", http.StatusUnauthorized)
		return
	}

	// Query Compass SQL for CO lines with this CFIN
	query := fmt.Sprintf(`
		SELECT
			oline.ORNO as order_number,
			oline.PONR as line_number,
			oline.POSX as line_suffix,
			oline.CFIN as cfin,
			oline.ITNO as item_number,
			oline.ITDS as item_description,
			oline.ORQT as order_quantity,
			oline.ORST as order_status,
			oline.WHLO as warehouse,
			oline.DWDT as requested_date,
			oline.CODT as confirmed_date,
			oline.PLDT as planned_date,
			oline.CUNO as customer_number
		FROM "default".OOLINE oline
		WHERE oline.CFIN = '%s'
		  AND oline.deleted = 'false'
		ORDER BY oline.ORNO, oline.PONR, oline.POSX
		LIMIT 100
	`, cfin)

	// Execute query with pagination (page size 100 means single page for LIMIT 100)
	jsonData, _, err := compassClient.ExecuteQueryWithPagination(r.Context(), query, 100, nil)
	if err != nil {
		log.Printf("Failed to query CFIN %s: %v", cfin, err)
		http.Error(w, fmt.Sprintf("Failed to query M3: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse JSON response into array of maps
	var rows []map[string]interface{}
	if err := json.Unmarshal(jsonData, &rows); err != nil {
		log.Printf("Failed to parse query results: %v", err)
		http.Error(w, "Failed to parse M3 response", http.StatusInternalServerError)
		return
	}

	response := CFINLookupResponse{
		CFIN:          cfin,
		Found:         len(rows) > 0,
		CustomerLines: make([]CustomerOrderLineResponse, 0),
	}

	// Parse rows
	for _, row := range rows {
		coLine := CustomerOrderLineResponse{}

		// Helper to safely get string values
		getString := func(key string) string {
			if val, ok := row[key].(string); ok {
				return val
			}
			return ""
		}

		// Helper to safely get int values
		getInt := func(key string) int {
			if val, ok := row[key].(float64); ok {
				return int(val)
			}
			if val, ok := row[key].(int); ok {
				return val
			}
			return 0
		}

		// Helper to safely get float values
		getFloat := func(key string) float64 {
			if val, ok := row[key].(float64); ok {
				return val
			}
			if val, ok := row[key].(int); ok {
				return float64(val)
			}
			return 0.0
		}

		coLine.OrderNumber = getString("order_number")
		coLine.LineNumber = getInt("line_number")
		coLine.LineSuffix = getInt("line_suffix")
		coLine.CFIN = getString("cfin")
		coLine.ItemNumber = getString("item_number")
		coLine.ItemDesc = getString("item_description")
		coLine.OrderQuantity = getFloat("order_quantity")
		coLine.OrderStatus = getString("order_status")
		coLine.Warehouse = getString("warehouse")
		coLine.RequestedDate = getInt("requested_date")
		coLine.ConfirmedDate = getInt("confirmed_date")
		coLine.PlannedDate = getInt("planned_date")
		coLine.CustomerNumber = getString("customer_number")

		response.CustomerLines = append(response.CustomerLines, coLine)
	}

	if !response.Found {
		response.Message = "No customer order lines found with this CFIN. The original order may be deleted, completed, or closed in M3."
	} else if len(response.CustomerLines) >= 100 {
		response.Message = "Results limited to 100 lines. This CFIN may have additional order lines."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
