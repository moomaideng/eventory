package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrefersOverrideThenEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.default"), []byte("PORT=8080\nDB_DSN=default-dsn\n"), 0o600); err != nil {
		t.Fatalf("write .env.default: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=7070\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("DB_DSN", "env-dsn")

	type sample struct {
		Port  string `mapstructure:"port"`
		DBDSN string `mapstructure:"db_dsn"`
	}
	cfg, err := Load[sample](dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "7070" {
		t.Errorf("Port = %q, want .env override 7070", cfg.Port)
	}
	if cfg.DBDSN != "env-dsn" {
		t.Errorf("DBDSN = %q, want process env env-dsn", cfg.DBDSN)
	}
}

func TestRequireNonEmpty(t *testing.T) {
	if err := RequireNonEmpty(map[string]string{"PORT": "8080", "DB_DSN": "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := RequireNonEmpty(map[string]string{"PORT": "8080", "DB_DSN": "  "})
	if err == nil {
		t.Fatal("expected error for blank DB_DSN")
	}
}
