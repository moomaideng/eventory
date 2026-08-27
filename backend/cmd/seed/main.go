package main

import (
	"log"

	"github.com/moomaideng/eventory/internal/seeds"
	appconfig "github.com/moomaideng/eventory/pkg/config"
	"github.com/moomaideng/eventory/pkg/database"
	"gorm.io/gorm"
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

	log.Println("Starting database seeding...")

	// Run all seeds inside a single transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, task := range seeds.All() {
			log.Printf("Running seed task: %s", task.Name)
			if err := task.Run(tx); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	log.Println("Database seed completed successfully")
}
