package infor

// UserProfileResponse is the root response from Infor User Management API
type UserProfileResponse struct {
	Response struct {
		UserList []UserProfile `json:"userlist"`
	} `json:"response"`
}

// UserProfile represents a user's profile information
type UserProfile struct {
	ID          string  `json:"id"`
	UserName    string  `json:"userName"`
	DisplayName string  `json:"displayName"`
	Name        Name    `json:"name"`
	Emails      []Email `json:"emails"`
	Title       string  `json:"title,omitempty"`
	Department  string  `json:"department,omitempty"`
	Groups      []Group `json:"groups,omitempty"`
}

// Name represents a user's given and family name
type Name struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

// Email represents an email address
type Email struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	Primary bool   `json:"primary"`
}

// Group represents a security role, accounting entity, or distribution group
type Group struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type"` // "Security Role", "Accounting Entity", "Distribution Group"
}

// M3UserInfo represents M3-specific user defaults and preferences from CRS650MI/GetUserInfo
type M3UserInfo struct {
	UserID           string `json:"userId"`           // ZZUSID - M3 User ID
	FullName         string `json:"fullName"`         // USFN - User full name
	DefaultCompany   string `json:"defaultCompany"`   // ZDCONO - Default company
	DefaultDivision  string `json:"defaultDivision"`  // ZDDIVI - Default division
	DefaultFacility  string `json:"defaultFacility"`  // ZDFACI - Default facility
	DefaultWarehouse string `json:"defaultWarehouse"` // ZZWHLO - Default warehouse
	LanguageCode     string `json:"languageCode"`     // ZDLANC - Language code
	DateFormat       string `json:"dateFormat"`       // ZDDTFM - Date format (YMD/MDY/DMY)
	DateSeparator    string `json:"dateSeparator"`    // DSEP - Date separator
	TimeSeparator    string `json:"timeSeparator"`    // TSEP - Time separator
	TimeZone         string `json:"timeZone"`         // TIZO - Time zone
}

// CombinedUserProfile combines Infor user management data with M3-specific user info
type CombinedUserProfile struct {
	UserProfile
	M3Info *M3UserInfo `json:"m3Info,omitempty"`
}
