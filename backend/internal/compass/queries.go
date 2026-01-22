package compass

import (
	"fmt"
	"strings"
)

// QueryBuilder builds SQL queries for Compass Data Fabric
type QueryBuilder struct {
	lastSyncDate int // YYYYMMDD format
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(lastSyncDate int) *QueryBuilder {
	return &QueryBuilder{
		lastSyncDate: lastSyncDate,
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

// BuildManufacturingOrdersQuery builds the query for MWOHED (Manufacturing Orders)
// JOINs with MPREAL to get linked CO numbers directly
// Only fetches MOs that are Released or Planned (not yet started: WHST <= '20')
// For full refresh, use GetFullRefreshDate() as the lastSyncDate parameter
func (qb *QueryBuilder) BuildManufacturingOrdersQuery() string {
	// Select MO fields with mo. prefix
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

		// Reference orders (often NULL - that's why we use MPREAL!)
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

		// MPREAL linking fields (to get CO number)
		"mpreal.DRDN as linked_co_number",
		"mpreal.DRDL as linked_co_line",
		"mpreal.DRDX as linked_co_suffix",
		"mpreal.PQTY as allocated_qty",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MWOHED mo
LEFT JOIN MPREAL mpreal
  ON mpreal.ARDN = mo.MFNO
  AND mpreal.AOCA = '101'
  AND mpreal.DOCA = '311'
  AND mpreal.deleted = 'false'
WHERE mo.deleted = 'false'
  AND mo.LMDT >= %d
  AND mo.WHST <= '20'
ORDER BY mo.STDT, mo.LMDT
`, strings.Join(fields, ", "), qb.lastSyncDate)

	return strings.TrimSpace(query)
}

// BuildPlannedOrdersWithCOLinksQuery builds the query for MMOPLP with MPREAL joins
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

		// Reference orders (often NULL)
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

		// MPREAL linking fields (to get CO number)
		"mpreal.DRDN as linked_co_number",
		"mpreal.DRDL as linked_co_line",
		"mpreal.DRDX as linked_co_suffix",
		"mpreal.PQTY as allocated_qty",
	}

	query := fmt.Sprintf(`
SELECT %s
FROM MMOPLP mop
LEFT JOIN MPREAL mpreal
  ON mop.PLPN = CAST(mpreal.ARDN AS BIGINT)
  AND mpreal.AOCA = '100'
  AND mpreal.DOCA = '311'
  AND mpreal.deleted = 'false'
WHERE mop.deleted = 'false'
  AND mop.LMDT >= %d
  AND mop.PSTS IN ('10', '20')
ORDER BY mop.PLDT, mop.LMDT
`, strings.Join(fields, ", "), qb.lastSyncDate)

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
// An open CO line is one that has not been fully allocated (status < 30)
func (qb *QueryBuilder) BuildOpenCustomerOrderLinesQuery() string {
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

	query := fmt.Sprintf(`
SELECT %s
FROM OOLINE
WHERE deleted = 'false'
  AND ORST < '30'
ORDER BY ORNO, PONR, POSX
`, strings.Join(fields, ", "))

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
