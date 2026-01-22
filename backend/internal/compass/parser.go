package compass

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// CompassResultSet represents the parsed Compass query results
type CompassResultSet struct {
	Records []map[string]interface{}
	Columns []string
}

// ParseResults parses raw Compass JSON results into a structured format
// Compass returns an array of row objects directly
func ParseResults(rawJSON []byte) (*CompassResultSet, error) {
	// First, try parsing as array of objects (actual Compass format)
	var rowObjects []map[string]interface{}
	if err := json.Unmarshal(rawJSON, &rowObjects); err == nil {
		// Extract column names from first row
		var columns []string
		if len(rowObjects) > 0 {
			for key := range rowObjects[0] {
				columns = append(columns, key)
			}
		}

		return &CompassResultSet{
			Records: rowObjects,
			Columns: columns,
		}, nil
	}

	// Fallback: try nested format (if API changes)
	var result struct {
		Data struct {
			Columns []string        `json:"columns"`
			Rows    [][]interface{} `json:"rows"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rawJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse compass results: %w", err)
	}

	// Convert rows to map[string]interface{} for easier access
	records := make([]map[string]interface{}, len(result.Data.Rows))
	for i, row := range result.Data.Rows {
		record := make(map[string]interface{})
		for j, col := range result.Data.Columns {
			if j < len(row) {
				record[col] = row[j]
			}
		}
		records[i] = record
	}

	return &CompassResultSet{
		Records: records,
		Columns: result.Data.Columns,
	}, nil
}

// CustomerOrderLineRecord represents a parsed CO line record
type CustomerOrderLineRecord struct {
	// Core identifiers
	CONO int
	DIVI string
	ORNO string
	PONR int
	POSX int

	// Item
	ITNO string
	ITDS string
	ORTY string
	ORST string

	// Location
	FACI string
	WHLO string

	// Quantities
	ORQT float64
	RNQT float64
	ALQT float64
	DLQT float64
	IVQT float64

	// Dates
	DWDT int
	CODT int
	PLDT int
	FDED int
	LDED int

	// Pricing
	SAPR float64
	NEPR float64
	LNAM float64
	CUCD string

	// Reference orders
	RORC int
	RORN string
	RORL int
	RORX int

	// Attribute model
	ATNR int64
	ATMO string

	// Attributes (will be built into JSONB)
	Attributes map[string]interface{}

	// M3 metadata
	RGDT int
	LMDT int
	LMTS int64
	CHID string

	// Customer
	CUNO string

	// Timestamps
	M3Timestamp string
	IsDeleted   bool
}

// ParseCustomerOrderLine converts a Compass record to a CustomerOrderLineRecord
func ParseCustomerOrderLine(record map[string]interface{}) (*CustomerOrderLineRecord, error) {
	col := &CustomerOrderLineRecord{
		CONO: getInt(record, "CONO"),
		DIVI: getString(record, "DIVI"),
		ORNO: getString(record, "ORNO"),
		PONR: getInt(record, "PONR"),
		POSX: getInt(record, "POSX"),

		ITNO: getString(record, "ITNO"),
		ITDS: getString(record, "ITDS"),
		ORTY: getString(record, "ORTY"),
		ORST: getString(record, "ORST"),

		FACI: getString(record, "FACI"),
		WHLO: getString(record, "WHLO"),

		ORQT: getFloat(record, "ORQT"),
		RNQT: getFloat(record, "RNQT"),
		ALQT: getFloat(record, "ALQT"),
		DLQT: getFloat(record, "DLQT"),
		IVQT: getFloat(record, "IVQT"),

		DWDT: getInt(record, "DWDT"),
		CODT: getInt(record, "CODT"),
		PLDT: getInt(record, "PLDT"),
		FDED: getInt(record, "FDED"),
		LDED: getInt(record, "LDED"),

		SAPR: getFloat(record, "SAPR"),
		NEPR: getFloat(record, "NEPR"),
		LNAM: getFloat(record, "LNAM"),
		CUCD: getString(record, "CUCD"),

		RORC: getInt(record, "RORC"),
		RORN: getString(record, "RORN"),
		RORL: getInt(record, "RORL"),
		RORX: getInt(record, "RORX"),

		ATNR: getInt64(record, "ATNR"),
		ATMO: getString(record, "ATMO"),

		CUNO: getString(record, "CUNO"),

		RGDT: getInt(record, "RGDT"),
		LMDT: getInt(record, "LMDT"),
		LMTS: getInt64(record, "LMTS"),
		CHID: getString(record, "CHID"),

		M3Timestamp: getString(record, "timestamp"),
		IsDeleted:   getString(record, "deleted") == "true",
	}

	// Build attributes JSONB
	col.Attributes = buildCOLineAttributes(record)

	return col, nil
}

// buildCOLineAttributes builds the JSONB attributes object from M3 fields
func buildCOLineAttributes(record map[string]interface{}) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Built-in numeric attributes
	builtinNumeric := make(map[string]interface{})
	for i := 1; i <= 5; i++ {
		fieldName := fmt.Sprintf("ATV%d", i)
		if val := getFloat(record, fieldName); val != 0 {
			builtinNumeric[fieldName] = val
		}
	}
	if len(builtinNumeric) > 0 {
		attributes["builtin_numeric"] = builtinNumeric
	}

	// Built-in string attributes
	builtinString := make(map[string]interface{})
	for _, i := range []string{"6", "7", "8", "9", "0"} {
		fieldName := fmt.Sprintf("ATV%s", i)
		if val := getString(record, fieldName); val != "" {
			builtinString[fieldName] = val
		}
	}
	if len(builtinString) > 0 {
		attributes["builtin_string"] = builtinString
	}

	// User-defined alpha fields
	userAlpha := make(map[string]interface{})
	for _, i := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		fieldName := fmt.Sprintf("UCA%s", i)
		if val := getString(record, fieldName); val != "" {
			userAlpha[fieldName] = val
		}
	}
	if len(userAlpha) > 0 {
		attributes["user_defined_alpha"] = userAlpha
	}

	// User-defined numeric fields
	userNumeric := make(map[string]interface{})
	for i := 1; i <= 6; i++ {
		fieldName := fmt.Sprintf("UDN%d", i)
		if val := getFloat(record, fieldName); val != 0 {
			userNumeric[fieldName] = val
		}
	}
	if len(userNumeric) > 0 {
		attributes["user_defined_numeric"] = userNumeric
	}

	// User-defined date fields
	userDates := make(map[string]interface{})
	for i := 1; i <= 3; i++ {
		fieldName := fmt.Sprintf("UID%d", i)
		if val := getInt(record, fieldName); val != 0 {
			userDates[fieldName] = val
		}
	}
	if len(userDates) > 0 {
		attributes["user_defined_dates"] = userDates
	}

	// User text field
	if uct1 := getString(record, "UCT1"); uct1 != "" {
		attributes["user_text"] = map[string]interface{}{"UCT1": uct1}
	}

	// Discount percentages
	discountPct := make(map[string]interface{})
	for i := 1; i <= 8; i++ {
		fieldName := fmt.Sprintf("DIP%d", i)
		if val := getFloat(record, fieldName); val != 0 {
			discountPct[fieldName] = val
		}
	}

	// Discount amounts
	discountAmt := make(map[string]interface{})
	for i := 1; i <= 8; i++ {
		fieldName := fmt.Sprintf("DIA%d", i)
		if val := getFloat(record, fieldName); val != 0 {
			discountAmt[fieldName] = val
		}
	}

	if len(discountPct) > 0 || len(discountAmt) > 0 {
		discounts := make(map[string]interface{})
		if len(discountPct) > 0 {
			discounts["percentages"] = discountPct
		}
		if len(discountAmt) > 0 {
			discounts["amounts"] = discountAmt
		}
		attributes["discounts"] = discounts
	}

	return attributes
}

// ManufacturingOrderRecord represents a parsed MO record
type ManufacturingOrderRecord struct {
	// Core identifiers
	CONO int
	DIVI string
	FACI string
	MFNO string
	PRNO string
	ITNO string

	// Status
	WHST string
	WHHS string
	WMST string
	MOHS string

	// Type and origin
	ORTY string
	GETP string

	// Quantities
	ORQT float64
	RVQT float64
	MAQT float64
	ORQA float64
	RVQA float64
	MAQA float64

	// Dates
	STDT int
	FIDT int
	RSDT int
	REFD int

	// Planning
	PRIO int
	RESP string
	PLGR string

	// Reference orders
	RORC int
	RORN string
	RORL int
	RORX int

	// Hierarchy
	PRHL string
	MFHL string
	PRLO string
	MFLO string
	LEVL int

	// Configuration
	CFIN int64
	ATNR int64

	// Project
	PROJ string
	ELNO string

	// M3 metadata
	RGDT int
	LMDT int
	LMTS int64

	// Timestamps
	M3Timestamp string
	IsDeleted   bool

	// Additional fields stored in attributes JSONB
	Attributes map[string]interface{}
}

// ParseManufacturingOrder converts a Compass record to a ManufacturingOrderRecord
func ParseManufacturingOrder(record map[string]interface{}) (*ManufacturingOrderRecord, error) {
	mo := &ManufacturingOrderRecord{
		CONO: getInt(record, "CONO"),
		DIVI: getString(record, "DIVI"),
		FACI: getString(record, "FACI"),
		MFNO: getString(record, "MFNO"),
		PRNO: getString(record, "PRNO"),
		ITNO: getString(record, "ITNO"),

		WHST: getString(record, "WHST"),
		WHHS: getString(record, "WHHS"),
		WMST: getString(record, "WMST"),
		MOHS: getString(record, "MOHS"),

		ORTY: getString(record, "ORTY"),
		GETP: getString(record, "GETP"),

		ORQT: getFloat(record, "ORQT"),
		RVQT: getFloat(record, "RVQT"),
		MAQT: getFloat(record, "MAQT"),
		ORQA: getFloat(record, "ORQA"),
		RVQA: getFloat(record, "RVQA"),
		MAQA: getFloat(record, "MAQA"),

		STDT: getInt(record, "STDT"),
		FIDT: getInt(record, "FIDT"),
		RSDT: getInt(record, "RSDT"),
		REFD: getInt(record, "REFD"),

		PRIO: getInt(record, "PRIO"),
		RESP: getString(record, "RESP"),
		PLGR: getString(record, "PLGR"),

		RORC: getInt(record, "RORC"),
		RORN: getString(record, "RORN"),
		RORL: getInt(record, "RORL"),
		RORX: getInt(record, "RORX"),

		PRHL: getString(record, "PRHL"),
		MFHL: getString(record, "MFHL"),
		PRLO: getString(record, "PRLO"),
		MFLO: getString(record, "MFLO"),
		LEVL: getInt(record, "LEVL"),

		CFIN: getInt64(record, "CFIN"),
		ATNR: getInt64(record, "ATNR"),

		PROJ: getString(record, "PROJ"),
		ELNO: getString(record, "ELNO"),

		RGDT: getInt(record, "RGDT"),
		LMDT: getInt(record, "LMDT"),
		LMTS: getInt64(record, "LMTS"),

		M3Timestamp: getString(record, "timestamp"),
		IsDeleted:   getString(record, "deleted") == "true",
	}

	// Build attributes JSONB for additional fields
	mo.Attributes = buildMOAttributes(record)

	return mo, nil
}

// buildMOAttributes builds the JSONB attributes object for MO
func buildMOAttributes(record map[string]interface{}) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Planning details
	planning := make(map[string]interface{})
	if val := getString(record, "WCLN"); val != "" {
		planning["production_line"] = val
	}
	if val := getFloat(record, "PRDY"); val != 0 {
		planning["production_days"] = val
	}
	if val := getString(record, "ACTP"); val != "" {
		planning["action_message"] = val
	}
	if len(planning) > 0 {
		attributes["planning"] = planning
	}

	// Routing details
	routing := make(map[string]interface{})
	if val := getString(record, "AOID"); val != "" {
		routing["alternative_routing"] = val
	}
	if val := getInt(record, "NUOP"); val != 0 {
		routing["num_operations"] = val
	}
	if val := getInt(record, "NUFO"); val != 0 {
		routing["finished_operations"] = val
	}
	if len(routing) > 0 {
		attributes["routing"] = routing
	}

	// Material/BOM
	material := make(map[string]interface{})
	if val := getString(record, "BDCD"); val != "" {
		material["explosion_method"] = val
	}
	if val := getString(record, "SCEX"); val != "" {
		material["subcontracting_exists"] = val
	}
	if val := getString(record, "STRT"); val != "" {
		material["structure_type"] = val
	}
	if val := getString(record, "ECVE"); val != "" {
		material["revision"] = val
	}
	if len(material) > 0 {
		attributes["material"] = material
	}

	// Text fields
	if txt1 := getString(record, "TXT1"); txt1 != "" {
		attributes["text_line_1"] = txt1
	}
	if txt2 := getString(record, "TXT2"); txt2 != "" {
		attributes["text_line_2"] = txt2
	}

	return attributes
}

// PlannedOrderRecord represents a parsed MOP record
type PlannedOrderRecord struct {
	// Core identifiers
	CONO int
	DIVI string
	FACI string
	PLPN int64
	PLPS int
	PRNO string
	ITNO string

	// Status
	PSTS string
	WHST string
	ACTP string

	// Type
	ORTY string

	// Quantities
	PPQT float64
	ORQA float64

	// Dates
	RELD int
	STDT int
	FIDT int
	PLDT int

	// Planning
	RESP string
	PRIP string
	PLGR string

	// Reference orders
	RORC int
	RORN string
	RORL int
	RORX int

	// Hierarchy
	PLLO int64
	PLHL int64

	// Configuration
	ATNR int64
	CFIN int64

	// Project
	PROJ string
	ELNO string

	// Messages
	Messages map[string]string

	// M3 metadata
	RGDT int
	LMDT int
	LMTS int64

	// Timestamps
	M3Timestamp string
	IsDeleted   bool

	// Additional fields stored in attributes
	Attributes map[string]interface{}
}

// ParsePlannedOrder converts a Compass record to a PlannedOrderRecord
func ParsePlannedOrder(record map[string]interface{}) (*PlannedOrderRecord, error) {
	mop := &PlannedOrderRecord{
		CONO: getInt(record, "CONO"),
		DIVI: getString(record, "DIVI"),
		FACI: getString(record, "FACI"),
		PLPN: getInt64(record, "PLPN"),
		PLPS: getInt(record, "PLPS"),
		PRNO: getString(record, "PRNO"),
		ITNO: getString(record, "ITNO"),

		PSTS: getString(record, "PSTS"),
		WHST: getString(record, "WHST"),
		ACTP: getString(record, "ACTP"),

		ORTY: getString(record, "ORTY"),

		PPQT: getFloat(record, "PPQT"),
		ORQA: getFloat(record, "ORQA"),

		RELD: getInt(record, "RELD"),
		STDT: getInt(record, "STDT"),
		FIDT: getInt(record, "FIDT"),
		PLDT: getInt(record, "PLDT"),

		RESP: getString(record, "RESP"),
		PRIP: getString(record, "PRIP"),
		PLGR: getString(record, "PLGR"),

		RORC: getInt(record, "RORC"),
		RORN: getString(record, "RORN"),
		RORL: getInt(record, "RORL"),
		RORX: getInt(record, "RORX"),

		PLLO: getInt64(record, "PLLO"),
		PLHL: getInt64(record, "PLHL"),

		ATNR: getInt64(record, "ATNR"),
		CFIN: getInt64(record, "CFIN"),

		PROJ: getString(record, "PROJ"),
		ELNO: getString(record, "ELNO"),

		RGDT: getInt(record, "RGDT"),
		LMDT: getInt(record, "LMDT"),
		LMTS: getInt64(record, "LMTS"),

		M3Timestamp: getString(record, "timestamp"),
		IsDeleted:   getString(record, "deleted") == "true",
	}

	// Build messages JSONB
	mop.Messages = buildMOPMessages(record)

	// Build attributes JSONB
	mop.Attributes = buildMOPAttributes(record)

	return mop, nil
}

// buildMOPMessages builds the messages JSONB object
func buildMOPMessages(record map[string]interface{}) map[string]string {
	messages := make(map[string]string)

	for i := 1; i <= 4; i++ {
		fieldName := fmt.Sprintf("MSG%d", i)
		if val := getString(record, fieldName); val != "" {
			messages[fieldName] = val
		}
	}

	return messages
}

// buildMOPAttributes builds the JSONB attributes object for MOP
func buildMOPAttributes(record map[string]interface{}) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Planning parameters
	planning := make(map[string]interface{})
	if val := getString(record, "WCLN"); val != "" {
		planning["production_line"] = val
	}
	if val := getFloat(record, "PRDY"); val != 0 {
		planning["production_days"] = val
	}
	if val := getString(record, "GETY"); val != "" {
		planning["origin_type"] = val
	}
	if val := getInt(record, "NUAU"); val != 0 {
		planning["number_auto_generated"] = val
	}
	if val := getString(record, "ORDP"); val != "" {
		planning["order_priority"] = val
	}
	if len(planning) > 0 {
		attributes["planning"] = planning
	}

	return attributes
}

// Helper functions to safely extract values from map

func getString(record map[string]interface{}, key string) string {
	if val, ok := record[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(record map[string]interface{}, key string) int {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		case string:
			// Parse string to int
			if parsed, err := strconv.Atoi(v); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func getInt64(record map[string]interface{}, key string) int64 {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		case string:
			// Parse string to int64
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func getFloat(record map[string]interface{}, key string) float64 {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getBool(record map[string]interface{}, key string) bool {
	if val, ok := record[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
		// Handle string "true"/"false"
		if str, ok := val.(string); ok {
			return str == "true"
		}
	}
	return false
}
