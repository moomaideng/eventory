package database

import (
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	_ = db.Exec("DROP TABLE IF EXISTS tournament_status_changes CASCADE").Error
	_ = db.Exec("ALTER TABLE IF EXISTS sponsorship_tiers DROP COLUMN IF EXISTS sort_order").Error

	if err := db.AutoMigrate(models.All()...); err != nil {
		return err
	}
	if err := db.Exec("DROP INDEX IF EXISTS idx_tournament_team_name").Error; err != nil {
		return err
	}
	_ = db.Exec("DROP INDEX IF EXISTS idx_tournaments_catalog_search").Error
	_ = db.Exec("DROP INDEX IF EXISTS idx_tournaments_public_schedule").Error

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_catalog_search
		ON tournaments USING GIN (
			(name || ' ' || game || ' ' || description) gin_trgm_ops
		)
		WHERE status != 'DRAFT'
	`).Error; err != nil {
		return err
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_public_schedule
		ON tournaments (start_at, id)
		WHERE status != 'DRAFT'
	`).Error
}
