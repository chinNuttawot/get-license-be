package http

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"get-license-be/internal/httperr"
	"get-license-be/internal/license"
	"get-license-be/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Server struct {
	app     *fiber.App
	service *license.Service
	store   *store.LicenseStore
}

func New(service *license.Service, store *store.LicenseStore) *Server {
	s := &Server{service: service, store: store}
	app := fiber.New(fiber.Config{
		BodyLimit:    10 * 1024 * 1024,
		ErrorHandler: errorHandler,
	})
	app.Use(cors.New())
	s.app = app
	s.routes()
	return s
}

func (s *Server) Listen(port string) error {
	return s.app.Listen(":" + port)
}

func (s *Server) routes() {
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "get-license-api", "timestamp": time.Now().UTC().Format(time.RFC3339Nano)})
	})
	s.app.Post("/api/client-license/import", s.importLicense)
	s.app.Get("/api/client-license/all", s.getAllLicenses)
	s.app.Patch("/api/client-license/:id/status", s.toggleLicenseStatus)
	s.app.Delete("/api/client-license/:id", s.deleteLicense)
	s.app.Get("/api/client-license/:id", s.getLicenseByID)
	s.app.Get("/api/client-license", s.getLatestLicense)
	s.app.Get("/api/client-license/", s.getLatestLicense)
}

func (s *Server) importLicense(c *fiber.Ctx) error {
	var req license.ImportRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.New(400, "Invalid JSON request body.")
	}
	if strings.TrimSpace(req.FileContent) == "" {
		return httperr.New(400, "fileContent is required. Please provide the base64 string of the .aglic file.")
	}

	bundle, err := s.service.DecryptBundle(req.FileContent)
	if err != nil {
		return err
	}
	rawID, decodedID, err := s.store.Import(req.FileContent, bundle)
	if err != nil {
		return httperr.New(500, "Failed to save license to database.")
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "License imported successfully",
		"data":    fiber.Map{"rawId": rawID, "decodedId": decodedID, "meta": bundle.Meta},
	})
}

func (s *Server) getLatestLicense(c *fiber.Ctx) error {
	item, err := s.store.Latest()
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"message": "No license found."})
	}
	if err != nil {
		return err
	}
	return c.JSON(s.service.LicenseDetail(item))
}

func (s *Server) getLicenseByID(c *fiber.Ctx) error {
	item, err := s.store.ByID(c.Params("id"))
	if err == sql.ErrNoRows {
		return httperr.New(404, "License not found.")
	}
	if err != nil {
		return err
	}
	return c.JSON(s.service.LicenseDetail(item))
}

func (s *Server) getAllLicenses(c *fiber.Ctx) error {
	data, err := s.store.All()
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}

func (s *Server) toggleLicenseStatus(c *fiber.Ctx) error {
	active, err := s.store.ToggleStatus(c.Params("id"))
	if err == sql.ErrNoRows {
		return httperr.New(404, "License not found")
	}
	if err != nil {
		return err
	}
	word := "deactivated"
	if active {
		word = "activated"
	}
	return c.JSON(fiber.Map{"message": "License " + word, "data": fiber.Map{"isActive": active}})
}

func (s *Server) deleteLicense(c *fiber.Ctx) error {
	ok, err := s.store.Delete(c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httperr.New(404, "License not found")
	}
	return c.JSON(fiber.Map{"message": "License deleted successfully"})
}

func errorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(httperr.Error); ok {
		return c.Status(e.Code).JSON(fiber.Map{"status": "fail", "error": e.Message, "message": e.Message})
	}
	log.Printf("request error: %v", err)
	return c.Status(500).JSON(fiber.Map{"status": "error", "error": "Internal server error", "message": err.Error()})
}
