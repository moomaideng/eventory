package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FieldType string

const (
	FieldTypeShortText      FieldType = "SHORT_TEXT"
	FieldTypeLongText       FieldType = "LONG_TEXT"
	FieldTypeNumber         FieldType = "NUMBER"
	FieldTypeSelectOne      FieldType = "SELECT_ONE"
	FieldTypeSelectMultiple FieldType = "SELECT_MULTIPLE"
)

type FormFieldOption struct {
	ID    string `bson:"id" json:"id"`
	Label string `bson:"label" json:"label"`
}

type FormFieldConfig struct {
	MinLength     *int              `bson:"min_length,omitempty" json:"minLength,omitempty"`
	MaxLength     *int              `bson:"max_length,omitempty" json:"maxLength,omitempty"`
	Min           *float64          `bson:"min,omitempty" json:"min,omitempty"`
	Max           *float64          `bson:"max,omitempty" json:"max,omitempty"`
	AllowDecimals *bool             `bson:"allow_decimals,omitempty" json:"allowDecimals,omitempty"`
	Options       []FormFieldOption `bson:"options,omitempty" json:"options,omitempty"`
	MinSelections *int              `bson:"min_selections,omitempty" json:"minSelections,omitempty"`
	MaxSelections *int              `bson:"max_selections,omitempty" json:"maxSelections,omitempty"`
}

type FormField struct {
	ID          string          `bson:"id" json:"id"`
	Type        FieldType       `bson:"type" json:"type" enum:"SHORT_TEXT,LONG_TEXT,NUMBER,SELECT_ONE,SELECT_MULTIPLE"`
	Label       string          `bson:"label" json:"label"`
	Description string          `bson:"description,omitempty" json:"description,omitempty"`
	Required    bool            `bson:"required" json:"required"`
	Order       int             `bson:"order" json:"order"`
	Config      FormFieldConfig `bson:"config,omitempty" json:"config,omitempty"`
}

type RegistrationForm struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	TournamentID uuid.UUID     `bson:"tournament_id" json:"tournamentId"`
	Title        string        `bson:"title" json:"title"`
	Description  string        `bson:"description,omitempty" json:"description,omitempty"`
	Fields       []FormField   `bson:"fields" json:"fields"`
	CreatedAt    time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updatedAt"`
}

type FieldAnswer struct {
	FieldID string `bson:"field_id" json:"fieldId"`
	Value   any    `bson:"value" json:"value"`
}

type FormResponse struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	TournamentID uuid.UUID     `bson:"tournament_id" json:"tournamentId"`
	AccountID    uuid.UUID     `bson:"account_id" json:"accountId"`
	Answers      []FieldAnswer `bson:"answers" json:"answers"`
}
