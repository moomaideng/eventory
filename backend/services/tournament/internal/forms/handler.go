package forms

import (
	"encoding/json"
	"errors"

	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrFieldRequired      = errors.New("field is required")
	ErrInvalidFieldValue  = errors.New("invalid field value")
	ErrInvalidFieldConfig = errors.New("invalid field configuration")
	ErrUnknownFieldType   = errors.New("unknown field type")
)

// FieldHandler encapsulates configuration and response validation logic for a specific field type.
type FieldHandler interface {
	Type() models.FieldType
	ValidateConfig(field models.FormField) error
	ValidateAnswer(field models.FormField, val any) error
}

func decodeConfig[T any](cfg any) (T, error) {
	var target T
	if cfg == nil {
		return target, nil
	}
	switch v := cfg.(type) {
	case T:
		return v, nil
	case *T:
		if v != nil {
			return *v, nil
		}
		return target, nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return target, err
	}
	if err := json.Unmarshal(data, &target); err != nil {
		return target, err
	}
	return target, nil
}
