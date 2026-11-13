package forms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/moomaideng/eventory/services/tournament/models"
)

const (
	DefaultShortTextMinLength = 0
	DefaultShortTextMaxLength = 200
	DefaultLongTextMinLength  = 0
	DefaultLongTextMaxLength  = 2000
)

func validateEmptyTextConfig(cfg any) error {
	if cfg == nil {
		return nil
	}
	switch cfg.(type) {
	case ShortTextConfig, *ShortTextConfig, LongTextConfig, *LongTextConfig:
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	var target struct{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&target); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	return nil
}

func validateTextAnswer(field models.FormField, val any, minLen, maxLen int) error {
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

	if strings.TrimSpace(strVal) == "" {
		if field.Required {
			return ErrFieldRequired
		}
		return nil
	}

	length := utf8.RuneCountInString(strVal)
	if length < minLen {
		return fmt.Errorf("%w: text length (%d) is less than minimum (%d)", ErrInvalidFieldValue, length, minLen)
	}
	if length > maxLen {
		return fmt.Errorf("%w: text length (%d) exceeds maximum (%d)", ErrInvalidFieldValue, length, maxLen)
	}

	return nil
}

// --- Short Text ---

type ShortTextConfig struct{}

type ShortTextHandler struct{}

func (h *ShortTextHandler) Type() models.FieldType {
	return models.FieldTypeShortText
}

func (h *ShortTextHandler) ValidateConfig(field models.FormField) error {
	return validateEmptyTextConfig(field.Config)
}

func (h *ShortTextHandler) ValidateAnswer(field models.FormField, val any) error {
	return validateTextAnswer(field, val, DefaultShortTextMinLength, DefaultShortTextMaxLength)
}

// --- Long Text ---

type LongTextConfig struct{}

type LongTextHandler struct{}

func (h *LongTextHandler) Type() models.FieldType {
	return models.FieldTypeLongText
}

func (h *LongTextHandler) ValidateConfig(field models.FormField) error {
	return validateEmptyTextConfig(field.Config)
}

func (h *LongTextHandler) ValidateAnswer(field models.FormField, val any) error {
	return validateTextAnswer(field, val, DefaultLongTextMinLength, DefaultLongTextMaxLength)
}

// --- Number ---

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

// --- Select One ---

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

// --- Select Multiple ---

type SelectMultipleConfig struct {
	Options       []models.FormFieldOption `json:"options" bson:"options"`
	MinSelections *int                     `json:"minSelections" bson:"min_selections"`
	MaxSelections *int                     `json:"maxSelections" bson:"max_selections"`
}

type SelectMultipleHandler struct{}

func (h *SelectMultipleHandler) Type() models.FieldType {
	return models.FieldTypeSelectMultiple
}

func (h *SelectMultipleHandler) ValidateConfig(field models.FormField) error {
	cfg, err := decodeConfig[SelectMultipleConfig](field.Config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFieldConfig, err)
	}
	options := cfg.Options
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

	cfg, _ := decodeConfig[SelectMultipleConfig](field.Config)
	validOptions := make(map[string]struct{}, len(cfg.Options))
	for _, opt := range cfg.Options {
		validOptions[opt.ID] = struct{}{}
	}

	seenSelections := make(map[string]struct{}, len(rawItems))
	for _, item := range rawItems {
		str, ok := item.(string)
		if !ok {
			return fmt.Errorf("%w: all items in selection must be strings", ErrInvalidFieldValue)
		}
		if _, valid := validOptions[str]; !valid {
			return fmt.Errorf("%w: selected option '%s' is not in configured options", ErrInvalidFieldValue, str)
		}
		if _, seen := seenSelections[str]; seen {
			return fmt.Errorf("%w: duplicate selection '%s'", ErrInvalidFieldValue, str)
		}
		seenSelections[str] = struct{}{}
	}

	count := len(rawItems)
	if cfg.MinSelections != nil && count < *cfg.MinSelections {
		return fmt.Errorf("%w: selected %d options, minimum required is %d", ErrInvalidFieldValue, count, *cfg.MinSelections)
	}
	if cfg.MaxSelections != nil && count > *cfg.MaxSelections {
		return fmt.Errorf("%w: selected %d options, maximum allowed is %d", ErrInvalidFieldValue, count, *cfg.MaxSelections)
	}

	return nil
}
