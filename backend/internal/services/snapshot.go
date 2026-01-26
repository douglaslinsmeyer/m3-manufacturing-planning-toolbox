package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// ProgressCallback is called to report progress during refresh operations
// Parameters: phase, stepNum, totalSteps, message, mopCount, moCount, coCount
type ProgressCallback func(phase string, stepNum, totalSteps int, message string, mopCount, moCount, coCount int)

// SnapshotService handles data refresh operations
type SnapshotService struct {
	compassClient    *compass.Client
	db               *db.Queries
	progressCallback ProgressCallback
	// Track counts for progress reporting
	mopCount int
	moCount  int
	coCount  int
}

// NewSnapshotService creates a new snapshot service
func NewSnapshotService(compassClient *compass.Client, database *db.Queries) *SnapshotService {
	return &SnapshotService{
		compassClient: compassClient,
		db:            database,
	}
}

// SetProgressCallback sets the callback function for progress updates
func (s *SnapshotService) SetProgressCallback(callback ProgressCallback) {
	s.progressCallback = callback
}

// reportProgress calls the progress callback if set
func (s *SnapshotService) reportProgress(phase string, stepNum, totalSteps int, message string) {
	if s.progressCallback != nil {
		s.progressCallback(phase, stepNum, totalSteps, message, s.mopCount, s.moCount, s.coCount)
	}
}

// RefreshAll performs a full data refresh from M3 with table truncation
// Strategy: Clear all M3 snapshot data, then pull MOPs/MOs with MPREAL links, then all open CO lines
// Filtered by company and facility context
func (s *SnapshotService) RefreshAll(ctx context.Context, company string, facility string) error {
	log.Printf("Starting full data refresh from M3 for company '%s' and facility '%s'...", company, facility)
	log.Println("Strategy: Truncate all M3 snapshot tables, load MOPs/MOs with MPREAL links, load all open CO lines")

	// Phase 0: Truncate all M3 snapshot tables
	s.reportProgress("truncate", 0, 5, "Preparing database for refresh")
	log.Println("Phase 0: Truncating all M3 snapshot tables for full refresh...")
	if err := s.db.TruncateAnalysisTables(ctx); err != nil {
		return fmt.Errorf("failed to truncate analysis tables: %w", err)
	}
	log.Println("✓ M3 snapshot tables truncated successfully")
	s.reportProgress("truncate", 1, 5, "Database prepared")

	// Phase 1: Load MOPs with CO links
	s.reportProgress("mops", 1, 5, "Loading planned manufacturing orders")
	log.Println("Phase 1: Refreshing planned manufacturing orders (MOPs) with CO links...")
	mopRefs, err := s.RefreshPlannedOrders(ctx, company, facility)
	if err != nil {
		return fmt.Errorf("failed to refresh MOPs: %w", err)
	}
	s.mopCount = len(mopRefs)
	s.reportProgress("mops", 1, 5, fmt.Sprintf("Loaded %d planned orders", s.mopCount))

	// Phase 2: Load MOs with CO links
	s.reportProgress("mos", 2, 5, "Loading manufacturing orders")
	log.Println("Phase 2: Refreshing manufacturing orders (MOs) with CO links...")
	moRefs, err := s.RefreshManufacturingOrders(ctx, company, facility)
	if err != nil {
		return fmt.Errorf("failed to refresh MOs: %w", err)
	}
	s.moCount = len(moRefs)
	s.reportProgress("mos", 2, 5, fmt.Sprintf("Loaded %d manufacturing orders", s.moCount))

	// Phase 3: Load all open CO lines (status < 30)
	s.reportProgress("cos", 3, 5, "Loading customer order lines")
	log.Println("Phase 3: Refreshing all open customer order lines (status < 30)...")
	coCount, err := s.RefreshOpenCustomerOrderLines(ctx, company, facility)
	if err != nil {
		return fmt.Errorf("failed to refresh CO lines: %w", err)
	}
	s.coCount = coCount
	s.reportProgress("cos", 3, 5, fmt.Sprintf("Loaded %d customer order lines", s.coCount))

	// Phase 4: Update unified production_orders view
	s.reportProgress("finalize", 4, 5, "Finalizing data refresh")
	log.Println("Phase 4: Updating unified production orders view...")
	if err := s.db.UpdateProductionOrdersFromMOPs(ctx); err != nil {
		return fmt.Errorf("failed to update production orders from MOPs: %w", err)
	}
	if err := s.db.UpdateProductionOrdersFromMOs(ctx); err != nil {
		return fmt.Errorf("failed to update production orders from MOs: %w", err)
	}

	// Phase 5: Run issue detection
	s.reportProgress("detection", 5, 5, "Running issue detectors")
	log.Println("Phase 5: Running issue detectors...")

	detectionService := NewDetectionService(s.db)
	detectionService.SetProgressCallback(s.progressCallback)

	// Get current job ID
	jobID, err := s.db.GetLatestRefreshJobID(ctx)
	if err != nil {
		log.Printf("Warning: Could not get job ID for detection: %v", err)
		// Continue without detection rather than failing the entire refresh
	} else {
		if err := detectionService.RunAllDetectors(ctx, jobID, company, facility); err != nil {
			log.Printf("Warning: Issue detection failed: %v", err)
			// Don't fail the entire refresh if detection fails
		}
	}

	s.reportProgress("complete", 5, 5, "Data refresh and detection completed")
	log.Printf("✓ Full data refresh and detection completed successfully for company '%s' and facility '%s'", company, facility)
	return nil
}

// RefreshOpenCustomerOrderLines refreshes all open CO lines (status < 30)
// Filtered by company and facility context
// This is more efficient than querying by specific order numbers when there are many orders
// Returns the count of records processed
func (s *SnapshotService) RefreshOpenCustomerOrderLines(ctx context.Context, company string, facility string) (int, error) {
	log.Printf("Refreshing all open customer order lines (status < 30) for company '%s' and facility '%s'...", company, facility)

	// Build query for all open CO lines with context filters
	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildOpenCustomerOrderLinesQuery()

	// Execute query
	log.Println("Submitting Compass query for open CO lines...")
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total CO line records", totalRecords)

	// Parse results
	log.Println("Parsing CO line results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d CO line records", len(resultSet.Records))

	// Transform to database records
	dbRecords := make([]*db.CustomerOrderLine, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		coLine, err := compass.ParseCustomerOrderLine(record)
		if err != nil {
			log.Printf("Warning: failed to parse CO line record: %v", err)
			continue
		}

		// Map all fields directly - all stored as strings
		dbRecord := &db.CustomerOrderLine{
			CONO: coLine.CONO,
			DIVI: coLine.DIVI,
			ORNO: coLine.ORNO,
			PONR: coLine.PONR,
			POSX: coLine.POSX,
			ITNO: coLine.ITNO,
			ITDS: coLine.ITDS,
			TEDS: coLine.TEDS,
			REPI: coLine.REPI,
			ORST: coLine.ORST,
			ORTY: coLine.ORTY,
			FACI: coLine.FACI,
			WHLO: coLine.WHLO,
			ORQT: coLine.ORQT,
			RNQT: coLine.RNQT,
			ALQT: coLine.ALQT,
			DLQT: coLine.DLQT,
			IVQT: coLine.IVQT,
			ORQA: coLine.ORQA,
			RNQA: coLine.RNQA,
			ALQA: coLine.ALQA,
			DLQA: coLine.DLQA,
			IVQA: coLine.IVQA,
			ALUN: coLine.ALUN,
			COFA: coLine.COFA,
			SPUN: coLine.SPUN,
			DWDT: coLine.DWDT,
			DWHM: coLine.DWHM,
			CODT: coLine.CODT,
			COHM: coLine.COHM,
			PLDT: coLine.PLDT,
			FDED: coLine.FDED,
			LDED: coLine.LDED,
			SAPR: coLine.SAPR,
			NEPR: coLine.NEPR,
			LNAM: coLine.LNAM,
			CUCD: coLine.CUCD,
			DIP1: coLine.DIP1,
			DIP2: coLine.DIP2,
			DIP3: coLine.DIP3,
			DIP4: coLine.DIP4,
			DIP5: coLine.DIP5,
			DIP6: coLine.DIP6,
			DIA1: coLine.DIA1,
			DIA2: coLine.DIA2,
			DIA3: coLine.DIA3,
			DIA4: coLine.DIA4,
			DIA5: coLine.DIA5,
			DIA6: coLine.DIA6,
			RORC: coLine.RORC,
			RORN: coLine.RORN,
			RORL: coLine.RORL,
			RORX: coLine.RORX,
			CUNO: coLine.CUNO,
			CUOR: coLine.CUOR,
			CUPO: coLine.CUPO,
			CUSX: coLine.CUSX,
			PRNO: coLine.PRNO,
			HDPR: coLine.HDPR,
			POPN: coLine.POPN,
			ALWT: coLine.ALWT,
			ALWQ: coLine.ALWQ,
			ADID: coLine.ADID,
			ROUT: coLine.ROUT,
			RODN: coLine.RODN,
			DSDT: coLine.DSDT,
			DSHM: coLine.DSHM,
			MODL: coLine.MODL,
			TEDL: coLine.TEDL,
			TEL2: coLine.TEL2,
			TEPA: coLine.TEPA,
			PACT: coLine.PACT,
			CUPA: coLine.CUPA,
			E0PA: coLine.E0PA,
			DSGP: coLine.DSGP,
			PUSN: coLine.PUSN,
			PUTP: coLine.PUTP,
			ATV1: coLine.ATV1,
			ATV2: coLine.ATV2,
			ATV3: coLine.ATV3,
			ATV4: coLine.ATV4,
			ATV5: coLine.ATV5,
			ATV6: coLine.ATV6,
			ATV7: coLine.ATV7,
			ATV8: coLine.ATV8,
			ATV9: coLine.ATV9,
			ATV0: coLine.ATV0,
			UCA1: coLine.UCA1,
			UCA2: coLine.UCA2,
			UCA3: coLine.UCA3,
			UCA4: coLine.UCA4,
			UCA5: coLine.UCA5,
			UCA6: coLine.UCA6,
			UCA7: coLine.UCA7,
			UCA8: coLine.UCA8,
			UCA9: coLine.UCA9,
			UCA0: coLine.UCA0,
			UDN1: coLine.UDN1,
			UDN2: coLine.UDN2,
			UDN3: coLine.UDN3,
			UDN4: coLine.UDN4,
			UDN5: coLine.UDN5,
			UDN6: coLine.UDN6,
			UID1: coLine.UID1,
			UID2: coLine.UID2,
			UID3: coLine.UID3,
			UCT1: coLine.UCT1,
			ATNR: coLine.ATNR,
			ATMO: coLine.ATMO,
			ATPR: coLine.ATPR,
			CFIN: coLine.CFIN,
			PROJ: coLine.PROJ,
			ELNO: coLine.ELNO,
			RGDT: coLine.RGDT,
			RGTM: coLine.RGTM,
			LMDT: coLine.LMDT,
			CHNO: coLine.CHNO,
			CHID: coLine.CHID,
			LMTS: coLine.LMTS,
			M3Timestamp: coLine.Timestamp,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	// Batch insert
	log.Printf("Inserting %d CO line records into database...", len(dbRecords))
	if err := s.db.BatchInsertCustomerOrderLines(ctx, dbRecords); err != nil {
		return 0, fmt.Errorf("failed to insert CO lines: %w", err)
	}

	log.Printf("CO lines refresh completed - inserted %d records", len(dbRecords))
	return len(dbRecords), nil
}

// RefreshCustomerOrderLinesByNumbers refreshes specific CO lines by order numbers
// DEPRECATED: This can cause issues with Compass when there are many order numbers
// Use RefreshOpenCustomerOrderLines instead
func (s *SnapshotService) RefreshCustomerOrderLinesByNumbers(ctx context.Context, orderNumbers []string, company string, facility string) error {
	if len(orderNumbers) == 0 {
		log.Println("No CO numbers to refresh")
		return nil
	}
	log.Printf("Refreshing %d specific customer order lines...", len(orderNumbers))

	// Build targeted query (no lastSyncDate needed - we want all lines for these orders)
	// Note: This deprecated method doesn't filter by context in the query builder call
	// because BuildCustomerOrderLinesByOrderNumbersQuery doesn't use context fields
	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildCustomerOrderLinesByOrderNumbersQuery(orderNumbers)

	// Execute query (DEPRECATED method - use batch refresh instead)
	log.Println("Submitting Compass query for CO lines...")
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total CO line records", totalRecords)

	// Parse results
	log.Println("Parsing CO line results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d CO line records", len(resultSet.Records))

	// Transform to database records
	dbRecords := make([]*db.CustomerOrderLine, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		coLine, err := compass.ParseCustomerOrderLine(record)
		if err != nil {
			log.Printf("Warning: failed to parse CO line record: %v", err)
			continue
		}

		// Map all fields directly - all stored as strings
		dbRecord := &db.CustomerOrderLine{
			CONO: coLine.CONO,
			DIVI: coLine.DIVI,
			ORNO: coLine.ORNO,
			PONR: coLine.PONR,
			POSX: coLine.POSX,
			ITNO: coLine.ITNO,
			ITDS: coLine.ITDS,
			TEDS: coLine.TEDS,
			REPI: coLine.REPI,
			ORST: coLine.ORST,
			ORTY: coLine.ORTY,
			FACI: coLine.FACI,
			WHLO: coLine.WHLO,
			ORQT: coLine.ORQT,
			RNQT: coLine.RNQT,
			ALQT: coLine.ALQT,
			DLQT: coLine.DLQT,
			IVQT: coLine.IVQT,
			ORQA: coLine.ORQA,
			RNQA: coLine.RNQA,
			ALQA: coLine.ALQA,
			DLQA: coLine.DLQA,
			IVQA: coLine.IVQA,
			ALUN: coLine.ALUN,
			COFA: coLine.COFA,
			SPUN: coLine.SPUN,
			DWDT: coLine.DWDT,
			DWHM: coLine.DWHM,
			CODT: coLine.CODT,
			COHM: coLine.COHM,
			PLDT: coLine.PLDT,
			FDED: coLine.FDED,
			LDED: coLine.LDED,
			SAPR: coLine.SAPR,
			NEPR: coLine.NEPR,
			LNAM: coLine.LNAM,
			CUCD: coLine.CUCD,
			DIP1: coLine.DIP1,
			DIP2: coLine.DIP2,
			DIP3: coLine.DIP3,
			DIP4: coLine.DIP4,
			DIP5: coLine.DIP5,
			DIP6: coLine.DIP6,
			DIA1: coLine.DIA1,
			DIA2: coLine.DIA2,
			DIA3: coLine.DIA3,
			DIA4: coLine.DIA4,
			DIA5: coLine.DIA5,
			DIA6: coLine.DIA6,
			RORC: coLine.RORC,
			RORN: coLine.RORN,
			RORL: coLine.RORL,
			RORX: coLine.RORX,
			CUNO: coLine.CUNO,
			CUOR: coLine.CUOR,
			CUPO: coLine.CUPO,
			CUSX: coLine.CUSX,
			PRNO: coLine.PRNO,
			HDPR: coLine.HDPR,
			POPN: coLine.POPN,
			ALWT: coLine.ALWT,
			ALWQ: coLine.ALWQ,
			ADID: coLine.ADID,
			ROUT: coLine.ROUT,
			RODN: coLine.RODN,
			DSDT: coLine.DSDT,
			DSHM: coLine.DSHM,
			MODL: coLine.MODL,
			TEDL: coLine.TEDL,
			TEL2: coLine.TEL2,
			TEPA: coLine.TEPA,
			PACT: coLine.PACT,
			CUPA: coLine.CUPA,
			E0PA: coLine.E0PA,
			DSGP: coLine.DSGP,
			PUSN: coLine.PUSN,
			PUTP: coLine.PUTP,
			ATV1: coLine.ATV1,
			ATV2: coLine.ATV2,
			ATV3: coLine.ATV3,
			ATV4: coLine.ATV4,
			ATV5: coLine.ATV5,
			ATV6: coLine.ATV6,
			ATV7: coLine.ATV7,
			ATV8: coLine.ATV8,
			ATV9: coLine.ATV9,
			ATV0: coLine.ATV0,
			UCA1: coLine.UCA1,
			UCA2: coLine.UCA2,
			UCA3: coLine.UCA3,
			UCA4: coLine.UCA4,
			UCA5: coLine.UCA5,
			UCA6: coLine.UCA6,
			UCA7: coLine.UCA7,
			UCA8: coLine.UCA8,
			UCA9: coLine.UCA9,
			UCA0: coLine.UCA0,
			UDN1: coLine.UDN1,
			UDN2: coLine.UDN2,
			UDN3: coLine.UDN3,
			UDN4: coLine.UDN4,
			UDN5: coLine.UDN5,
			UDN6: coLine.UDN6,
			UID1: coLine.UID1,
			UID2: coLine.UID2,
			UID3: coLine.UID3,
			UCT1: coLine.UCT1,
			ATNR: coLine.ATNR,
			ATMO: coLine.ATMO,
			ATPR: coLine.ATPR,
			CFIN: coLine.CFIN,
			PROJ: coLine.PROJ,
			ELNO: coLine.ELNO,
			RGDT: coLine.RGDT,
			RGTM: coLine.RGTM,
			LMDT: coLine.LMDT,
			CHNO: coLine.CHNO,
			CHID: coLine.CHID,
			LMTS: coLine.LMTS,
			M3Timestamp: coLine.Timestamp,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	// Batch insert
	log.Printf("Inserting %d CO line records into database...", len(dbRecords))
	if err := s.db.BatchInsertCustomerOrderLines(ctx, dbRecords); err != nil {
		return fmt.Errorf("failed to insert CO lines: %w", err)
	}

	log.Printf("CO lines refresh completed - inserted %d records", len(dbRecords))
	return nil
}

// RefreshManufacturingOrders refreshes MO data from Compass with MPREAL joins
// Filtered by company and facility context
// Returns list of unique CO numbers referenced by MOs
func (s *SnapshotService) RefreshManufacturingOrders(ctx context.Context, company string, facility string) ([]string, error) {
	log.Printf("Refreshing manufacturing orders for company '%s' and facility '%s'...", company, facility)

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query with context filters
	qb := compass.NewQueryBuilder(fullRefreshDate, company, facility)
	query := qb.BuildManufacturingOrdersQuery()

	// Execute query (DEPRECATED method - use batch refresh instead)
	log.Println("Submitting Compass query for MOs...")
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total MO records", totalRecords)

	// Parse results
	log.Println("Parsing MO results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d MO records", len(resultSet.Records))

	// Transform to database records
	dbRecords := make([]*db.ManufacturingOrder, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		mo, err := compass.ParseManufacturingOrder(record)
		if err != nil {
			log.Printf("Warning: failed to parse MO record: %v", err)
			continue
		}

		// Extract CO link fields from MPREAL join (all returned as strings)
		linkedCONumber := getRecordString(record, "linked_co_number")
		linkedCOLine := getRecordString(record, "linked_co_line")
		linkedCOSuffix := getRecordString(record, "linked_co_suffix")
		allocatedQty := getRecordString(record, "allocated_qty")

		// Create database record - all M3 fields stored as strings
		dbRecord := &db.ManufacturingOrder{
			// Core Identifiers
			CONO:          intToString(mo.CONO),
			DIVI:          mo.DIVI,
			FACI:          mo.FACI,
			MFNO:          mo.MFNO,
			PRNO:          mo.PRNO,
			ITNO:          mo.ITNO,

			// Status
			WHST:          mo.WHST,
			WHHS:          mo.WHHS,
			WMST:          mo.WMST,
			MOHS:          mo.MOHS,

			// Quantities
			ORQT:          floatToString(mo.ORQT),
			MAQT:          floatToString(mo.MAQT),
			ORQA:          floatToString(mo.ORQA),
			RVQT:          floatToString(mo.RVQT),
			RVQA:          floatToString(mo.RVQA),
			MAQA:          floatToString(mo.MAQA),

			// Dates
			STDT:          intToString(mo.STDT),
			FIDT:          intToString(mo.FIDT),
			MSTI:          intToString(mo.MSTI),
			MFTI:          intToString(mo.MFTI),
			FSTD:          intToString(mo.FSTD),
			FFID:          intToString(mo.FFID),
			RSDT:          intToString(mo.RSDT),
			REFD:          intToString(mo.REFD),
			RPDT:          intToString(mo.RPDT),

			// Planning
			PRIO:          intToString(mo.PRIO),
			RESP:          mo.RESP,
			PLGR:          mo.PLGR,
			WCLN:          mo.WCLN,
			PRDY:          intToString(mo.PRDY),

			// Warehouse/Location
			WHLO:          mo.WHLO,
			WHSL:          mo.WHSL,
			BANO:          mo.BANO,

			// Reference Orders
			RORC:          intToString(mo.RORC),
			RORN:          mo.RORN,
			RORL:          intToString(mo.RORL),
			RORX:          intToString(mo.RORX),

			// Hierarchy
			PRHL:          mo.PRHL,
			MFHL:          mo.MFHL,
			PRLO:          mo.PRLO,
			MFLO:          mo.MFLO,
			LEVL:          intToString(mo.LEVL),

			// Configuration
			CFIN:          int64ToString(mo.CFIN),
			ATNR:          int64ToString(mo.ATNR),

			// Order Type
			ORTY:          mo.ORTY,
			GETP:          mo.GETP,

			// Material/BOM
			BDCD:          mo.BDCD,
			SCEX:          mo.SCEX,
			STRT:          mo.STRT,
			ECVE:          mo.ECVE,

			// Routing
			AOID:          mo.AOID,
			NUOP:          intToString(mo.NUOP),
			NUFO:          intToString(mo.NUFO),

			// Action/Text
			ACTP:          mo.ACTP,
			TXT1:          mo.TXT1,
			TXT2:          mo.TXT2,

			// Project
			PROJ:          mo.PROJ,
			ELNO:          mo.ELNO,

			// M3 Audit
			RGDT:          intToString(mo.RGDT),
			RGTM:          intToString(mo.RGTM),
			LMDT:          intToString(mo.LMDT),
			LMTS:          int64ToString(mo.LMTS),
			CHNO:          intToString(mo.CHNO),
			CHID:          mo.CHID,

			// Metadata
			M3Timestamp:   int64ToString(mo.Timestamp),

			// CO Link
			LinkedCONumber: linkedCONumber,
			LinkedCOLine:   linkedCOLine,
			LinkedCOSuffix: linkedCOSuffix,
			AllocatedQty:   allocatedQty,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	// Batch insert
	log.Printf("Inserting %d MO records into database...", len(dbRecords))
	if err := s.db.BatchInsertManufacturingOrders(ctx, dbRecords); err != nil {
		return nil, fmt.Errorf("failed to insert MOs: %w", err)
	}

	// Extract unique CO numbers from linked_co_number field
	uniqueCONumbers := make(map[string]bool)
	for _, record := range resultSet.Records {
		if coNum, ok := record["linked_co_number"].(string); ok && coNum != "" {
			uniqueCONumbers[coNum] = true
		}
	}

	coNumberList := make([]string, 0, len(uniqueCONumbers))
	for coNum := range uniqueCONumbers {
		coNumberList = append(coNumberList, coNum)
	}

	log.Printf("MO refresh completed - found %d unique CO references", len(coNumberList))
	return coNumberList, nil
}

// RefreshPlannedOrders refreshes MOP data from Compass with MPREAL joins
// Filtered by company and facility context
// Returns list of unique CO numbers referenced by MOPs
func (s *SnapshotService) RefreshPlannedOrders(ctx context.Context, company string, facility string) ([]string, error) {
	log.Printf("Refreshing planned manufacturing orders (with CO links via MPREAL) for company '%s' and facility '%s'...", company, facility)

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query with MPREAL join and context filters
	qb := compass.NewQueryBuilder(fullRefreshDate, company, facility)
	query := qb.BuildPlannedOrdersWithCOLinksQuery()

	// Execute query (DEPRECATED method - use batch refresh instead)
	log.Println("Submitting Compass query for MOPs...")
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total MOP records", totalRecords)

	// Parse results
	log.Println("Parsing MOP results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d MOP records", len(resultSet.Records))

	// Debug: Print field names from first record
	if len(resultSet.Records) > 0 {
		log.Println("=== MOP Record Field Names ===")
		for key := range resultSet.Records[0] {
			log.Printf("Field: %s", key)
		}
		log.Println("==============================")
	}

	// Transform to database records
	dbRecords := make([]*db.PlannedManufacturingOrder, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		mop, err := compass.ParsePlannedOrder(record)
		if err != nil {
			log.Printf("Warning: failed to parse MOP record: %v", err)
			continue
		}

		// Build messages JSONB
		messagesJSON, _ := json.Marshal(mop.Messages)

		// Extract CO link fields from MPREAL join (all strings)
		linkedCONumber := getRecordString(record, "linked_co_number")
		linkedCOLine := getRecordString(record, "linked_co_line")
		linkedCOSuffix := getRecordString(record, "linked_co_suffix")
		allocatedQty := getRecordString(record, "allocated_qty")

		// Create database record - all M3 fields stored as strings
		dbRecord := &db.PlannedManufacturingOrder{
			// Core Identifiers
			CONO:          intToString(mop.CONO),
			DIVI:          mop.DIVI,
			FACI:          mop.FACI,
			PLPN:          int64ToString(mop.PLPN),
			PLPS:          intToString(mop.PLPS),
			PRNO:          mop.PRNO,
			ITNO:          mop.ITNO,

			// Status
			PSTS:          mop.PSTS,
			WHST:          mop.WHST,
			ACTP:          mop.ACTP,

			// Order Type
			ORTY:          mop.ORTY,
			GETY:          mop.GETY,

			// Quantities
			PPQT:          floatToString(mop.PPQT),
			ORQA:          floatToString(mop.ORQA),

			// Dates
			RELD:          intToString(mop.RELD),
			STDT:          intToString(mop.STDT),
			FIDT:          intToString(mop.FIDT),
			MSTI:          intToString(mop.MSTI),
			MFTI:          intToString(mop.MFTI),
			PLDT:          intToString(mop.PLDT),

			// Planning
			RESP:          mop.RESP,
			PRIP:          intToString(mop.PRIP),
			PLGR:          mop.PLGR,
			WCLN:          mop.WCLN,
			PRDY:          intToString(mop.PRDY),

			// Warehouse
			WHLO:          mop.WHLO,

			// Reference Orders
			RORC:          intToString(mop.RORC),
			RORN:          mop.RORN,
			RORL:          intToString(mop.RORL),
			RORX:          intToString(mop.RORX),
			RORH:          mop.RORH,

			// Hierarchy
			PLLO:          mop.PLLO,
			PLHL:          mop.PLHL,

			// Configuration
			ATNR:          int64ToString(mop.ATNR),
			CFIN:          int64ToString(mop.CFIN),

			// Project
			PROJ:          mop.PROJ,
			ELNO:          mop.ELNO,

			// Messages
			Messages:      messagesJSON,

			// Planning Parameters
			NUAU:          intToString(mop.NUAU),
			ORDP:          mop.ORDP,

			// M3 Audit
			RGDT:          intToString(mop.RGDT),
			RGTM:          intToString(mop.RGTM),
			LMDT:          intToString(mop.LMDT),
			LMTS:          int64ToString(mop.LMTS),
			CHNO:          intToString(mop.CHNO),
			CHID:          mop.CHID,

			// Metadata
			M3Timestamp:   int64ToString(mop.Timestamp),

			// CO Link
			LinkedCONumber: linkedCONumber,
			LinkedCOLine:   linkedCOLine,
			LinkedCOSuffix: linkedCOSuffix,
			AllocatedQty:   allocatedQty,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	// Extract unique CO numbers while transforming
	uniqueCONumbers := make(map[string]bool)
	for _, record := range resultSet.Records {
		if coNum, ok := record["linked_co_number"].(string); ok && coNum != "" {
			uniqueCONumbers[coNum] = true
		}
	}

	// Batch insert
	log.Printf("Inserting %d MOP records into database...", len(dbRecords))
	if err := s.db.BatchInsertPlannedOrders(ctx, dbRecords); err != nil {
		return nil, fmt.Errorf("failed to insert MOPs: %w", err)
	}

	// Convert to slice
	coNumberList := make([]string, 0, len(uniqueCONumbers))
	for coNum := range uniqueCONumbers {
		coNumberList = append(coNumberList, coNum)
	}

	log.Printf("MOP refresh completed - found %d unique CO references", len(coNumberList))
	return coNumberList, nil
}

// Helper functions to convert parser types to strings for storage

func intToString(val int) string {
	if val == 0 {
		return ""
	}
	return strconv.Itoa(val)
}

func int64ToString(val int64) string {
	if val == 0 {
		return ""
	}
	return strconv.FormatInt(val, 10)
}

func floatToString(val float64) string {
	if val == 0 {
		return ""
	}
	return strconv.FormatFloat(val, 'f', -1, 64)
}

// Helper functions to extract join fields that may be strings or numbers from Compass

func getRecordString(record map[string]interface{}, key string) string {
	if val, ok := record[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getRecordInt(record map[string]interface{}, key string) int {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func getRecordFloat(record map[string]interface{}, key string) float64 {
	if val, ok := record[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

// ========================================
// Parallel Batching Types and Helpers
// ========================================

// BatchRange represents an ID range for parallel batch processing
type BatchRange struct {
	MinID        interface{} // int64 for PLPN, string for MFNO/ORNO
	MaxID        interface{}
	BatchNumber  int
	TotalBatches int
}

// RangeMetadata represents MIN/MAX/COUNT from pre-query
type RangeMetadata struct {
	MinID        interface{} // int64 for PLPN, string for MFNO/ORNO
	MaxID        interface{}
	TotalRecords int
}

// CalculateBatchRanges determines optimal batch ranges for parallel processing
// Returns nil if no batching needed (dataset size <= batchSize)
// Uses over-partitioning to handle ID gaps from deleted records
func CalculateBatchRanges(meta RangeMetadata, batchSize int, overPartitionFactor float64) []BatchRange {
	// No partitioning needed for small datasets
	if meta.TotalRecords <= batchSize {
		log.Printf("Dataset size (%d) <= batch size (%d), no partitioning needed",
			meta.TotalRecords, batchSize)
		return nil
	}

	// Calculate number of batches with over-partitioning for ID gaps
	estimatedBatches := int(float64(meta.TotalRecords) / float64(batchSize))
	if estimatedBatches < 1 {
		estimatedBatches = 1
	}

	actualBatches := int(float64(estimatedBatches) * overPartitionFactor)
	if actualBatches < 2 {
		actualBatches = 2
	}

	log.Printf("Partitioning %d records into %d batches (estimated: %d, over-partition factor: %.1f)",
		meta.TotalRecords, actualBatches, estimatedBatches, overPartitionFactor)

	var ranges []BatchRange

	switch minID := meta.MinID.(type) {
	case int64:
		// Numeric ID partitioning (PLPN for MOPs)
		maxID := meta.MaxID.(int64)
		idRange := maxID - minID
		stepSize := idRange / int64(actualBatches)
		if stepSize < 1 {
			stepSize = 1
		}

		for i := 0; i < actualBatches; i++ {
			rangeMin := minID + (int64(i) * stepSize)
			rangeMax := rangeMin + stepSize
			if i == actualBatches-1 {
				rangeMax = maxID + 1 // Include max in last batch
			}

			ranges = append(ranges, BatchRange{
				MinID:        rangeMin,
				MaxID:        rangeMax,
				BatchNumber:  i + 1,
				TotalBatches: actualBatches,
			})
		}

	case string:
		// String ID partitioning (MFNO for MOs, ORNO for COs)
		maxID := meta.MaxID.(string)

		// Parse numeric portions for range calculation
		minNum, prefix := parseNumericSuffix(minID)
		maxNum, _ := parseNumericSuffix(maxID)

		if minNum == 0 && maxNum == 0 {
			// Failed to parse, use simple string division
			log.Printf("Warning: Could not parse numeric portion of string IDs, creating single batch")
			return nil
		}

		numRange := maxNum - minNum
		stepSize := numRange / actualBatches
		if stepSize < 1 {
			stepSize = 1
		}

		digitWidth := len(minID) - len(prefix)

		for i := 0; i < actualBatches; i++ {
			rangeMinNum := minNum + (i * stepSize)
			rangeMaxNum := rangeMinNum + stepSize
			if i == actualBatches-1 {
				rangeMaxNum = maxNum + 1
			}

			// Reconstruct string IDs with original prefix and zero-padding
			ranges = append(ranges, BatchRange{
				MinID:        formatStringID(prefix, rangeMinNum, digitWidth),
				MaxID:        formatStringID(prefix, rangeMaxNum, digitWidth),
				BatchNumber:  i + 1,
				TotalBatches: actualBatches,
			})
		}
	}

	return ranges
}

// parseNumericSuffix extracts the numeric portion from a string ID
// Examples: "M001234" → (1234, "M"), "001234" → (1234, "")
func parseNumericSuffix(id string) (int, string) {
	if id == "" {
		return 0, ""
	}

	// Find where digits start
	digitStart := -1
	for i, ch := range id {
		if ch >= '0' && ch <= '9' {
			digitStart = i
			break
		}
	}

	if digitStart == -1 {
		// No digits found
		return 0, id
	}

	prefix := id[:digitStart]
	numStr := id[digitStart:]

	// Parse numeric portion
	num := 0
	fmt.Sscanf(numStr, "%d", &num)

	return num, prefix
}

// formatStringID reconstructs a string ID with prefix and zero-padding
// Example: formatStringID("M", 1234, 6) → "M001234"
func formatStringID(prefix string, num int, digitWidth int) string {
	if digitWidth <= 0 {
		digitWidth = 6 // Default width
	}

	format := fmt.Sprintf("%%s%%0%dd", digitWidth)
	return fmt.Sprintf(format, prefix, num)
}

// LoadSystemSettingInt loads an integer setting from database with default fallback
func LoadSystemSettingInt(database *db.Queries, key string, defaultValue int) int {
	ctx := context.Background()
	settings, err := database.GetSystemSettings(ctx)
	if err != nil {
		log.Printf("Warning: Failed to load system settings for %s, using default: %d", key, defaultValue)
		return defaultValue
	}

	for _, setting := range settings {
		if setting.SettingKey == key {
			var value int
			if _, err := fmt.Sscanf(setting.SettingValue, "%d", &value); err == nil {
				return value
			}
			log.Printf("Warning: Invalid value for %s, using default: %d", key, defaultValue)
			return defaultValue
		}
	}

	log.Printf("Warning: Setting %s not found, using default: %d", key, defaultValue)
	return defaultValue
}

// LoadSystemSettingFloat loads a float setting from database with default fallback
func LoadSystemSettingFloat(database *db.Queries, key string, defaultValue float64) float64 {
	ctx := context.Background()
	settings, err := database.GetSystemSettings(ctx)
	if err != nil {
		log.Printf("Warning: Failed to load system settings for %s, using default: %.2f", key, defaultValue)
		return defaultValue
	}

	for _, setting := range settings {
		if setting.SettingKey == key {
			var value float64
			if _, err := fmt.Sscanf(setting.SettingValue, "%f", &value); err == nil {
				return value
			}
			log.Printf("Warning: Invalid value for %s, using default: %.2f", key, defaultValue)
			return defaultValue
		}
	}

	log.Printf("Warning: Setting %s not found, using default: %.2f", key, defaultValue)
	return defaultValue
}

// ========================================
// Range Query Methods (MIN/MAX/COUNT)
// ========================================

// QueryMOPRange queries for MIN/MAX/COUNT of planned manufacturing orders
// Returns metadata for batch range calculation
func (s *SnapshotService) QueryMOPRange(ctx context.Context, company, facility string) (*RangeMetadata, error) {
	log.Printf("Querying MOP range (MIN/MAX/COUNT) for company '%s', facility '%s'...", company, facility)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildMOPRangeQuery()

	// Execute lightweight aggregate query (always returns 1 row)
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, _, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query MOP range: %w", err)
	}

	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range results: %w", err)
	}

	if len(resultSet.Records) == 0 {
		log.Println("MOP range query returned 0 records")
		return &RangeMetadata{TotalRecords: 0}, nil
	}

	record := resultSet.Records[0]
	meta := &RangeMetadata{
		MinID:        compass.GetInt64(record, "min_id"),
		MaxID:        compass.GetInt64(record, "max_id"),
		TotalRecords: compass.GetInt(record, "total_records"),
	}

	log.Printf("MOP range: MIN=%v, MAX=%v, COUNT=%d", meta.MinID, meta.MaxID, meta.TotalRecords)
	return meta, nil
}

// QueryMORange queries for MIN/MAX/COUNT of manufacturing orders
func (s *SnapshotService) QueryMORange(ctx context.Context, company, facility string) (*RangeMetadata, error) {
	log.Printf("Querying MO range (MIN/MAX/COUNT) for company '%s', facility '%s'...", company, facility)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildMORangeQuery()

	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, _, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query MO range: %w", err)
	}

	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range results: %w", err)
	}

	if len(resultSet.Records) == 0 {
		log.Println("MO range query returned 0 records")
		return &RangeMetadata{TotalRecords: 0}, nil
	}

	record := resultSet.Records[0]
	meta := &RangeMetadata{
		MinID:        compass.GetString(record, "min_id"),
		MaxID:        compass.GetString(record, "max_id"),
		TotalRecords: compass.GetInt(record, "total_records"),
	}

	log.Printf("MO range: MIN=%v, MAX=%v, COUNT=%d", meta.MinID, meta.MaxID, meta.TotalRecords)
	return meta, nil
}

// QueryCORange queries for MIN/MAX/COUNT of customer order lines
func (s *SnapshotService) QueryCORange(ctx context.Context, company, facility string) (*RangeMetadata, error) {
	log.Printf("Querying CO range (MIN/MAX/COUNT) for company '%s', facility '%s'...", company, facility)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildCORangeQuery()

	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, _, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query CO range: %w", err)
	}

	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range results: %w", err)
	}

	if len(resultSet.Records) == 0 {
		log.Println("CO range query returned 0 records")
		return &RangeMetadata{TotalRecords: 0}, nil
	}

	record := resultSet.Records[0]
	meta := &RangeMetadata{
		MinID:        compass.GetString(record, "min_id"),
		MaxID:        compass.GetString(record, "max_id"),
		TotalRecords: compass.GetInt(record, "total_records"),
	}

	log.Printf("CO range: MIN=%v, MAX=%v, COUNT=%d", meta.MinID, meta.MaxID, meta.TotalRecords)
	return meta, nil
}

// ========================================
// Batch Refresh Methods (ID Range Filtered)
// ========================================

// RefreshMOPBatch fetches a single batch of MOPs by PLPN ID range
// Optimized for Spark: Uses WHERE PLPN >= X AND PLPN < Y predicate pushdown
func (s *SnapshotService) RefreshMOPBatch(ctx context.Context, company, facility string, minID, maxID int64) (int, error) {
	log.Printf("Fetching MOP batch: PLPN >= %d AND PLPN < %d", minID, maxID)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildMOPBatchQuery(minID, maxID)

	// Load page size from settings
	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)

	// Execute with automatic pagination
	results, recordCount, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch MOP batch: %w", err)
	}

	if recordCount == 0 {
		log.Printf("MOP batch [%d-%d] returned 0 records (ID gap)", minID, maxID)
		return 0, nil
	}

	// Parse results
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse MOP batch: %w", err)
	}

	// Transform to DB records (same logic as RefreshPlannedOrders)
	dbRecords := make([]*db.PlannedManufacturingOrder, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		mop, err := compass.ParsePlannedOrder(record)
		if err != nil {
			log.Printf("Warning: failed to parse MOP: %v", err)
			continue
		}

		// Build messages JSONB
		messagesJSON, _ := json.Marshal(mop.Messages)

		// Extract CO link fields from MPREAL join
		linkedCONumber := getRecordString(record, "linked_co_number")
		linkedCOLine := getRecordString(record, "linked_co_line")
		linkedCOSuffix := getRecordString(record, "linked_co_suffix")
		allocatedQty := getRecordString(record, "allocated_qty")

		dbRecord := &db.PlannedManufacturingOrder{
			CONO:           intToString(mop.CONO),
			DIVI:           mop.DIVI,
			FACI:           mop.FACI,
			PLPN:           int64ToString(mop.PLPN),
			PLPS:           intToString(mop.PLPS),
			PRNO:           mop.PRNO,
			ITNO:           mop.ITNO,
			PSTS:           mop.PSTS,
			WHST:           mop.WHST,
			ACTP:           mop.ACTP,
			ORTY:           mop.ORTY,
			GETY:           mop.GETY,
			PPQT:           floatToString(mop.PPQT),
			ORQA:           floatToString(mop.ORQA),
			RELD:           intToString(mop.RELD),
			STDT:           intToString(mop.STDT),
			FIDT:           intToString(mop.FIDT),
			MSTI:           intToString(mop.MSTI),
			MFTI:           intToString(mop.MFTI),
			PLDT:           intToString(mop.PLDT),
			RESP:           mop.RESP,
			PRIP:           intToString(mop.PRIP),
			PLGR:           mop.PLGR,
			WCLN:           mop.WCLN,
			PRDY:           intToString(mop.PRDY),
			WHLO:           mop.WHLO,
			RORC:           intToString(mop.RORC),
			RORN:           mop.RORN,
			RORL:           intToString(mop.RORL),
			RORX:           intToString(mop.RORX),
			RORH:           mop.RORH,
			PLLO:           mop.PLLO,
			PLHL:           mop.PLHL,
			ATNR:           int64ToString(mop.ATNR),
			CFIN:           int64ToString(mop.CFIN),
			PROJ:           mop.PROJ,
			ELNO:           mop.ELNO,
			Messages:       messagesJSON,
			NUAU:           intToString(mop.NUAU),
			ORDP:           mop.ORDP,
			RGDT:           intToString(mop.RGDT),
			RGTM:           intToString(mop.RGTM),
			LMDT:           intToString(mop.LMDT),
			LMTS:           int64ToString(mop.LMTS),
			CHNO:           intToString(mop.CHNO),
			CHID:           mop.CHID,
			M3Timestamp:    int64ToString(mop.Timestamp),
			LinkedCONumber: linkedCONumber,
			LinkedCOLine:   linkedCOLine,
			LinkedCOSuffix: linkedCOSuffix,
			AllocatedQty:   allocatedQty,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	// Batch insert
	log.Printf("Inserting %d MOP records from batch [%d-%d]...", len(dbRecords), minID, maxID)
	if err := s.db.BatchInsertPlannedOrders(ctx, dbRecords); err != nil {
		return 0, fmt.Errorf("failed to insert MOP batch: %w", err)
	}

	log.Printf("MOP batch [%d-%d] completed: %d records inserted", minID, maxID, len(dbRecords))
	return len(dbRecords), nil
}

// RefreshMOBatch fetches a single batch of MOs by MFNO string range
func (s *SnapshotService) RefreshMOBatch(ctx context.Context, company, facility string, minID, maxID string) (int, error) {
	log.Printf("Fetching MO batch: MFNO >= '%s' AND MFNO < '%s'", minID, maxID)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildMOBatchQuery(minID, maxID)

	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, recordCount, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch MO batch: %w", err)
	}

	if recordCount == 0 {
		log.Printf("MO batch [%s-%s] returned 0 records (ID gap)", minID, maxID)
		return 0, nil
	}

	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse MO batch: %w", err)
	}

	// Transform to DB records (same logic as RefreshManufacturingOrders)
	dbRecords := make([]*db.ManufacturingOrder, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		mo, err := compass.ParseManufacturingOrder(record)
		if err != nil {
			log.Printf("Warning: failed to parse MO: %v", err)
			continue
		}

		linkedCONumber := getRecordString(record, "linked_co_number")
		linkedCOLine := getRecordString(record, "linked_co_line")
		linkedCOSuffix := getRecordString(record, "linked_co_suffix")
		allocatedQty := getRecordString(record, "allocated_qty")

		dbRecord := &db.ManufacturingOrder{
			CONO:           intToString(mo.CONO),
			DIVI:           mo.DIVI,
			FACI:           mo.FACI,
			MFNO:           mo.MFNO,
			PRNO:           mo.PRNO,
			ITNO:           mo.ITNO,
			WHST:           mo.WHST,
			WHHS:           mo.WHHS,
			WMST:           mo.WMST,
			MOHS:           mo.MOHS,
			ORQT:           floatToString(mo.ORQT),
			MAQT:           floatToString(mo.MAQT),
			ORQA:           floatToString(mo.ORQA),
			RVQT:           floatToString(mo.RVQT),
			RVQA:           floatToString(mo.RVQA),
			MAQA:           floatToString(mo.MAQA),
			STDT:           intToString(mo.STDT),
			FIDT:           intToString(mo.FIDT),
			MSTI:           intToString(mo.MSTI),
			MFTI:           intToString(mo.MFTI),
			FSTD:           intToString(mo.FSTD),
			FFID:           intToString(mo.FFID),
			RSDT:           intToString(mo.RSDT),
			REFD:           intToString(mo.REFD),
			RPDT:           intToString(mo.RPDT),
			PRIO:           intToString(mo.PRIO),
			RESP:           mo.RESP,
			PLGR:           mo.PLGR,
			WCLN:           mo.WCLN,
			PRDY:           intToString(mo.PRDY),
			WHLO:           mo.WHLO,
			WHSL:           mo.WHSL,
			BANO:           mo.BANO,
			RORC:           intToString(mo.RORC),
			RORN:           mo.RORN,
			RORL:           intToString(mo.RORL),
			RORX:           intToString(mo.RORX),
			PRHL:           mo.PRHL,
			MFHL:           mo.MFHL,
			PRLO:           mo.PRLO,
			MFLO:           mo.MFLO,
			LEVL:           intToString(mo.LEVL),
			CFIN:           int64ToString(mo.CFIN),
			ATNR:           int64ToString(mo.ATNR),
			ORTY:           mo.ORTY,
			GETP:           mo.GETP,
			BDCD:           mo.BDCD,
			SCEX:           mo.SCEX,
			STRT:           mo.STRT,
			ECVE:           mo.ECVE,
			AOID:           mo.AOID,
			NUOP:           intToString(mo.NUOP),
			NUFO:           intToString(mo.NUFO),
			ACTP:           mo.ACTP,
			TXT1:           mo.TXT1,
			TXT2:           mo.TXT2,
			PROJ:           mo.PROJ,
			ELNO:           mo.ELNO,
			RGDT:           intToString(mo.RGDT),
			RGTM:           intToString(mo.RGTM),
			LMDT:           intToString(mo.LMDT),
			LMTS:           int64ToString(mo.LMTS),
			CHNO:           intToString(mo.CHNO),
			CHID:           mo.CHID,
			M3Timestamp:    int64ToString(mo.Timestamp),
			LinkedCONumber: linkedCONumber,
			LinkedCOLine:   linkedCOLine,
			LinkedCOSuffix: linkedCOSuffix,
			AllocatedQty:   allocatedQty,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	log.Printf("Inserting %d MO records from batch [%s-%s]...", len(dbRecords), minID, maxID)
	if err := s.db.BatchInsertManufacturingOrders(ctx, dbRecords); err != nil {
		return 0, fmt.Errorf("failed to insert MO batch: %w", err)
	}

	log.Printf("MO batch [%s-%s] completed: %d records inserted", minID, maxID, len(dbRecords))
	return len(dbRecords), nil
}

// RefreshCOBatch fetches a single batch of customer order lines by ORNO string range
func (s *SnapshotService) RefreshCOBatch(ctx context.Context, company, facility string, minID, maxID string) (int, error) {
	log.Printf("Fetching CO batch: ORNO >= '%s' AND ORNO < '%s'", minID, maxID)

	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildCOBatchQuery(minID, maxID)

	pageSize := LoadSystemSettingInt(s.db, "compass_page_size", 10000)
	results, recordCount, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch CO batch: %w", err)
	}

	if recordCount == 0 {
		log.Printf("CO batch [%s-%s] returned 0 records (ID gap)", minID, maxID)
		return 0, nil
	}

	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CO batch: %w", err)
	}

	// Transform to DB records (same logic as RefreshOpenCustomerOrderLines)
	dbRecords := make([]*db.CustomerOrderLine, 0, len(resultSet.Records))
	for _, record := range resultSet.Records {
		coLine, err := compass.ParseCustomerOrderLine(record)
		if err != nil {
			log.Printf("Warning: failed to parse CO line: %v", err)
			continue
		}

		dbRecord := &db.CustomerOrderLine{
			CONO:        coLine.CONO,
			DIVI:        coLine.DIVI,
			ORNO:        coLine.ORNO,
			PONR:        coLine.PONR,
			POSX:        coLine.POSX,
			ITNO:        coLine.ITNO,
			ITDS:        coLine.ITDS,
			TEDS:        coLine.TEDS,
			REPI:        coLine.REPI,
			ORST:        coLine.ORST,
			ORTY:        coLine.ORTY,
			FACI:        coLine.FACI,
			WHLO:        coLine.WHLO,
			ORQT:        coLine.ORQT,
			RNQT:        coLine.RNQT,
			ALQT:        coLine.ALQT,
			DLQT:        coLine.DLQT,
			IVQT:        coLine.IVQT,
			ORQA:        coLine.ORQA,
			RNQA:        coLine.RNQA,
			ALQA:        coLine.ALQA,
			DLQA:        coLine.DLQA,
			IVQA:        coLine.IVQA,
			ALUN:        coLine.ALUN,
			COFA:        coLine.COFA,
			SPUN:        coLine.SPUN,
			DWDT:        coLine.DWDT,
			DWHM:        coLine.DWHM,
			CODT:        coLine.CODT,
			COHM:        coLine.COHM,
			PLDT:        coLine.PLDT,
			FDED:        coLine.FDED,
			LDED:        coLine.LDED,
			SAPR:        coLine.SAPR,
			NEPR:        coLine.NEPR,
			LNAM:        coLine.LNAM,
			CUCD:        coLine.CUCD,
			DIP1:        coLine.DIP1,
			DIP2:        coLine.DIP2,
			DIP3:        coLine.DIP3,
			DIP4:        coLine.DIP4,
			DIP5:        coLine.DIP5,
			DIP6:        coLine.DIP6,
			DIA1:        coLine.DIA1,
			DIA2:        coLine.DIA2,
			DIA3:        coLine.DIA3,
			DIA4:        coLine.DIA4,
			DIA5:        coLine.DIA5,
			DIA6:        coLine.DIA6,
			RORC:        coLine.RORC,
			RORN:        coLine.RORN,
			RORL:        coLine.RORL,
			RORX:        coLine.RORX,
			CUNO:        coLine.CUNO,
			CUOR:        coLine.CUOR,
			CUPO:        coLine.CUPO,
			CUSX:        coLine.CUSX,
			PRNO:        coLine.PRNO,
			HDPR:        coLine.HDPR,
			POPN:        coLine.POPN,
			ALWT:        coLine.ALWT,
			ALWQ:        coLine.ALWQ,
			ADID:        coLine.ADID,
			ROUT:        coLine.ROUT,
			RODN:        coLine.RODN,
			DSDT:        coLine.DSDT,
			DSHM:        coLine.DSHM,
			MODL:        coLine.MODL,
			TEDL:        coLine.TEDL,
			TEL2:        coLine.TEL2,
			TEPA:        coLine.TEPA,
			PACT:        coLine.PACT,
			CUPA:        coLine.CUPA,
			E0PA:        coLine.E0PA,
			DSGP:        coLine.DSGP,
			PUSN:        coLine.PUSN,
			PUTP:        coLine.PUTP,
			ATV1:        coLine.ATV1,
			ATV2:        coLine.ATV2,
			ATV3:        coLine.ATV3,
			ATV4:        coLine.ATV4,
			ATV5:        coLine.ATV5,
			ATV6:        coLine.ATV6,
			ATV7:        coLine.ATV7,
			ATV8:        coLine.ATV8,
			ATV9:        coLine.ATV9,
			ATV0:        coLine.ATV0,
			UCA1:        coLine.UCA1,
			UCA2:        coLine.UCA2,
			UCA3:        coLine.UCA3,
			UCA4:        coLine.UCA4,
			UCA5:        coLine.UCA5,
			UCA6:        coLine.UCA6,
			UCA7:        coLine.UCA7,
			UCA8:        coLine.UCA8,
			UCA9:        coLine.UCA9,
			UCA0:        coLine.UCA0,
			UDN1:        coLine.UDN1,
			UDN2:        coLine.UDN2,
			UDN3:        coLine.UDN3,
			UDN4:        coLine.UDN4,
			UDN5:        coLine.UDN5,
			UDN6:        coLine.UDN6,
			UID1:        coLine.UID1,
			UID2:        coLine.UID2,
			UID3:        coLine.UID3,
			UCT1:        coLine.UCT1,
			ATNR:        coLine.ATNR,
			ATMO:        coLine.ATMO,
			ATPR:        coLine.ATPR,
			CFIN:        coLine.CFIN,
			PROJ:        coLine.PROJ,
			ELNO:        coLine.ELNO,
			RGDT:        coLine.RGDT,
			RGTM:        coLine.RGTM,
			LMDT:        coLine.LMDT,
			CHNO:        coLine.CHNO,
			CHID:        coLine.CHID,
			LMTS:        coLine.LMTS,
			M3Timestamp: coLine.Timestamp,
		}

		dbRecords = append(dbRecords, dbRecord)
	}

	log.Printf("Inserting %d CO line records from batch [%s-%s]...", len(dbRecords), minID, maxID)
	if err := s.db.BatchInsertCustomerOrderLines(ctx, dbRecords); err != nil {
		return 0, fmt.Errorf("failed to insert CO batch: %w", err)
	}

	log.Printf("CO batch [%s-%s] completed: %d records inserted", minID, maxID, len(dbRecords))
	return len(dbRecords), nil
}
