package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"get-license-be/internal/license"

	"github.com/google/uuid"
)

type LicenseStore struct {
	db *sql.DB
}

func NewLicenseStore(db *sql.DB) *LicenseStore {
	return &LicenseStore{db: db}
}

func (s *LicenseStore) Import(fileContent string, bundle license.Bundle) (string, string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	rawID := uuid.NewString()
	decodedID := uuid.NewString()
	now := time.Now().UTC()
	tokensJSON, _ := json.Marshal(bundle.Tokens)

	if _, err := tx.Exec(`INSERT INTO licenses (id, content, created_at) VALUES ($1, $2, $3)`, rawID, fileContent, now); err != nil {
		return "", "", err
	}
	_, err = tx.Exec(`INSERT INTO decoded_licenses (id, company, license_type, expiry, issued_at, tokens, is_active, is_deleted, created_at, license_id)
		VALUES ($1, $2, $3, $4, $5, $6, true, false, $7, $8)`,
		decodedID, bundle.Meta.Company, bundle.Meta.LicenseType, bundle.Meta.Expiry, bundle.Meta.IssuedAt, string(tokensJSON), now, rawID)
	if err != nil {
		return "", "", err
	}
	if err := tx.Commit(); err != nil {
		return "", "", err
	}
	return rawID, decodedID, nil
}

func (s *LicenseStore) Latest() (license.DecodedLicense, error) {
	return s.fetchOne(`SELECT id, company, license_type, expiry, issued_at, tokens, is_active, is_deleted, created_at, license_id
		FROM decoded_licenses ORDER BY created_at DESC LIMIT 1`)
}

func (s *LicenseStore) ByID(id string) (license.DecodedLicense, error) {
	return s.fetchOne(`SELECT id, company, license_type, expiry, issued_at, tokens, is_active, is_deleted, created_at, license_id
		FROM decoded_licenses WHERE id = $1 AND is_deleted = false`, id)
}

func (s *LicenseStore) All() ([]map[string]any, error) {
	rows, err := s.db.Query(`SELECT id, company, license_type, expiry, issued_at, tokens, is_active, created_at
		FROM decoded_licenses WHERE is_deleted = false ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []map[string]any{}
	for rows.Next() {
		var id, company, licenseType string
		var expiry, issuedAt, createdAt time.Time
		var tokensRaw []byte
		var isActive bool
		if err := rows.Scan(&id, &company, &licenseType, &expiry, &issuedAt, &tokensRaw, &isActive, &createdAt); err != nil {
			return nil, err
		}
		var tokens []string
		_ = json.Unmarshal(tokensRaw, &tokens)
		data = append(data, map[string]any{
			"id": id, "company": company, "licenseType": licenseType, "expiry": expiry, "issuedAt": issuedAt,
			"importedAt": createdAt, "tokenCount": len(tokens), "isActive": isActive,
		})
	}
	return data, rows.Err()
}

func (s *LicenseStore) ToggleStatus(id string) (bool, error) {
	var active bool
	err := s.db.QueryRow(`UPDATE decoded_licenses SET is_active = NOT is_active WHERE id = $1 AND is_deleted = false RETURNING is_active`, id).Scan(&active)
	return active, err
}

func (s *LicenseStore) Delete(id string) (bool, error) {
	res, err := s.db.Exec(`UPDATE decoded_licenses SET is_deleted = true WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *LicenseStore) fetchOne(query string, args ...any) (license.DecodedLicense, error) {
	var item license.DecodedLicense
	var tokensRaw []byte
	err := s.db.QueryRow(query, args...).Scan(&item.ID, &item.Company, &item.LicenseType, &item.Expiry, &item.IssuedAt, &tokensRaw, &item.IsActive, &item.IsDeleted, &item.CreatedAt, &item.LicenseID)
	if err != nil {
		return item, err
	}
	_ = json.Unmarshal(tokensRaw, &item.Tokens)
	return item, nil
}
