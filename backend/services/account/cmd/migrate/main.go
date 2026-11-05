package main

import (
	"flag"
	"log"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/account/config"
	"github.com/moomaideng/eventory/services/account/models"
)

func main() {
	reset := flag.Bool("reset", false, "Drop the public schema before migrating")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if *reset {
		log.Println("resetting database schema...")
		if err := database.ResetPublicSchema(db); err != nil {
			log.Fatalf("failed to reset schema: %v", err)
		}
	}

	if err := db.AutoMigrate(models.All()...); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Preserve existing categories as a first tag. Old visibility flags are no
	// longer used: creating a sponsor profile is consent to directory listing.
	if db.Migrator().HasColumn(&models.SponsorProfile{}, "category") {
		if err := db.Exec(`UPDATE sponsor_profiles SET tags = jsonb_build_array(LOWER(TRIM(category)))
			WHERE tags = '[]'::jsonb AND category IS NOT NULL AND TRIM(category) <> ''`).Error; err != nil {
			log.Fatalf("failed to migrate sponsor categories to tags: %v", err)
		}
	}
	log.Println("account database migration completed")
}
