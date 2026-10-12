package config

import (
	"fmt"
	"strings"
)

const DefaultSlowRequestMS = 1000

// Logging holds shared logging settings for every service (LOGGING_* env).
// Embed with mapstructure:",squash" so dotenv flat keys match these tags.
type Logging struct {
	Level         string `mapstructure:"logging_level"`
	Format        string `mapstructure:"logging_format"`
	SlowRequestMS int    `mapstructure:"logging_slow_request_ms"`
}

// WithDefaults normalizes, fills empty fields from environment, and validates.
func (l Logging) WithDefaults(environment string) (Logging, error) {
	isProd := environment == "production"

	l.Level = strings.ToLower(strings.TrimSpace(l.Level))
	l.Format = strings.ToLower(strings.TrimSpace(l.Format))
	if l.Level == "warning" {
		l.Level = "warn"
	}

	if l.Level == "" {
		if isProd {
			l.Level = "info"
		} else {
			l.Level = "debug"
		}
	}
	if l.Format == "" {
		if isProd {
			l.Format = "json"
		} else {
			l.Format = "text"
		}
	}
	if l.SlowRequestMS <= 0 {
		l.SlowRequestMS = DefaultSlowRequestMS
	}

	switch l.Level {
	case "debug", "info", "warn", "error":
	default:
		return l, fmt.Errorf("logging_level %q is invalid", l.Level)
	}
	switch l.Format {
	case "json", "text":
	default:
		return l, fmt.Errorf("logging_format %q is invalid", l.Format)
	}
	return l, nil
}

// IsJSON reports whether the validated format is JSON.
func (l Logging) IsJSON() bool {
	return l.Format == "json"
}
