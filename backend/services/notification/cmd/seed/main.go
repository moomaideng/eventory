package main

import (
	"fmt"
	"log"
	"time"

	shareddatabase "github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/notification/config"
	notificationdatabase "github.com/moomaideng/eventory/services/notification/database"
	"github.com/moomaideng/eventory/services/notification/models"
	"gorm.io/gorm/clause"
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

	now := time.Now().UTC().Truncate(time.Second)
	readAt := now.Add(-2 * time.Hour)

	sampleNotifications := []models.Notification{
		// Alex Dev: mix of unread + read notifications across all 4 types
		{
			ID:        seedids.NotificationID(1),
			UserID:    seedids.AlexID,
			Type:      models.NotificationTypeAnnouncementCreated,
			Title:     "New Announcement: ASEAN Dota 2 Invitational 2026",
			Message:   "Quarterfinals 3 & 4 broadcast schedule is now live on Main Stage A and Stream Booth 2.",
			Link:      fmt.Sprintf("/tournaments/%s/announcements", seedids.TournRealOngoingDota2),
			IsRead:    false,
			CreatedAt: now.Add(-15 * time.Minute),
			UpdatedAt: now.Add(-15 * time.Minute),
		},
		{
			ID:        seedids.NotificationID(2),
			UserID:    seedids.AlexID,
			Type:      models.NotificationTypePaymentConfirmed,
			Title:     "Sponsorship Payment Confirmed",
			Message:   "Your THB 40,000.00 Platinum sponsorship for ASEAN Dota 2 Invitational 2026 was confirmed.",
			Link:      fmt.Sprintf("/payments/%s/receipt", seedids.InvoiceID(14)),
			IsRead:    false,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        seedids.NotificationID(3),
			UserID:    seedids.AlexID,
			Type:      models.NotificationTypeRegistrationApproved,
			Title:     "Tournament Roster Highlight",
			Message:   "SEA Valorant Challengers Open now has 4 accepted squads and 2 locked teams waiting in your review queue.",
			Link:      fmt.Sprintf("/tournaments/%s", seedids.TournRealRegValorant),
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-12 * time.Hour),
			UpdatedAt: readAt,
		},
		{
			ID:        seedids.NotificationID(4),
			UserID:    seedids.AlexID,
			Type:      models.NotificationTypeRegistrationRejected,
			Title:     "Refund Processed on Test Queue",
			Message:   "Rejected entry Silom Strikers (TM050009) in [TEST · Review] Queue, Forms & Refunds has been refunded.",
			Link:      fmt.Sprintf("/tournaments/%s", seedids.TournTestReviewQueue),
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: readAt,
		},

		// Bob Dev: Organizer notifications
		{
			ID:        seedids.NotificationID(5),
			UserID:    seedids.BobID,
			Type:      models.NotificationTypePaymentConfirmed,
			Title:     "New Sponsorship Received",
			Message:   "HyperVolt Energy confirmed a THB 35,000.00 Title Partner pledge for Bangkok Cyber Arena 2026.",
			Link:      fmt.Sprintf("/tournaments/%s", seedids.TournRealCrowdBCA),
			IsRead:    false,
			CreatedAt: now.Add(-25 * time.Minute),
			UpdatedAt: now.Add(-25 * time.Minute),
		},
		{
			ID:        seedids.NotificationID(6),
			UserID:    seedids.BobID,
			Type:      models.NotificationTypeAnnouncementCreated,
			Title:     "Announcement Posted on ASEAN Dota 2 Invitational 2026",
			Message:   "Alex Dev posted: Quarterfinals 3 & 4 Broadcast Times.",
			Link:      fmt.Sprintf("/tournaments/%s/announcements", seedids.TournRealOngoingDota2),
			IsRead:    false,
			CreatedAt: now.Add(-6 * time.Hour),
			UpdatedAt: now.Add(-6 * time.Hour),
		},
		{
			ID:        seedids.NotificationID(7),
			UserID:    seedids.BobID,
			Type:      models.NotificationTypeRegistrationApproved,
			Title:     "Admin Review Logged",
			Message:   "Aurora Bangkok was approved in SEA Valorant Challengers Open.",
			Link:      fmt.Sprintf("/tournaments/%s", seedids.TournRealRegValorant),
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-36 * time.Hour),
			UpdatedAt: readAt,
		},

		// Carter Dev: Sponsor notifications
		{
			ID:        seedids.NotificationID(8),
			UserID:    seedids.CarterID,
			Type:      models.NotificationTypePaymentConfirmed,
			Title:     "Sponsorship Receipt Ready",
			Message:   "Your THB 20,000.00 Platinum pledge for [TEST · Crowdfunding] Goal Reached is confirmed.",
			Link:      fmt.Sprintf("/payments/%s/receipt", seedids.InvoiceID(1)),
			IsRead:    false,
			CreatedAt: now.Add(-35 * time.Minute),
			UpdatedAt: now.Add(-35 * time.Minute),
		},
		{
			ID:        seedids.NotificationID(9),
			UserID:    seedids.CarterID,
			Type:      models.NotificationTypeAnnouncementCreated,
			Title:     "Sponsored Event Update: Thailand Mobile Legends Championship S4",
			Message:   "Siam Titans clinched the Season 4 Championship! View final bracket & sponsor credits.",
			Link:      fmt.Sprintf("/tournaments/%s/announcements", seedids.TournRealCompletedMLBB),
			IsRead:    true,
			ReadAt:    &readAt,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: readAt,
		},
	}

	// Preserve read/unread state toggled during development.
	for _, n := range sampleNotifications {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&n).Error; err != nil {
			log.Fatalf("failed to seed notification %s: %v", n.ID, err)
		}
	}

	log.Println("Notification database seed completed successfully")
}
