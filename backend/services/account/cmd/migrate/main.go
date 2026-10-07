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

	log.Println("account database migration completed")
}
