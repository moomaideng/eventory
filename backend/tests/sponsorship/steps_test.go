package sponsorship

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

type tierScenario struct {
	app        *apptest.App
	client     *apptest.Client
	resp       *apptest.Response
	owner      *apptest.TestUser
	other      *apptest.TestUser
	tournament models.Tournament
	tierIDs    map[string]uuid.UUID
}

type tierItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (s *tierScenario) organizer(name string) (*apptest.TestUser, error) {
	u := apptest.NewTestUser()
	account := accountmodels.Account{ID: u.ID, Email: u.Email, Handle: u.Handle, DisplayName: u.DisplayName, Status: accountmodels.AccountStatusActive}
	if err := s.app.DB.Create(&account).Error; err != nil {
		return nil, err
	}
	profile := accountmodels.OrganizerProfile{AccountID: u.ID, OrganizerName: name, OrganizerEmail: u.Email}
	return u, s.app.DB.Create(&profile).Error
}

func (s *tierScenario) publishedTournament() error {
	owner, err := s.organizer("Tier Host")
	if err != nil {
		return err
	}
	s.owner = owner
	now := time.Now().UTC()
	s.tournament = models.Tournament{
		ID: uuid.New(), Name: "Sponsor Cup " + apptest.UniqueSuffix(), Description: "Sponsorship tier test", Game: "Chess", Location: "Online",
		StartsAt: now.Add(72 * time.Hour), EndsAt: now.Add(80 * time.Hour), RegistrationDeadline: now.Add(48 * time.Hour),
		Currency: "THB", RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 1, Capacity: 8,
		Status: models.TournamentStatusRegistration,
		Staff:  []models.TournamentStaff{{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner}},
	}
	return s.app.DB.Create(&s.tournament).Error
}

func (s *tierScenario) tiersPath() string {
	return "/tournaments/" + s.tournament.ID.String() + "/sponsorship-tiers"
}

func (s *tierScenario) do(method, path string, body any, user *apptest.TestUser) error {
	headers := map[string]string{}
	if user != nil {
		headers = user.AuthHeaders()
	}
	resp, err := s.client.Do(method, path, body, headers)
	s.resp = resp
	return err
}

func splitList(raw string) []string {
	items := []string{}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ';' || r == ',' }) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func (s *tierScenario) create(user *apptest.TestUser, name string, price, slots int, perks string) error {
	body := map[string]any{"name": name, "price": price, "slots": slots, "perks": splitList(perks)}
	if err := s.do(http.MethodPost, s.tiersPath(), body, user); err != nil {
		return err
	}
	var created tierItem
	if s.resp.StatusCode == http.StatusCreated && s.resp.JSON(&created) == nil {
		s.tierIDs[created.Name] = created.ID
	}
	return nil
}

func (s *tierScenario) draftTournament() error {
	if err := s.publishedTournament(); err != nil {
		return err
	}
	s.tournament.Status = models.TournamentStatusDraft
	return s.app.DB.Model(&s.tournament).Update("status", models.TournamentStatusDraft).Error
}

func (s *tierScenario) given(name string, price int) error {
	prevStatus := s.tournament.Status
	if prevStatus != models.TournamentStatusDraft {
		if err := s.app.DB.Model(&s.tournament).Update("status", models.TournamentStatusDraft).Error; err != nil {
			return err
		}
		defer func() {
			_ = s.app.DB.Model(&s.tournament).Update("status", prevStatus).Error
		}()
	}
	if err := s.create(s.owner, name, price, 2, "Logo on stream"); err != nil {
		return err
	}
	if s.resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("setup tier %q: status %d: %s", name, s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *tierScenario) rename(from, to string) error {
	body := map[string]any{"name": to, "price": 10000, "slots": 2, "perks": []string{"Logo on stream"}}
	s.tierIDs[to] = s.tierIDs[from]
	return s.do(http.MethodPut, s.tiersPath()+"/"+s.tierIDs[from].String(), body, s.owner)
}

func (s *tierScenario) remove(user *apptest.TestUser, name string) error {
	return s.do(http.MethodDelete, s.tiersPath()+"/"+s.tierIDs[name].String(), nil, user)
}

func (s *tierScenario) visibleTiers(expected string) error {
	prevStatus := s.tournament.Status
	if prevStatus == models.TournamentStatusDraft {
		if err := s.app.DB.Model(&s.tournament).Update("status", models.TournamentStatusRegistration).Error; err != nil {
			return err
		}
		defer func() {
			_ = s.app.DB.Model(&s.tournament).Update("status", prevStatus).Error
		}()
	}
	resp, err := s.client.Do(http.MethodGet, s.tiersPath(), nil, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("list tiers: status %d: %s", resp.StatusCode, resp.Body)
	}
	var list struct {
		Items []tierItem `json:"items"`
	}
	if err := resp.JSON(&list); err != nil {
		return err
	}
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	if got, want := strings.Join(names, ", "), strings.Join(splitList(expected), ", "); got != want {
		return fmt.Errorf("visible tiers = %q, want %q", got, want)
	}
	return nil
}

func (s *tierScenario) managedTiers(expected string) error {
	var list struct {
		Items []tierItem `json:"items"`
	}
	if err := s.resp.JSON(&list); err != nil {
		return err
	}
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	if got, want := strings.Join(names, ", "), strings.Join(splitList(expected), ", "); got != want {
		return fmt.Errorf("managed tiers = %q, want %q", got, want)
	}
	return nil
}

func (s *tierScenario) status(code int) error {
	if s.resp == nil {
		return fmt.Errorf("no response recorded")
	}
	if s.resp.StatusCode != code {
		return fmt.Errorf("status = %d, want %d: %s", s.resp.StatusCode, code, s.resp.Body)
	}
	return nil
}

func initializeScenario(sc *godog.ScenarioContext, app *apptest.App) {
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}

	sc.Step(`^an organizer owns a published tournament open for sponsorship$`, s.draftTournament)
	sc.Step(`^another organizer exists$`, func() error {
		other, err := s.organizer("Rival Host")
		s.other = other
		return err
	})
	sc.Step(`^the organizer has a tier named "([^"]*)"$`, func(name string) error { return s.given(name, 10000) })
	sc.Step(`^the organizer has tiers "([^"]*)" costing (\d+) and "([^"]*)" costing (\d+)$`, func(a string, priceA int, b string, priceB int) error {
		if err := s.given(a, priceA); err != nil {
			return err
		}
		return s.given(b, priceB)
	})
	sc.Step(`^the tournament status is "([^"]*)"$`, func(status string) error {
		s.tournament.Status = models.TournamentStatus(status)
		return s.app.DB.Model(&s.tournament).Update("status", status).Error
	})
	sc.Step(`^the tournament is unpublished$`, func() error {
		s.tournament.Status = models.TournamentStatusDraft
		return s.app.DB.Model(&s.tournament).Update("status", models.TournamentStatusDraft).Error
	})
	sc.Step(`^the (organizer|other organizer) creates a tier named "([^"]*)" costing (\d+) with (\d+) slots and perks "([^"]*)"$`, func(who, name string, price, slots int, perks string) error {
		user := s.owner
		if who == "other organizer" {
			user = s.other
		}
		return s.create(user, name, price, slots, perks)
	})
	sc.Step(`^the organizer renames tier "([^"]*)" to "([^"]*)"$`, s.rename)
	sc.Step(`^the organizer deletes tier "([^"]*)"$`, func(name string) error { return s.remove(s.owner, name) })
	sc.Step(`^the other organizer deletes tier "([^"]*)"$`, func(name string) error { return s.remove(s.other, name) })
	sc.Step(`^a visitor lists the tournament's tiers$`, func() error { return s.do(http.MethodGet, s.tiersPath(), nil, nil) })
	sc.Step(`^the (organizer|other organizer) lists tiers to manage$`, func(who string) error {
		user := s.owner
		if who == "other organizer" {
			user = s.other
		}
		return s.do(http.MethodGet, s.tiersPath()+"/manage", nil, user)
	})
	sc.Step(`^the managed tiers should be exactly "([^"]*)"$`, s.managedTiers)
	sc.Step(`^the response status should be (\d+)$`, s.status)
	sc.Step(`^sponsors should see exactly "([^"]*)" on the tournament$`, s.visibleTiers)
}
