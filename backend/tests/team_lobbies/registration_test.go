package team_lobbies

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	handlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestRegistrationJoin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	a := apptest.NewApp(t, pg.DSN)
	client := apptest.NewClient(a.TournamentBaseURL())
	user := func() *apptest.TestUser {
		u := apptest.NewTestUser()
		if err := a.DB.Create(&accountmodels.Account{ID: u.ID, Email: u.Email, Handle: u.Handle, DisplayName: u.DisplayName, Status: accountmodels.AccountStatusActive}).Error; err != nil {
			t.Fatal(err)
		}
		return u
	}
	owner, captain, applicant := user(), user(), user()
	profile := accountmodels.OrganizerProfile{ID: uuid.New(), AccountID: owner.ID, OrganizerName: "Test organizer", OrganizerEmail: owner.Email}
	if err := a.DB.Create(&profile).Error; err != nil {
		t.Fatal(err)
	}
	tournament := models.Tournament{ID: uuid.New(), OrganizerID: profile.ID, Name: "Registration test", Description: "Test", Game: "Test", Location: "Online", StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(80 * time.Hour), RegistrationDeadline: time.Now().Add(48 * time.Hour), Capacity: 16, Currency: "THB", RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 2, MaxTeamSize: 3, Status: models.TournamentStatusRegistrationOpen, Published: true}
	if err := a.DB.Create(&tournament).Error; err != nil {
		t.Fatal(err)
	}
	formPath := fmt.Sprintf("/tournaments/%s/registration-form", tournament.ID)
	form := models.RegistrationForm{Questions: []models.RegistrationQuestion{
		{ID: "game_id", Label: "Game ID", Type: "TEXT", Required: true, Pattern: `^[a-z]{3,20}$`},
		{ID: "rank", Label: "Rank", Type: "DROPDOWN", Required: true, Options: []string{"Gold", "Silver"}},
		{ID: "proof", Label: "Proof", Type: "FILE", Required: true, AllowedTypes: []string{"application/pdf"}, MaxFileBytes: 1024},
	}, ConsentNotice: "Share with organizer for review."}
	request := func(method, path string, body any, who *apptest.TestUser, status int) *apptest.Response {
		t.Helper()
		resp, err := client.Do(method, path, body, who.AuthHeaders())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != status {
			t.Fatalf("%s %s: want %d, got %d: %s", method, path, status, resp.StatusCode, resp.Body)
		}
		return resp
	}
	formBody := map[string]any{"questions": form.Questions, "consentNotice": form.ConsentNotice}
	// Creation includes the form atomically, so a published tournament never
	// briefly accepts participants with an unintended empty questionnaire.
	createTournament := handlers.CreateTournamentRequest{Name: "Tournament with form", Game: "Test", Location: "Online", StartAt: tournament.StartsAt, EndAt: tournament.EndsAt, RegistrationDeadline: tournament.RegistrationDeadline, RegistrationMode: "TEAM", MinTeamSize: 2, MaxTeamSize: 3, Capacity: 16, RegistrationForm: &models.RegistrationFormConfig{Questions: form.Questions, ConsentNotice: form.ConsentNotice}}
	createTournament.RegistrationForm.ConsentNotice = strings.Repeat("ก", 700)
	createdResp := request("POST", "/tournaments", createTournament, owner, 201)
	var created handlers.TournamentResponse
	if err := createdResp.JSON(&created); err != nil {
		t.Fatal(err)
	}
	var createdForm models.RegistrationForm
	if err := a.DB.First(&createdForm, "tournament_id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}
	if createdForm.Version != 1 || len(createdForm.Questions) != 3 || createdForm.ConsentNotice != createTournament.RegistrationForm.ConsentNotice {
		t.Fatal("creation did not save the configured form")
	}
	createTournament.Name = "Invalid form must not publish"
	createTournament.RegistrationForm.Questions = []models.RegistrationQuestion{{ID: "bad", Label: "Bad pattern", Type: "TEXT", Pattern: "["}}
	request("POST", "/tournaments", createTournament, owner, 422)
	var invalidCount int64
	if err := a.DB.Model(&models.Tournament{}).Where("name = ?", createTournament.Name).Count(&invalidCount).Error; err != nil {
		t.Fatal(err)
	}
	if invalidCount != 0 {
		t.Fatal("invalid form published a tournament")
	}
	request("PUT", formPath, formBody, captain, 403)
	request("PUT", formPath, formBody, owner, 200)
	formVersion := 1
	valid := func() models.RegistrationSubmission {
		submission := models.RegistrationSubmission{FormVersion: formVersion, Consent: true, Answers: []models.RegistrationAnswer{
			{QuestionID: "game_id", Value: "player"}, {QuestionID: "rank", Value: "Gold"}, {QuestionID: "proof", File: &models.RegistrationFile{Name: "proof.pdf", Data: []byte("%PDF-1.4\nTest proof")}},
		}}
		if formVersion == 2 {
			submission.Answers = append(submission.Answers, models.RegistrationAnswer{QuestionID: "contact", Value: "contact@example.com"})
		}
		return submission
	}
	createPath := fmt.Sprintf("/tournaments/%s/lobbies", tournament.ID)
	request("POST", createPath, handlers.CreateTeamLobbyRequest{Name: "No consent", Registration: models.RegistrationSubmission{Answers: []models.RegistrationAnswer{}}}, captain, 409)
	resp := request("POST", createPath, handlers.CreateTeamLobbyRequest{Name: "Squad", Registration: valid()}, captain, 201)
	var lobby handlers.TeamLobbyResponse
	if err := resp.JSON(&lobby); err != nil {
		t.Fatal(err)
	}
	joinPath := fmt.Sprintf("/lobbies/%s/join", lobby.InviteCode)
	checks := []struct {
		name   string
		edit   func(*models.RegistrationSubmission)
		status int
	}{
		{"missing consent", func(s *models.RegistrationSubmission) { s.Consent = false }, 422},
		{"missing required", func(s *models.RegistrationSubmission) { s.Answers = s.Answers[1:] }, 422},
		{"invalid pattern", func(s *models.RegistrationSubmission) { s.Answers[0].Value = "!!!" }, 422},
		{"invalid dropdown", func(s *models.RegistrationSubmission) { s.Answers[1].Value = "Diamond" }, 422},
		{"wrong file contents", func(s *models.RegistrationSubmission) { s.Answers[2].File.Data = []byte("not a PDF") }, 422},
		{"oversize file", func(s *models.RegistrationSubmission) {
			s.Answers[2].File.Data = []byte("%PDF-" + strings.Repeat("x", 1024))
		}, 422},
		{"unknown question", func(s *models.RegistrationSubmission) {
			s.Answers = append(s.Answers, models.RegistrationAnswer{QuestionID: "unknown", Value: "x"})
		}, 422},
		{"duplicate question", func(s *models.RegistrationSubmission) { s.Answers = append(s.Answers, s.Answers[0]) }, 422},
		{"stale form", func(s *models.RegistrationSubmission) { s.FormVersion = 0 }, 409},
	}
	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			submission := valid()
			tc.edit(&submission)
			request("POST", joinPath, submission, applicant, tc.status)
			var count int64
			if err := a.DB.Model(&models.TournamentTeamMember{}).Where("account_id = ?", applicant.ID).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatal("invalid submission persisted membership")
			}
		})
	}
	resp = request("POST", joinPath, valid(), applicant, 200)
	if strings.Contains(string(resp.Body), "proof.pdf") || strings.Contains(string(resp.Body), "registrationAnswers") {
		t.Fatal("private answers leaked through roster response")
	}
	var member models.TournamentTeamMember
	if err := a.DB.First(&member, "account_id = ?", applicant.ID).Error; err != nil {
		t.Fatal(err)
	}
	if member.ConsentedAt == nil || member.ConsentNotice != form.ConsentNotice || len(member.RegistrationAnswers) != 3 {
		t.Fatalf("registration did not persist atomically: %+v", member)
	}
	answersPath := fmt.Sprintf("/lobbies/members/%s/registration", member.ID)
	request("GET", answersPath, nil, captain, 403)
	request("GET", answersPath, nil, applicant, 200)
	request("GET", answersPath, nil, owner, 200)
	request("POST", joinPath, valid(), applicant, 409)
	// Existing members stay valid under their original form snapshot.
	updatedQuestions := append([]models.RegistrationQuestion{}, form.Questions...)
	updatedQuestions[0].Label = "Updated game ID"
	updatedQuestions = append(updatedQuestions, models.RegistrationQuestion{ID: "contact", Label: "Contact", Type: "TEXT", Required: true})
	request("PUT", formPath, map[string]any{"questions": updatedQuestions, "consentNotice": "Updated consent notice."}, owner, 200)
	request("POST", joinPath, valid(), user(), 409) // in-progress old form must be refreshed
	formVersion = 2
	missingNewAnswer := valid()
	missingNewAnswer.Answers = missingNewAnswer.Answers[:3]
	request("POST", joinPath, missingNewAnswer, user(), 422)
	if err := a.DB.First(&member, "id = ?", member.ID).Error; err != nil {
		t.Fatal(err)
	}
	if member.FormVersion != 1 || member.ConsentNotice != form.ConsentNotice || len(member.RegistrationAnswers) != 3 || len(member.RegistrationQuestions) != 3 || member.RegistrationQuestions[0].Label != "Game ID" {
		t.Fatal("editing the form changed an existing participant's submission")
	}
	old := request("GET", answersPath, nil, owner, 200)
	if strings.Contains(string(old.Body), "Updated game ID") || !strings.Contains(string(old.Body), "Game ID") {
		t.Fatal("private answers did not return original question labels")
	}

	// Two actual HTTP requests compete for the final slot.
	users := []*apptest.TestUser{user(), user()}
	statuses := make(chan int, 2)
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(1)
		go func(u *apptest.TestUser) {
			defer wg.Done()
			r, e := client.Do(http.MethodPost, joinPath, valid(), u.AuthHeaders())
			if e != nil {
				failures <- e
				return
			}
			statuses <- r.StatusCode
		}(u)
	}
	wg.Wait()
	close(statuses)
	close(failures)
	for e := range failures {
		t.Fatal(e)
	}
	ok, rejected := 0, 0
	for status := range statuses {
		if status == 200 {
			ok++
		} else if status == 409 {
			rejected++
		} else {
			t.Fatalf("unexpected final-slot response %d", status)
		}
	}
	if ok != 1 || rejected != 1 {
		t.Fatalf("final slot: %d accepted, %d rejected", ok, rejected)
	}
	var count int64
	if err := a.DB.Model(&models.TournamentTeamMember{}).Where("tournament_team_id = ?", lobby.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("final roster size = %d, want 3", count)
	}
	// OPEN status alone must not admit a participant after the deadline.
	if err := a.DB.Model(&tournament).Update("registration_deadline", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	for _, attempt := range []struct {
		path string
		body any
	}{
		{createPath, handlers.CreateTeamLobbyRequest{Name: "Too late", Registration: valid()}},
		{joinPath, valid()},
	} {
		lateUser := user()
		resp := request("POST", attempt.path, attempt.body, lateUser, 409)
		if !strings.Contains(string(resp.Body), "registration deadline has passed") {
			t.Fatalf("deadline failure needs an actionable message: %s", resp.Body)
		}
		var saved int64
		if err := a.DB.Model(&models.TournamentTeamMember{}).Where("account_id = ?", lateUser.ID).Count(&saved).Error; err != nil {
			t.Fatal(err)
		}
		if saved != 0 {
			t.Fatal("expired registration created a membership")
		}
	}
	request("POST", fmt.Sprintf("/lobbies/%s/invite/regenerate", lobby.ID), nil, captain, 200)
	request("POST", joinPath, valid(), user(), 404)
}
