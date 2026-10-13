package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

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
// env overrides work with Unmarshal. Nested structs use mapstructure:",squash"
// with flat keys (e.g. logging_level) so dotenv stays flat.
func bindEnvs(v *viper.Viper, sample any) error {
	t := reflect.TypeOf(sample)
	return bindEnvsPrefix(v, t)
}

func bindEnvsPrefix(v *viper.Viper, t reflect.Type) error {
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

		tag := field.Tag.Get("mapstructure")
		if tag == "-" {
			continue
		}

		name, squash := parseMapstructureTag(tag)
		if squash {
			if err := bindEnvsPrefix(v, field.Type); err != nil {
				return err
			}
			continue
		}
		if name == "" {
			continue
		}

		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && !isBindLeafStruct(ft) {
			return fmt.Errorf("nested config field %q must use mapstructure:\",squash\" with flat keys (dotenv has no nesting)", name)
		}

		if err := v.BindEnv(name); err != nil {
			return fmt.Errorf("bind env %q: %w", name, err)
		}
	}
	return nil
}

func parseMapstructureTag(tag string) (name string, squash bool) {
	if tag == "" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, opt := range parts[1:] {
		if strings.TrimSpace(opt) == "squash" {
			squash = true
		}
	}
	if name == "squash" {
		return "", true
	}
	return name, squash
}

func isBindLeafStruct(t reflect.Type) bool {
	switch t {
	case reflect.TypeFor[time.Time]():
		return true
	default:
		return false
	}
}
