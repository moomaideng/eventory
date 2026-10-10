package tournament_status

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

type statusOverrideScenario struct {
	client     *apptest.Client
	organizer  *apptest.TestUser
	other      *apptest.TestUser
	tournament *models.Tournament
	resp       *apptest.Response
	lastReason string
	accountIDs []uuid.UUID
}

// persistActiveAccount inserts an ACTIVE account so the auth middleware can
// resolve the signed test JWT to a real account.
func (s *statusOverrideScenario) persistActiveAccount(user *apptest.TestUser) error {
	account := models.Account{
		ID:          user.ID,
		Email:       user.Email,
		Handle:      user.Handle,
		DisplayName: user.DisplayName,
		Status:      models.AccountStatusActive,
	}
	if err := app.DB.Create(&account).Error; err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	s.accountIDs = append(s.accountIDs, user.ID)
	return nil
}

func (s *statusOverrideScenario) newOrganizer() (*apptest.TestUser, models.OrganizerProfile, error) {
	user := apptest.NewTestUser()
	profile := models.OrganizerProfile{}
	if err := s.persistActiveAccount(user); err != nil {
		return nil, profile, err
	}
	profile = models.OrganizerProfile{
		ID:             uuid.New(),
		AccountID:      user.ID,
		OrganizerName:  "BDD Status Organizer " + apptest.UniqueSuffix(),
		OrganizerEmail: user.Email,
	}
	if err := app.DB.Create(&profile).Error; err != nil {
		return nil, profile, fmt.Errorf("create organizer profile: %w", err)
	}
	return user, profile, nil
}

func (s *statusOverrideScenario) anOrganizerOwnsATournament(status string) error {
	user, profile, err := s.newOrganizer()
	if err != nil {
		return err
	}
	s.organizer = user

	now := time.Now().UTC()
	s.tournament = &models.Tournament{
		ID:                   uuid.New(),
		OrganizerID:          profile.ID,
		Name:                 "BDD Status Tournament " + apptest.UniqueSuffix(),
		Description:          "Tournament created for status override acceptance tests.",
		Game:                 "Valorant",
		Location:             "Online",
		StartsAt:             now.Add(72 * time.Hour),
		EndsAt:               now.Add(80 * time.Hour),
		RegistrationDeadline: now.Add(48 * time.Hour),
		Currency:             "THB",
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          2,
		MaxTeamSize:          3,
		Capacity:             16,
		Status:               models.TournamentStatus(status),
		// Drafts are the only unpublished state; everything else is public.
		Published: status != string(models.TournamentStatusDraft),
	}
	if err := app.DB.Create(s.tournament).Error; err != nil {
		return fmt.Errorf("create tournament: %w", err)
	}
	return nil
}

func (s *statusOverrideScenario) anotherOrganizerExists() error {
	user, _, err := s.newOrganizer()
	if err != nil {
		return err
	}
	s.other = user
	return nil
}

func (s *statusOverrideScenario) override(headers map[string]string, body map[string]any) error {
	var err error
	s.resp, err = s.client.Do(
		http.MethodPatch,
		"/tournaments/"+s.tournament.ID.String()+"/status",
		body,
		headers,
	)
	return err
}

func (s *statusOverrideScenario) organizerOverrides(target, reason string) error {
	s.lastReason = reason
	return s.override(s.organizer.AuthHeaders(), map[string]any{"status": target, "reason": reason})
}

// overrideWithReasonLength exercises the 10-500 character rule, including the
// empty reason, which the schema rejects as a missing required field.
func (s *statusOverrideScenario) overrideWithReasonLength(target string, length int) error {
	body := map[string]any{"status": target}
	if length > 0 {
		body["reason"] = strings.Repeat("a", length)
	}
	return s.override(s.organizer.AuthHeaders(), body)
}

func (s *statusOverrideScenario) anonymousOverrides(target, reason string) error {
	return s.override(nil, map[string]any{"status": target, "reason": reason})
}

func (s *statusOverrideScenario) otherOrganizerOverrides(target, reason string) error {
	return s.override(s.other.AuthHeaders(), map[string]any{"status": target, "reason": reason})
}

func (s *statusOverrideScenario) readHistory(user *apptest.TestUser) error {
	var err error
	s.resp, err = s.client.Do(
		http.MethodGet,
		"/tournaments/"+s.tournament.ID.String()+"/status/history",
		nil,
		user.AuthHeaders(),
	)
	return err
}

func (s *statusOverrideScenario) responseStatus(expected int) error {
	if s.resp == nil {
		return fmt.Errorf("no response captured")
	}
	if s.resp.StatusCode != expected {
		return fmt.Errorf("expected HTTP %d, got %d: %s", expected, s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *statusOverrideScenario) errorDetail(expected string) error {
	var body struct {
		Detail string `json:"detail"`
	}
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	if body.Detail != expected {
		return fmt.Errorf("expected error %q, got %q", expected, body.Detail)
	}
	return nil
}

func (s *statusOverrideScenario) tournamentStatusShouldBe(expected string) error {
	var stored models.Tournament
	if err := app.DB.First(&stored, "id = ?", s.tournament.ID).Error; err != nil {
		return err
	}
	if string(stored.Status) != expected {
		return fmt.Errorf("expected stored status %q, got %q", expected, stored.Status)
	}
	return nil
}

func (s *statusOverrideScenario) auditTrailShouldRecord(from, to string) error {
	var changes []models.TournamentStatusChange
	if err := app.DB.Where("tournament_id = ?", s.tournament.ID).Find(&changes).Error; err != nil {
		return err
	}
	if len(changes) != 1 {
		return fmt.Errorf("expected exactly one recorded change, got %d", len(changes))
	}
	change := changes[0]
	if string(change.FromStatus) != from || string(change.ToStatus) != to {
		return fmt.Errorf("expected %s -> %s, recorded %s -> %s", from, to, change.FromStatus, change.ToStatus)
	}
	if change.ActorAccountID != s.organizer.ID {
		return fmt.Errorf("expected actor %s, recorded %s", s.organizer.ID, change.ActorAccountID)
	}
	if change.Reason != s.lastReason {
		return fmt.Errorf("expected reason %q, recorded %q", s.lastReason, change.Reason)
	}
	if change.CreatedAt.IsZero() {
		return fmt.Errorf("expected the change to be timestamped")
	}
	return nil
}

func (s *statusOverrideScenario) noStatusChangeRecorded() error {
	var count int64
	if err := app.DB.Model(&models.TournamentStatusChange{}).
		Where("tournament_id = ?", s.tournament.ID).Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("expected no recorded change, found %d", count)
	}
	return nil
}

func (s *statusOverrideScenario) history() (handlers.TournamentStatusHistoryOutput, error) {
	var output handlers.TournamentStatusHistoryOutput
	return output, s.resp.JSON(&output.Body)
}

func (s *statusOverrideScenario) currentStatusShouldBe(expected string) error {
	output, err := s.history()
	if err != nil {
		return err
	}
	if output.Body.CurrentStatus != expected {
		return fmt.Errorf("expected current status %q, got %q", expected, output.Body.CurrentStatus)
	}
	return nil
}

func (s *statusOverrideScenario) allowedTransitionsShouldBe(expected string) error {
	output, err := s.history()
	if err != nil {
		return err
	}
	actual := strings.Join(output.Body.AllowedTransitions, ", ")
	if actual != strings.TrimSpace(expected) {
		return fmt.Errorf("expected allowed transitions %q, got %q", expected, actual)
	}
	return nil
}

func (s *statusOverrideScenario) historyShouldList(count int) error {
	output, err := s.history()
	if err != nil {
		return err
	}
	if len(output.Body.Items) != count {
		return fmt.Errorf("expected %d changes, got %d", count, len(output.Body.Items))
	}
	for i := 1; i < len(output.Body.Items); i++ {
		if output.Body.Items[i-1].CreatedAt.Before(output.Body.Items[i].CreatedAt) {
			return fmt.Errorf("expected newest change first, got %v then %v",
				output.Body.Items[i-1].CreatedAt, output.Body.Items[i].CreatedAt)
		}
	}
	return nil
}

// The audit trail loads the acting account, so guard against its private
// contact details reaching the response.
func (s *statusOverrideScenario) historyShouldNotExposeEmail() error {
	if strings.Contains(string(s.resp.Body), s.organizer.Email) {
		return fmt.Errorf("history exposed the organizer email: %s", s.resp.Body)
	}
	return nil
}

func (s *statusOverrideScenario) cleanup() error {
	if s.tournament != nil {
		if err := app.DB.Where("id = ?", s.tournament.ID).Delete(&models.Tournament{}).Error; err != nil {
			return err
		}
	}
	if len(s.accountIDs) > 0 {
		return app.DB.Where("id IN ?", s.accountIDs).Delete(&models.Account{}).Error
	}
	return nil
}

// InitializeScenario registers the status override steps with godog.
func InitializeScenario(sc *godog.ScenarioContext) {
	var s *statusOverrideScenario
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		s = &statusOverrideScenario{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return ctx, s.cleanup()
	})

	sc.Step(`^an organizer owns a tournament in "([^"]*)" status$`, func(status string) error {
		return s.anOrganizerOwnsATournament(status)
	})
	sc.Step(`^another organizer exists$`, func() error { return s.anotherOrganizerExists() })

	sc.Step(`^the organizer overrides the status to "([^"]*)" with reason "([^"]*)"$`, func(target, reason string) error {
		return s.organizerOverrides(target, reason)
	})
	sc.Step(`^the organizer overrides the status to "([^"]*)" with a reason of (\d+) characters$`, func(target string, length int) error {
		return s.overrideWithReasonLength(target, length)
	})
	sc.Step(`^an anonymous visitor overrides the status to "([^"]*)" with reason "([^"]*)"$`, func(target, reason string) error {
		return s.anonymousOverrides(target, reason)
	})
	sc.Step(`^the other organizer overrides the status to "([^"]*)" with reason "([^"]*)"$`, func(target, reason string) error {
		return s.otherOrganizerOverrides(target, reason)
	})
	sc.Step(`^the organizer reads the status history$`, func() error { return s.readHistory(s.organizer) })
	sc.Step(`^the other organizer reads the status history$`, func() error { return s.readHistory(s.other) })

	sc.Step(`^the response status should be (\d+)$`, func(code int) error { return s.responseStatus(code) })
	sc.Step(`^the error detail should be "([^"]*)"$`, func(detail string) error { return s.errorDetail(detail) })
	sc.Step(`^the tournament status should be "([^"]*)"$`, func(status string) error {
		return s.tournamentStatusShouldBe(status)
	})
	sc.Step(`^the audit trail should record "([^"]*)" to "([^"]*)" with the acting organizer and a timestamp$`, func(from, to string) error {
		return s.auditTrailShouldRecord(from, to)
	})
	sc.Step(`^no status change should be recorded$`, func() error { return s.noStatusChangeRecorded() })
	sc.Step(`^the current status should be "([^"]*)"$`, func(status string) error {
		return s.currentStatusShouldBe(status)
	})
	sc.Step(`^the allowed transitions should be "([^"]*)"$`, func(expected string) error {
		return s.allowedTransitionsShouldBe(expected)
	})
	sc.Step(`^the history should list (\d+) changes, newest first$`, func(count int) error {
		return s.historyShouldList(count)
	})
	sc.Step(`^the history should not expose the organizer's email$`, func() error {
		return s.historyShouldNotExposeEmail()
	})
}
