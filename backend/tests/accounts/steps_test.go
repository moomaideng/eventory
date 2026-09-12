package accounts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

type accountScenarioContext struct {
	client      *apptest.Client
	currentUser *apptest.TestUser
	secondUser  *apptest.TestUser
	resp        *apptest.Response
}

func (s *accountScenarioContext) theEventoryAPIServiceIsRunning() error {
	if app == nil || app.Server == nil {
		return fmt.Errorf("application server is not running")
	}
	return nil
}

func (s *accountScenarioContext) aNewAuthenticatedUserWithEmail(email string) error {
	id := uuid.New()
	s.currentUser = &apptest.TestUser{
		ID:    id,
		Email: email,
		Token: apptest.GenerateTestJWT(id, email),
	}
	return nil
}

func (s *accountScenarioContext) aNewAuthenticatedUser() error {
	s.currentUser = apptest.NewTestUser()
	return nil
}

func (s *accountScenarioContext) anotherAuthenticatedUser() error {
	s.secondUser = apptest.NewTestUser()
	return nil
}

func (s *accountScenarioContext) anExistingUserOnboardedWithHandle(handle string) error {
	s.currentUser = apptest.NewTestUser()
	body := handlers.CreateAccountRequest{
		DisplayName: "Existing Champion",
		Handle:      &handle,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected 201 when creating initial user, got %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func (s *accountScenarioContext) anotherUserOnboardedWithHandle(handle string) error {
	s.secondUser = apptest.NewTestUser()
	body := handlers.CreateAccountRequest{
		DisplayName: "Challenger Two",
		Handle:      &handle,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.secondUser.AuthHeaders())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected 201 when creating second user, got %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func (s *accountScenarioContext) theSecondUserUpdatesTheirProfileWithHandle(handle string) error {
	body := handlers.UpdateAccountRequest{
		Handle: &handle,
	}
	resp, err := s.client.Do(http.MethodPatch, "/accounts/me", body, s.secondUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) anExistingUserOnboardedWithDisplayNameAndHandle(name, handle string) error {
	s.currentUser = apptest.NewTestUser()
	body := handlers.CreateAccountRequest{
		DisplayName: name,
		Handle:      &handle,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected 201 when creating initial user, got %d: %s", resp.StatusCode, resp.Body)
	}

	// Complete onboarding so account status transitions from onboarding to active
	patchBody := handlers.UpdateAccountRequest{
		DisplayName: &name,
	}
	patchResp, err := s.client.Do(http.MethodPatch, "/accounts/me", patchBody, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	if patchResp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 when activating onboarded user, got %d: %s", patchResp.StatusCode, patchResp.Body)
	}

	return nil
}

func (s *accountScenarioContext) theUserSubmitsOnboardingRequestWithDisplayNameAndHandle(displayName, handle string) error {
	body := handlers.CreateAccountRequest{
		DisplayName: displayName,
		Handle:      &handle,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserSubmitsOnboardingRequestWithDisplayNameAndEmptyHandle(displayName string) error {
	body := handlers.CreateAccountRequest{
		DisplayName: displayName,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theSecondUserSubmitsOnboardingRequestWithHandle(handle string) error {
	body := handlers.CreateAccountRequest{
		DisplayName: "Second Imposter",
		Handle:      &handle,
	}
	resp, err := s.client.Do(http.MethodPost, "/accounts", body, s.secondUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserRequestsTheirCurrentAccountDetailsViaGetMe() error {
	resp, err := s.client.Do(http.MethodGet, "/accounts/me", nil, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserUpdatesTheirProfileWithDisplayNameAndPhone(name, phone string) error {
	body := handlers.UpdateAccountRequest{
		DisplayName: &name,
		Phone:       &phone,
	}
	resp, err := s.client.Do(http.MethodPatch, "/accounts/me", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserUpdatesTheirProfileWithDisplayName(name string) error {
	body := handlers.UpdateAccountRequest{
		DisplayName: &name,
	}
	resp, err := s.client.Do(http.MethodPatch, "/accounts/me", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserUpsertsTheirOrganizerProfile(orgName, email string) error {
	body := handlers.UpsertOrganizerProfileRequest{
		OrganizerName:  orgName,
		OrganizerEmail: &email,
	}
	resp, err := s.client.Do(http.MethodPut, "/accounts/me/organizer-profile", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserRetrievesTheirOrganizerProfile() error {
	resp, err := s.client.Do(http.MethodGet, "/accounts/me/organizer-profile", nil, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserUpsertsTheirSponsorProfile(sponsorName, email string) error {
	body := handlers.UpsertSponsorProfileRequest{
		SponsorName:  sponsorName,
		SponsorEmail: &email,
	}
	resp, err := s.client.Do(http.MethodPut, "/accounts/me/sponsor-profile", body, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theUserRetrievesTheirSponsorProfile() error {
	resp, err := s.client.Do(http.MethodGet, "/accounts/me/sponsor-profile", nil, s.currentUser.AuthHeaders())
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) anUnauthenticatedRequestIsSentToGetMe() error {
	resp, err := s.client.Do(http.MethodGet, "/accounts/me", nil, nil)
	if err != nil {
		return err
	}
	s.resp = resp
	return nil
}

func (s *accountScenarioContext) theResponseStatusCodeShouldBe(expectedStatus int) error {
	if s.resp == nil {
		return fmt.Errorf("no HTTP response captured")
	}
	if s.resp.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d with body: %s", expectedStatus, s.resp.StatusCode, string(s.resp.Body))
	}
	return nil
}

func (s *accountScenarioContext) theCreatedAccountShouldHaveDisplayName(expectedName string) error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if out.DisplayName != expectedName {
		return fmt.Errorf("expected displayName %q, got %q", expectedName, out.DisplayName)
	}
	return nil
}

func (s *accountScenarioContext) theCreatedAccountShouldHaveANonEmptyHandle() error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if strings.TrimSpace(out.Handle) == "" {
		return fmt.Errorf("expected non-empty handle, got empty string")
	}
	return nil
}

func (s *accountScenarioContext) theAccountHandleShouldBeValid() error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if len(out.Handle) < 3 {
		return fmt.Errorf("expected valid handle >= 3 chars, got %q", out.Handle)
	}
	return nil
}

func (s *accountScenarioContext) theReturnedAccountEmailShouldMatchUsersRegisteredEmail() error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if !strings.EqualFold(out.Email, s.currentUser.Email) {
		return fmt.Errorf("expected email %q, got %q", s.currentUser.Email, out.Email)
	}
	return nil
}

func (s *accountScenarioContext) theReturnedAccountDisplayNameShouldBe(expected string) error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if out.DisplayName != expected {
		return fmt.Errorf("expected displayName %q, got %q", expected, out.DisplayName)
	}
	return nil
}

func (s *accountScenarioContext) theUpdatedAccountDisplayNameShouldBe(expected string) error {
	return s.theReturnedAccountDisplayNameShouldBe(expected)
}

func (s *accountScenarioContext) theUpdatedAccountPhoneShouldBe(expected string) error {
	var out handlers.AccountResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse account response: %w", err)
	}
	if out.Phone == nil || *out.Phone != expected {
		return fmt.Errorf("expected phone %q, got %v", expected, out.Phone)
	}
	return nil
}

func (s *accountScenarioContext) theOrganizerProfileNameShouldBe(expected string) error {
	var out handlers.OrganizerProfileResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse organizer response: %w", err)
	}
	if out.OrganizerName != expected {
		return fmt.Errorf("expected organizerName %q, got %q", expected, out.OrganizerName)
	}
	return nil
}

func (s *accountScenarioContext) theOrganizerContactEmailShouldBe(expected string) error {
	var out handlers.OrganizerProfileResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse organizer response: %w", err)
	}
	if out.OrganizerEmail != expected {
		return fmt.Errorf("expected organizerEmail %q, got %q", expected, out.OrganizerEmail)
	}
	return nil
}

func (s *accountScenarioContext) theSponsorProfileNameShouldBe(expected string) error {
	var out handlers.SponsorProfileResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse sponsor response: %w", err)
	}
	if out.SponsorName != expected {
		return fmt.Errorf("expected sponsorName %q, got %q", expected, out.SponsorName)
	}
	return nil
}

func (s *accountScenarioContext) theSponsorContactEmailShouldBe(expected string) error {
	var out handlers.SponsorProfileResponse
	if err := s.resp.JSON(&out); err != nil {
		return fmt.Errorf("failed to parse sponsor response: %w", err)
	}
	if out.SponsorEmail != expected {
		return fmt.Errorf("expected sponsorEmail %q, got %q", expected, out.SponsorEmail)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var s *accountScenarioContext

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		s = &accountScenarioContext{
			client: apptest.NewClient(app.BaseURL()),
		}
		// Clean tables between scenarios to guarantee absolute isolation
		app.DB.Exec("DELETE FROM sponsor_profiles; DELETE FROM organizer_profiles; DELETE FROM accounts;")
		return ctx, nil
	})

	sc.Step(`^the Eventory API service is running$`, func() error {
		return s.theEventoryAPIServiceIsRunning()
	})
	sc.Step(`^a new authenticated user with email "([^"]*)"$`, func(email string) error {
		return s.aNewAuthenticatedUserWithEmail(email)
	})
	sc.Step(`^a new authenticated user$`, func() error {
		return s.aNewAuthenticatedUser()
	})
	sc.Step(`^another authenticated user$`, func() error {
		return s.anotherAuthenticatedUser()
	})
	sc.Step(`^an existing user onboarded with handle "([^"]*)"$`, func(handle string) error {
		return s.anExistingUserOnboardedWithHandle(handle)
	})
	sc.Step(`^another user onboarded with handle "([^"]*)"$`, func(handle string) error {
		return s.anotherUserOnboardedWithHandle(handle)
	})
	sc.Step(`^the second user updates their profile with handle "([^"]*)"$`, func(handle string) error {
		return s.theSecondUserUpdatesTheirProfileWithHandle(handle)
	})
	sc.Step(`^an existing user onboarded with display name "([^"]*)" and handle "([^"]*)"$`, func(name, handle string) error {
		return s.anExistingUserOnboardedWithDisplayNameAndHandle(name, handle)
	})
	sc.Step(`^the user submits onboarding request with display name "([^"]*)" and handle "([^"]*)"$`, func(name, handle string) error {
		return s.theUserSubmitsOnboardingRequestWithDisplayNameAndHandle(name, handle)
	})
	sc.Step(`^the user submits onboarding request with display name "([^"]*)" and empty handle$`, func(name string) error {
		return s.theUserSubmitsOnboardingRequestWithDisplayNameAndEmptyHandle(name)
	})
	sc.Step(`^the second user submits onboarding request with handle "([^"]*)"$`, func(handle string) error {
		return s.theSecondUserSubmitsOnboardingRequestWithHandle(handle)
	})
	sc.Step(`^the user requests their current account details via "GET /me"$`, func() error {
		return s.theUserRequestsTheirCurrentAccountDetailsViaGetMe()
	})
	sc.Step(`^the user updates their profile with display name "([^"]*)" and phone "([^"]*)"$`, func(name, phone string) error {
		return s.theUserUpdatesTheirProfileWithDisplayNameAndPhone(name, phone)
	})
	sc.Step(`^the user updates their profile with display name "([^"]*)"$`, func(name string) error {
		return s.theUserUpdatesTheirProfileWithDisplayName(name)
	})
	sc.Step(`^the user upserts their organizer profile with organization "([^"]*)" and email "([^"]*)"$`, func(name, email string) error {
		return s.theUserUpsertsTheirOrganizerProfile(name, email)
	})
	sc.Step(`^the user retrieves their organizer profile$`, func() error {
		return s.theUserRetrievesTheirOrganizerProfile()
	})
	sc.Step(`^the user upserts their sponsor profile with company "([^"]*)" and email "([^"]*)"$`, func(name, email string) error {
		return s.theUserUpsertsTheirSponsorProfile(name, email)
	})
	sc.Step(`^the user retrieves their sponsor profile$`, func() error {
		return s.theUserRetrievesTheirSponsorProfile()
	})
	sc.Step(`^an unauthenticated request is sent to "GET /me"$`, func() error {
		return s.anUnauthenticatedRequestIsSentToGetMe()
	})
	sc.Step(`^the response status code should be (\d+)$`, func(code int) error {
		return s.theResponseStatusCodeShouldBe(code)
	})
	sc.Step(`^the created account should have display name "([^"]*)"$`, func(name string) error {
		return s.theCreatedAccountShouldHaveDisplayName(name)
	})
	sc.Step(`^the created account should have a non-empty handle$`, func() error {
		return s.theCreatedAccountShouldHaveANonEmptyHandle()
	})
	sc.Step(`^the account handle should be valid$`, func() error {
		return s.theAccountHandleShouldBeValid()
	})
	sc.Step(`^the returned account email should match the user's registered email$`, func() error {
		return s.theReturnedAccountEmailShouldMatchUsersRegisteredEmail()
	})
	sc.Step(`^the returned account display name should be "([^"]*)"$`, func(name string) error {
		return s.theReturnedAccountDisplayNameShouldBe(name)
	})
	sc.Step(`^the updated account display name should be "([^"]*)"$`, func(name string) error {
		return s.theUpdatedAccountDisplayNameShouldBe(name)
	})
	sc.Step(`^the updated account phone should be "([^"]*)"$`, func(phone string) error {
		return s.theUpdatedAccountPhoneShouldBe(phone)
	})
	sc.Step(`^the organizer profile name should be "([^"]*)"$`, func(name string) error {
		return s.theOrganizerProfileNameShouldBe(name)
	})
	sc.Step(`^the organizer contact email should be "([^"]*)"$`, func(email string) error {
		return s.theOrganizerContactEmailShouldBe(email)
	})
	sc.Step(`^the sponsor profile name should be "([^"]*)"$`, func(name string) error {
		return s.theSponsorProfileNameShouldBe(name)
	})
	sc.Step(`^the sponsor contact email should be "([^"]*)"$`, func(email string) error {
		return s.theSponsorContactEmailShouldBe(email)
	})
}
