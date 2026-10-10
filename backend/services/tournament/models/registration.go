package models

import (
	"time"

	"github.com/google/uuid"
)

// RegistrationQuestion is configured by the organizer, never by an applicant.
type RegistrationQuestion struct {
	ID           string   `json:"id" maxLength:"64"`
	Label        string   `json:"label" maxLength:"160"`
	Type         string   `json:"type" enum:"TEXT,DROPDOWN,FILE"`
	Required     bool     `json:"required"`
	Options      []string `json:"options"`
	Pattern      string   `json:"pattern" maxLength:"256"`
	HelpText     string   `json:"helpText" maxLength:"500"`
	AllowedTypes []string `json:"allowedTypes"`
	MaxFileBytes int      `json:"maxFileBytes"`
}

type RegistrationForm struct {
	TournamentID  uuid.UUID              `gorm:"type:uuid;primaryKey" json:"-"`
	Version       int                    `json:"version"`
	Questions     []RegistrationQuestion `gorm:"serializer:json;type:jsonb;not null" json:"questions"`
	ConsentNotice string                 `gorm:"type:text;not null" json:"consentNotice"`
	UpdatedAt     time.Time              `json:"-"`
}

type RegistrationFormConfig struct {
	Questions     []RegistrationQuestion `json:"questions" maxItems:"30"`
	ConsentNotice string                 `json:"consentNotice" minLength:"1" maxLength:"2000"`
}

// Files travel with the submission and are persisted privately, with membership.
// No public file URL or unclaimed upload is created.
type RegistrationFile struct {
	Name string `json:"name" maxLength:"255"`
	Data []byte `json:"data" doc:"Base64-encoded PDF, JPEG or PNG file"`
}

type RegistrationAnswer struct {
	QuestionID string            `json:"questionId" maxLength:"64"`
	Value      string            `json:"value" maxLength:"2000"`
	File       *RegistrationFile `json:"file,omitempty"`
}

type RegistrationSubmission struct {
	FormVersion int                  `json:"formVersion" minimum:"0"`
	Answers     []RegistrationAnswer `json:"answers" maxItems:"30"`
	Consent     bool                 `json:"consent"`
}

const DefaultConsentNotice = "I agree to share my registration answers and uploaded files with the tournament organizer for participant review and tournament administration."

func EmptyRegistrationForm(id uuid.UUID) *RegistrationForm {
	return &RegistrationForm{TournamentID: id, Questions: []RegistrationQuestion{}, ConsentNotice: DefaultConsentNotice}
}
