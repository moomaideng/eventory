package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration properties for the backend service.
type Config struct {
	Port        string
	DBDSN       string
	SupabaseURL string
	Environment string
	CORSOrigins []string
}

// Load reads optional local defaults from .env and lets process environment
// variables override them. Production and containers should inject environment
// variables instead of copying an .env file into the image.
func Load() (Config, error) {
	return load(".env")
}

func load(configFile string) (Config, error) {
	v := viper.New()
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("SUPABASE_URL", "")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")

	v.SetConfigFile(configFile)
	v.SetConfigType("dotenv")
	if err := v.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) && !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read config file %q: %w", configFile, err)
		}
	}

	// Environment variables have higher precedence than values from .env.
	v.AutomaticEnv()

	return Config{
		Port:        v.GetString("PORT"),
		DBDSN:       v.GetString("DB_DSN"),
		SupabaseURL: v.GetString("SUPABASE_URL"),
		Environment: v.GetString("ENVIRONMENT"),
		CORSOrigins: splitCommaSeparated(v.GetString("CORS_ALLOWED_ORIGINS")),
	}, nil
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
