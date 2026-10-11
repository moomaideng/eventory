package forms

import (
	"fmt"
	"math"

	"github.com/moomaideng/eventory/services/tournament/models"
)

type NumberConfig struct {
	Min           *float64 `json:"min" bson:"min"`
	Max           *float64 `json:"max" bson:"max"`
	AllowDecimals *bool    `json:"allowDecimals" bson:"allow_decimals"`
}

type NumberHandler struct{}

func (h *NumberHandler) Type() models.FieldType {
	return models.FieldTypeNumber
}

func (h *NumberHandler) ValidateConfig(field models.FormField) error {
	cfg, err := decodeConfig[NumberConfig](field.Config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	if cfg.Min != nil && cfg.Max != nil && *cfg.Min > *cfg.Max {
		return fmt.Errorf("%w: min cannot exceed max", ErrInvalidFieldConfig)
	}
	return nil
}

func (h *NumberHandler) ValidateAnswer(field models.FormField, val any) error {
	if val == nil {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	var num float64
	switch v := val.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case int32:
		num = float64(v)
	default:
		return fmt.Errorf("%w: expected number", ErrInvalidFieldValue)
	}

	cfg, _ := decodeConfig[NumberConfig](field.Config)
	if cfg.AllowDecimals != nil && !*cfg.AllowDecimals {
		if num != math.Trunc(num) {
			return fmt.Errorf("%w: decimals are not allowed for this field", ErrInvalidFieldValue)
		}
	}

	if cfg.Min != nil && num < *cfg.Min {
		return fmt.Errorf("%w: value (%v) is less than minimum (%v)", ErrInvalidFieldValue, num, *cfg.Min)
	}
	if cfg.Max != nil && num > *cfg.Max {
		return fmt.Errorf("%w: value (%v) exceeds maximum (%v)", ErrInvalidFieldValue, num, *cfg.Max)
	}

	return nil
}
