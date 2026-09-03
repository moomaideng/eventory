package main

import (
	"flag"
	"log"

	"github.com/moomaideng/eventory/internal/models"
	appconfig "github.com/moomaideng/eventory/pkg/config"
	"github.com/moomaideng/eventory/pkg/database"
)

func main() {
	config, err := appconfig.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.ConnectPostgres(config.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	reset := flag.Bool("reset", false, "Drop the public schema before migrating")
	flag.Parse()

	if *reset {
		log.Println("resetting database schema...")
		if err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error; err != nil {
			log.Fatalf("failed to reset schema: %v", err)
		}
	}

	if err := db.AutoMigrate(models.All()...); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// basically partial keyword search + case sensitive
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		log.Fatalf("failed to enable pg_trgm: %v", err)
	}
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_catalog_search
		ON tournaments USING GIN (
			(name || ' ' || game || ' ' || description) gin_trgm_ops
		)
		WHERE published = true
	`).Error; err != nil {
		log.Fatalf("failed to create tournament search index: %v", err)
	}
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tournaments_public_schedule
		ON tournaments (start_at, id)
		WHERE published = true
	`).Error; err != nil {
		log.Fatalf("failed to create tournament schedule index: %v", err)
	}

	log.Println("database migration completed")
}
