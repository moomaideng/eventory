// Package testutil provides shared infrastructure for integration tests.
package testutil

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresTransaction opens TEST_DATABASE_DSN and rolls back the transaction
// before closing its connection pool when the test and its subtests finish.
// Callers must isolate their fixtures using temporary tables or a test schema.
func PostgresTransaction(t testing.TB) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is required for PostgreSQL query tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}
