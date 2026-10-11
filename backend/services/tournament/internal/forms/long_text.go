package forms

import (
	"fmt"
	"unicode/utf8"

	"github.com/moomaideng/eventory/services/tournament/models"
)

type LongTextHandler struct{}

func (h *LongTextHandler) Type() models.FieldType {
	return models.FieldTypeLongText
}

func (h *LongTextHandler) ValidateConfig(field models.FormField) error {
	cfg := field.Config
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

func (h *LongTextHandler) ValidateAnswer(field models.FormField, val any) error {
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

	length := utf8.RuneCountInString(strVal)
	cfg := field.Config
	if cfg.MinLength != nil && length < *cfg.MinLength {
		return fmt.Errorf("%w: text length (%d) is less than minimum (%d)", ErrInvalidFieldValue, length, *cfg.MinLength)
	}
	if cfg.MaxLength != nil && length > *cfg.MaxLength {
		return fmt.Errorf("%w: text length (%d) exceeds maximum (%d)", ErrInvalidFieldValue, length, *cfg.MaxLength)
	}

	return nil
}
