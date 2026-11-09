package config

import (
	"errors"
	"net"

	sharedconfig "github.com/moomaideng/eventory/internal/config"
)

const (
	serviceDir = "services/tournament"
	// ServiceName identifies this service in logs and traces.
	ServiceName = "tournament"
)

// Config holds tournament-service settings.
type Config struct {
	BrokerURL         string               `mapstructure:"broker_url"`
	PaymentGRPCAddr   string               `mapstructure:"payment_grpc_addr"`
	PaymentServiceKey string               `mapstructure:"payment_service_key"`
	HTTPPort          string               `mapstructure:"http_port"`
	DBDSN             string               `mapstructure:"db_dsn"`
	SupabaseURL       string               `mapstructure:"supabase_url"`
	Environment       string               `mapstructure:"environment"`
	CORSOrigins       []string             `mapstructure:"cors_allowed_origins"`
	AccountGRPCAddr   string               `mapstructure:"account_grpc_addr"`
	MongoURI          string               `mapstructure:"mongo_uri"`
	MongoDBName       string               `mapstructure:"mongo_db_name"`
	Logging           sharedconfig.Logging `mapstructure:",squash"`
	OTel              sharedconfig.OTel    `mapstructure:",squash"`
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

	if host, port, err := net.SplitHostPort(cfg.PaymentGRPCAddr); err != nil || host == "" || port == "" {
		return Config{}, errors.New("PAYMENT_GRPC_ADDR must be a host:port address")
	}
	if len(cfg.PaymentServiceKey) < 32 || (cfg.Environment == "production" && cfg.PaymentServiceKey == "local-payment-service-key-not-for-production") {
		return Config{}, errors.New("PAYMENT_SERVICE_KEY must be at least 32 characters and use a private production value")
	}

	logging, err := cfg.Logging.WithDefaults(cfg.Environment)
	if err != nil {
		return Config{}, err
	}
	cfg.Logging = logging
	return cfg, nil
}
