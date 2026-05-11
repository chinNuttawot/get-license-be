package main

import (
	"log"

	"get-license-be/internal/config"
	api "get-license-be/internal/http"
	"get-license-be/internal/license"
	"get-license-be/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DB)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	crypto, err := license.NewCrypto(cfg.PublicKey, cfg.BundleKey)
	if err != nil {
		log.Fatalf("crypto setup failed: %v", err)
	}
	service := license.NewService(crypto)
	licenseStore := store.NewLicenseStore(db)
	server := api.New(service, licenseStore)

	log.Printf("Get License API (Go Fiber) running on http://localhost:%s", cfg.Port)
	log.Fatal(server.Listen(cfg.Port))
}
