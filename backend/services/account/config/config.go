package config

import sharedconfig "github.com/moomaideng/eventory/internal/config"

const serviceDir = "services/account"

// Config holds account-service settings.
type Config struct {
	HTTPPort    string   `mapstructure:"http_port"`
	GRPCPort    string   `mapstructure:"grpc_port"`
	DBDSN       string   `mapstructure:"db_dsn"`
	SupabaseURL string   `mapstructure:"supabase_url"`
	Environment string   `mapstructure:"environment"`
	CORSOrigins []string `mapstructure:"cors_allowed_origins"`
}

// Load reads services/account/.env.default, then .env, then process env.
func Load() (Config, error) {
	cfg, err := sharedconfig.Load[Config](serviceDir)
	if err != nil {
		return Config{}, err
	}

	if err := sharedconfig.RequireNonEmpty(map[string]string{
		"HTTP_PORT": cfg.HTTPPort,
		"GRPC_PORT": cfg.GRPCPort,
		"DB_DSN":    cfg.DBDSN,
	}); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
