package store

import (
	"database/sql"

	"get-license-be/internal/config"

	_ "github.com/lib/pq"
)

func Open(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.ConnString())
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS licenses (id uuid PRIMARY KEY, content text NOT NULL, created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS decoded_licenses (
			id uuid PRIMARY KEY,
			company varchar NOT NULL,
			license_type varchar NOT NULL,
			expiry timestamp NOT NULL,
			issued_at timestamp NOT NULL,
			tokens jsonb NOT NULL,
			is_active boolean NOT NULL DEFAULT true,
			is_deleted boolean NOT NULL DEFAULT false,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
			license_id uuid NULL REFERENCES licenses(id) ON DELETE CASCADE,
			software_version varchar NULL,
			external_license_id varchar NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
