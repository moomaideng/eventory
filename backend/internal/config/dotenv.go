package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Load reads serviceDir/.env.default, then optional .env, overlays process
// env, and unmarshals into T (mapstructure tags).
func Load[T any](serviceDir string) (T, error) {
	var cfg T

	v, err := mergeDotEnv(serviceDir)
	if err != nil {
		return cfg, err
	}
	if err := bindEnvs(v, cfg); err != nil {
		return cfg, err
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

// RequireNonEmpty fails if any named field is blank after trimming.
func RequireNonEmpty(fields map[string]string) error {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func mergeDotEnv(serviceDir string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("dotenv")

	for _, name := range []string{".env.default", ".env"} {
		path := filepath.Join(serviceDir, name)
		v.SetConfigFile(path)
		if err := v.MergeInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if errors.As(err, &notFound) || os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read config file %q: %w", path, err)
		}
	}

	v.AutomaticEnv()
	return v, nil
}

// bindEnvs registers BindEnv for every mapstructure-tagged field so process
// env overrides work with Unmarshal.
func bindEnvs(v *viper.Viper, sample any) error {
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("config type must be a struct, got %s", t.Kind())
	}
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}
		key := field.Tag.Get("mapstructure")
		if key == "" || key == "-" {
			continue
		}
		if err := v.BindEnv(key); err != nil {
			return fmt.Errorf("bind env %q: %w", key, err)
		}
	}
	return nil
}
