package forms

import (
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
