package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	sharedconfig "github.com/moomaideng/eventory/internal/config"
)

const (
	serviceDir = "services/notification"
	// ServiceName identifies this service in logs and traces.
	ServiceName = "notification"
)

type Config struct {
	HTTPPort           string               `mapstructure:"http_port"`
	DBDSN              string               `mapstructure:"db_dsn"`
	AMQPURL            string               `mapstructure:"amqp_url"`
	AMQPHost           string               `mapstructure:"amqp_host"`
	AMQPPort           string               `mapstructure:"amqp_port"`
	AMQPUsername       string               `mapstructure:"amqp_username"`
	AMQPPassword       string               `mapstructure:"amqp_password"`
	SMTPHost           string               `mapstructure:"smtp_host"`
	SMTPPort           string               `mapstructure:"smtp_port"`
	SMTPUsername       string               `mapstructure:"smtp_username"`
	SMTPPassword       string               `mapstructure:"smtp_password"`
	SMTPFromAddress    string               `mapstructure:"smtp_from_address"`
	SMTPFromName       string               `mapstructure:"smtp_from_name"`
	SMTPStartTLSRaw    string               `mapstructure:"smtp_starttls"`
	PublicWebURL       string               `mapstructure:"public_web_url"`
	Environment        string               `mapstructure:"environment"`
	RetryDelaysRaw     string               `mapstructure:"retry_delays"`
	ConsumerPrefetch   int                  `mapstructure:"consumer_prefetch"`
	ProcessingLeaseRaw string               `mapstructure:"processing_lease"`
	SMTPTimeoutRaw     string               `mapstructure:"smtp_timeout"`
	BrokerReconnectRaw string               `mapstructure:"broker_reconnect_delay"`
	Logging            sharedconfig.Logging `mapstructure:",squash"`
	OTel               sharedconfig.OTel    `mapstructure:",squash"`

	SMTPStartTLS         bool            `mapstructure:"-"`
	RetryDelays          []time.Duration `mapstructure:"-"`
	ProcessingLease      time.Duration   `mapstructure:"-"`
	SMTPTimeout          time.Duration   `mapstructure:"-"`
	BrokerReconnectDelay time.Duration   `mapstructure:"-"`
	BrokerURL            string          `mapstructure:"-"`
}

func Load() (Config, error) {
	cfg, err := sharedconfig.Load[Config](serviceDir)
	if err != nil {
		return Config{}, err
	}
	if err := sharedconfig.RequireNonEmpty(map[string]string{
		"HTTP_PORT": cfg.HTTPPort, "DB_DSN": cfg.DBDSN,
		"SMTP_HOST": cfg.SMTPHost, "SMTP_PORT": cfg.SMTPPort,
		"SMTP_FROM_ADDRESS": cfg.SMTPFromAddress, "PUBLIC_WEB_URL": cfg.PublicWebURL,
	}); err != nil {
		return Config{}, err
	}
	if cfg.AMQPURL != "" {
		cfg.BrokerURL = cfg.AMQPURL
	} else {
		if err := sharedconfig.RequireNonEmpty(map[string]string{
			"AMQP_HOST": cfg.AMQPHost, "AMQP_PORT": cfg.AMQPPort,
			"AMQP_USERNAME": cfg.AMQPUsername, "AMQP_PASSWORD": cfg.AMQPPassword,
		}); err != nil {
			return Config{}, err
		}
		if _, err := strconv.Atoi(cfg.AMQPPort); err != nil {
			return Config{}, fmt.Errorf("AMQP_PORT must be numeric: %w", err)
		}
		brokerURL := &url.URL{Scheme: "amqp", User: url.UserPassword(cfg.AMQPUsername, cfg.AMQPPassword), Host: net.JoinHostPort(cfg.AMQPHost, cfg.AMQPPort), Path: "/"}
		cfg.BrokerURL = brokerURL.String()
	}
	if cfg.ConsumerPrefetch <= 0 {
		return Config{}, fmt.Errorf("CONSUMER_PREFETCH must be positive")
	}
	if _, err := strconv.Atoi(cfg.SMTPPort); err != nil {
		return Config{}, fmt.Errorf("SMTP_PORT must be numeric: %w", err)
	}
	if address, err := mail.ParseAddress(cfg.SMTPFromAddress); err != nil || !strings.EqualFold(address.Address, cfg.SMTPFromAddress) {
		return Config{}, fmt.Errorf("SMTP_FROM_ADDRESS must be a plain valid email address")
	}
	publicURL, err := url.ParseRequestURI(cfg.PublicWebURL)
	if err != nil || (publicURL.Scheme != "http" && publicURL.Scheme != "https") || publicURL.Host == "" {
		return Config{}, fmt.Errorf("PUBLIC_WEB_URL must be an absolute http or https URL")
	}
	cfg.SMTPStartTLS, err = strconv.ParseBool(cfg.SMTPStartTLSRaw)
	if err != nil {
		return Config{}, fmt.Errorf("SMTP_STARTTLS must be true or false: %w", err)
	}
	cfg.RetryDelays, err = parseDurations(cfg.RetryDelaysRaw)
	if err != nil || len(cfg.RetryDelays) == 0 {
		return Config{}, fmt.Errorf("RETRY_DELAYS must contain durations: %w", err)
	}
	cfg.ProcessingLease, err = time.ParseDuration(cfg.ProcessingLeaseRaw)
	if err != nil || cfg.ProcessingLease <= 0 {
		return Config{}, fmt.Errorf("PROCESSING_LEASE must be a positive duration")
	}
	cfg.SMTPTimeout, err = time.ParseDuration(cfg.SMTPTimeoutRaw)
	if err != nil || cfg.SMTPTimeout <= 0 {
		return Config{}, fmt.Errorf("SMTP_TIMEOUT must be a positive duration")
	}
	cfg.BrokerReconnectDelay, err = time.ParseDuration(cfg.BrokerReconnectRaw)
	if err != nil || cfg.BrokerReconnectDelay <= 0 {
		return Config{}, fmt.Errorf("BROKER_RECONNECT_DELAY must be a positive duration")
	}
	logging, err := cfg.Logging.WithDefaults(cfg.Environment)
	if err != nil {
		return Config{}, err
	}
	cfg.Logging = logging
	return cfg, nil
}

// LoadDatabase loads only the configuration required by the migration command.
func LoadDatabase() (Config, error) {
	cfg, err := sharedconfig.Load[Config](serviceDir)
	if err != nil {
		return Config{}, err
	}
	if err := sharedconfig.RequireNonEmpty(map[string]string{"DB_DSN": cfg.DBDSN}); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func parseDurations(value string) ([]time.Duration, error) {
	parts := strings.Split(value, ",")
	out := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		delay, err := time.ParseDuration(strings.TrimSpace(part))
		if err != nil || delay <= 0 {
			return nil, fmt.Errorf("invalid retry delay %q", part)
		}
		out = append(out, delay)
	}
	return out, nil
}
