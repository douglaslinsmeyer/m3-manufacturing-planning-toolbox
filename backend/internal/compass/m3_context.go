package compass

import (
	"context"
	"fmt"
	"strings"

	"github.com/pinggolf/m3-planning-tools/internal/m3api"
)

// User context from MMS060MI/GetUserInfo
type UserInfo struct {
	UserID       string `json:"USID"`
	Company      string `json:"CONO"`
	Division     string `json:"DIVI"`
	Facility     string `json:"FACI"`
	Warehouse    string `json:"WHLO"`
	FullName     string `json:"NAME"`
	Language     string `json:"LANC"`
	DateFormat   string `json:"DTFM"`
	TimeZone     string `json:"TIZO"`
}

// M3 Company from MNS095MI/Lst
type M3Company struct {
	CompanyNumber string `json:"CONO"`
	CompanyName   string `json:"CONM"`
	Currency      string `json:"LOCD"`
	Database      string `json:"DIVI"` // Database name
}

// M3 Division from MNS100MI/LstDivisions
type M3Division struct {
	CompanyNumber string `json:"CONO"`
	Division      string `json:"DIVI"`
	DivisionName  string `json:"DINM"`
	Facility      string `json:"FACI"`
	Warehouse     string `json:"WHLO"`
}

// M3 Facility from CRS008MI/ListFacility
type M3Facility struct {
	CompanyNumber string `json:"CONO"`
	Facility      string `json:"FACI"`
	FacilityName  string `json:"FACN"`
	Division      string `json:"DIVI"`
	Warehouse     string `json:"WHLO"`
}

// M3 Warehouse from MMS005MI/LstWarehouses
type M3Warehouse struct {
	CompanyNumber string `json:"CONO"`
	Warehouse     string `json:"WHLO"`
	WarehouseName string `json:"WHNM"`
	Division      string `json:"DIVI"`
	Facility      string `json:"FACI"`
}

// GetUserInfo retrieves authenticated user's default context from MMS060MI/GetUserInfo
// Note: This requires an M3 API client, not Compass
func GetUserInfo(ctx context.Context, m3Client *m3api.Client) (*UserInfo, error) {
	// Call MMS060MI/GetUserInfo
	record, err := m3Client.GetSingleRecord(ctx, "MMS060MI", "GetUserInfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Parse response
	userInfo := &UserInfo{}

	if val, ok := record["USID"].(string); ok {
		userInfo.UserID = strings.TrimSpace(val)
	}
	if val, ok := record["CONO"].(string); ok {
		userInfo.Company = strings.TrimSpace(val)
	}
	if val, ok := record["DIVI"].(string); ok {
		userInfo.Division = strings.TrimSpace(val)
	}
	if val, ok := record["FACI"].(string); ok {
		userInfo.Facility = strings.TrimSpace(val)
	}
	if val, ok := record["WHLO"].(string); ok {
		userInfo.Warehouse = strings.TrimSpace(val)
	}
	if val, ok := record["NAME"].(string); ok {
		userInfo.FullName = strings.TrimSpace(val)
	}
	if val, ok := record["LANC"].(string); ok {
		userInfo.Language = strings.TrimSpace(val)
	}
	if val, ok := record["DTFM"].(string); ok {
		userInfo.DateFormat = strings.TrimSpace(val)
	}
	if val, ok := record["TIZO"].(string); ok {
		userInfo.TimeZone = strings.TrimSpace(val)
	}

	return userInfo, nil
}

// ListCompanies retrieves all companies from MNS095MI/Lst
func ListCompanies(ctx context.Context, m3Client *m3api.Client) ([]M3Company, error) {
	// Call MNS095MI/Lst
	records, err := m3Client.GetMultipleRecords(ctx, "MNS095MI", "Lst", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}

	companies := make([]M3Company, 0, len(records))
	for _, record := range records {
		company := M3Company{}

		if val, ok := record["CONO"].(string); ok {
			company.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["CONM"].(string); ok {
			company.CompanyName = strings.TrimSpace(val)
		}
		if val, ok := record["LOCD"].(string); ok {
			company.Currency = strings.TrimSpace(val)
		}

		// Skip company 1 (system/template company)
		if company.CompanyNumber != "1" && company.CompanyNumber != "" {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// ListDivisions retrieves divisions for a company from MNS100MI/LstDivisions
func ListDivisions(ctx context.Context, m3Client *m3api.Client, companyNumber string) ([]M3Division, error) {
	// Call MNS100MI/LstDivisions with CONO parameter
	params := map[string]string{
		"CONO": companyNumber,
	}

	records, err := m3Client.GetMultipleRecords(ctx, "MNS100MI", "LstDivisions", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list divisions for company %s: %w", companyNumber, err)
	}

	divisions := make([]M3Division, 0, len(records))
	for _, record := range records {
		division := M3Division{}

		if val, ok := record["CONO"].(string); ok {
			division.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["DIVI"].(string); ok {
			division.Division = strings.TrimSpace(val)
		}
		if val, ok := record["DINM"].(string); ok {
			division.DivisionName = strings.TrimSpace(val)
		}
		if val, ok := record["FACI"].(string); ok {
			division.Facility = strings.TrimSpace(val)
		}
		if val, ok := record["WHLO"].(string); ok {
			division.Warehouse = strings.TrimSpace(val)
		}

		// Skip system divisions and empty divisions
		if division.Division != "" && division.Division != "991" && division.Division != "992" {
			divisions = append(divisions, division)
		}
	}

	return divisions, nil
}

// ListFacilities retrieves all facilities from CRS008MI/ListFacility
func ListFacilities(ctx context.Context, m3Client *m3api.Client) ([]M3Facility, error) {
	// Call CRS008MI/ListFacility
	records, err := m3Client.GetMultipleRecords(ctx, "CRS008MI", "ListFacility", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list facilities: %w", err)
	}

	facilities := make([]M3Facility, 0, len(records))
	for _, record := range records {
		facility := M3Facility{}

		if val, ok := record["CONO"].(string); ok {
			facility.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["FACI"].(string); ok {
			facility.Facility = strings.TrimSpace(val)
		}
		if val, ok := record["FACN"].(string); ok {
			facility.FacilityName = strings.TrimSpace(val)
		}
		if val, ok := record["DIVI"].(string); ok {
			facility.Division = strings.TrimSpace(val)
		}
		if val, ok := record["WHLO"].(string); ok {
			facility.Warehouse = strings.TrimSpace(val)
		}

		if facility.Facility != "" {
			facilities = append(facilities, facility)
		}
	}

	return facilities, nil
}

// ListWarehouses retrieves warehouses for a company from MMS005MI/LstWarehouses
func ListWarehouses(ctx context.Context, m3Client *m3api.Client, companyNumber string) ([]M3Warehouse, error) {
	// Call MMS005MI/LstWarehouses with CONO parameter
	params := map[string]string{
		"CONO": companyNumber,
	}

	records, err := m3Client.GetMultipleRecords(ctx, "MMS005MI", "LstWarehouses", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list warehouses for company %s: %w", companyNumber, err)
	}

	warehouses := make([]M3Warehouse, 0, len(records))
	for _, record := range records {
		warehouse := M3Warehouse{}

		if val, ok := record["CONO"].(string); ok {
			warehouse.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["WHLO"].(string); ok {
			warehouse.Warehouse = strings.TrimSpace(val)
		}
		if val, ok := record["WHNM"].(string); ok {
			warehouse.WarehouseName = strings.TrimSpace(val)
		}
		if val, ok := record["DIVI"].(string); ok {
			warehouse.Division = strings.TrimSpace(val)
		}
		if val, ok := record["FACI"].(string); ok {
			warehouse.Facility = strings.TrimSpace(val)
		}

		if warehouse.Warehouse != "" {
			warehouses = append(warehouses, warehouse)
		}
	}

	return warehouses, nil
}
