package forms

import (
	"fmt"

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
