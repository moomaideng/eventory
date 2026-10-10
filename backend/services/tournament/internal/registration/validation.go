package registration

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/moomaideng/eventory/services/tournament/models"
)

var ErrInvalid = errors.New("invalid registration")
var ErrFormChanged = errors.New("registration form changed; reopen the form and try again")

const MaxFileBytes = 5 * 1024 * 1024
const MaxTotalFileBytes = 10 * 1024 * 1024

func invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalid, message) }

func ValidateForm(form *models.RegistrationForm) error {
	if strings.TrimSpace(form.ConsentNotice) == "" || len(form.ConsentNotice) > 2000 {
		return invalid("a consent notice of up to 2000 characters is required")
	}
	if len(form.Questions) > 30 {
		return invalid("at most 30 questions are allowed")
	}
	ids := map[string]bool{}
	for _, q := range form.Questions {
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`).MatchString(q.ID) || ids[q.ID] || strings.TrimSpace(q.Label) == "" {
			return invalid("question IDs must be unique and labels must not be blank")
		}
		ids[q.ID] = true
		switch q.Type {
		case "TEXT":
			if _, err := regexp.Compile(q.Pattern); err != nil {
				return invalid(q.Label + ": invalid text pattern")
			}
		case "DROPDOWN":
			if len(q.Options) == 0 || len(q.Options) > 100 {
				return invalid(q.Label + ": provide dropdown options")
			}
			seen := map[string]bool{}
			for _, option := range q.Options {
				if strings.TrimSpace(option) == "" || seen[option] {
					return invalid(q.Label + ": options must be nonblank and unique")
				}
				seen[option] = true
			}
		case "FILE":
			if q.MaxFileBytes < 1 || q.MaxFileBytes > MaxFileBytes || len(q.AllowedTypes) == 0 {
				return invalid(q.Label + ": choose file types and a size limit up to 5 MB")
			}
			for _, mime := range q.AllowedTypes {
				if !slices.Contains([]string{"application/pdf", "image/png", "image/jpeg"}, mime) {
					return invalid(q.Label + ": unsupported file type")
				}
			}
		default:
			return invalid(q.Label + ": unsupported question type")
		}
	}
	return nil
}

func ValidateSubmission(form *models.RegistrationForm, submission models.RegistrationSubmission) error {
	if submission.FormVersion != form.Version {
		return ErrFormChanged
	}
	if !submission.Consent {
		return invalid("participant-data consent is required")
	}
	answers := map[string]models.RegistrationAnswer{}
	for _, a := range submission.Answers {
		if _, exists := answers[a.QuestionID]; exists {
			return invalid("duplicate question answer")
		}
		answers[a.QuestionID] = a
	}
	total := 0
	for _, q := range form.Questions {
		a := answers[q.ID]
		delete(answers, q.ID)
		if q.Type == "FILE" {
			if a.Value != "" {
				return invalid(q.Label + ": use a file answer")
			}
			if a.File == nil {
				if q.Required {
					return invalid(q.Label + ": a file is required")
				}
				continue
			}
			size := len(a.File.Data)
			total += size
			if strings.TrimSpace(a.File.Name) == "" || strings.ContainsAny(a.File.Name, "/\\\r\n") || len(a.File.Name) > 255 || size == 0 || size > q.MaxFileBytes || total > MaxTotalFileBytes {
				return invalid(q.Label + ": invalid file name or size")
			}
			if !slices.Contains(q.AllowedTypes, http.DetectContentType(a.File.Data)) {
				return invalid(q.Label + ": file content type is not allowed")
			}
			continue
		}
		if a.File != nil {
			return invalid(q.Label + ": file answers are not allowed")
		}
		value := strings.TrimSpace(a.Value)
		if value == "" {
			if q.Required {
				return invalid(q.Label + ": this answer is required")
			}
			continue
		}
		if len([]rune(a.Value)) > 2000 {
			return invalid(q.Label + ": answer is too long")
		}
		if q.Type == "DROPDOWN" && !slices.Contains(q.Options, a.Value) {
			return invalid(q.Label + ": select a configured option")
		}
		if q.Type == "TEXT" && q.Pattern != "" {
			pattern, err := regexp.Compile(q.Pattern)
			if err != nil || !pattern.MatchString(value) {
				return invalid(q.Label + ": answer does not match the required format")
			}
		}
	}
	if len(answers) > 0 {
		return invalid("unknown question answer")
	}
	return nil
}
