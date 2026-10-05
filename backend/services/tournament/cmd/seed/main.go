package main

import (
	"log"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/internal/seeds"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("Starting tournament database seeding...")
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
	log.Println("Tournament database seed completed successfully")
}
