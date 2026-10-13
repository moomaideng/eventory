package application_review

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

type reviewScenarioContext struct {
	client         *apptest.Client
	reviewer       *apptest.TestUser // who sends the review request
	owner          *apptest.TestUser // tournament owner, used to read the dashboard
	tournament     *models.Tournament
	team           *models.TournamentTeam // the application under review
	statusBefore   models.TournamentTeamStatus
	acceptedBefore int
	resp           *apptest.Response
}

func (s *reviewScenarioContext) theEventoryAPIServiceIsRunning() error {
	if app == nil || app.TournamentServer == nil {
		return fmt.Errorf("application server is not running")
	}
	return nil
}

func (s *reviewScenarioContext) persistActiveAccount(user *apptest.TestUser) error {
	account := accountmodels.Account{
		ID:          user.ID,
		Email:       user.Email,
		Handle:      user.Handle,
		DisplayName: user.DisplayName,
		Status:      accountmodels.AccountStatusActive,
	}
	if err := app.DB.Create(&account).Error; err != nil {
		return fmt.Errorf("create active account %s: %w", user.Email, err)
	}
	return nil
}

// createTournament seeds a published team tournament with registration open.
// Its owner becomes the default reviewer.
func (s *reviewScenarioContext) createTournament(capacity int) error {
	owner := apptest.NewTestUser()
	if err := s.persistActiveAccount(owner); err != nil {
		return err
	}
	profile := accountmodels.OrganizerProfile{
		AccountID:      owner.ID,
		OrganizerName:  "BDD Review Organizer",
		OrganizerEmail: owner.Email,
	}
	if err := app.DB.Create(&profile).Error; err != nil {
		return fmt.Errorf("create organizer profile: %w", err)
	}

	now := time.Now().UTC()
	s.tournament = &models.Tournament{
		ID: uuid.New(),
		Staff: []models.TournamentStaff{
			{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner},
		},
		Name:                 "BDD Review Tournament " + uuid.NewString()[:8],
		Description:          "Tournament created for application review acceptance tests.",
		Game:                 "Valorant",
		Location:             "Online",
		StartsAt:             now.Add(72 * time.Hour),
		EndsAt:               now.Add(80 * time.Hour),
		RegistrationDeadline: now.Add(48 * time.Hour),
		Currency:             "THB",
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          1,
		MaxTeamSize:          3,
		Capacity:             capacity,
		Status:               models.TournamentStatusRegistrationOpen,
		Published:            true,
	}
	if err := app.DB.Create(s.tournament).Error; err != nil {
		return fmt.Errorf("create tournament: %w", err)
	}
	s.owner = owner
	s.reviewer = owner
	return nil
}

// seedTeam writes a team with the given status straight into the database, so
// these scenarios test the review decision rather than the lobby/payment flow.
func (s *reviewScenarioContext) seedTeam(status models.TournamentTeamStatus) (*models.TournamentTeam, error) {
	captain := apptest.NewTestUser()
	if err := s.persistActiveAccount(captain); err != nil {
		return nil, err
	}
	team := models.TournamentTeam{
		ID:           uuid.New(),
		TournamentID: s.tournament.ID,
		Name:         "Team " + uuid.NewString()[:8],
		InviteCode:   uuid.NewString()[:8],
		Status:       status,
	}
	if status != models.TournamentTeamStatusForming {
		lockedAt := time.Now().UTC()
		team.LockedAt = &lockedAt
	}
	if err := app.DB.Create(&team).Error; err != nil {
		return nil, fmt.Errorf("create %s team: %w", status, err)
	}
	member := models.TournamentTeamMember{
		TournamentTeamID: team.ID,
		AccountID:        captain.ID,
		Role:             models.TournamentTeamMemberRoleCaptain,
	}
	if err := app.DB.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("create team captain: %w", err)
	}
	return &team, nil
}

// ---- Given steps ----

func (s *reviewScenarioContext) anOrganizerManagesATournamentThatStillHasOpenSpots() error {
	return s.createTournament(8)
}

func (s *reviewScenarioContext) aTeamHasLockedItsApplicationForThatTournament() error {
	team, err := s.seedTeam(models.TournamentTeamStatusLocked)
	s.team = team
	return err
}

func (s *reviewScenarioContext) aTeamApplicationHasAlreadyBeenAccepted() error {
	team, err := s.seedTeam(models.TournamentTeamStatusAccepted)
	s.team = team
	return err
}

func (s *reviewScenarioContext) anOrganizerReviewsAnApplicationWhere(situation string) error {
	switch situation {
	case "the tournament has already reached its capacity":
		if err := s.createTournament(1); err != nil {
			return err
		}
		if _, err := s.seedTeam(models.TournamentTeamStatusAccepted); err != nil {
			return err
		}
		return s.aTeamHasLockedItsApplicationForThatTournament()
	case "the team has not locked its application yet":
		if err := s.createTournament(8); err != nil {
			return err
		}
		team, err := s.seedTeam(models.TournamentTeamStatusForming)
		s.team = team
		return err
	case "the application has already been accepted":
		if err := s.createTournament(8); err != nil {
			return err
		}
		return s.aTeamApplicationHasAlreadyBeenAccepted()
	case "the tournament is managed by another organizer":
		if err := s.createTournament(8); err != nil {
			return err
		}
		if err := s.aTeamHasLockedItsApplicationForThatTournament(); err != nil {
			return err
		}
		stranger := apptest.NewTestUser()
		if err := s.persistActiveAccount(stranger); err != nil {
			return err
		}
		s.reviewer = stranger
		return nil
	}
	return fmt.Errorf("unknown review situation %q", situation)
}

// ---- When steps ----

// review sends the decision and snapshots the state beforehand so the Then
// steps can compare against it.
func (s *reviewScenarioContext) review(decision, reason string) error {
	var err error
	if s.statusBefore, err = s.teamStatus(); err != nil {
		return err
	}
	if s.acceptedBefore, err = s.acceptedEntries(); err != nil {
		return err
	}

	body := map[string]any{"decision": decision}
	if reason != "" {
		body["reason"] = reason
	}
	path := fmt.Sprintf("/tournaments/%s/entries/%s/review", s.tournament.ID, s.team.ID)
	s.resp, err = s.client.Do(http.MethodPost, path, body, s.reviewer.AuthHeaders())
	return err
}

func (s *reviewScenarioContext) theOrganizerApprovesTheApplication() error {
	return s.review("ACCEPT", "")
}

func (s *reviewScenarioContext) theOrganizerRejectsTheApplicationWithAReason() error {
	return s.review("REJECT", "The roster does not meet the tournament requirements")
}

func (s *reviewScenarioContext) theOrganizerRejectsTheApplication() error {
	return s.review("REJECT", "")
}

// ---- Then steps ----

func (s *reviewScenarioContext) teamStatus() (models.TournamentTeamStatus, error) {
	var team models.TournamentTeam
	if err := app.DB.First(&team, "id = ?", s.team.ID).Error; err != nil {
		return "", fmt.Errorf("load team: %w", err)
	}
	return team.Status, nil
}

func (s *reviewScenarioContext) acceptedEntries() (int, error) {
	path := fmt.Sprintf("/tournaments/%s/dashboard", s.tournament.ID)
	resp, err := s.client.Do(http.MethodGet, path, nil, s.owner.AuthHeaders())
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dashboard: expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	var dashboard struct {
		Summary struct {
			Metrics struct {
				AcceptedEntries *int `json:"acceptedEntries"`
			} `json:"metrics"`
		} `json:"summary"`
	}
	if err := resp.JSON(&dashboard); err != nil {
		return 0, err
	}
	// A pointer distinguishes "missing field" from a genuine count of zero, so a
	// wrong path fails loudly instead of silently reading 0.
	if dashboard.Summary.Metrics.AcceptedEntries == nil {
		return 0, fmt.Errorf("dashboard response has no summary.metrics.acceptedEntries: %s", resp.Body)
	}
	return *dashboard.Summary.Metrics.AcceptedEntries, nil
}

func (s *reviewScenarioContext) expectDecision(want models.TournamentTeamStatus) error {
	if s.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected the review to succeed, got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	got, err := s.teamStatus()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("expected application status %s, got %s", want, got)
	}
	return nil
}

func (s *reviewScenarioContext) theApplicationShouldBeAccepted() error {
	return s.expectDecision(models.TournamentTeamStatusAccepted)
}

func (s *reviewScenarioContext) theApplicationShouldBeRejected() error {
	return s.expectDecision(models.TournamentTeamStatusRejected)
}

func (s *reviewScenarioContext) expectAcceptedCountChange(delta int) error {
	got, err := s.acceptedEntries()
	if err != nil {
		return err
	}
	if want := s.acceptedBefore + delta; got != want {
		return fmt.Errorf("expected accepted count %d (before %d, change %+d), got %d", want, s.acceptedBefore, delta, got)
	}
	return nil
}

func (s *reviewScenarioContext) theAcceptedCountShouldIncrease() error {
	return s.expectAcceptedCountChange(1)
}

func (s *reviewScenarioContext) theAcceptedCountShouldStayTheSame() error {
	return s.expectAcceptedCountChange(0)
}

func (s *reviewScenarioContext) theAcceptedCountShouldDecrease() error {
	return s.expectAcceptedCountChange(-1)
}

func (s *reviewScenarioContext) theApprovalShouldBeBlocked() error {
	switch s.resp.StatusCode {
	case http.StatusConflict, http.StatusNotFound:
	default:
		return fmt.Errorf("expected approval to be blocked with 409 or 404, got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	// A bare router 404 ("404 page not found") means the endpoint does not
	// exist. A real rejection is a JSON error body carrying its own status.
	var problem struct {
		Status int `json:"status"`
	}
	if err := s.resp.JSON(&problem); err != nil || problem.Status != s.resp.StatusCode {
		return fmt.Errorf("expected a JSON error body for status %d, got: %s", s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *reviewScenarioContext) theApplicationShouldKeepItsPreviousStatus() error {
	got, err := s.teamStatus()
	if err != nil {
		return err
	}
	if got != s.statusBefore {
		return fmt.Errorf("expected application to stay %s, got %s", s.statusBefore, got)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	s := &reviewScenarioContext{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = reviewScenarioContext{client: apptest.NewClient(app.TournamentBaseURL())}
		cleanup := "TRUNCATE tournament_staff, tournament_team_members, tournament_fundings, tournament_status_changes, tournament_teams, tournaments, sponsor_profiles, organizer_profiles, accounts CASCADE"
		if err := app.DB.Exec(cleanup).Error; err != nil {
			return ctx, fmt.Errorf("clean scenario database: %w", err)
		}
		return ctx, nil
	})

	// Background
	sc.Step(`^the Eventory API service is running$`, s.theEventoryAPIServiceIsRunning)

	// Given
	sc.Step(`^an organizer manages a tournament that still has open spots$`, s.anOrganizerManagesATournamentThatStillHasOpenSpots)
	sc.Step(`^a team has locked its application for that tournament$`, s.aTeamHasLockedItsApplicationForThatTournament)
	sc.Step(`^a team application has already been accepted$`, s.aTeamApplicationHasAlreadyBeenAccepted)
	sc.Step(`^an organizer reviews an application where (.+)$`, s.anOrganizerReviewsAnApplicationWhere)

	// When
	sc.Step(`^the organizer approves the application$`, s.theOrganizerApprovesTheApplication)
	sc.Step(`^the organizer rejects the application with a reason$`, s.theOrganizerRejectsTheApplicationWithAReason)
	sc.Step(`^the organizer rejects the application$`, s.theOrganizerRejectsTheApplication)

	// Then
	sc.Step(`^the application should be accepted$`, s.theApplicationShouldBeAccepted)
	sc.Step(`^the application should be rejected$`, s.theApplicationShouldBeRejected)
	sc.Step(`^the accepted count on the dashboard should increase$`, s.theAcceptedCountShouldIncrease)
	sc.Step(`^the accepted count on the dashboard should stay the same$`, s.theAcceptedCountShouldStayTheSame)
	sc.Step(`^the accepted count on the dashboard should decrease$`, s.theAcceptedCountShouldDecrease)
	sc.Step(`^the approval should be blocked$`, s.theApprovalShouldBeBlocked)
	sc.Step(`^the application should keep its previous status$`, s.theApplicationShouldKeepItsPreviousStatus)
}
