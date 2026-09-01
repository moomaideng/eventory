package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadFromEnvironment(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

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
