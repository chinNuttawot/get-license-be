package license

import (
	"database/sql"
	"time"
)

type ImportRequest struct {
	FileContent string `json:"fileContent"`
}

type BundleEnvelope struct {
	V    int    `json:"v"`
	Alg  string `json:"alg"`
	IV   string `json:"iv"`
	Tag  string `json:"tag"`
	Data string `json:"data"`
}

type Bundle struct {
	Version int      `json:"version"`
	Meta    Meta     `json:"meta"`
	Tokens  []string `json:"tokens"`
}

type Meta struct {
	Company     string    `json:"company"`
	LicenseType string    `json:"licenseType"`
	Expiry      time.Time `json:"expiry"`
	IssuedAt    time.Time `json:"issuedAt"`
}

type DecodedLicense struct {
	ID          string
	Company     string
	LicenseType string
	Expiry      time.Time
	IssuedAt    time.Time
	Tokens      []string
	IsActive    bool
	IsDeleted   bool
	CreatedAt   time.Time
	LicenseID   sql.NullString
}
