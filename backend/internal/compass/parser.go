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
	// M3 Core Identifiers (all strings)
	CONO, DIVI, ORNO, PONR, POSX string

	// M3 Item Information
	ITNO, ITDS, TEDS, REPI string

	// M3 Status/Type
	ORST, ORTY string

	// M3 Facility/Warehouse
	FACI, WHLO string

	// M3 Quantities (all strings)
	ORQT, RNQT, ALQT, DLQT, IVQT string
	ORQA, RNQA, ALQA, DLQA, IVQA string

	// M3 Units
	ALUN, COFA, SPUN string

	// M3 Delivery Dates (all strings)
	DWDT, DWHM, CODT, COHM, PLDT, FDED, LDED string

	// M3 Pricing
	SAPR, NEPR, LNAM, CUCD string

	// M3 Discounts
	DIP1, DIP2, DIP3, DIP4, DIP5, DIP6 string
	DIA1, DIA2, DIA3, DIA4, DIA5, DIA6 string

	// M3 Reference Orders
	RORC, RORN, RORL, RORX string

	// M3 Customer References
	CUNO, CUOR, CUPO, CUSX string

	// Enrichment: Customer Name (from OCUSMA)
	CustomerName string

	// M3 Product/Model
	PRNO, HDPR, POPN, ALWT, ALWQ string

	// M3 Delivery/Route
	ADID, ROUT, RODN, DSDT, DSHM, MODL, TEDL, TEL2 string

	// M3 Packaging
	TEPA, PACT, CUPA string

	// M3 Partner/EDI
	E0PA, DSGP, PUSN, PUTP string

	// M3 Joint Delivery
	JDCD string

	// M3 Delivery Number
	DLIX string

	// M3 Order Type
	ORTP string

	// Enrichment: CO Type Description (from OOTYPE)
	COTypeDescription string

	// Enrichment: Delivery Method (from OOHEAD)
	DeliveryMethod string

	// M3 Attributes (ATV1-ATV0)
	ATV1, ATV2, ATV3, ATV4, ATV5 string
	ATV6, ATV7, ATV8, ATV9, ATV0 string

	// M3 User-Defined Alpha (UCA1-UCA0)
	UCA1, UCA2, UCA3, UCA4, UCA5 string
	UCA6, UCA7, UCA8, UCA9, UCA0 string

	// M3 User-Defined Numeric (UDN1-UDN6)
	UDN1, UDN2, UDN3, UDN4, UDN5, UDN6 string

	// M3 User-Defined Date (UID1-UID3)
	UID1, UID2, UID3 string

	// M3 User-Defined Text
	UCT1 string

	// M3 Configuration
	ATNR, ATMO, ATPR, CFIN string

	// M3 Project
	PROJ, ELNO string

	// M3 Audit
	RGDT, RGTM, LMDT, CHNO, CHID, LMTS string

	// Metadata
	Timestamp string
	Deleted   string
}

// ParseCustomerOrderLine converts a Compass record to a CustomerOrderLineRecord
// All fields extracted as strings to match Data Fabric source format
func ParseCustomerOrderLine(record map[string]interface{}) (*CustomerOrderLineRecord, error) {
	col := &CustomerOrderLineRecord{
		// Core Identifiers
		CONO: getStringFromAny(record, "CONO"),
		DIVI: getString(record, "DIVI"),
		ORNO: getString(record, "ORNO"),
		PONR: getStringFromAny(record, "PONR"),
		POSX: getStringFromAny(record, "POSX"),

		// Item Information
		ITNO: getString(record, "ITNO"),
		ITDS: getString(record, "ITDS"),
		TEDS: getString(record, "TEDS"),
		REPI: getString(record, "REPI"),

		// Status/Type
		ORST: getString(record, "ORST"),
		ORTY: getString(record, "ORTY"),

		// Facility/Warehouse
		FACI: getString(record, "FACI"),
		WHLO: getString(record, "WHLO"),

		// Quantities
		ORQT: getStringFromAny(record, "ORQT"),
		RNQT: getStringFromAny(record, "RNQT"),
		ALQT: getStringFromAny(record, "ALQT"),
		DLQT: getStringFromAny(record, "DLQT"),
		IVQT: getStringFromAny(record, "IVQT"),
		ORQA: getStringFromAny(record, "ORQA"),
		RNQA: getStringFromAny(record, "RNQA"),
		ALQA: getStringFromAny(record, "ALQA"),
		DLQA: getStringFromAny(record, "DLQA"),
		IVQA: getStringFromAny(record, "IVQA"),

		// Units
		ALUN: getString(record, "ALUN"),
		COFA: getStringFromAny(record, "COFA"),
		SPUN: getString(record, "SPUN"),

		// Delivery Dates
		DWDT: getStringFromAny(record, "DWDT"),
		DWHM: getStringFromAny(record, "DWHM"),
		CODT: getStringFromAny(record, "CODT"),
		COHM: getStringFromAny(record, "COHM"),
		PLDT: getStringFromAny(record, "PLDT"),
		FDED: getStringFromAny(record, "FDED"),
		LDED: getStringFromAny(record, "LDED"),

		// Pricing
		SAPR: getStringFromAny(record, "SAPR"),
		NEPR: getStringFromAny(record, "NEPR"),
		LNAM: getStringFromAny(record, "LNAM"),
		CUCD: getString(record, "CUCD"),

		// Discounts
		DIP1: getStringFromAny(record, "DIP1"),
		DIP2: getStringFromAny(record, "DIP2"),
		DIP3: getStringFromAny(record, "DIP3"),
		DIP4: getStringFromAny(record, "DIP4"),
		DIP5: getStringFromAny(record, "DIP5"),
		DIP6: getStringFromAny(record, "DIP6"),
		DIA1: getStringFromAny(record, "DIA1"),
		DIA2: getStringFromAny(record, "DIA2"),
		DIA3: getStringFromAny(record, "DIA3"),
		DIA4: getStringFromAny(record, "DIA4"),
		DIA5: getStringFromAny(record, "DIA5"),
		DIA6: getStringFromAny(record, "DIA6"),

		// Reference Orders
		RORC: getStringFromAny(record, "RORC"),
		RORN: getString(record, "RORN"),
		RORL: getStringFromAny(record, "RORL"),
		RORX: getStringFromAny(record, "RORX"),

		// Customer References
		CUNO: getString(record, "CUNO"),
		CUOR: getString(record, "CUOR"),
		CUPO: getStringFromAny(record, "CUPO"),
		CUSX: getStringFromAny(record, "CUSX"),

		// Enrichment: Customer Name
		CustomerName: getString(record, "customer_name"),

		// Product/Model
		PRNO: getString(record, "PRNO"),
		HDPR: getString(record, "HDPR"),
		POPN: getString(record, "POPN"),
		ALWT: getStringFromAny(record, "ALWT"),
		ALWQ: getString(record, "ALWQ"),

		// Delivery/Route
		ADID: getString(record, "ADID"),
		ROUT: getString(record, "ROUT"),
		RODN: getStringFromAny(record, "RODN"),
		DSDT: getStringFromAny(record, "DSDT"),
		DSHM: getStringFromAny(record, "DSHM"),
		MODL: getString(record, "MODL"),
		TEDL: getString(record, "TEDL"),
		TEL2: getString(record, "TEL2"),

		// Packaging
		TEPA: getString(record, "TEPA"),
		PACT: getString(record, "PACT"),
		CUPA: getString(record, "CUPA"),

		// Partner/EDI
		E0PA: getString(record, "E0PA"),
		DSGP: getString(record, "DSGP"),
		PUSN: getString(record, "PUSN"),
		PUTP: getStringFromAny(record, "PUTP"),

		// Joint Delivery
		JDCD: getString(record, "JDCD"),

		// Delivery Number
		DLIX: getString(record, "DLIX"),

		// Order Type
		ORTP: getString(record, "ORTP"),

		// Enrichment: CO Type Description
		COTypeDescription: getString(record, "co_type_description"),

		// Enrichment: Delivery Method
		DeliveryMethod: getString(record, "delivery_method"),

		// Attributes (ATV1-ATV0)
		ATV1: getStringFromAny(record, "ATV1"),
		ATV2: getStringFromAny(record, "ATV2"),
		ATV3: getStringFromAny(record, "ATV3"),
		ATV4: getStringFromAny(record, "ATV4"),
		ATV5: getStringFromAny(record, "ATV5"),
		ATV6: getString(record, "ATV6"),
		ATV7: getString(record, "ATV7"),
		ATV8: getString(record, "ATV8"),
		ATV9: getString(record, "ATV9"),
		ATV0: getString(record, "ATV0"),

		// User-Defined Alpha (UCA1-UCA0)
		UCA1: getString(record, "UCA1"),
		UCA2: getString(record, "UCA2"),
		UCA3: getString(record, "UCA3"),
		UCA4: getString(record, "UCA4"),
		UCA5: getString(record, "UCA5"),
		UCA6: getString(record, "UCA6"),
		UCA7: getString(record, "UCA7"),
		UCA8: getString(record, "UCA8"),
		UCA9: getString(record, "UCA9"),
		UCA0: getString(record, "UCA0"),

		// User-Defined Numeric (UDN1-UDN6)
		UDN1: getStringFromAny(record, "UDN1"),
		UDN2: getStringFromAny(record, "UDN2"),
		UDN3: getStringFromAny(record, "UDN3"),
		UDN4: getStringFromAny(record, "UDN4"),
		UDN5: getStringFromAny(record, "UDN5"),
		UDN6: getStringFromAny(record, "UDN6"),

		// User-Defined Date (UID1-UID3)
		UID1: getStringFromAny(record, "UID1"),
		UID2: getStringFromAny(record, "UID2"),
		UID3: getStringFromAny(record, "UID3"),

		// User-Defined Text
		UCT1: getString(record, "UCT1"),

		// Configuration
		ATNR: getStringFromAny(record, "ATNR"),
		ATMO: getString(record, "ATMO"),
		ATPR: getString(record, "ATPR"),
		CFIN: getStringFromAny(record, "CFIN"),

		// Project
		PROJ: getString(record, "PROJ"),
		ELNO: getString(record, "ELNO"),

		// Audit
		RGDT:      getStringFromAny(record, "RGDT"),
		RGTM:      getStringFromAny(record, "RGTM"),
		LMDT:      getStringFromAny(record, "LMDT"),
		CHNO:      getStringFromAny(record, "CHNO"),
		CHID:      getString(record, "CHID"),
		LMTS:      getStringFromAny(record, "LMTS"),
		Timestamp: getString(record, "timestamp"),
		Deleted:   getString(record, "deleted"),
	}

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
	MSTI int
	MFTI int
	FSTD int
	FFID int
	RSDT int
	REFD int
	RPDT int

	// Planning
	PRIO int
	RESP string
	PLGR string
	WCLN string
	PRDY int

	// Warehouse/Location
	WHLO string
	WHSL string
	BANO string

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

	// Material/BOM
	BDCD string
	SCEX string
	STRT string
	ECVE string

	// Routing
	AOID string
	NUOP int
	NUFO int

	// Action/Text
	ACTP string
	TXT1 string
	TXT2 string

	// Project
	PROJ string
	ELNO string

	// M3 metadata
	RGDT int
	RGTM int
	LMDT int
	CHNO int
	CHID string
	LMTS int64

	// Timestamps
	Timestamp int64
	Deleted   string

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
		MSTI: getInt(record, "MSTI"),
		MFTI: getInt(record, "MFTI"),
		FSTD: getInt(record, "FSTD"),
		FFID: getInt(record, "FFID"),
		RSDT: getInt(record, "RSDT"),
		REFD: getInt(record, "REFD"),
		RPDT: getInt(record, "RPDT"),

		PRIO: getInt(record, "PRIO"),
		RESP: getString(record, "RESP"),
		PLGR: getString(record, "PLGR"),
		WCLN: getString(record, "WCLN"),
		PRDY: getInt(record, "PRDY"),

		WHLO: getString(record, "WHLO"),
		WHSL: getString(record, "WHSL"),
		BANO: getString(record, "BANO"),

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

		BDCD: getString(record, "BDCD"),
		SCEX: getString(record, "SCEX"),
		STRT: getString(record, "STRT"),
		ECVE: getString(record, "ECVE"),

		AOID: getString(record, "AOID"),
		NUOP: getInt(record, "NUOP"),
		NUFO: getInt(record, "NUFO"),

		ACTP: getString(record, "ACTP"),
		TXT1: getString(record, "TXT1"),
		TXT2: getString(record, "TXT2"),

		PROJ: getString(record, "PROJ"),
		ELNO: getString(record, "ELNO"),

		RGDT: getInt(record, "RGDT"),
		RGTM: getInt(record, "RGTM"),
		LMDT: getInt(record, "LMDT"),
		CHNO: getInt(record, "CHNO"),
		CHID: getString(record, "CHID"),
		LMTS: getInt64(record, "LMTS"),

		Timestamp: getInt64(record, "timestamp"),
		Deleted:   getString(record, "deleted"),
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
	GETY string

	// Quantities
	PPQT float64
	ORQA float64

	// Dates
	RELD int
	STDT int
	FIDT int
	MSTI int
	MFTI int
	PLDT int

	// Planning
	RESP string
	PRIP int
	PLGR string
	WCLN string
	PRDY int

	// Warehouse
	WHLO string

	// Reference orders
	RORC int
	RORN string
	RORL int
	RORX int
	RORH string

	// Hierarchy
	PLLO string
	PLHL string

	// Planning parameters
	NUAU int
	ORDP string

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
	RGTM int
	LMDT int
	CHNO int
	CHID string
	LMTS int64

	// Timestamps
	Timestamp int64
	Deleted   string

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
		GETY: getString(record, "GETY"),

		PPQT: getFloat(record, "PPQT"),
		ORQA: getFloat(record, "ORQA"),

		RELD: getInt(record, "RELD"),
		STDT: getInt(record, "STDT"),
		FIDT: getInt(record, "FIDT"),
		MSTI: getInt(record, "MSTI"),
		MFTI: getInt(record, "MFTI"),
		PLDT: getInt(record, "PLDT"),

		RESP: getString(record, "RESP"),
		PRIP: getInt(record, "PRIP"),
		PLGR: getString(record, "PLGR"),
		WCLN: getString(record, "WCLN"),
		PRDY: getInt(record, "PRDY"),

		WHLO: getString(record, "WHLO"),

		RORC: getInt(record, "RORC"),
		RORN: getString(record, "RORN"),
		RORL: getInt(record, "RORL"),
		RORX: getInt(record, "RORX"),
		RORH: getString(record, "RORH"),

		PLLO: getString(record, "PLLO"),
		PLHL: getString(record, "PLHL"),

		NUAU: getInt(record, "NUAU"),
		ORDP: getString(record, "ORDP"),

		ATNR: getInt64(record, "ATNR"),
		CFIN: getInt64(record, "CFIN"),

		PROJ: getString(record, "PROJ"),
		ELNO: getString(record, "ELNO"),

		RGDT: getInt(record, "RGDT"),
		RGTM: getInt(record, "RGTM"),
		LMDT: getInt(record, "LMDT"),
		CHNO: getInt(record, "CHNO"),
		CHID: getString(record, "CHID"),
		LMTS: getInt64(record, "LMTS"),

		Timestamp: getInt64(record, "timestamp"),
		Deleted:   getString(record, "deleted"),
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

// getStringFromAny converts any value to string (handles int, float, string from Data Fabric)
func getStringFromAny(record map[string]interface{}, key string) string {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			if v == 0 {
				return ""
			}
			// Remove decimal if it's a whole number
			if v == float64(int64(v)) {
				return strconv.FormatInt(int64(v), 10)
			}
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			if v == 0 {
				return ""
			}
			return strconv.Itoa(v)
		case int64:
			if v == 0 {
				return ""
			}
			return strconv.FormatInt(v, 10)
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

// ========================================
// Public Helpers for External Packages
// ========================================

// GetString safely extracts a string value from a Compass record
func GetString(record map[string]interface{}, key string) string {
	return getString(record, key)
}

// GetInt safely extracts an integer value from a Compass record
func GetInt(record map[string]interface{}, key string) int {
	return getInt(record, key)
}

// GetInt64 safely extracts an int64 value from a Compass record
func GetInt64(record map[string]interface{}, key string) int64 {
	return getInt64(record, key)
}

// GetFloat safely extracts a float64 value from a Compass record
func GetFloat(record map[string]interface{}, key string) float64 {
	return getFloat(record, key)
}
