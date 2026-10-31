package forms

import (
	"fmt"
	"strings"

	"github.com/moomaideng/eventory/services/tournament/models"
)

// FieldRegistry manages handlers for all supported field types.
type FieldRegistry struct {
	handlers map[models.FieldType]FieldHandler
}

// NewDefaultRegistry instantiates a registry pre-populated with the five standard field handlers.
func NewDefaultRegistry() *FieldRegistry {
	r := &FieldRegistry{
		handlers: make(map[models.FieldType]FieldHandler),
	}
	r.Register(&ShortTextHandler{})
	r.Register(&LongTextHandler{})
	r.Register(&NumberHandler{})
	r.Register(&SelectOneHandler{})
	r.Register(&SelectMultipleHandler{})
	return r
}

// Register registers or replaces a field handler for a given type.
func (r *FieldRegistry) Register(h FieldHandler) {
	r.handlers[h.Type()] = h
}

// Get looks up a handler by field type.
func (r *FieldRegistry) Get(t models.FieldType) (FieldHandler, bool) {
	h, ok := r.handlers[t]
	return h, ok
}

// ValidateFieldConfig verifies that the field has a known type and valid configuration.
func (r *FieldRegistry) ValidateFieldConfig(field models.FormField) error {
	handler, ok := r.Get(field.Type)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownFieldType, field.Type)
	}
	return handler.ValidateConfig(field)
}

// ValidateFieldAnswer verifies that the answer conforms to the field's type and rules.
func (r *FieldRegistry) ValidateFieldAnswer(field models.FormField, val any) error {
	handler, ok := r.Get(field.Type)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownFieldType, field.Type)
	}
	return handler.ValidateAnswer(field, val)
}

// ValidateFormAnswers validates a slice of answers against form field definitions.
func (r *FieldRegistry) ValidateFormAnswers(fields []models.FormField, answers []models.FieldAnswer) error {
	formFieldsMap := make(map[string]models.FormField, len(fields))
	for _, f := range fields {
		formFieldsMap[f.ID] = f
	}

	submittedMap := make(map[string]any, len(answers))
	for _, a := range answers {
		fieldID := strings.TrimSpace(a.FieldID)
		if _, exists := formFieldsMap[fieldID]; !exists {
			return fmt.Errorf("%w: '%s'", ErrUnrecognizedField, fieldID)
		}
		if _, duplicate := submittedMap[fieldID]; duplicate {
			return fmt.Errorf("%w: '%s'", ErrDuplicateFieldAnswer, fieldID)
		}
		submittedMap[fieldID] = a.Value
	}

	for _, field := range fields {
		val, answered := submittedMap[field.ID]
		if !answered {
			val = nil
		}
		if err := r.ValidateFieldAnswer(field, val); err != nil {
			return fmt.Errorf("field '%s': %w", field.ID, err)
		}
	}
	return nil
}
