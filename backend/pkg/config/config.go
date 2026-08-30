package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port  string
	DBDSN string
}

func Load() (Config, error) {
	viper.SetConfigFile(".env")
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return Config{}, fmt.Errorf("read .env: %w", err)
		}
	}

	return Config{
		Port:  viper.GetString("PORT"),
		DBDSN: viper.GetString("DB_DSN"),
	}, nil
}
