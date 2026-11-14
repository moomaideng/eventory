package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Temporary tables shadow the application tables on this transaction's connection.
// No seed data is changed, even when running against a development database.
func TestSponsorDirectoryFiltersPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run sponsor directory PostgreSQL tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "connect to the test database")
	pool, err := db.DB()
	require.NoError(t, err, "get the test connection pool")
	t.Cleanup(func() { _ = pool.Close() })
	tx := db.Begin()
	require.NoError(t, tx.Error, "start the isolated fixture transaction")
	t.Cleanup(func() { tx.Rollback() })
	for _, statement := range []string{
		"CREATE TEMP TABLE accounts (id uuid PRIMARY KEY, status text) ON COMMIT DROP",
		"CREATE TEMP TABLE sponsor_profiles (account_id uuid PRIMARY KEY, sponsor_name text, sponsor_email text, tags jsonb, privacy_consented_at timestamptz, created_at timestamptz, updated_at timestamptz) ON COMMIT DROP",
		"INSERT INTO accounts VALUES ('00000000-0000-0000-0000-000000000001','ACTIVE'), ('00000000-0000-0000-0000-000000000002','ACTIVE'), ('00000000-0000-0000-0000-000000000003','SUSPENDED')",
		`INSERT INTO sponsor_profiles (account_id, sponsor_name, sponsor_email, tags) VALUES ('00000000-0000-0000-0000-000000000001','Alpha','a@example.com','["gaming","valorant"]'), ('00000000-0000-0000-0000-000000000002','GameHub','b@example.com','["gaming","chess"]'), ('00000000-0000-0000-0000-000000000003','Hidden','c@example.com','["secret"]')`,
	} {
		if err := tx.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := NewAccountRepository(tx)
	for _, tc := range []struct {
		name, search, tags string
		total              int64
	}{
		{"partial name", "game", "", 1},
		{"partial tag", "valo", "", 1},
		{"name or tag", "gam", "", 2},
		{"all tags", "", "gaming,valorant", 1},
		{"no common sponsor", "", "valorant,chess", 0},
		{"search and tags", "alpha", "chess", 0},
		{"literal wildcard", "%", "", 0},
		{"hidden account", "", "secret", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items, total, err := repo.ListVisibleSponsorProfiles(context.Background(), tc.search, tc.tags, 1, 1)
			if err != nil || total != tc.total {
				t.Fatalf("total=%d err=%v", total, err)
			}
			if len(items) > 1 {
				t.Fatal("pagination ignored")
			}
		})
	}
	tags, err := repo.ListSponsorTags(context.Background())
	if err != nil || len(tags) != 3 {
		t.Fatalf("expected all 3 visible tags, got %v: %v", tags, err)
	}
}
