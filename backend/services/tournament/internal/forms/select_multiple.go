package forms

import (
	"fmt"
	"strings"

	"github.com/moomaideng/eventory/services/tournament/models"
)

type SelectMultipleHandler struct{}

func (h *SelectMultipleHandler) Type() models.FieldType {
	return models.FieldTypeSelectMultiple
}

func (h *SelectMultipleHandler) ValidateConfig(field models.FormField) error {
	options := field.Config.Options
	if len(options) == 0 {
		return fmt.Errorf("%w: select_multiple field must have at least one option", ErrInvalidFieldConfig)
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

	cfg := field.Config
	if cfg.MinSelections != nil && *cfg.MinSelections < 0 {
		return fmt.Errorf("%w: min_selections cannot be negative", ErrInvalidFieldConfig)
	}
	if cfg.MaxSelections != nil && *cfg.MaxSelections < 0 {
		return fmt.Errorf("%w: max_selections cannot be negative", ErrInvalidFieldConfig)
	}
	if cfg.MinSelections != nil && cfg.MaxSelections != nil && *cfg.MinSelections > *cfg.MaxSelections {
		return fmt.Errorf("%w: min_selections cannot exceed max_selections", ErrInvalidFieldConfig)
	}
	if cfg.MaxSelections != nil && *cfg.MaxSelections > len(options) {
		return fmt.Errorf("%w: max_selections cannot exceed total options count", ErrInvalidFieldConfig)
	}

	return nil
}

func (h *SelectMultipleHandler) ValidateAnswer(field models.FormField, val any) error {
	if val == nil {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	var rawItems []any
	switch v := val.(type) {
	case []string:
		rawItems = make([]any, len(v))
		for i, s := range v {
			rawItems[i] = s
		}
	case []any:
		rawItems = v
	default:
		return fmt.Errorf("%w: expected array of option ids", ErrInvalidFieldValue)
	}

	if len(rawItems) == 0 {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	validOptions := make(map[string]struct{}, len(field.Config.Options))
	for _, opt := range field.Config.Options {
		validOptions[opt.ID] = struct{}{}
	}

	seenSelections := make(map[string]struct{}, len(rawItems))
	for _, item := range rawItems {
		str, ok := item.(string)
		if !ok {
			return fmt.Errorf("%w: array element must be string option id", ErrInvalidFieldValue)
		}
		if _, valid := validOptions[str]; !valid {
			return fmt.Errorf("%w: selected option '%s' is not valid", ErrInvalidFieldValue, str)
		}
		if _, duplicate := seenSelections[str]; duplicate {
			return fmt.Errorf("%w: duplicate selection '%s'", ErrInvalidFieldValue, str)
		}
		seenSelections[str] = struct{}{}
	}

	cfg := field.Config
	count := len(rawItems)
	if cfg.MinSelections != nil && count < *cfg.MinSelections {
		return fmt.Errorf("%w: selected items count (%d) is less than minimum (%d)", ErrInvalidFieldValue, count, *cfg.MinSelections)
	}
	if cfg.MaxSelections != nil && count > *cfg.MaxSelections {
		return fmt.Errorf("%w: selected items count (%d) exceeds maximum (%d)", ErrInvalidFieldValue, count, *cfg.MaxSelections)
	}

	return nil
}
