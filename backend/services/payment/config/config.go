package config

import (
	"fmt"
	sharedconfig "github.com/moomaideng/eventory/internal/config"
)

type Config struct {
	HTTPPort    string   `mapstructure:"http_port"`
	DBDSN       string   `mapstructure:"db_dsn"`
	SupabaseURL string   `mapstructure:"supabase_url"`
	Environment string   `mapstructure:"environment"`
	CORSOrigins []string `mapstructure:"cors_allowed_origins"`
	ServiceKey  string   `mapstructure:"payment_service_key"`
}

func Load() (Config, error) {
	cfg, err := sharedconfig.Load[Config]("services/payment")
	if err != nil {
		return Config{}, err
	}
	if err := sharedconfig.RequireNonEmpty(map[string]string{"HTTP_PORT": cfg.HTTPPort, "DB_DSN": cfg.DBDSN}); err != nil {
		return Config{}, err
	}
	if len(cfg.ServiceKey) < 32 {
		return Config{}, fmt.Errorf("PAYMENT_SERVICE_KEY must contain at least 32 characters")
	}
	if cfg.Environment == "production" && cfg.ServiceKey == "local-payment-service-key-not-for-production" {
		return Config{}, fmt.Errorf("production requires a private PAYMENT_SERVICE_KEY")
	}
	return cfg, nil
}
