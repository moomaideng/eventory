package forms

import (
	"fmt"
	"strings"

	"github.com/moomaideng/eventory/services/tournament/models"
)

type SelectOneConfig struct {
	Options []models.FormFieldOption `json:"options" bson:"options"`
}

type SelectOneHandler struct{}

func (h *SelectOneHandler) Type() models.FieldType {
	return models.FieldTypeSelectOne
}

func (h *SelectOneHandler) ValidateConfig(field models.FormField) error {
	cfg, err := decodeConfig[SelectOneConfig](field.Config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	options := cfg.Options
	if len(options) == 0 {
		return fmt.Errorf("%w: select_one field must have at least one option", ErrInvalidFieldConfig)
	}

	seenIDs := make(map[string]struct{}, len(options))
	for _, opt := range options {
		id := strings.TrimSpace(opt.ID)
		if id == "" {
			return fmt.Errorf("%w: option id cannot be empty", ErrInvalidFieldConfig)
		}
		if strings.TrimSpace(opt.Label) == "" {
			return fmt.Errorf("%w: option label cannot be empty", ErrInvalidFieldConfig)
		}
		if _, exists := seenIDs[id]; exists {
			return fmt.Errorf("%w: duplicate option id '%s'", ErrInvalidFieldConfig, id)
		}
		seenIDs[id] = struct{}{}
	}
	return nil
}

func (h *SelectOneHandler) ValidateAnswer(field models.FormField, val any) error {
	if val == nil {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("%w: expected option id string", ErrInvalidFieldValue)
	}

	if strVal == "" {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	cfg, _ := decodeConfig[SelectOneConfig](field.Config)
	for _, opt := range cfg.Options {
		if opt.ID == strVal {
			return nil
		}
	}

	return fmt.Errorf("%w: selected option '%s' is not valid", ErrInvalidFieldValue, strVal)
}
