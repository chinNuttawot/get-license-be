package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	PublicKey string
	BundleKey string
	DB        DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		Port:      env("PORT", "4001"),
		PublicKey: os.Getenv("PUBLIC_KEY"),
		BundleKey: os.Getenv("BUNDLE_KEY"),
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     env("DB_PORT", "5432"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  sslMode(),
		},
	}
}

func (c DBConfig) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sslMode() string {
	if strings.EqualFold(os.Getenv("DB_SSL"), "false") {
		return "disable"
	}
	return "require"
}
