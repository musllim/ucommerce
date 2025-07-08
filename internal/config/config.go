// internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	MigrationsDir string
}

func Load() (*Config, error) {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "file:ecommerce.db?cache=shared&_fk=1"
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	return &Config{
		ServerAddress: serverAddress,
		DatabaseURL:   databaseURL,
		MigrationsDir: migrationsDir,
	}, nil
}
