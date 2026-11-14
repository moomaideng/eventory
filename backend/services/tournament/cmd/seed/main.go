package main

import (
	"context"
	"log"
	"time"

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

	if cfg.MongoURI == "" {
		log.Println("Warning: MONGO_URI is empty; skipping registration form seeding")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		mongoClient, err := database.ConnectMongo(ctx, cfg.MongoURI)
		if err != nil {
			log.Fatalf("failed to connect to mongodb: %v", err)
		}
		defer func() { _ = mongoClient.Disconnect(context.Background()) }()

		dbName := cfg.MongoDBName
		if dbName == "" {
			dbName = "tournament_db"
		}
		log.Println("Running seed task: Registration Forms & Responses (MongoDB)")
		if err := seeds.SeedForms(ctx, mongoClient.Database(dbName)); err != nil {
			log.Fatalf("failed to seed mongodb forms: %v", err)
		}
	}

	log.Println("Tournament database seed completed successfully")
}
