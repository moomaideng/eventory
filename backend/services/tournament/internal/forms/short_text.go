package forms

import (
	"fmt"
	"unicode/utf8"

	"github.com/moomaideng/eventory/services/tournament/models"
)

type ShortTextConfig struct {
	MinLength *int `json:"minLength" bson:"min_length"`
	MaxLength *int `json:"maxLength" bson:"max_length"`
}

type ShortTextHandler struct{}

func (h *ShortTextHandler) Type() models.FieldType {
	return models.FieldTypeShortText
}

func (h *ShortTextHandler) ValidateConfig(field models.FormField) error {
	cfg, err := decodeConfig[ShortTextConfig](field.Config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	if cfg.MinLength != nil && *cfg.MinLength < 0 {
		return fmt.Errorf("%w: min_length cannot be negative", ErrInvalidFieldConfig)
	}
	if cfg.MaxLength != nil && *cfg.MaxLength < 0 {
		return fmt.Errorf("%w: max_length cannot be negative", ErrInvalidFieldConfig)
	}
	if cfg.MinLength != nil && cfg.MaxLength != nil && *cfg.MinLength > *cfg.MaxLength {
		return fmt.Errorf("%w: min_length cannot exceed max_length", ErrInvalidFieldConfig)
	}
	return nil
}

func (h *ShortTextHandler) ValidateAnswer(field models.FormField, val any) error {
	if val == nil {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("%w: expected string", ErrInvalidFieldValue)
	}

	if strVal == "" {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	cfg, _ := decodeConfig[ShortTextConfig](field.Config)
	length := utf8.RuneCountInString(strVal)
	if cfg.MinLength != nil && length < *cfg.MinLength {
		return fmt.Errorf("%w: text length (%d) is less than minimum (%d)", ErrInvalidFieldValue, length, *cfg.MinLength)
	}
	if cfg.MaxLength != nil && length > *cfg.MaxLength {
		return fmt.Errorf("%w: text length (%d) exceeds maximum (%d)", ErrInvalidFieldValue, length, *cfg.MaxLength)
	}

	return nil
}
