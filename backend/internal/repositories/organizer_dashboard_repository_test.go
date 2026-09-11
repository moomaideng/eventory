package repositories_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
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
	// All fixtures, including their schema, disappear when this transaction rolls back.
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
	owner := models.Account{ID: uuid.New(), Email: "owner@example.com", Handle: "owner", DisplayName: "Owner"}
	other := models.Account{ID: uuid.New(), Email: "other@example.com", Handle: "other", DisplayName: "Other"}
	for _, account := range []*models.Account{&owner, &other} {
		if err := tx.Create(account).Error; err != nil {
			t.Fatal(err)
		}
	}
	organizer := models.OrganizerProfile{ID: uuid.New(), AccountID: owner.ID, OrganizerName: "Organizer"}
	if err := tx.Create(&organizer).Error; err != nil {
		t.Fatal(err)
	}
	tournament := models.Tournament{
		ID: uuid.New(), OrganizerID: organizer.ID, Name: "Private draft", Game: "Chess",
		StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour), RegistrationDeadline: time.Now(),
		Capacity: 8, Status: models.TournamentStatusRegistrationOpen, Published: false,
	}
	if err := tx.Create(&tournament).Error; err != nil {
		t.Fatal(err)
	}
	repo := repositories.NewOrganizerDashboardRepository(tx)
	ctx := context.Background()
	items, total, err := repo.ListOwned(ctx, owner.ID, 1, 12)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != tournament.ID || items[0].Published {
		t.Fatalf("owner should see their draft: total=%d items=%+v err=%v", total, items, err)
	}
	items, total, err = repo.ListOwned(ctx, other.ID, 1, 12)
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("non-owner list must be empty: total=%d items=%+v err=%v", total, items, err)
	}
	if _, err := repo.GetOwned(ctx, owner.ID, tournament.ID); err != nil {
		t.Fatalf("owner cannot read dashboard: %v", err)
	}
	if _, err := repo.GetOwned(ctx, other.ID, tournament.ID); !errors.Is(err, repositories.ErrTournamentNotFound) {
		t.Fatalf("non-owner dashboard: error=%v, want not found", err)
	}
	if _, err := repo.GetOwned(ctx, owner.ID, uuid.New()); !errors.Is(err, repositories.ErrTournamentNotFound) {
		t.Fatalf("missing dashboard: error=%v, want not found", err)
	}
	items, total, err = repo.ListOwned(ctx, owner.ID, 2, 1)
	if err != nil || total != 1 || len(items) != 0 {
		t.Fatalf("pagination: total=%d items=%+v err=%v", total, items, err)
	}
}
