package main

import (
	"context"
	"flag"
	"log"

	"github.com/google/uuid"
	shareddatabase "github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/notification/config"
	notificationdatabase "github.com/moomaideng/eventory/services/notification/database"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
)

func main() {
	reset := flag.Bool("reset", false, "Drop the public schema before migrating")
	redriveEvent := flag.String("redrive-event", "", "Reset one FAILED event so its DLQ message can be republished")
	flag.Parse()
	cfg, err := config.LoadDatabase()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	db, err := shareddatabase.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if *reset {
		if err := shareddatabase.ResetPublicSchema(db); err != nil {
			log.Fatalf("failed to reset schema: %v", err)
		}
	}
	if err := notificationdatabase.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate notification database: %v", err)
	}
	if *redriveEvent != "" {
		eventID, err := uuid.Parse(*redriveEvent)
		if err != nil {
			log.Fatalf("invalid redrive event ID: %v", err)
		}
		reset, err := repositories.NewDeliveryRepository(db).ResetFailed(context.Background(), eventID)
		if err != nil {
			log.Fatalf("failed to reset notification delivery: %v", err)
		}
		if !reset {
			log.Fatalf("no FAILED delivery found for event_id=%s", eventID)
		}
		log.Printf("notification delivery reset; republish the matching DLQ message event_id=%s", eventID)
	}
	log.Println("notification database migration completed")
}
