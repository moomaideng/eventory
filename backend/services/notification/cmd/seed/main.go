package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	shareddatabase "github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/notification/config"
	notificationdatabase "github.com/moomaideng/eventory/services/notification/database"
	"github.com/moomaideng/eventory/services/notification/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	devUserID = uuid.MustParse("99999999-0000-4000-8000-000000000001")
)

func main() {
	cfg, err := config.LoadDatabase()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := shareddatabase.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := notificationdatabase.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate database: %v", err)
	}

	now := time.Now().UTC()
	sampleNotifications := []models.Notification{
		{
			ID:        uuid.MustParse("88888888-0000-4000-8000-000000000001"),
			UserID:    devUserID,
			Type:      models.NotificationTypeRegistrationApproved,
			Title:     "Registration Approved",
			Message:   "Your team Phoenix Squad has been approved for Cyber Championship 2026!",
			Link:      "/tournaments",
			IsRead:    false,
			CreatedAt: now.Add(-5 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
		},
		{
			ID:        uuid.MustParse("88888888-0000-4000-8000-000000000002"),
			UserID:    devUserID,
			Type:      models.NotificationTypePaymentConfirmed,
			Title:     "Payment Confirmed",
			Message:   "Your entry fee payment of 500.00 THB for Cyber Championship 2026 was confirmed.",
			Link:      "/payments",
			IsRead:    false,
			CreatedAt: now.Add(-30 * time.Minute),
			UpdatedAt: now.Add(-30 * time.Minute),
		},
		{
			ID:        uuid.MustParse("88888888-0000-4000-8000-000000000003"),
			UserID:    devUserID,
			Type:      models.NotificationTypeRegistrationApproved,
			Title:     "Registration Approved",
			Message:   "Your registration for Summer Invitational 2026 has been approved!",
			Link:      "/tournaments",
			IsRead:    true,
			ReadAt:    &now,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:        uuid.MustParse("88888888-0000-4000-8000-000000000004"),
			UserID:    devUserID,
			Type:      models.NotificationTypeRegistrationRejected,
			Title:     "Registration Rejected",
			Message:   "Your application for Winter Clash was rejected. Reason: Team roster incomplete.",
			Link:      "/tournaments",
			IsRead:    true,
			ReadAt:    &now,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-12 * time.Hour),
		},
		{
			ID:        uuid.MustParse("88888888-0000-4000-8000-000000000005"),
			UserID:    devUserID,
			Type:      models.NotificationTypePaymentConfirmed,
			Title:     "Payment Confirmed",
			Message:   "Your payment of 1,200.00 THB for Premier Pass was confirmed.",
			Link:      "/payments",
			IsRead:    true,
			ReadAt:    &now,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: now.Add(-47 * time.Hour),
		},
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, item := range sampleNotifications {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to seed sample notifications: %v", err)
	}

	log.Printf("seeded %d sample notifications successfully", len(sampleNotifications))
}
