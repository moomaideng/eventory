package team_lobbies

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

var inviteCodePattern = regexp.MustCompile(`^[A-Z2-9]{6}$`)

type teamLobbyScenarioContext struct {
	client          *apptest.Client
	actor           *apptest.TestUser
	captain         *apptest.TestUser
	otherCompetitor *apptest.TestUser
	tournament      *models.Tournament
	lobby           handlers.TeamLobbyResponse
	resp            *apptest.Response
}

func (s *teamLobbyScenarioContext) theEventoryAPIServiceIsRunning() error {
	if app == nil || app.Server == nil {
		return fmt.Errorf("application server is not running")
	}
	return nil
}

func (s *teamLobbyScenarioContext) aPublishedTeamTournamentWithOpenRegistration() error {
	return s.createPublishedTeamTournament()
}

func (s *teamLobbyScenarioContext) createPublishedTeamTournament() error {
	organizer := apptest.NewTestUser()
	if err := s.persistActiveAccount(organizer); err != nil {
		return err
	}
	organizerProfile := models.OrganizerProfile{
		ID:             uuid.New(),
		AccountID:      organizer.ID,
		OrganizerName:  "BDD Tournament Organizer",
		OrganizerEmail: organizer.Email,
	}
	if err := app.DB.Create(&organizerProfile).Error; err != nil {
		return fmt.Errorf("create organizer profile: %w", err)
	}

	now := time.Now().UTC()
	s.tournament = &models.Tournament{
		ID:                   uuid.New(),
		OrganizerID:          organizerProfile.ID,
		Name:                 "BDD Tournament " + apptest.UniqueSuffix(),
		Description:          "Tournament created for team lobby acceptance tests.",
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
		Status:               models.TournamentStatusRegistrationOpen,
		Published:            true,
	}
	if err := app.DB.Create(s.tournament).Error; err != nil {
		return fmt.Errorf("create tournament: %w", err)
	}
	return nil
}

func (s *teamLobbyScenarioContext) anActiveCompetitor() error {
	s.actor = apptest.NewTestUser()
	return s.persistActiveAccount(s.actor)
}

func (s *teamLobbyScenarioContext) anotherActiveCompetitor() error {
	s.otherCompetitor = apptest.NewTestUser()
	return s.persistActiveAccount(s.otherCompetitor)
}

func (s *teamLobbyScenarioContext) persistActiveAccount(user *apptest.TestUser) error {
	account := models.Account{
		ID:          user.ID,
		Email:       user.Email,
		Handle:      user.Handle,
		DisplayName: user.DisplayName,
		Status:      models.AccountStatusActive,
	}
	if err := app.DB.Create(&account).Error; err != nil {
		return fmt.Errorf("create active account %s: %w", user.Email, err)
	}
	return nil
}

func (s *teamLobbyScenarioContext) aCaptainHasCreatedAFormingTeamLobby() error {
	if err := s.aPublishedTeamTournamentWithOpenRegistration(); err != nil {
		return err
	}
	s.captain = apptest.NewTestUser()
	if err := s.persistActiveAccount(s.captain); err != nil {
		return err
	}
	if err := s.createLobby(s.captain, "Captain's Squad"); err != nil {
		return err
	}
	if s.resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create forming lobby: expected 201, got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *teamLobbyScenarioContext) theCompetitorCreatesATeamLobbyNamed(name string) error {
	return s.createLobby(s.actor, name)
}

func (s *teamLobbyScenarioContext) theCompetitorTriesToCreateATeamLobbyWithoutEnteringATeamName() error {
	return s.createLobby(s.actor, "")
}

func (s *teamLobbyScenarioContext) createLobby(user *apptest.TestUser, name string) error {
	if s.tournament == nil {
		return fmt.Errorf("no tournament configured")
	}
	resp, err := s.client.Do(
		http.MethodPost,
		fmt.Sprintf("/tournaments/%s/lobbies", s.tournament.ID),
		handlers.CreateTeamLobbyRequest{Name: name},
		user.AuthHeaders(),
	)
	if err != nil {
		return err
	}
	s.resp = resp
	if resp.StatusCode == http.StatusCreated {
		if err := resp.JSON(&s.lobby); err != nil {
			return fmt.Errorf("parse created lobby: %w", err)
		}
	}
	return nil
}

func (s *teamLobbyScenarioContext) theOtherCompetitorJoinsUsingCurrentInviteCode() error {
	resp, err := s.client.Do(
		http.MethodPost,
		fmt.Sprintf("/lobbies/%s/join", s.lobby.InviteCode),
		nil,
		s.otherCompetitor.AuthHeaders(),
	)
	if err != nil {
		return err
	}
	s.resp = resp
	if resp.StatusCode == http.StatusOK {
		if err := resp.JSON(&s.lobby); err != nil {
			return fmt.Errorf("parse joined lobby: %w", err)
		}
	}
	return nil
}

func (s *teamLobbyScenarioContext) theCompetitorTriesToJoinWithUnknownInviteCode(inviteCode string) error {
	resp, err := s.client.Do(http.MethodPost, fmt.Sprintf("/lobbies/%s/join", inviteCode), nil, s.actor.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *teamLobbyScenarioContext) responseStatusShouldBe(expected int) error {
	if s.resp == nil {
		return fmt.Errorf("no HTTP response captured")
	}
	if s.resp.StatusCode != expected {
		return fmt.Errorf("expected status %d, got %d with body: %s", expected, s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *teamLobbyScenarioContext) theTeamLobbyShouldBeCreatedSuccessfully() error {
	return s.responseStatusShouldBe(http.StatusCreated)
}

func (s *teamLobbyScenarioContext) theTeamLobbyCreationShouldFail() error {
	return s.responseStatusShouldBe(http.StatusUnprocessableEntity)
}

func (s *teamLobbyScenarioContext) theOtherCompetitorShouldJoinTheTeamLobbySuccessfully() error {
	return s.responseStatusShouldBe(http.StatusOK)
}

func (s *teamLobbyScenarioContext) joiningTheTeamLobbyShouldFail() error {
	return s.responseStatusShouldBe(http.StatusNotFound)
}

func (s *teamLobbyScenarioContext) theLobbyShouldBeNamed(expected string) error {
	if s.lobby.Name != expected {
		return fmt.Errorf("expected lobby name %q, got %q", expected, s.lobby.Name)
	}
	return nil
}

func (s *teamLobbyScenarioContext) theLobbyStatusShouldBe(expected string) error {
	if s.lobby.Status != expected {
		return fmt.Errorf("expected lobby status %q, got %q", expected, s.lobby.Status)
	}
	return nil
}

func (s *teamLobbyScenarioContext) theCompetitorShouldBeTheLobbyCaptain() error {
	if s.lobby.CaptainID != s.actor.ID || s.lobby.ViewerRole != "CAPTAIN" {
		return fmt.Errorf("expected competitor %s to be captain, got captain %s and viewer role %q", s.actor.ID, s.lobby.CaptainID, s.lobby.ViewerRole)
	}
	for _, member := range s.lobby.Members {
		if member.AccountID == s.actor.ID && member.Role == "CAPTAIN" {
			return nil
		}
	}
	return fmt.Errorf("captain membership for competitor %s was not returned", s.actor.ID)
}

func (s *teamLobbyScenarioContext) theLobbyShouldHaveAValidSixCharacterInviteCode() error {
	if !inviteCodePattern.MatchString(s.lobby.InviteCode) {
		return fmt.Errorf("expected a valid six-character invite code, got %q", s.lobby.InviteCode)
	}
	return nil
}

func (s *teamLobbyScenarioContext) theOtherCompetitorsProfileShouldBeAttachedToTheTeamEntry() error {
	for _, member := range s.lobby.Members {
		if member.AccountID != s.otherCompetitor.ID {
			continue
		}
		if member.Handle != s.otherCompetitor.Handle || member.DisplayName != s.otherCompetitor.DisplayName {
			return fmt.Errorf(
				"expected attached profile %q/%q, got %q/%q",
				s.otherCompetitor.Handle,
				s.otherCompetitor.DisplayName,
				member.Handle,
				member.DisplayName,
			)
		}
		if member.Role != "MEMBER" {
			return fmt.Errorf("expected attached competitor role MEMBER, got %q", member.Role)
		}
		return nil
	}
	return fmt.Errorf("competitor profile %s was not attached to the team entry", s.otherCompetitor.ID)
}

func (s *teamLobbyScenarioContext) theLobbyShouldContainMembers(expected int) error {
	if len(s.lobby.Members) != expected {
		return fmt.Errorf("expected %d members in response, got %d", expected, len(s.lobby.Members))
	}
	return nil
}

func (s *teamLobbyScenarioContext) theCompetitorShouldNotBeAttachedToATournamentTeam() error {
	var count int64
	if err := app.DB.Model(&models.TournamentTeamMember{}).
		Where("account_id = ?", s.actor.ID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("count competitor team memberships: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("expected competitor to have no tournament team membership, got %d", count)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	s := &teamLobbyScenarioContext{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = teamLobbyScenarioContext{client: apptest.NewClient(app.BaseURL())}
		cleanup := "TRUNCATE tournament_team_members, tournament_fundings, tournament_status_changes, tournament_teams, tournaments, sponsor_profiles, organizer_profiles, accounts CASCADE"
		if err := app.DB.Exec(cleanup).Error; err != nil {
			return ctx, fmt.Errorf("clean scenario database: %w", err)
		}
		return ctx, nil
	})

	sc.Step(`^the Eventory API service is running$`, s.theEventoryAPIServiceIsRunning)
	sc.Step(`^a published team tournament exists with registration open$`, s.aPublishedTeamTournamentWithOpenRegistration)
	sc.Step(`^the competitor has signed in and completed onboarding$`, s.anActiveCompetitor)
	sc.Step(`^another competitor has signed in and completed onboarding$`, s.anotherActiveCompetitor)
	sc.Step(`^a captain has created a forming team lobby$`, s.aCaptainHasCreatedAFormingTeamLobby)
	sc.Step(`^the competitor creates a team lobby named "([^"]*)"$`, s.theCompetitorCreatesATeamLobbyNamed)
	sc.Step(`^the competitor tries to create a team lobby without entering a team name$`, s.theCompetitorTriesToCreateATeamLobbyWithoutEnteringATeamName)
	sc.Step(`^the other competitor joins the lobby using its current invite code$`, s.theOtherCompetitorJoinsUsingCurrentInviteCode)
	sc.Step(`^the competitor tries to join with unknown invite code "([^"]*)"$`, s.theCompetitorTriesToJoinWithUnknownInviteCode)
	sc.Step(`^the team lobby should be created successfully$`, s.theTeamLobbyShouldBeCreatedSuccessfully)
	sc.Step(`^the team lobby creation should fail$`, s.theTeamLobbyCreationShouldFail)
	sc.Step(`^the other competitor should join the team lobby successfully$`, s.theOtherCompetitorShouldJoinTheTeamLobbySuccessfully)
	sc.Step(`^joining the team lobby should fail$`, s.joiningTheTeamLobbyShouldFail)
	sc.Step(`^the lobby should be named "([^"]*)"$`, s.theLobbyShouldBeNamed)
	sc.Step(`^the lobby status should be "([^"]*)"$`, s.theLobbyStatusShouldBe)
	sc.Step(`^the competitor should be the lobby captain$`, s.theCompetitorShouldBeTheLobbyCaptain)
	sc.Step(`^the lobby should have a valid six-character invite code$`, s.theLobbyShouldHaveAValidSixCharacterInviteCode)
	sc.Step(`^the other competitor's profile should be attached to the team entry$`, s.theOtherCompetitorsProfileShouldBeAttachedToTheTeamEntry)
	sc.Step(`^the lobby should contain (\d+) members$`, s.theLobbyShouldContainMembers)
	sc.Step(`^the competitor should not be attached to a tournament team$`, s.theCompetitorShouldNotBeAttachedToATournamentTeam)
}
