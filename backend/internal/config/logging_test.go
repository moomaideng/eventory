package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSquashedLoggingFromDotenvAndEnv(t *testing.T) {
	dir := t.TempDir()
	content := "ENVIRONMENT=development\nLOGGING_FORMAT=json\nLOGGING_SLOW_REQUEST_MS=2500\n"
	if err := os.WriteFile(filepath.Join(dir, ".env.default"), []byte(content), 0o600); err != nil {
		t.Fatalf("write .env.default: %v", err)
	}

	t.Setenv("LOGGING_LEVEL", "warn")

	type sample struct {
		Environment string  `mapstructure:"environment"`
		Logging     Logging `mapstructure:",squash"`
	}
	cfg, err := Load[sample](dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Logging, err = cfg.Logging.WithDefaults(cfg.Environment)
	if err != nil {
		t.Fatalf("WithDefaults() error = %v", err)
	}

	if cfg.Logging.Level != "warn" {
		t.Errorf("Level = %q, want warn from process env", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Format = %q, want json from dotenv", cfg.Logging.Format)
	}
	if cfg.Logging.SlowRequestMS != 2500 {
		t.Errorf("SlowRequestMS = %d, want 2500 from dotenv", cfg.Logging.SlowRequestMS)
	}
}

func TestLoggingWithDefaults(t *testing.T) {
	dev, err := Logging{}.WithDefaults("development")
	if err != nil {
		t.Fatal(err)
	}
	if dev.Level != "debug" {
		t.Errorf("dev Level = %q, want debug", dev.Level)
	}
	if dev.Format != "text" {
		t.Errorf("dev Format = %q, want text", dev.Format)
	}
	if dev.SlowRequestMS != DefaultSlowRequestMS {
		t.Errorf("SlowRequestMS = %d, want %d", dev.SlowRequestMS, DefaultSlowRequestMS)
	}
	if dev.IsJSON() {
		t.Error("development default format should be text")
	}

	prod, err := Logging{}.WithDefaults("production")
	if err != nil {
		t.Fatal(err)
	}
	if prod.Level != "info" {
		t.Errorf("prod Level = %q, want info", prod.Level)
	}
	if prod.Format != "json" {
		t.Errorf("prod Format = %q, want json", prod.Format)
	}
	if !prod.IsJSON() {
		t.Error("production default format should be json")
	}
}

func TestLoggingWithDefaultsRejectsInvalid(t *testing.T) {
	if _, err := (Logging{Level: "verbose"}).WithDefaults("development"); err == nil {
		t.Fatal("expected error for invalid level")
	}
	if _, err := (Logging{Format: "auto"}).WithDefaults("development"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestLoggingWarningAlias(t *testing.T) {
	got, err := (Logging{Level: "WARNING", Format: "JSON"}).WithDefaults("development")
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != "warn" {
		t.Errorf("Level = %q, want warn", got.Level)
	}
	if got.Format != "json" {
		t.Errorf("Format = %q, want json", got.Format)
	}
}
