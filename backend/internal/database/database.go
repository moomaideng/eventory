package database

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

const defaultMaxOpenConns = 20

func maxOpenConns() (int, error) {
	raw := strings.TrimSpace(os.Getenv("DB_MAX_OPEN_CONNS"))
	if raw == "" {
		return defaultMaxOpenConns, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("DB_MAX_OPEN_CONNS must be a positive integer, got %q", raw)
	}
	return n, nil
}

// ConnectPostgres opens a GORM connection and instruments it with the OTel
// GORM plugin. The plugin is a harmless no-op without a global TracerProvider
// (see internal/telemetry), so it's always on rather than wired per service.
func ConnectPostgres(dsn string) (*gorm.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DB_DSN is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	maxOpen, err := maxOpenConns()
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxOpen)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("register otel gorm plugin: %w", err)
	}
	return db, nil
}

// ResetPublicSchema drops every object in public and recreates the empty schema.
func ResetPublicSchema(db *gorm.DB) error {
	return db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error
}
