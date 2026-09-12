package apptest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/moomaideng/eventory/internal/models"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const postgresImage = "postgres:18-alpine"

// Postgres manages the connection string for a disposable Postgres test instance.
type Postgres struct {
	DSN string
}

// StartPostgres launches an isolated container (or reuses TEST_DB_DSN if provided)
// and runs AutoMigrate for all registered database models.
func StartPostgres(ctx context.Context, tb testing.TB) *Postgres {
	tb.Helper()

	// 1. Direct DSN override for environments without Docker
	if dsn := os.Getenv("TEST_DB_DSN"); dsn != "" {
		if err := migrateUp(dsn); err != nil {
			tb.Fatalf("apptest: migrate up on TEST_DB_DSN: %v", err)
		}
		return &Postgres{DSN: dsn}
	}

	// 2. Start disposable Docker container via testcontainers-go
	container, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase("eventory_test"),
		postgres.WithUsername("eventory_test"),
		postgres.WithPassword("eventory_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		tb.Fatalf("apptest: start postgres container: %v", err)
	}

	tb.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			tb.Logf("apptest: terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatalf("apptest: read postgres connection string: %v", err)
	}

	if err := migrateUp(dsn); err != nil {
		tb.Fatalf("apptest: migrate up: %v", err)
	}

	return &Postgres{DSN: dsn}
}

func migrateUp(dsn string) error {
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return db.AutoMigrate(models.All()...)
}
