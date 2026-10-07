package database

import (
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(models.All()...); err != nil {
		return err
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_catalog_search
		ON tournaments USING GIN (
			(name || ' ' || game || ' ' || description) gin_trgm_ops
		)
		WHERE published = true
	`).Error; err != nil {
		return err
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_public_schedule
		ON tournaments (start_at, id)
		WHERE published = true
	`).Error
}
