package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

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
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("register otel gorm plugin: %w", err)
	}
	return db, nil
}

// ResetPublicSchema drops every object in public and recreates the empty schema.
func ResetPublicSchema(db *gorm.DB) error {
	return db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error
}
