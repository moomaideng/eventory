package config

import sharedconfig "github.com/moomaideng/eventory/internal/config"

const serviceDir = "services/tournament"

// Config holds tournament-service settings.
type Config struct {
	HTTPPort        string   `mapstructure:"http_port"`
	DBDSN           string   `mapstructure:"db_dsn"`
	SupabaseURL     string   `mapstructure:"supabase_url"`
	Environment     string   `mapstructure:"environment"`
	CORSOrigins     []string `mapstructure:"cors_allowed_origins"`
	AccountGRPCAddr string   `mapstructure:"account_grpc_addr"`
	MongoURI        string   `mapstructure:"mongo_uri"`
	MongoDBName     string   `mapstructure:"mongo_db_name"`
}

// Load reads services/tournament/.env.default, then .env, then process env.
func Load() (Config, error) {
	cfg, err := sharedconfig.Load[Config](serviceDir)
	if err != nil {
		return Config{}, err
	}

	if err := sharedconfig.RequireNonEmpty(map[string]string{
		"HTTP_PORT":         cfg.HTTPPort,
		"DB_DSN":            cfg.DBDSN,
		"ACCOUNT_GRPC_ADDR": cfg.AccountGRPCAddr,
	}); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
