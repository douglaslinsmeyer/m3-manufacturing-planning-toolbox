package compass

import (
	"fmt"
	"strings"
)

// QueryBuilder builds SQL queries for Compass Data Fabric
type QueryBuilder struct {
	lastSyncDate int    // YYYYMMDD format
	company      string // Company number (CONO)
	facility     string // Facility (FACI)
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(lastSyncDate int, company string, facility string) *QueryBuilder {
	return &QueryBuilder{
		lastSyncDate: lastSyncDate,
		company:      company,
		facility:     facility,
	}
}

// BuildCustomerOrderLinesQuery builds the query for OOLINE (Customer Order Lines)
func (qb *QueryBuilder) BuildCustomerOrderLinesQuery() string {
	fields := []string{
		// Core identifiers
		"CONO", "DIVI", "ORNO", "PONR", "POSX",

		// Item information
		"ITNO", "ITDS", "ORTY", "ORST",

		// Location
		"FACI", "WHLO",

		// Quantities (basic U/M)
		"ORQT", "RNQT", "ALQT", "DLQT", "IVQT",

		// Quantities (alternate U/M)
		"ORQA", "RNQA", "ALQA", "DLQA", "IVQA",

		// Unit of measure
		"ALUN", "COFA", "SPUN",

		// Dates - requested
		"DWDT", "DWHM",

		// Dates - confirmed
		"CODT", "COHM",

		// Dates - planning
		"PLDT", "FDED", "LDED",

		// Pricing
		"SAPR", "NEPR", "LNAM", "CUCD",

		// Discounts (percentage) - only first 6 exist in this environment
		"DIP1", "DIP2", "DIP3", "DIP4", "DIP5", "DIP6",

		// Discounts (amount) - only first 6 exist in this environment
		"DIA1", "DIA2", "DIA3", "DIA4", "DIA5", "DIA6",

		// Reference orders (CRITICAL for linking!)
		"RORC", "RORN", "RORL", "RORX",

		// Built-in attributes (numeric)
		"ATV1", "ATV2", "ATV3", "ATV4", "ATV5",

		// Built-in attributes (string)
		"ATV6", "ATV7", "ATV8", "ATV9", "ATV0",

		// User-defined attributes (alpha)
		"UCA1", "UCA2", "UCA3", "UCA4", "UCA5",
		"UCA6", "UCA7", "UCA8", "UCA9", "UCA0",

		// User-defined attributes (numeric)
		"UDN1", "UDN2", "UDN3", "UDN4", "UDN5", "UDN6",

		// User-defined attributes (dates)
		"UID1", "UID2", "UID3",

		// User-defined attributes (text)
		"UCT1",

		// Attribute model
		"ATNR", "ATMO", "ATPR",

		// Configuration
		"CFIN",

		// Customer information
		"CUNO",

		// M3 audit fields
		"RGDT", "RGTM", "LMDT", "CHNO", "CHID", "LMTS",

		// Data Lake metadata
		"timestamp", "deleted",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM OOLINE
WHERE deleted = 'false'
  AND LMDT >= %d
ORDER BY LMDT, LMTS
`, strings.Join(fields, ", "), qb.lastSyncDate)

	return strings.TrimSpace(query)
}

// BuildManufacturingOrdersQuery builds the query for MWOHED with MPREAL supply chain resolution
// Uses SCNB (Supply Chain Number) to handle multi-level chains of any depth: MO → DO → ... → CO
// All records in a supply chain share the same SCNB, so we self-join MPREAL to find the CO link
// Only fetches MOs that are Released or Planned (not yet started: WHST <= '20')
// Filtered by company and facility context
// For full refresh, use GetFullRefreshDate() as the lastSyncDate parameter
func (qb *QueryBuilder) BuildManufacturingOrdersQuery() string {
	fields := []string{
		// Core identifiers
		"mo.CONO", "mo.DIVI", "mo.FACI", "mo.MFNO", "mo.PRNO", "mo.ITNO",

		// Status
		"mo.WHST", "mo.WHHS", "mo.WMST", "mo.MOHS",

		// Order type and origin
		"mo.ORTY", "mo.GETP",

		// Quantities
		"mo.ORQT", "mo.RVQT", "mo.MAQT", "mo.ORQA", "mo.RVQA", "mo.MAQA",

		// Dates
		"mo.STDT", "mo.FIDT", "mo.MSTI", "mo.MFTI", "mo.FSTD", "mo.FFID",
		"mo.RSDT", "mo.REFD", "mo.RPDT",

		// Planning
		"mo.PRIO", "mo.RESP", "mo.PLGR", "mo.WCLN", "mo.PRDY",

		// Warehouse
		"mo.WHLO", "mo.WHSL", "mo.BANO",

		// Reference orders
		"mo.RORC", "mo.RORN", "mo.RORL", "mo.RORX",

		// Hierarchy
		"mo.PRHL", "mo.MFHL", "mo.PRLO", "mo.MFLO", "mo.LEVL",

		// Configuration
		"mo.CFIN", "mo.ATNR",

		// Material
		"mo.BDCD", "mo.SCEX", "mo.STRT", "mo.ECVE",

		// Project
		"mo.PROJ", "mo.ELNO",

		// Routing
		"mo.AOID", "mo.NUOP", "mo.NUFO",

		// Action
		"mo.ACTP",

		// Text
		"mo.TXT1", "mo.TXT2",

		// M3 audit
		"mo.RGDT", "mo.RGTM", "mo.LMDT", "mo.CHNO", "mo.CHID", "mo.LMTS",

		// Data Lake
		"mo.timestamp", "mo.deleted",

		// CO link (direct or indirect via DO/PO)
		"COALESCE(mpreal_direct.DRDN, co_link.DRDN) as linked_co_number",
		"COALESCE(mpreal_direct.DRDL, co_link.DRDL) as linked_co_line",
		"COALESCE(mpreal_direct.DRDX, co_link.DRDX) as linked_co_suffix",
		"COALESCE(mpreal_direct.PQTY, co_link.PQTY) as allocated_qty",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MWOHED mo
-- Direct link: MO → CO
LEFT JOIN MPREAL mpreal_direct
  ON mpreal_direct.ARDN = mo.MFNO
  AND mpreal_direct.AOCA = '101'
  AND mpreal_direct.DOCA = '311'
  AND mpreal_direct.deleted = 'false'
-- Indirect link step 1: MO → DO/PO
LEFT JOIN MPREAL mpreal_mo
  ON mpreal_mo.ARDN = mo.MFNO
  AND mpreal_mo.AOCA = '101'
  AND mpreal_mo.DOCA IN ('510', '511')
  AND mpreal_mo.deleted = 'false'
  AND mpreal_direct.DRDN IS NULL
-- Indirect link step 2: DO/PO → CO
LEFT JOIN MPREAL co_link
  ON co_link.ARDN = mpreal_mo.DRDN
  AND co_link.AOCA IN ('500', '501')
  AND co_link.DOCA = '311'
  AND co_link.deleted = 'false'
WHERE mo.deleted = 'false'
  AND mo.LMDT >= %d
  AND mo.WHST <= '20'
  AND mo.CONO = '%s'
  AND mo.FACI = '%s'
ORDER BY mo.STDT, mo.LMDT
`, strings.Join(fields, ", "), qb.lastSyncDate, qb.company, qb.facility)

	return strings.TrimSpace(query)
}

// BuildPlannedOrdersWithCOLinksQuery builds the query for MMOPLP with MPREAL supply chain resolution
// Uses SCNB (Supply Chain Number) to handle multi-level chains of any depth: MOP → DO → ... → CO
// All records in a supply chain share the same SCNB, so we self-join MPREAL to find the CO link
// Filtered by company and facility context, only includes firmed planned orders (PSTS = '20')
// For full refresh, use GetFullRefreshDate() as the lastSyncDate parameter
func (qb *QueryBuilder) BuildPlannedOrdersWithCOLinksQuery() string {
	fields := []string{
		// Core identifiers
		"mop.CONO", "mop.FACI", "mop.PLPN", "mop.PLPS", "mop.PRNO", "mop.ITNO",

		// Status
		"mop.PSTS", "mop.WHST", "mop.ACTP",

		// Type
		"mop.ORTY", "mop.GETY",

		// Quantities
		"mop.PPQT", "mop.ORQA",

		// Dates
		"mop.RELD", "mop.STDT", "mop.FIDT", "mop.MSTI", "mop.MFTI", "mop.PLDT",

		// Planning
		"mop.RESP", "mop.PRIP", "mop.PLGR", "mop.WCLN", "mop.PRDY",

		// Warehouse
		"mop.WHLO",

		// Reference orders
		"mop.RORC", "mop.RORN", "mop.RORL", "mop.RORX", "mop.RORH",

		// Hierarchy
		"mop.PLLO", "mop.PLHL",

		// Configuration
		"mop.ATNR", "mop.CFIN",

		// Project
		"mop.PROJ", "mop.ELNO",

		// Messages
		"mop.MSG1", "mop.MSG2", "mop.MSG3", "mop.MSG4",

		// Planning params
		"mop.NUAU", "mop.ORDP",

		// M3 audit
		"mop.RGDT", "mop.RGTM", "mop.LMDT", "mop.CHNO", "mop.CHID", "mop.LMTS",

		// Data Lake
		"mop.timestamp", "mop.deleted",

		// CO link (direct or indirect via DO/PO)
		"COALESCE(mpreal_direct.DRDN, co_link.DRDN) as linked_co_number",
		"COALESCE(mpreal_direct.DRDL, co_link.DRDL) as linked_co_line",
		"COALESCE(mpreal_direct.DRDX, co_link.DRDX) as linked_co_suffix",
		"COALESCE(mpreal_direct.PQTY, co_link.PQTY) as allocated_qty",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MMOPLP mop
-- Direct link: MOP → CO
LEFT JOIN MPREAL mpreal_direct
  ON mop.PLPN = CAST(mpreal_direct.ARDN AS BIGINT)
  AND mpreal_direct.AOCA = '100'
  AND mpreal_direct.DOCA = '311'
  AND mpreal_direct.deleted = 'false'
-- Indirect link step 1: MOP → DO/PO
LEFT JOIN MPREAL mpreal_mop
  ON mop.PLPN = CAST(mpreal_mop.ARDN AS BIGINT)
  AND mpreal_mop.AOCA = '100'
  AND mpreal_mop.DOCA IN ('510', '511')
  AND mpreal_mop.deleted = 'false'
  AND mpreal_direct.DRDN IS NULL
-- Indirect link step 2: DO/PO → CO
LEFT JOIN MPREAL co_link
  ON co_link.ARDN = mpreal_mop.DRDN
  AND co_link.AOCA IN ('500', '501')
  AND co_link.DOCA = '311'
  AND co_link.deleted = 'false'
WHERE mop.deleted = 'false'
  AND mop.LMDT >= %d
  AND mop.PSTS = '20'
  AND mop.CONO = '%s'
  AND mop.FACI = '%s'
ORDER BY mop.PLDT, mop.LMDT
`, strings.Join(fields, ", "), qb.lastSyncDate, qb.company, qb.facility)

	return strings.TrimSpace(query)
}

// BuildPlannedOrdersQuery builds the query for MMOPLP (Planned Manufacturing Orders)
func (qb *QueryBuilder) BuildPlannedOrdersQuery() string {
	fields := []string{
		// Core identifiers (DIVI doesn't exist in MMOPLP)
		"CONO", "FACI", "PLPN", "PLPS", "PRNO", "ITNO",

		// Status
		"PSTS", "WHST", "ACTP",

		// Order type
		"ORTY", "GETY",

		// Quantities
		"PPQT", "ORQA",

		// Dates
		"RELD", "STDT", "FIDT", "MSTI", "MFTI", "PLDT",

		// Planning
		"RESP", "PRIP", "PLGR", "WCLN", "PRDY",

		// Warehouse
		"WHLO",

		// Reference orders (CRITICAL for linking!)
		"RORC", "RORN", "RORL", "RORX", "RORH",

		// Hierarchy
		"PLLO", "PLHL",

		// Configuration
		"ATNR", "CFIN",

		// Project
		"PROJ", "ELNO",

		// Messages
		"MSG1", "MSG2", "MSG3", "MSG4",

		// Planning parameters
		"NUAU", "ORDP",

		// M3 audit fields
		"RGDT", "RGTM", "LMDT", "CHNO", "CHID", "LMTS",

		// Data Lake metadata
		"timestamp", "deleted",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MMOPLP
WHERE deleted = 'false'
  AND LMDT >= %d
ORDER BY LMDT, LMTS
`, strings.Join(fields, ", "), qb.lastSyncDate)

	return strings.TrimSpace(query)
}

// BuildCustomerOrderLinesByOrderNumbersQuery builds a targeted query for specific CO numbers
// Only fetches CO lines referenced by the MOPs/MOs we loaded (via MPREAL)
// DEPRECATED: This approach causes issues when there are many order numbers
// Use BuildOpenCustomerOrderLinesQuery instead
func (qb *QueryBuilder) BuildCustomerOrderLinesByOrderNumbersQuery(orderNumbers []string) string {
	if len(orderNumbers) == 0 {
		return ""
	}

	fields := []string{
		"CONO", "DIVI", "ORNO", "PONR", "POSX",
		"ITNO", "ITDS", "ORTY", "ORST",
		"FACI", "WHLO",
		"ORQT", "RNQT", "ALQT", "DLQT", "IVQT",
		"ORQA", "RNQA", "ALQA", "DLQA", "IVQA",
		"ALUN", "COFA", "SPUN",
		"DWDT", "DWHM", "CODT", "COHM", "PLDT", "FDED", "LDED",
		"SAPR", "NEPR", "LNAM", "CUCD",
		"DIP1", "DIP2", "DIP3", "DIP4", "DIP5", "DIP6",
		"DIA1", "DIA2", "DIA3", "DIA4", "DIA5", "DIA6",
		"RORC", "RORN", "RORL", "RORX",
		"ATV1", "ATV2", "ATV3", "ATV4", "ATV5",
		"ATV6", "ATV7", "ATV8", "ATV9", "ATV0",
		"UCA1", "UCA2", "UCA3", "UCA4", "UCA5",
		"UCA6", "UCA7", "UCA8", "UCA9", "UCA0",
		"UDN1", "UDN2", "UDN3", "UDN4", "UDN5", "UDN6",
		"UID1", "UID2", "UID3",
		"UCT1",
		"ATNR", "ATMO", "ATPR",
		"CFIN",
		"CUNO",
		"RGDT", "RGTM", "LMDT", "CHNO", "CHID", "LMTS",
		"timestamp", "deleted",
	}

	// Build IN clause (quote all order numbers)
	quotedOrders := make([]string, len(orderNumbers))
	for i, orderNum := range orderNumbers {
		quotedOrders[i] = fmt.Sprintf("'%s'", orderNum)
	}

	query := fmt.Sprintf(`
SELECT %s
FROM OOLINE
WHERE deleted = 'false'
  AND ORNO IN (%s)
ORDER BY ORNO, PONR, POSX
`, strings.Join(fields, ", "), strings.Join(quotedOrders, ", "))

	return strings.TrimSpace(query)
}

// BuildOpenCustomerOrderLinesQuery builds a query for all open CO lines
// An open CO line is Reserved (status >= 20 and < 30) - excludes quotations/preliminary
// Filtered by company and facility context
func (qb *QueryBuilder) BuildOpenCustomerOrderLinesQuery() string {
	fields := []string{
		// Core identifiers
		"CONO", "DIVI", "ORNO", "PONR", "POSX",

		// Item information (ALL)
		"ITNO", "ITDS", "TEDS", "REPI",

		// Status/Type
		"ORST", "ORTY",

		// Facility/Warehouse
		"FACI", "WHLO",

		// Quantities - Basic U/M
		"ORQT", "RNQT", "ALQT", "DLQT", "IVQT",

		// Quantities - Alternate U/M
		"ORQA", "RNQA", "ALQA", "DLQA", "IVQA",

		// Units
		"ALUN", "COFA", "SPUN",

		// Delivery Dates (critical for planning)
		"DWDT", "DWHM", "CODT", "COHM", "PLDT", "FDED", "LDED",

		// Pricing (basic)
		"SAPR", "NEPR", "LNAM", "CUCD",

		// Discounts (6 main ones)
		"DIP1", "DIP2", "DIP3", "DIP4", "DIP5", "DIP6",
		"DIA1", "DIA2", "DIA3", "DIA4", "DIA5", "DIA6",

		// Reference Orders (critical for linking)
		"RORC", "RORN", "RORL", "RORX",

		// Customer References (ALL)
		"CUNO", "CUOR", "CUPO", "CUSX",

		// Enrichment: Customer Name (from OCUSMA)
		"customer_name",

		// Product/Model (ALL)
		"PRNO", "HDPR", "POPN", "ALWT", "ALWQ",

		// Delivery/Route (ALL)
		"ADID", "ROUT", "RODN", "DSDT", "DSHM", "MODL", "TEDL", "TEL2",

		// Packaging (ALL)
		"TEPA", "PACT", "CUPA",

		// Partner/EDI (ALL)
		"E0PA", "DSGP", "PUSN", "PUTP",

		// Joint Delivery
		"JDCD",

		// Delivery Number (from MHDISL)
		"DLIX",

		// Order Type (from OOHEAD)
		"ORTP",

		// Enrichment: CO Type Description (from OOTYPE)
		"co_type_description",

		// Enrichment: Delivery Method (from OOHEAD)
		"delivery_method",

		// Attributes (ATV1-ATV0)
		"ATV1", "ATV2", "ATV3", "ATV4", "ATV5",
		"ATV6", "ATV7", "ATV8", "ATV9", "ATV0",

		// User-Defined Fields (often used for custom workflows)
		"UCA1", "UCA2", "UCA3", "UCA4", "UCA5",
		"UCA6", "UCA7", "UCA8", "UCA9", "UCA0",
		"UDN1", "UDN2", "UDN3", "UDN4", "UDN5", "UDN6",
		"UID1", "UID2", "UID3",
		"UCT1",

		// Configuration (for configured products)
		"ATNR", "ATMO", "ATPR", "CFIN",

		// Project
		"PROJ", "ELNO",

		// M3 Audit
		"RGDT", "RGTM", "LMDT", "CHNO", "CHID", "LMTS",

		// Data Lake
		"timestamp", "deleted",
	}

	// Build field list with table aliases
	var fieldList []string
	for _, field := range fields {
		// Special handling for fields from joined tables
		if field == "DLIX" {
			fieldList = append(fieldList, "dl.DLIX")
		} else if field == "ORTP" {
			fieldList = append(fieldList, "oh.ORTP")
		} else if field == "customer_name" {
			fieldList = append(fieldList, "cu.CUNM as customer_name")
		} else if field == "co_type_description" {
			fieldList = append(fieldList, "ootype.TX40 as co_type_description")
		} else if field == "delivery_method" {
			fieldList = append(fieldList, "oh.MODL as delivery_method")
		} else {
			// All other fields come from OOLINE with ol. prefix
			fieldList = append(fieldList, "ol."+field)
		}
	}

	query := fmt.Sprintf(`
SELECT %s
FROM OOLINE ol
LEFT JOIN MHDISL dl
  ON ol.ORNO = dl.RIDN
  AND ol.PONR = dl.RIDL
  AND ol.POSX = dl.RIDX
  AND dl.deleted = 'false'
LEFT JOIN OOHEAD oh
  ON ol.ORNO = oh.ORNO
  AND oh.deleted = 'false'
LEFT JOIN OCUSMA cu
  ON ol.CUNO = cu.CUNO
  AND cu.deleted = 'false'
LEFT JOIN OOTYPE ootype
  ON oh.ORTP = ootype.ORTP
  AND ootype.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORST >= '20'
  AND ol.ORST < '30'
  AND ol.CONO = '%s'
  AND ol.FACI = '%s'
ORDER BY ol.ORNO, ol.PONR, ol.POSX
`, strings.Join(fieldList, ", "), qb.company, qb.facility)

	return strings.TrimSpace(query)
}

// BuildCustomerOrdersQuery builds the query for OOHEAD (Customer Order Header)
func (qb *QueryBuilder) BuildCustomerOrdersQuery() string {
	fields := []string{
		// Core identifiers
		"CONO", "DIVI", "ORNO",

		// Customer
		"CUNO", "CUNM",

		// Order information
		"ORTY", "ORDT",

		// Dates
		"RLDT", "CODT",

		// Status
		"ORST",

		// Financial
		"CUCD", "NTAM",

		// Warehouse
		"WHLO",

		// Sales
		"SMCD",

		// M3 audit fields
		"RGDT", "LMDT", "LMTS",

		// Data Lake metadata
		"timestamp", "deleted",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM OOHEAD
WHERE deleted = 'false'
  AND LMDT >= %d
ORDER BY LMDT, LMTS
`, strings.Join(fields, ", "), qb.lastSyncDate)

	return strings.TrimSpace(query)
}

// BuildMPREALQuery builds the query for MPREAL (Pre-Allocation/Supply Chain Links)
// Loads all pre-allocation records for supply chain resolution
// Filtered by company context
// For full refresh, use GetFullRefreshDate() as the lastSyncDate parameter
func (qb *QueryBuilder) BuildMPREALQuery() string {
	fields := []string{
		// Core identifiers
		"CONO", "WHLO", "ITNO",

		// Acquisition (source) order
		"AOCA", "ARDN", "ARDL", "ARDX",

		// Demand (destination) order
		"DOCA", "DRDN", "DRDL", "DRDX",

		// Quantity
		"PQTY", "PQTR",

		// Supply Chain Number (CRITICAL for multi-level linking!)
		"SCNB",

		// Planning
		"RESP", "PATY",

		// M3 audit
		"RGDT", "RGTM", "LMDT", "CHNO", "CHID", "LMTS",

		// Data Lake
		"timestamp", "deleted",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MPREAL
WHERE deleted = 'false'
  AND LMDT >= %d
  AND CONO = '%s'
ORDER BY SCNB, AOCA, DOCA
`, strings.Join(fields, ", "), qb.lastSyncDate, qb.company)

	return strings.TrimSpace(query)
}

// GetFullRefreshDate returns a date far in the past for full refresh
func GetFullRefreshDate() int {
	return 20200101 // January 1, 2020
}

// FormatDateForQuery converts a YYYYMMDD integer to a string for SQL
func FormatDateForQuery(date int) string {
	return fmt.Sprintf("%d", date)
}

// ParseM3Date converts M3 date format (YYYYMMDD int) to a date string for PostgreSQL
func ParseM3Date(m3Date int) string {
	if m3Date == 0 {
		return ""
	}
	year := m3Date / 10000
	month := (m3Date % 10000) / 100
	day := m3Date % 100
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// ParseM3Time converts M3 time format (HHMM or HHMMSS int) to a time string
func ParseM3Time(m3Time int) string {
	if m3Time == 0 {
		return ""
	}

	if m3Time < 10000 {
		// HHMM format
		hour := m3Time / 100
		minute := m3Time % 100
		return fmt.Sprintf("%02d:%02d:00", hour, minute)
	}

	// HHMMSS format
	hour := m3Time / 10000
	minute := (m3Time % 10000) / 100
	second := m3Time % 100
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

