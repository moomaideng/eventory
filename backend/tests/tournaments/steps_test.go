package tournaments

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

type discoveryScenario struct {
	app           *apptest.App
	client        *apptest.Client
	resp          *apptest.Response
	tournaments   map[string]models.Tournament
	accountIDs    []uuid.UUID
	acceptedID    uuid.UUID
	privateValues []string
}

func (s *discoveryScenario) account() (models.Account, error) {
	id := uuid.New()
	a := models.Account{ID: id, Email: id.String() + "@test.eventory.gg", Handle: "bdd_" + apptest.UniqueSuffix(), DisplayName: "Private Player", Status: models.AccountStatusActive}
	if err := s.app.DB.Create(&a).Error; err != nil {
		return a, err
	}
	s.accountIDs = append(s.accountIDs, id)
	s.privateValues = append(s.privateValues, a.Email, a.Handle)
	return a, nil
}

func (s *discoveryScenario) fixtures() error {
	a, err := s.account()
	if err != nil {
		return err
	}
	org := models.OrganizerProfile{ID: uuid.New(), AccountID: a.ID, OrganizerName: "Discovery Host", OrganizerEmail: "private-host@test.eventory.gg"}
	if err := s.app.DB.Create(&org).Error; err != nil {
		return err
	}
	s.privateValues = append(s.privateValues, org.OrganizerEmail)
	for _, row := range []struct {
		name, starts string
		fee          int64
		published    bool
	}{
		{"Free Cup", "2026-09-30T23:59:59+07:00", 0, true},
		{"Morning Cup", "2026-10-01T00:00:00+07:00", 100, true},
		{"Late Cup", "2026-10-01T23:59:59.999999+07:00", 500, true},
		{"Next Day Cup", "2026-10-02T00:00:00+07:00", 250, true},
		{"Premium Cup", "2026-10-02T12:00:00+07:00", 1000, true},
		{"Secret Draft", "2026-10-01T12:00:00+07:00", 200, false},
	} {
		starts, err := time.Parse(time.RFC3339Nano, row.starts)
		if err != nil {
			return err
		}
		tournament := models.Tournament{ID: uuid.New(), OrganizerID: org.ID, Name: row.name,
			Description: "Rules: best of three; no cheating.", Game: "Valorant", Location: "Online",
			StartsAt: starts, EndsAt: starts.Add(2 * time.Hour), RegistrationDeadline: starts.Add(-24 * time.Hour),
			EntryFee: row.fee, Currency: "THB", Capacity: 16, RegistrationMode: models.TournamentRegistrationModeTeam,
			MinTeamSize: 2, MaxTeamSize: 5, Status: models.TournamentStatusRegistrationOpen, Published: row.published}
		if !row.published {
			tournament.Status = models.TournamentStatusDraft
		}
		if err := s.app.DB.Create(&tournament).Error; err != nil {
			return err
		}
		tournament.Organizer = org
		s.tournaments[row.name] = tournament
	}
	return nil
}

// Delete only this scenario's fixtures, including when a step fails.
func (s *discoveryScenario) cleanup() error {
	ids := make([]uuid.UUID, 0, len(s.tournaments))
	for _, tournament := range s.tournaments {
		ids = append(ids, tournament.ID)
	}
	if len(ids) > 0 {
		if err := s.app.DB.Where("id IN ?", ids).Delete(&models.Tournament{}).Error; err != nil {
			return err
		}
	}
	if len(s.accountIDs) > 0 {
		return s.app.DB.Where("id IN ?", s.accountIDs).Delete(&models.Account{}).Error
	}
	return nil
}

func (s *discoveryScenario) search(query string) error {
	var err error
	s.resp, err = s.client.Do(http.MethodGet, "/tournaments?"+query, nil, nil)
	return err
}

func (s *discoveryScenario) details(name string) error {
	id := uuid.New()
	if tournament, exists := s.tournaments[name]; exists {
		id = tournament.ID
	}
	var err error
	s.resp, err = s.client.Do(http.MethodGet, "/tournaments/"+id.String(), nil, nil)
	return err
}

func (s *discoveryScenario) status(expected int) error {
	if s.resp == nil {
		return fmt.Errorf("no response captured")
	}
	if s.resp.StatusCode != expected {
		return fmt.Errorf("expected HTTP %d, got %d: %s", expected, s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *discoveryScenario) catalog(names string, total int) error {
	var body handlers.TournamentListBody
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	actual := make([]string, 0, len(body.Items))
	for _, item := range body.Items {
		actual = append(actual, item.Name)
	}
	if strings.Join(actual, ", ") != names || body.Total != int64(total) {
		return fmt.Errorf("expected catalog %q / total %d, got %v / total %d", names, total, actual, body.Total)
	}
	return nil
}

func (s *discoveryScenario) teams(name string) error {
	tournament := s.tournaments[name]
	for _, status := range []models.TournamentTeamStatus{models.TournamentTeamStatusAccepted, models.TournamentTeamStatusForming, models.TournamentTeamStatusLocked, models.TournamentTeamStatusRejected} {
		team := models.TournamentTeam{ID: uuid.New(), TournamentID: tournament.ID, Name: string(status) + " Team", Status: status, InviteCode: uuid.NewString()[:32]}
		if err := s.app.DB.Create(&team).Error; err != nil {
			return err
		}
		s.privateValues = append(s.privateValues, team.InviteCode)
		if status == models.TournamentTeamStatusAccepted {
			s.acceptedID = team.ID
		}
		for _, role := range []models.TournamentTeamMemberRole{models.TournamentTeamMemberRoleCaptain, models.TournamentTeamMemberRoleMember} {
			a, err := s.account()
			if err != nil {
				return err
			}
			member := models.TournamentTeamMember{ID: uuid.New(), TournamentTeamID: team.ID, AccountID: a.ID, Role: role}
			if err := s.app.DB.Create(&member).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *discoveryScenario) funding(name string, goal, raised int) error {
	funding := models.TournamentFunding{ID: uuid.New(), TournamentID: s.tournaments[name].ID, GoalAmount: int64(goal), RaisedAmount: int64(raised), SupporterCount: 4}
	return s.app.DB.Create(&funding).Error
}

func (s *discoveryScenario) publicDetails(name string) error {
	var body handlers.TournamentDetailsBody
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	want := s.tournaments[name]
	got := body.Tournament
	if got.ID != want.ID || got.Name != want.Name || got.Description != want.Description || got.Game != want.Game || got.Location != want.Location || got.OrganizerName != want.Organizer.OrganizerName ||
		!got.StartAt.Equal(want.StartsAt) || !got.EndAt.Equal(want.EndsAt) || !got.RegistrationDeadline.Equal(want.RegistrationDeadline) ||
		got.EntryFee != want.EntryFee || got.Currency != want.Currency || got.Capacity != want.Capacity || got.Status != string(want.Status) ||
		got.RegistrationMode != string(want.RegistrationMode) || got.MinTeamSize != want.MinTeamSize || got.MaxTeamSize != want.MaxTeamSize {
		return fmt.Errorf("public details do not match stored tournament: %+v", got)
	}
	return nil
}

func (s *discoveryScenario) acceptedTeam() error {
	var body handlers.TournamentDetailsBody
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	if body.Tournament.RegisteredCount != 1 || len(body.Teams) != 1 || body.Teams[0].ID != s.acceptedID || body.Teams[0].Name != "ACCEPTED Team" || body.Teams[0].MemberCount != 2 {
		return fmt.Errorf("expected one accepted team with two members, got count %d, teams %+v", body.Tournament.RegisteredCount, body.Teams)
	}
	return nil
}

func (s *discoveryScenario) fundingStats(goal, raised, remaining int, percentage float64) error {
	var body handlers.TournamentDetailsBody
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	f := body.Funding
	if f.GoalAmount != int64(goal) || f.RaisedAmount != int64(raised) || f.RemainingAmount != int64(remaining) || math.Abs(f.Percentage-percentage) > 0.0001 || f.SupporterCount != 4 || f.Currency != "THB" {
		return fmt.Errorf("unexpected funding statistics: %+v", f)
	}
	return nil
}

func (s *discoveryScenario) privateData() error {
	for _, value := range append(s.privateValues, `"inviteCode"`, `"accountId"`, `"email"`, `"members"`, `"organizerEmail"`) {
		if strings.Contains(string(s.resp.Body), value) {
			return fmt.Errorf("private data %q exposed: %s", value, s.resp.Body)
		}
	}
	return nil
}

func (s *discoveryScenario) noData() error {
	var body map[string]json.RawMessage
	if err := s.resp.JSON(&body); err != nil {
		return err
	}
	for _, key := range []string{"items", "tournament", "teams", "funding", "description", "organizerEmail"} {
		if _, found := body[key]; found {
			return fmt.Errorf("error response exposed %s", key)
		}
	}
	for name := range s.tournaments {
		if strings.Contains(string(s.resp.Body), name) {
			return fmt.Errorf("error response exposed tournament %q", name)
		}
	}
	return s.privateData()
}

func initializeScenario(sc *godog.ScenarioContext, app *apptest.App) {
	var s *discoveryScenario
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		s = &discoveryScenario{app: app, client: apptest.NewClient(app.BaseURL()), tournaments: make(map[string]models.Tournament)}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return ctx, s.cleanup()
	})
	sc.Step(`^tournaments with different schedules, fees and publication states exist$`, func() error { return s.fixtures() })
	sc.Step(`^a visitor searches tournaments with "([^"]*)"$`, func(query string) error { return s.search(query) })
	sc.Step(`^the response status should be (\d+)$`, func(code int) error { return s.status(code) })
	sc.Step(`^the catalog should contain exactly "([^"]*)" with total (\d+)$`, func(names string, total int) error { return s.catalog(names, total) })
	sc.Step(`^"([^"]*)" has teams in every registration status$`, func(name string) error { return s.teams(name) })
	sc.Step(`^"([^"]*)" has a funding goal of (\d+) with (\d+) raised and 4 supporters$`, func(name string, goal, raised int) error { return s.funding(name, goal, raised) })
	sc.Step(`^a visitor opens the details for "([^"]*)"$`, func(name string) error { return s.details(name) })
	sc.Step(`^the public details should match "([^"]*)"$`, func(name string) error { return s.publicDetails(name) })
	sc.Step(`^only the accepted team and its 2 members should be counted$`, func() error { return s.acceptedTeam() })
	sc.Step(`^funding should show goal (\d+), raised (\d+), remaining (\d+) and percentage ([\d.]+)$`, func(goal, raised, remaining int, percentage float64) error {
		return s.fundingStats(goal, raised, remaining, percentage)
	})
	sc.Step(`^team invite codes and member account data should be private$`, func() error { return s.privateData() })
	sc.Step(`^no tournament data should be exposed$`, func() error { return s.noData() })
	sc.Step(`^the error detail should be "([^"]*)"$`, func(expected string) error {
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
	})
	sc.Step(`^the catalog registration count should be (\d+)$`, func(expected int) error {
		var body handlers.TournamentListBody
		if err := s.resp.JSON(&body); err != nil {
			return err
		}
		if len(body.Items) != 1 || body.Items[0].RegisteredCount != expected {
			return fmt.Errorf("unexpected catalog counts: %+v", body.Items)
		}
		return nil
	})
	sc.Step(`^the details should contain empty teams and zero funding$`, func() error {
		var body handlers.TournamentDetailsBody
		if err := s.resp.JSON(&body); err != nil {
			return err
		}
		if body.Teams == nil || len(body.Teams) != 0 || body.Tournament.RegisteredCount != 0 || !reflect.DeepEqual(body.Funding, handlers.TournamentFundingResponse{Currency: "THB"}) {
			return fmt.Errorf("expected empty teams and zero funding, got %+v", body)
		}
		return nil
	})
}
