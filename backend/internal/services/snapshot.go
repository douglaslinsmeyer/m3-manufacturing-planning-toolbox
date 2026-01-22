package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// SnapshotService handles data refresh operations
type SnapshotService struct {
	compassClient *compass.Client
	db            *db.Queries
}

// NewSnapshotService creates a new snapshot service
func NewSnapshotService(compassClient *compass.Client, database *db.Queries) *SnapshotService {
	return &SnapshotService{
		compassClient: compassClient,
		db:            database,
	}
}

// RefreshAll performs a full data refresh from M3 with table truncation
// Strategy: Clear all M3 snapshot data, then pull MOPs/MOs with MPREAL links, then all open CO lines
func (s *SnapshotService) RefreshAll(ctx context.Context) error {
	log.Println("Starting full data refresh from M3...")
	log.Println("Strategy: Truncate all M3 snapshot tables, load MOPs/MOs with MPREAL links, load all open CO lines")

	// Phase 0: Truncate all M3 snapshot tables
	log.Println("Phase 0: Truncating all M3 snapshot tables for full refresh...")
	if err := s.db.TruncateAnalysisTables(ctx); err != nil {
		return fmt.Errorf("failed to truncate analysis tables: %w", err)
	}
	log.Println("✓ M3 snapshot tables truncated successfully")

	// Phase 1: Load MOPs with CO links
	log.Println("Phase 1: Refreshing planned manufacturing orders (MOPs) with CO links...")
	_, err := s.RefreshPlannedOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh MOPs: %w", err)
	}

	// Phase 2: Load MOs with CO links
	log.Println("Phase 2: Refreshing manufacturing orders (MOs) with CO links...")
	_, err = s.RefreshManufacturingOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh MOs: %w", err)
	}

	// Phase 3: Load all open CO lines (status < 30)
	log.Println("Phase 3: Refreshing all open customer order lines (status < 30)...")
	if err := s.RefreshOpenCustomerOrderLines(ctx); err != nil {
		return fmt.Errorf("failed to refresh CO lines: %w", err)
	}

	// Phase 4: Update unified production_orders view
	log.Println("Phase 4: Updating unified production orders view...")
	if err := s.db.UpdateProductionOrdersFromMOPs(ctx); err != nil {
		return fmt.Errorf("failed to update production orders from MOPs: %w", err)
	}
	if err := s.db.UpdateProductionOrdersFromMOs(ctx); err != nil {
		return fmt.Errorf("failed to update production orders from MOs: %w", err)
	}

	log.Println("✓ Full data refresh completed successfully")
	return nil
}

// RefreshOpenCustomerOrderLines refreshes all open CO lines (status < 30)
// This is more efficient than querying by specific order numbers when there are many orders
func (s *SnapshotService) RefreshOpenCustomerOrderLines(ctx context.Context) error {
	log.Println("Refreshing all open customer order lines (status < 30)...")

	// Build query for all open CO lines
	qb := compass.NewQueryBuilder(0)
	query := qb.BuildOpenCustomerOrderLinesQuery()

	// Execute query
	log.Println("Submitting Compass query for open CO lines...")
	results, err := s.compassClient.ExecuteQuery(ctx, query, 100000)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

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

		// Convert to database record
		attributesJSON, _ := json.Marshal(coLine.Attributes)

		dbRecord := &db.CustomerOrderLine{
			CONO:         coLine.CONO,
			DIVI:         coLine.DIVI,
			OrderNumber:  coLine.ORNO,
			LineNumber:   fmt.Sprintf("%d", coLine.PONR),
			LineSuffix:   fmt.Sprintf("%d", coLine.POSX),
			ItemNumber:   coLine.ITNO,
			ItemDesc:     coLine.ITDS,
			Status:       coLine.ORST,
			RORC:         coLine.RORC,
			RORN:         coLine.RORN,
			RORL:         coLine.RORL,
			RORX:         coLine.RORX,
			OrderedQty:   coLine.ORQT,
			DeliveredQty: coLine.DLQT,
			Attributes:   attributesJSON,
		}

		// Set dates if valid
		if coLine.DWDT != 0 {
			dbRecord.DWDT = sql.NullInt32{Int32: int32(coLine.DWDT), Valid: true}
		}
		if coLine.CODT != 0 {
			dbRecord.CODT = sql.NullInt32{Int32: int32(coLine.CODT), Valid: true}
		}
		if coLine.PLDT != 0 {
			dbRecord.PLDT = sql.NullInt32{Int32: int32(coLine.PLDT), Valid: true}
		}
		if coLine.LMDT != 0 {
			dbRecord.LMDT = sql.NullInt32{Int32: int32(coLine.LMDT), Valid: true}
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

// RefreshCustomerOrderLinesByNumbers refreshes specific CO lines by order numbers
// DEPRECATED: This can cause issues with Compass when there are many order numbers
// Use RefreshOpenCustomerOrderLines instead
func (s *SnapshotService) RefreshCustomerOrderLinesByNumbers(ctx context.Context, orderNumbers []string) error {
	if len(orderNumbers) == 0 {
		log.Println("No CO numbers to refresh")
		return nil
	}
	log.Printf("Refreshing %d specific customer order lines...", len(orderNumbers))

	// Build targeted query (no lastSyncDate needed - we want all lines for these orders)
	qb := compass.NewQueryBuilder(0)
	query := qb.BuildCustomerOrderLinesByOrderNumbersQuery(orderNumbers)

	// Execute query
	log.Println("Submitting Compass query for CO lines...")
	results, err := s.compassClient.ExecuteQuery(ctx, query, 100000)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

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

		// Convert to database record
		attributesJSON, _ := json.Marshal(coLine.Attributes)

		dbRecord := &db.CustomerOrderLine{
			CONO:         coLine.CONO,
			DIVI:         coLine.DIVI,
			OrderNumber:  coLine.ORNO,
			LineNumber:   fmt.Sprintf("%d", coLine.PONR),
			LineSuffix:   fmt.Sprintf("%d", coLine.POSX),
			ItemNumber:   coLine.ITNO,
			ItemDesc:     coLine.ITDS,
			Status:       coLine.ORST,
			RORC:         coLine.RORC,
			RORN:         coLine.RORN,
			RORL:         coLine.RORL,
			RORX:         coLine.RORX,
			OrderedQty:   coLine.ORQT,
			DeliveredQty: coLine.DLQT,
			Attributes:   attributesJSON,
		}

		// Set dates if valid
		if coLine.DWDT != 0 {
			dbRecord.DWDT = sql.NullInt32{Int32: int32(coLine.DWDT), Valid: true}
		}
		if coLine.CODT != 0 {
			dbRecord.CODT = sql.NullInt32{Int32: int32(coLine.CODT), Valid: true}
		}
		if coLine.PLDT != 0 {
			dbRecord.PLDT = sql.NullInt32{Int32: int32(coLine.PLDT), Valid: true}
		}
		if coLine.LMDT != 0 {
			dbRecord.LMDT = sql.NullInt32{Int32: int32(coLine.LMDT), Valid: true}
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
// Returns list of unique CO numbers referenced by MOs
func (s *SnapshotService) RefreshManufacturingOrders(ctx context.Context) ([]string, error) {
	log.Println("Refreshing manufacturing orders...")

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query
	qb := compass.NewQueryBuilder(fullRefreshDate)
	query := qb.BuildManufacturingOrdersQuery()

	// Execute query
	log.Println("Submitting Compass query for MOs...")
	results, err := s.compassClient.ExecuteQuery(ctx, query, 100000)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

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

		// Convert to database record
		attributesJSON, _ := json.Marshal(mo.Attributes)

		dbRecord := &db.ManufacturingOrder{
			CONO:            mo.CONO,
			DIVI:            mo.DIVI,
			Facility:        mo.FACI,
			MONumber:        mo.MFNO,
			ProductNumber:   mo.PRNO,
			ItemNumber:      mo.ITNO,
			Status:          mo.WHST,
			WHHS:            mo.WHHS,
			WMST:            mo.WMST,
			MOHS:            mo.MOHS,
			OrderedQty:      mo.ORQT,
			ManufacturedQty: mo.MAQT,
			RORC:            mo.RORC,
			RORN:            mo.RORN,
			RORL:            mo.RORL,
			RORX:            mo.RORX,
			PRHL:            mo.PRHL,
			MFHL:            mo.MFHL,
			LEVL:            mo.LEVL,
			Attributes:      attributesJSON,
		}

		// Set dates if valid
		if mo.STDT != 0 {
			dbRecord.STDT = sql.NullInt32{Int32: int32(mo.STDT), Valid: true}
		}
		if mo.FIDT != 0 {
			dbRecord.FIDT = sql.NullInt32{Int32: int32(mo.FIDT), Valid: true}
		}
		if mo.RSDT != 0 {
			dbRecord.RSDT = sql.NullInt32{Int32: int32(mo.RSDT), Valid: true}
		}
		if mo.REFD != 0 {
			dbRecord.REFD = sql.NullInt32{Int32: int32(mo.REFD), Valid: true}
		}
		if mo.LMDT != 0 {
			dbRecord.LMDT = sql.NullInt32{Int32: int32(mo.LMDT), Valid: true}
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
// Returns list of unique CO numbers referenced by MOPs
func (s *SnapshotService) RefreshPlannedOrders(ctx context.Context) ([]string, error) {
	log.Println("Refreshing planned manufacturing orders (with CO links via MPREAL)...")

	// Use full refresh date - no incremental loading
	fullRefreshDate := compass.GetFullRefreshDate()
	log.Printf("Using full refresh date: %d", fullRefreshDate)

	// Build query with MPREAL join
	qb := compass.NewQueryBuilder(fullRefreshDate)
	query := qb.BuildPlannedOrdersWithCOLinksQuery()

	// Execute query
	log.Println("Submitting Compass query for MOPs...")
	results, err := s.compassClient.ExecuteQuery(ctx, query, 100000)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

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

		// Convert to database record
		messagesJSON, _ := json.Marshal(mop.Messages)
		attributesJSON, _ := json.Marshal(mop.Attributes)

		dbRecord := &db.PlannedManufacturingOrder{
			CONO:       mop.CONO,
			DIVI:       mop.DIVI,
			MOPNumber:  fmt.Sprintf("%d", mop.PLPN),
			PLPS:       mop.PLPS,
			Facility:   mop.FACI,
			ItemNumber: mop.ITNO,
			Status:     mop.WHST,
			PSTS:       mop.PSTS,
			WHST:       mop.WHST,
			PlannedQty: mop.PPQT,
			RORC:       mop.RORC,
			RORN:       mop.RORN,
			RORL:       mop.RORL,
			RORX:       mop.RORX,
			Messages:   messagesJSON,
			Attributes: attributesJSON,
		}

		// Set dates if valid
		if mop.STDT != 0 {
			dbRecord.STDT = sql.NullInt32{Int32: int32(mop.STDT), Valid: true}
		}
		if mop.FIDT != 0 {
			dbRecord.FIDT = sql.NullInt32{Int32: int32(mop.FIDT), Valid: true}
		}
		if mop.PLDT != 0 {
			dbRecord.PLDT = sql.NullInt32{Int32: int32(mop.PLDT), Valid: true}
		}
		if mop.LMDT != 0 {
			dbRecord.LMDT = sql.NullInt32{Int32: int32(mop.LMDT), Valid: true}
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
