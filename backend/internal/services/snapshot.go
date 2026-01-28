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

// reportSubProgress reports a sub-operation within the current phase
func (s *SnapshotService) reportSubProgress(operation string, recordCount int) {
	if s.progressCallback != nil {
		s.progressCallback("", 0, 0, operation, s.mopCount, s.moCount, s.coCount)
	}
}

// RefreshOpenCustomerOrderLines refreshes all open CO lines (status < 30)
// Filtered by environment, company and facility context
// This is more efficient than querying by specific order numbers when there are many orders
// Returns the count of records processed
func (s *SnapshotService) RefreshOpenCustomerOrderLines(ctx context.Context, environment, company string, facility string) (int, error) {
	log.Printf("Refreshing all open customer order lines (status < 30) for environment '%s', company '%s' and facility '%s'...", environment, company, facility)

	// Build query for all open CO lines with context filters
	qb := compass.NewQueryBuilder(0, company, facility)
	query := qb.BuildOpenCustomerOrderLinesQuery()

	// Execute query
	log.Println("Submitting Compass query for open CO lines...")
	s.reportSubProgress("Querying Compass SQL for customer order lines...", 0)
	pageSize := LoadSystemSettingInt(s.db, environment, "compass_batch_size", 50000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(
		ctx, query, pageSize,
		func(page, totalPages, pageRecords, totalFetched, totalRecords int) {
			operation := fmt.Sprintf("Loading page %d/%d from Compass SQL (%d records, %d total)",
				page, totalPages, pageRecords, totalFetched)
			s.reportSubProgress(operation, totalFetched)
		},
	)
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
	s.reportSubProgress(fmt.Sprintf("Processing %d customer order line records...", len(resultSet.Records)), 0)

	// Transform to database records
	const parseProgressInterval = 5000
	dbRecords := make([]*db.CustomerOrderLine, 0, len(resultSet.Records))
	for i, record := range resultSet.Records {
		coLine, err := compass.ParseCustomerOrderLine(record)
		if err != nil {
			log.Printf("Warning: failed to parse CO line record: %v", err)
			continue
		}

		// Map all fields directly - all stored as strings
		dbRecord := &db.CustomerOrderLine{
			Environment: environment,
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
			CustomerName: coLine.CustomerName,
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
			JDCD: coLine.JDCD,
			DLIX: coLine.DLIX,
			ORTP: coLine.ORTP,
			COTypeDescription: coLine.COTypeDescription,
			DeliveryMethod: coLine.DeliveryMethod,
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

		// Report parsing progress every N records
		if (i+1)%parseProgressInterval == 0 || (i+1) == len(resultSet.Records) {
			operation := fmt.Sprintf("Parsed %d/%d records (%d%%)",
				i+1, len(resultSet.Records),
				((i+1)*100)/len(resultSet.Records))
			s.reportSubProgress(operation, i+1)
		}
	}

	// Batch insert
	log.Printf("Inserting %d CO line records into database...", len(dbRecords))
	s.reportSubProgress(fmt.Sprintf("Inserting %d customer order lines into database...", len(dbRecords)), len(dbRecords))

	insertCallback := func(inserted, total int) {
		operation := fmt.Sprintf("Inserted %d/%d customer order lines into database (%d%%)",
			inserted, total, (inserted*100)/total)
		s.reportSubProgress(operation, inserted)
	}

	if err := s.db.BatchInsertCustomerOrderLines(ctx, dbRecords, insertCallback); err != nil {
		return 0, fmt.Errorf("failed to insert CO lines: %w", err)
	}

	log.Printf("CO lines refresh completed - inserted %d records", len(dbRecords))
	return len(dbRecords), nil
}

// RefreshCustomerOrderLinesByNumbers refreshes specific CO lines by order numbers
// DEPRECATED: This can cause issues with Compass when there are many order numbers
// Use RefreshOpenCustomerOrderLines instead
func (s *SnapshotService) RefreshCustomerOrderLinesByNumbers(ctx context.Context, environment string, orderNumbers []string, company string, facility string) error {
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
	pageSize := LoadSystemSettingInt(s.db, environment, "compass_batch_size", 50000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(ctx, query, pageSize, nil)
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
			Environment: environment,
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
			CustomerName: coLine.CustomerName,
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
			JDCD: coLine.JDCD,
			DLIX: coLine.DLIX,
			ORTP: coLine.ORTP,
			COTypeDescription: coLine.COTypeDescription,
			DeliveryMethod: coLine.DeliveryMethod,
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
	if err := s.db.BatchInsertCustomerOrderLines(ctx, dbRecords, nil); err != nil {
		return fmt.Errorf("failed to insert CO lines: %w", err)
	}

	log.Printf("CO lines refresh completed - inserted %d records", len(dbRecords))
	return nil
}

// RefreshManufacturingOrders refreshes MO data from Compass with MPREAL joins
// Filtered by environment, company and facility context
// Returns list of unique CO numbers referenced by MOs
func (s *SnapshotService) RefreshManufacturingOrders(ctx context.Context, environment, company string, facility string) (int, error) {
	log.Printf("Refreshing manufacturing orders for environment '%s', company '%s' and facility '%s'...", environment, company, facility)

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query with context filters
	qb := compass.NewQueryBuilder(fullRefreshDate, company, facility)
	query := qb.BuildManufacturingOrdersQuery()

	// Execute query (DEPRECATED method - use batch refresh instead)
	log.Println("Submitting Compass query for MOs...")
	s.reportSubProgress("Querying Compass SQL for manufacturing orders...", 0)
	pageSize := LoadSystemSettingInt(s.db, environment, "compass_batch_size", 50000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(
		ctx, query, pageSize,
		func(page, totalPages, pageRecords, totalFetched, totalRecords int) {
			operation := fmt.Sprintf("Loading page %d/%d from Compass SQL (%d records, %d total)",
				page, totalPages, pageRecords, totalFetched)
			s.reportSubProgress(operation, totalFetched)
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total MO records", totalRecords)

	// Parse results
	log.Println("Parsing MO results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d MO records", len(resultSet.Records))
	s.reportSubProgress(fmt.Sprintf("Processing %d manufacturing order records...", len(resultSet.Records)), 0)

	// Transform to database records
	const parseProgressInterval = 5000
	dbRecords := make([]*db.ManufacturingOrder, 0, len(resultSet.Records))
	for i, record := range resultSet.Records {
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
			Environment:   environment,
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

		// Report parsing progress every N records
		if (i+1)%parseProgressInterval == 0 || (i+1) == len(resultSet.Records) {
			operation := fmt.Sprintf("Parsed %d/%d records (%d%%)",
				i+1, len(resultSet.Records),
				((i+1)*100)/len(resultSet.Records))
			s.reportSubProgress(operation, i+1)
		}
	}

	// Batch insert
	log.Printf("Inserting %d MO records into database...", len(dbRecords))
	s.reportSubProgress(fmt.Sprintf("Inserting %d manufacturing orders into database...", len(dbRecords)), len(dbRecords))

	insertCallback := func(inserted, total int) {
		operation := fmt.Sprintf("Inserted %d/%d manufacturing orders into database (%d%%)",
			inserted, total, (inserted*100)/total)
		s.reportSubProgress(operation, inserted)
	}

	if err := s.db.BatchInsertManufacturingOrders(ctx, dbRecords, insertCallback); err != nil {
		return 0, fmt.Errorf("failed to insert MOs: %w", err)
	}

	// Extract unique CO numbers from linked_co_number field
	log.Printf("MO refresh completed - inserted %d records", len(dbRecords))
	return len(dbRecords), nil
}

// RefreshPlannedOrders refreshes MOP data from Compass with MPREAL joins
// Filtered by environment, company and facility context
// Returns list of unique CO numbers referenced by MOPs
func (s *SnapshotService) RefreshPlannedOrders(ctx context.Context, environment, company string, facility string) (int, error) {
	log.Printf("Refreshing planned manufacturing orders (with CO links via MPREAL) for environment '%s', company '%s' and facility '%s'...", environment, company, facility)

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query with MPREAL join and context filters
	qb := compass.NewQueryBuilder(fullRefreshDate, company, facility)
	query := qb.BuildPlannedOrdersWithCOLinksQuery()

	// Execute query (DEPRECATED method - use batch refresh instead)
	log.Println("Submitting Compass query for MOPs...")
	s.reportSubProgress("Querying Compass SQL for planned orders...", 0)
	pageSize := LoadSystemSettingInt(s.db, environment, "compass_batch_size", 50000)
	results, totalRecords, err := s.compassClient.ExecuteQueryWithPagination(
		ctx, query, pageSize,
		func(page, totalPages, pageRecords, totalFetched, totalRecords int) {
			operation := fmt.Sprintf("Loading page %d/%d from Compass SQL (%d records, %d total)",
				page, totalPages, pageRecords, totalFetched)
			s.reportSubProgress(operation, totalFetched)
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Printf("Query returned %d total MOP records", totalRecords)

	// Parse results
	log.Println("Parsing MOP results...")
	resultSet, err := compass.ParseResults(results)
	if err != nil {
		return 0, fmt.Errorf("failed to parse results: %w", err)
	}

	log.Printf("Received %d MOP records", len(resultSet.Records))
	s.reportSubProgress(fmt.Sprintf("Processing %d planned order records...", len(resultSet.Records)), 0)

	// Debug: Print field names from first record
	if len(resultSet.Records) > 0 {
		log.Println("=== MOP Record Field Names ===")
		for key := range resultSet.Records[0] {
			log.Printf("Field: %s", key)
		}
		log.Println("==============================")
	}

	// Transform to database records
	const parseProgressInterval = 5000
	dbRecords := make([]*db.PlannedManufacturingOrder, 0, len(resultSet.Records))
	for i, record := range resultSet.Records {
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
			Environment:   environment,
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

		// Report parsing progress every N records
		if (i+1)%parseProgressInterval == 0 || (i+1) == len(resultSet.Records) {
			operation := fmt.Sprintf("Parsed %d/%d records (%d%%)",
				i+1, len(resultSet.Records),
				((i+1)*100)/len(resultSet.Records))
			s.reportSubProgress(operation, i+1)
		}
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
	s.reportSubProgress(fmt.Sprintf("Inserting %d planned orders into database...", len(dbRecords)), len(dbRecords))

	insertCallback := func(inserted, total int) {
		operation := fmt.Sprintf("Inserted %d/%d planned orders into database (%d%%)",
			inserted, total, (inserted*100)/total)
		s.reportSubProgress(operation, inserted)
	}

	if err := s.db.BatchInsertPlannedOrders(ctx, dbRecords, insertCallback); err != nil {
		return 0, fmt.Errorf("failed to insert MOPs: %w", err)
	}

	log.Printf("MOP refresh completed - inserted %d records", len(dbRecords))
	return len(dbRecords), nil
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


// LoadSystemSettingInt loads an integer setting from database with default fallback for a specific environment
func LoadSystemSettingInt(database *db.Queries, environment, key string, defaultValue int) int {
	ctx := context.Background()
	settings, err := database.GetSystemSettings(ctx, environment)
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

// LoadSystemSettingFloat loads a float setting from database with default fallback for a specific environment
func LoadSystemSettingFloat(database *db.Queries, environment, key string, defaultValue float64) float64 {
	ctx := context.Background()
	settings, err := database.GetSystemSettings(ctx, environment)
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


