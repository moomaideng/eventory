package repositories_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOrganizerDashboardOwnershipPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the PostgreSQL ownership integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	schema := "dashboard_test_" + uuid.New().String()[:8]
	if err := tx.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}

	organizerID := uuid.New()
	otherOrganizerID := uuid.New()
	tournament := models.Tournament{
		ID: uuid.New(), OrganizerID: organizerID, Name: "Private draft", Game: "Chess",
		StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour), RegistrationDeadline: time.Now(),
		Capacity: 8, Status: models.TournamentStatusRegistrationOpen, Published: false,
	}
	if err := tx.Create(&tournament).Error; err != nil {
		t.Fatal(err)
	}
	repo := repositories.NewOrganizerDashboardRepository(tx)
	ctx := context.Background()
	items, total, err := repo.ListOwned(ctx, organizerID, 1, 12)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != tournament.ID || items[0].Published {
		t.Fatalf("owner should see their draft: total=%d items=%+v err=%v", total, items, err)
	}
	items, total, err = repo.ListOwned(ctx, otherOrganizerID, 1, 12)
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("non-owner list must be empty: total=%d items=%+v err=%v", total, items, err)
	}
	if _, err := repo.GetOwned(ctx, organizerID, tournament.ID); err != nil {
		t.Fatalf("owner cannot read dashboard: %v", err)
	}
	if _, err := repo.GetOwned(ctx, otherOrganizerID, tournament.ID); !errors.Is(err, repositories.ErrTournamentNotFound) {
		t.Fatalf("non-owner dashboard: error=%v, want not found", err)
	}
	if _, err := repo.GetOwned(ctx, organizerID, uuid.New()); !errors.Is(err, repositories.ErrTournamentNotFound) {
		t.Fatalf("missing dashboard: error=%v, want not found", err)
	}
	items, total, err = repo.ListOwned(ctx, organizerID, 2, 1)
	if err != nil || total != 1 || len(items) != 0 {
		t.Fatalf("pagination: total=%d items=%+v err=%v", total, items, err)
	}
}
