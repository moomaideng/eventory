package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_DSN", "postgresql://example")
	t.Setenv("SUPABASE_URL", "https://example.supabase.co")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://eventory.example, https://admin.eventory.example")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Port != "9090" {
		t.Errorf("Port = %q, want %q", config.Port, "9090")
	}
	if config.DBDSN != "postgresql://example" {
		t.Errorf("DBDSN = %q, want %q", config.DBDSN, "postgresql://example")
	}
	if config.SupabaseURL != "https://example.supabase.co" {
		t.Errorf("SupabaseURL = %q, want %q", config.SupabaseURL, "https://example.supabase.co")
	}
	if config.Environment != "production" {
		t.Errorf("Environment = %q, want %q", config.Environment, "production")
	}

	wantOrigins := []string{"https://eventory.example", "https://admin.eventory.example"}
	if !reflect.DeepEqual(config.CORSOrigins, wantOrigins) {
		t.Errorf("CORSOrigins = %#v, want %#v", config.CORSOrigins, wantOrigins)
	}
}

func TestLoadFromDotEnv(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), ".env")
	contents := []byte(`PORT=7070
DB_DSN=postgresql://dotenv
SUPABASE_URL=https://dotenv.supabase.co
ENVIRONMENT=local
CORS_ALLOWED_ORIGINS=http://localhost:3000, http://localhost:3001
`)
	if err := os.WriteFile(configFile, contents, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	config, err := load(configFile)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if config.Port != "7070" {
		t.Errorf("Port = %q, want %q", config.Port, "7070")
	}
	if config.DBDSN != "postgresql://dotenv" {
		t.Errorf("DBDSN = %q, want %q", config.DBDSN, "postgresql://dotenv")
	}
	if config.SupabaseURL != "https://dotenv.supabase.co" {
		t.Errorf("SupabaseURL = %q, want %q", config.SupabaseURL, "https://dotenv.supabase.co")
	}
	if config.Environment != "local" {
		t.Errorf("Environment = %q, want %q", config.Environment, "local")
	}

	wantOrigins := []string{"http://localhost:3000", "http://localhost:3001"}
	if !reflect.DeepEqual(config.CORSOrigins, wantOrigins) {
		t.Errorf("CORSOrigins = %#v, want %#v", config.CORSOrigins, wantOrigins)
	}
}

func TestEnvironmentOverridesDotEnv(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(configFile, []byte("PORT=7070\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("PORT", "9090")

	config, err := load(configFile)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if config.Port != "9090" {
		t.Errorf("Port = %q, want environment override %q", config.Port, "9090")
	}
}
