package registration

import (
	"strings"
	"testing"

	"github.com/moomaideng/eventory/services/tournament/models"
)

func TestValidateFormConsentCharacterLimit(t *testing.T) {
	for _, count := range []int{700, 2000, 2001} {
		form := &models.RegistrationForm{ConsentNotice: strings.Repeat("ก", count)}
		err := ValidateForm(form)
		if (err != nil) != (count > 2000) {
			t.Fatalf("%d Thai characters: error = %v", count, err)
		}
	}
}

func TestValidateFormDropdownOptionsCanBeSubmitted(t *testing.T) {
	for _, count := range []int{2000, 2001} {
		option := strings.Repeat("ก", count)
		form := &models.RegistrationForm{Version: 1, ConsentNotice: "Consent", Questions: []models.RegistrationQuestion{
			{ID: "rank", Label: "Rank", Type: "DROPDOWN", Required: true, Options: []string{option, "Other"}},
		}}
		err := ValidateForm(form)
		if count > 2000 {
			if err == nil {
				t.Fatal("accepted an option longer than the answer limit")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateSubmission(form, models.RegistrationSubmission{FormVersion: 1, Consent: true, Answers: []models.RegistrationAnswer{{QuestionID: "rank", Value: option}}}); err != nil {
			t.Fatalf("accepted option cannot be submitted: %v", err)
		}
	}
}

func TestValidateFormRequiresTwoDistinctDropdownOptions(t *testing.T) {
	for _, options := range [][]string{nil, {"Gold"}, {"Gold", "Gold"}, {"Gold", ""}} {
		form := &models.RegistrationForm{ConsentNotice: "Consent", Questions: []models.RegistrationQuestion{{ID: "rank", Label: "Rank", Type: "DROPDOWN", Options: options}}}
		if err := ValidateForm(form); err == nil {
			t.Fatalf("accepted incomplete dropdown options: %v", options)
		}
	}
}
