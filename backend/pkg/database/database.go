package database

import (
	"errors"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DB_DSN is required")
	}

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
