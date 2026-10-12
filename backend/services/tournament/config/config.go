package config

import (
	"errors"
	sharedconfig "github.com/moomaideng/eventory/internal/config"
	"net/url"
)

const serviceDir = "services/tournament"

// Config holds tournament-service settings.
type Config struct {
	PaymentURL        string   `mapstructure:"payment_url"`
	PaymentServiceKey string   `mapstructure:"payment_service_key"`
	HTTPPort          string   `mapstructure:"http_port"`
	DBDSN             string   `mapstructure:"db_dsn"`
	SupabaseURL       string   `mapstructure:"supabase_url"`
	Environment       string   `mapstructure:"environment"`
	CORSOrigins       []string `mapstructure:"cors_allowed_origins"`
	AccountGRPCAddr   string   `mapstructure:"account_grpc_addr"`
	MongoURI          string   `mapstructure:"mongo_uri"`
	MongoDBName       string   `mapstructure:"mongo_db_name"`
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

	paymentURL, err := url.Parse(cfg.PaymentURL)
	if err != nil || paymentURL.Host == "" || (paymentURL.Scheme != "http" && paymentURL.Scheme != "https") {
		return Config{}, errors.New("PAYMENT_URL must be an HTTP service URL")
	}
	if len(cfg.PaymentServiceKey) < 32 || (cfg.Environment == "production" && cfg.PaymentServiceKey == "local-payment-service-key-not-for-production") {
		return Config{}, errors.New("PAYMENT_SERVICE_KEY must be at least 32 characters and use a private production value")
	}
	return cfg, nil
}
