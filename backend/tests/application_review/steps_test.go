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

// Statuses are spelled out here on purpose rather than taken from the model
// constants: team and solo applications must use the same ACCEPTED/REJECTED
// spelling, and a literal makes a drift between them fail these scenarios.
const (
	statusAccepted = "ACCEPTED"
	statusRejected = "REJECTED"
)

type reviewScenarioContext struct {
	client         *apptest.Client
	reviewer       *apptest.TestUser // who sends the review request
	owner          *apptest.TestUser // tournament owner, used to read the dashboard
	tournament     *models.Tournament
	team           *models.TournamentTeam   // the application under review, in a team tournament
	solo           *models.SoloRegistration // the application under review, in a solo tournament
	statusBefore   string
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
	return s.createTournamentIn(models.TournamentRegistrationModeTeam, capacity)
}

// createTournamentIn seeds a published tournament in the given registration
// mode with registration open. Its owner becomes the default reviewer.
func (s *reviewScenarioContext) createTournamentIn(mode models.TournamentRegistrationMode, capacity int) error {
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

	maxTeamSize := 3
	if mode == models.TournamentRegistrationModeSolo {
		maxTeamSize = 1
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
		RegistrationMode:     mode,
		MinTeamSize:          1,
		MaxTeamSize:          maxTeamSize,
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

// seedSolo writes a solo registration with the given status straight into the
// database, for the same reason as seedTeam.
func (s *reviewScenarioContext) seedSolo(status models.SoloRegistrationStatus) (*models.SoloRegistration, error) {
	player := apptest.NewTestUser()
	if err := s.persistActiveAccount(player); err != nil {
		return nil, err
	}
	solo := models.SoloRegistration{
		ID:           uuid.New(),
		TournamentID: s.tournament.ID,
		AccountID:    player.ID,
		Status:       status,
	}
	if status != models.SoloPaymentPending {
		registeredAt := time.Now().UTC()
		solo.RegisteredAt = &registeredAt
	}
	if err := app.DB.Create(&solo).Error; err != nil {
		return nil, fmt.Errorf("create %s solo registration: %w", status, err)
	}
	return &solo, nil
}

// makeReviewerAnotherOrganizer lets a stranger, who does not manage the
// tournament, send the review request.
func (s *reviewScenarioContext) makeReviewerAnotherOrganizer() error {
	stranger := apptest.NewTestUser()
	if err := s.persistActiveAccount(stranger); err != nil {
		return err
	}
	s.reviewer = stranger
	return nil
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
		return s.makeReviewerAnotherOrganizer()
	}
	return fmt.Errorf("unknown review situation %q", situation)
}

func (s *reviewScenarioContext) anOrganizerManagesASoloTournamentThatStillHasOpenSpots() error {
	return s.createTournamentIn(models.TournamentRegistrationModeSolo, 8)
}

func (s *reviewScenarioContext) aSoloPlayerIsAwaitingApprovalForThatTournament() error {
	solo, err := s.seedSolo(models.SoloAwaitingApproval)
	s.solo = solo
	return err
}

func (s *reviewScenarioContext) aSoloPlayerApplicationHasAlreadyBeenAccepted() error {
	solo, err := s.seedSolo(models.SoloAccepted)
	s.solo = solo
	return err
}

func (s *reviewScenarioContext) anOrganizerReviewsASoloApplicationWhere(situation string) error {
	switch situation {
	case "the tournament has already reached its capacity":
		if err := s.createTournamentIn(models.TournamentRegistrationModeSolo, 1); err != nil {
			return err
		}
		if _, err := s.seedSolo(models.SoloAccepted); err != nil {
			return err
		}
		return s.aSoloPlayerIsAwaitingApprovalForThatTournament()
	case "the player has not completed payment yet":
		if err := s.createTournamentIn(models.TournamentRegistrationModeSolo, 8); err != nil {
			return err
		}
		solo, err := s.seedSolo(models.SoloPaymentPending)
		s.solo = solo
		return err
	case "the application has already been accepted":
		if err := s.createTournamentIn(models.TournamentRegistrationModeSolo, 8); err != nil {
			return err
		}
		return s.aSoloPlayerApplicationHasAlreadyBeenAccepted()
	case "the tournament is managed by another organizer":
		if err := s.createTournamentIn(models.TournamentRegistrationModeSolo, 8); err != nil {
			return err
		}
		if err := s.aSoloPlayerIsAwaitingApprovalForThatTournament(); err != nil {
			return err
		}
		return s.makeReviewerAnotherOrganizer()
	}
	return fmt.Errorf("unknown solo review situation %q", situation)
}

// anAdminHasBeenAddedToThatTournamentsStaff makes an ADMIN of the tournament
// the reviewer. The dashboard is still read as the owner.
func (s *reviewScenarioContext) anAdminHasBeenAddedToThatTournamentsStaff() error {
	admin := apptest.NewTestUser()
	if err := s.persistActiveAccount(admin); err != nil {
		return err
	}
	staff := models.TournamentStaff{
		TournamentID: s.tournament.ID,
		AccountID:    admin.ID,
		Role:         models.TournamentStaffRoleAdmin,
	}
	if err := app.DB.Create(&staff).Error; err != nil {
		return fmt.Errorf("add admin to tournament staff: %w", err)
	}
	s.reviewer = admin
	return nil
}

// ---- When steps ----

// entryID is the application under review, whichever kind it is.
func (s *reviewScenarioContext) entryID() uuid.UUID {
	if s.solo != nil {
		return s.solo.ID
	}
	return s.team.ID
}

// review sends the decision and snapshots the state beforehand so the Then
// steps can compare against it.
func (s *reviewScenarioContext) review(decision, reason string) error {
	var err error
	if s.statusBefore, err = s.applicationStatus(); err != nil {
		return err
	}
	if s.acceptedBefore, err = s.acceptedEntries(); err != nil {
		return err
	}

	body := map[string]any{"decision": decision}
	if reason != "" {
		body["reason"] = reason
	}
	path := fmt.Sprintf("/tournaments/%s/entries/%s/review", s.tournament.ID, s.entryID())
	s.resp, err = s.client.Do(http.MethodPost, path, body, s.reviewer.AuthHeaders())
	return err
}

func (s *reviewScenarioContext) theOrganizerApprovesTheApplication() error {
	return s.review("ACCEPT", "")
}

func (s *reviewScenarioContext) theAdminApprovesTheApplication() error {
	return s.review("ACCEPT", "")
}

func (s *reviewScenarioContext) theOrganizerRejectsTheApplicationWithAReason() error {
	return s.review("REJECT", "The roster does not meet the tournament requirements")
}

func (s *reviewScenarioContext) theOrganizerRejectsTheApplication() error {
	return s.review("REJECT", "")
}

// ---- Then steps ----

// applicationStatus reads the current status of the application under review
// straight from the database.
func (s *reviewScenarioContext) applicationStatus() (string, error) {
	if s.solo != nil {
		var solo models.SoloRegistration
		if err := app.DB.First(&solo, "id = ?", s.solo.ID).Error; err != nil {
			return "", fmt.Errorf("load solo registration: %w", err)
		}
		return string(solo.Status), nil
	}
	var team models.TournamentTeam
	if err := app.DB.First(&team, "id = ?", s.team.ID).Error; err != nil {
		return "", fmt.Errorf("load team: %w", err)
	}
	return string(team.Status), nil
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

func (s *reviewScenarioContext) expectDecision(want string) error {
	if s.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected the review to succeed, got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	got, err := s.applicationStatus()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("expected application status %s, got %s", want, got)
	}
	return nil
}

func (s *reviewScenarioContext) theApplicationShouldBeAccepted() error {
	return s.expectDecision(statusAccepted)
}

func (s *reviewScenarioContext) theApplicationShouldBeRejected() error {
	return s.expectDecision(statusRejected)
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
	got, err := s.applicationStatus()
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
		// CASCADE also clears solo_registrations and the review audit tables,
		// which all reference tournaments or the rows truncated here.
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
	sc.Step(`^an organizer manages a solo tournament that still has open spots$`, s.anOrganizerManagesASoloTournamentThatStillHasOpenSpots)
	sc.Step(`^a solo player is awaiting approval for that tournament$`, s.aSoloPlayerIsAwaitingApprovalForThatTournament)
	sc.Step(`^a solo player application has already been accepted$`, s.aSoloPlayerApplicationHasAlreadyBeenAccepted)
	sc.Step(`^an organizer reviews a solo application where (.+)$`, s.anOrganizerReviewsASoloApplicationWhere)
	sc.Step(`^an admin has been added to that tournament's staff$`, s.anAdminHasBeenAddedToThatTournamentsStaff)

	// When
	sc.Step(`^the organizer approves the application$`, s.theOrganizerApprovesTheApplication)
	sc.Step(`^the admin approves the application$`, s.theAdminApprovesTheApplication)
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
