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

	db, err := database.Open(config.DBDSN)
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

	log.Println("database migration completed")
}
