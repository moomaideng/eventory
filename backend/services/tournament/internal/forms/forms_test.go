package forms

import (
	"testing"

	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func TestShortTextHandler(t *testing.T) {
	h := &ShortTextHandler{}
	assert.Equal(t, models.FieldTypeShortText, h.Type())

	t.Run("config validation", func(t *testing.T) {
		assert.NoError(t, h.ValidateConfig(models.FormField{}))

		field := models.FormField{
			Config: ShortTextConfig{MinLength: intPtr(-1)},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)

		field = models.FormField{
			Config: ShortTextConfig{MinLength: intPtr(10), MaxLength: intPtr(5)},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)

		field = models.FormField{
			Config: ShortTextConfig{MinLength: intPtr(2), MaxLength: intPtr(10)},
		}
		assert.NoError(t, h.ValidateConfig(field))

		// raw map config (from JSON)
		field = models.FormField{
			Config: map[string]any{"minLength": 2, "maxLength": 10},
		}
		assert.NoError(t, h.ValidateConfig(field))
	})

	t.Run("answer validation", func(t *testing.T) {
		field := models.FormField{
			Required: true,
			Config:   ShortTextConfig{MinLength: intPtr(3), MaxLength: intPtr(8)},
		}

		assert.ErrorIs(t, h.ValidateAnswer(field, nil), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, ""), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, 123), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, "ab"), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, "toolongvalue"), ErrInvalidFieldValue)
		assert.NoError(t, h.ValidateAnswer(field, "hello"))

		// Unicode check
		assert.NoError(t, h.ValidateAnswer(field, "สวัสดี")) // 6 runes
	})
}

func TestNumberHandler(t *testing.T) {
	h := &NumberHandler{}
	assert.Equal(t, models.FieldTypeNumber, h.Type())

	t.Run("config validation", func(t *testing.T) {
		field := models.FormField{
			Config: NumberConfig{Min: floatPtr(50), Max: floatPtr(10)},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)

		field = models.FormField{
			Config: NumberConfig{Min: floatPtr(10), Max: floatPtr(50)},
		}
		assert.NoError(t, h.ValidateConfig(field))

		// raw map config (from JSON)
		field = models.FormField{
			Config: map[string]any{"min": 10.0, "max": 50.0},
		}
		assert.NoError(t, h.ValidateConfig(field))
	})

	t.Run("answer validation", func(t *testing.T) {
		field := models.FormField{
			Required: true,
			Config: NumberConfig{
				Min:           floatPtr(18),
				Max:           floatPtr(99),
				AllowDecimals: boolPtr(false),
			},
		}

		assert.ErrorIs(t, h.ValidateAnswer(field, nil), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, "eighteen"), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, 18.5), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, 17), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, 100), ErrInvalidFieldValue)
		assert.NoError(t, h.ValidateAnswer(field, 18))
		assert.NoError(t, h.ValidateAnswer(field, 25.0))
	})
}

func TestSelectOneHandler(t *testing.T) {
	h := &SelectOneHandler{}
	assert.Equal(t, models.FieldTypeSelectOne, h.Type())

	t.Run("config validation", func(t *testing.T) {
		assert.ErrorIs(t, h.ValidateConfig(models.FormField{}), ErrInvalidFieldConfig)

		field := models.FormField{
			Config: SelectOneConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_1", Label: "Option 1"},
					{ID: "opt_1", Label: "Option 2"},
				},
			},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)

		field = models.FormField{
			Config: SelectOneConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_1", Label: "Option 1"},
					{ID: "opt_2", Label: "Option 2"},
				},
			},
		}
		assert.NoError(t, h.ValidateConfig(field))
	})

	t.Run("answer validation", func(t *testing.T) {
		field := models.FormField{
			Required: true,
			Config: SelectOneConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_s", Label: "Small"},
					{ID: "opt_m", Label: "Medium"},
				},
			},
		}

		assert.ErrorIs(t, h.ValidateAnswer(field, nil), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, "opt_l"), ErrInvalidFieldValue)
		assert.NoError(t, h.ValidateAnswer(field, "opt_s"))
		assert.NoError(t, h.ValidateAnswer(field, "opt_m"))
	})
}

func TestSelectMultipleHandler(t *testing.T) {
	h := &SelectMultipleHandler{}
	assert.Equal(t, models.FieldTypeSelectMultiple, h.Type())

	t.Run("config validation", func(t *testing.T) {
		field := models.FormField{
			Config: SelectMultipleConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_1", Label: "Option 1"},
				},
				MinSelections: intPtr(2),
				MaxSelections: intPtr(1),
			},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)

		field = models.FormField{
			Config: SelectMultipleConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_1", Label: "Option 1"},
				},
				MaxSelections: intPtr(5),
			},
		}
		assert.ErrorIs(t, h.ValidateConfig(field), ErrInvalidFieldConfig)
	})

	t.Run("answer validation", func(t *testing.T) {
		field := models.FormField{
			Required: true,
			Config: SelectMultipleConfig{
				Options: []models.FormFieldOption{
					{ID: "opt_1", Label: "Option 1"},
					{ID: "opt_2", Label: "Option 2"},
					{ID: "opt_3", Label: "Option 3"},
				},
				MinSelections: intPtr(1),
				MaxSelections: intPtr(2),
			},
		}

		assert.ErrorIs(t, h.ValidateAnswer(field, nil), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, []any{}), ErrFieldRequired)
		assert.ErrorIs(t, h.ValidateAnswer(field, []any{"opt_4"}), ErrInvalidFieldValue)
		assert.ErrorIs(t, h.ValidateAnswer(field, []any{"opt_1", "opt_1"}), ErrInvalidFieldValue) // duplicate
		assert.ErrorIs(t, h.ValidateAnswer(field, []any{"opt_1", "opt_2", "opt_3"}), ErrInvalidFieldValue) // max 2
		assert.NoError(t, h.ValidateAnswer(field, []any{"opt_1"}))
		assert.NoError(t, h.ValidateAnswer(field, []any{"opt_1", "opt_2"}))
	})
}

func TestFieldRegistry(t *testing.T) {
	r := NewDefaultRegistry()

	_, ok := r.Get(models.FieldTypeShortText)
	assert.True(t, ok)

	err := r.ValidateFieldConfig(models.FormField{Type: "UNKNOWN_TYPE"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownFieldType)
}
