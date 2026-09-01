package config

import (
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

// Load reads configuration values from the process environment.
func Load() (Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SUPABASE_URL", "")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")
	viper.AutomaticEnv()

	return Config{
		Port:        viper.GetString("PORT"),
		DBDSN:       viper.GetString("DB_DSN"),
		SupabaseURL: viper.GetString("SUPABASE_URL"),
		Environment: viper.GetString("ENVIRONMENT"),
		CORSOrigins: splitCommaSeparated(viper.GetString("CORS_ALLOWED_ORIGINS")),
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
