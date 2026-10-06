package database

import (
	"errors"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(dsn string) (*gorm.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DB_DSN is required")
	}

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// ResetPublicSchema drops every object in public and recreates the empty schema.
func ResetPublicSchema(db *gorm.DB) error {
	return db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error
}
