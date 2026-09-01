package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration properties for the backend service.
type Config struct {
	Port        string
	DBDSN       string
	SupabaseURL string
	Environment string
}

// Load reads configuration values from environment variables or .env file.
func Load() (Config, error) {
	viper.SetConfigFile(".env")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SUPABASE_URL", "")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read .env: %w", err)
	}

	return Config{
		Port:        viper.GetString("PORT"),
		DBDSN:       viper.GetString("DB_DSN"),
		SupabaseURL: viper.GetString("SUPABASE_URL"),
		Environment: viper.GetString("ENVIRONMENT"),
	}, nil
}
