package announcements

import (
	"bytes"
	"context"
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

// The default announcement is Thai on purpose: the length limits count
// characters, and a Thai letter takes three bytes, so this would fail if the
// limits ever counted bytes instead.
const (
	defaultTitle = "ย้ายสถานที่แข่งขัน"
	defaultBody  = "การแข่งขันจะจัดที่ห้อง 301 แทนห้อง 205 ขออภัยในความไม่สะดวก"
)

// announcementItem is one entry of the public list, as the API returns it.
type announcementItem struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournamentId"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	AuthorName   string    `json:"authorName"`
	CreatedAt    time.Time `json:"createdAt"`
}

type announcementScenarioContext struct {
	client              *apptest.Client
	owner               *apptest.TestUser // tournament owner
	poster              *apptest.TestUser // who sends the post request
	tournament          *models.Tournament
	title               string // what was last sent
	body                string
	expectedBlockStatus int   // status a blocked post must answer with
	countBefore         int64 // announcements stored before the last post
	resp                *apptest.Response
	items               []announcementItem // the last list read
}

func (s *announcementScenarioContext) theEventoryAPIServiceIsRunning() error {
	if app == nil || app.TournamentServer == nil {
		return fmt.Errorf("application server is not running")
	}
	return nil
}

func (s *announcementScenarioContext) persistActiveAccount(user *apptest.TestUser) error {
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

// createTournament seeds a tournament owned by a new account, which becomes
// the default poster. A tournament that is not published is a draft.
func (s *announcementScenarioContext) createTournament(published bool) error {
	owner := apptest.NewTestUser()
	if err := s.persistActiveAccount(owner); err != nil {
		return err
	}

	status := models.TournamentStatusRegistrationOpen
	if !published {
		status = models.TournamentStatusDraft
	}

	now := time.Now().UTC()
	s.tournament = &models.Tournament{
		ID: uuid.New(),
		Staff: []models.TournamentStaff{
			{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner},
		},
		Name:                 "BDD Announcement Tournament " + uuid.NewString()[:8],
		Description:          "Tournament created for announcement acceptance tests.",
		Game:                 "Valorant",
		Location:             "Online",
		StartsAt:             now.Add(72 * time.Hour),
		EndsAt:               now.Add(80 * time.Hour),
		RegistrationDeadline: now.Add(48 * time.Hour),
		Currency:             "THB",
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          1,
		MaxTeamSize:          3,
		Capacity:             8,
		Status:               status,
		Published:            published,
	}
	if err := app.DB.Create(s.tournament).Error; err != nil {
		return fmt.Errorf("create tournament: %w", err)
	}
	s.owner = owner
	s.poster = owner
	return nil
}

// addStaff adds a new account to the tournament's staff with the given role.
func (s *announcementScenarioContext) addStaff(role models.TournamentStaffRole) (*apptest.TestUser, error) {
	member := apptest.NewTestUser()
	if err := s.persistActiveAccount(member); err != nil {
		return nil, err
	}
	staff := models.TournamentStaff{
		TournamentID: s.tournament.ID,
		AccountID:    member.ID,
		Role:         role,
	}
	if err := app.DB.Create(&staff).Error; err != nil {
		return nil, fmt.Errorf("add %s to tournament staff: %w", role, err)
	}
	return member, nil
}

func (s *announcementScenarioContext) announcementsPath() string {
	return fmt.Sprintf("/tournaments/%s/announcements", s.tournament.ID)
}

// announcementCount reads how many announcements are stored for the
// tournament straight from the database.
func (s *announcementScenarioContext) announcementCount() (int64, error) {
	var count int64
	err := app.DB.Model(&models.TournamentAnnouncement{}).
		Where("tournament_id = ?", s.tournament.ID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count announcements: %w", err)
	}
	return count, nil
}

// post sends an announcement as the current poster and snapshots the stored
// count beforehand so the Then steps can compare against it.
func (s *announcementScenarioContext) post(title, body string) error {
	var err error
	if s.countBefore, err = s.announcementCount(); err != nil {
		return err
	}
	s.title, s.body = title, body
	payload := map[string]any{"title": title, "body": body}
	s.resp, err = s.client.Do(http.MethodPost, s.announcementsPath(), payload, s.poster.AuthHeaders())
	return err
}

// fetchList reads the public list without signing in, and keeps the items when
// the list is returned.
func (s *announcementScenarioContext) fetchList() error {
	resp, err := s.client.Do(http.MethodGet, s.announcementsPath(), nil, nil)
	if err != nil {
		return err
	}
	s.resp = resp
	s.items = nil
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var list struct {
		Items []announcementItem `json:"items"`
	}
	if err := resp.JSON(&list); err != nil {
		return fmt.Errorf("decode announcement list: %w", err)
	}
	s.items = list.Items
	return nil
}

// ---- Given steps ----

func (s *announcementScenarioContext) anOrganizerManagesAPublishedTournament() error {
	return s.createTournament(true)
}

func (s *announcementScenarioContext) anOrganizerManagesATournamentThatIsNotPublishedYet() error {
	return s.createTournament(false)
}

func (s *announcementScenarioContext) anAdminHasBeenAddedToThatTournamentsStaff() error {
	admin, err := s.addStaff(models.TournamentStaffRoleAdmin)
	if err != nil {
		return err
	}
	s.poster = admin
	return nil
}

// expectPosted fails unless the last post was accepted.
func (s *announcementScenarioContext) expectPosted() error {
	if s.resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("expected the announcement to be posted (201), got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *announcementScenarioContext) theOrganizerHasPostedAnAnnouncement() error {
	if err := s.post(defaultTitle, defaultBody); err != nil {
		return err
	}
	return s.expectPosted()
}

func (s *announcementScenarioContext) theOrganizerHasPostedTwoAnnouncementsOneAfterTheOther() error {
	for _, title := range []string{"First announcement", "Second announcement"} {
		if err := s.post(title, "Details of the "+strings.ToLower(title)); err != nil {
			return err
		}
		if err := s.expectPosted(); err != nil {
			return err
		}
		// Keep the two timestamps clearly apart so the order is not a coin toss.
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

// aUserPreparesAnAnnouncementWhere sets up one way a post can fail. The
// validation failures answer 422; a poster without the right to post is told
// the tournament was not found (404), never "forbidden".
func (s *announcementScenarioContext) aUserPreparesAnAnnouncementWhere(situation string) error {
	if err := s.createTournament(true); err != nil {
		return err
	}
	s.title, s.body = defaultTitle, defaultBody
	s.expectedBlockStatus = http.StatusUnprocessableEntity

	switch situation {
	case "the title is empty":
		s.title = ""
	case "the title contains only spaces":
		s.title = "   "
	case "the message is empty":
		s.body = ""
	case "the title is longer than 150 characters":
		s.title = strings.Repeat("a", 151)
	case "the message is longer than 2000 characters":
		s.body = strings.Repeat("a", 2001)
	case "the user does not manage the tournament":
		s.expectedBlockStatus = http.StatusNotFound
		stranger := apptest.NewTestUser()
		if err := s.persistActiveAccount(stranger); err != nil {
			return err
		}
		s.poster = stranger
	case "the user is only a referee of the tournament":
		s.expectedBlockStatus = http.StatusNotFound
		referee, err := s.addStaff(models.TournamentStaffRoleReferee)
		if err != nil {
			return err
		}
		s.poster = referee
	default:
		return fmt.Errorf("unknown announcement situation %q", situation)
	}
	return nil
}

// ---- When steps ----

func (s *announcementScenarioContext) theOrganizerPostsAnAnnouncementWithATitleAndAMessage() error {
	return s.post(defaultTitle, defaultBody)
}

func (s *announcementScenarioContext) theAdminPostsAnAnnouncementWithATitleAndAMessage() error {
	return s.post(defaultTitle, defaultBody)
}

func (s *announcementScenarioContext) theUserPostsTheAnnouncement() error {
	return s.post(s.title, s.body)
}

func (s *announcementScenarioContext) aVisitorWhoIsNotSignedInOpensTheAnnouncementList() error {
	return s.fetchList()
}

// ---- Then steps ----

func (s *announcementScenarioContext) theAnnouncementShouldAppearFirstInTheAnnouncementList() error {
	if err := s.expectPosted(); err != nil {
		return err
	}
	if err := s.fetchList(); err != nil {
		return err
	}
	if s.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected the list to load (200), got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	if len(s.items) == 0 {
		return fmt.Errorf("expected the posted announcement in the list, got an empty list")
	}
	first := s.items[0]
	if first.Title != s.title || first.Body != s.body {
		return fmt.Errorf("expected first announcement %q / %q, got %q / %q", s.title, s.body, first.Title, first.Body)
	}
	return nil
}

func (s *announcementScenarioContext) theAnnouncementShouldShowWhoPostedItAndWhen() error {
	if len(s.items) == 0 {
		return fmt.Errorf("no announcement was read from the list")
	}
	first := s.items[0]
	if first.AuthorName != s.poster.DisplayName {
		return fmt.Errorf("expected author %q, got %q", s.poster.DisplayName, first.AuthorName)
	}
	if first.TournamentID != s.tournament.ID.String() {
		return fmt.Errorf("expected tournament %s, got %s", s.tournament.ID, first.TournamentID)
	}
	if first.CreatedAt.IsZero() || time.Since(first.CreatedAt) > time.Minute {
		return fmt.Errorf("expected a posting time from the last minute, got %s", first.CreatedAt)
	}
	// Only the name is public: the poster's internal account ID must not leak.
	if bytes.Contains(s.resp.Body, []byte(s.poster.ID.String())) {
		return fmt.Errorf("the public list exposes the poster's account ID: %s", s.resp.Body)
	}
	return nil
}

func (s *announcementScenarioContext) theNewestAnnouncementShouldBeListedFirst() error {
	if s.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected the list to load (200), got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	if len(s.items) != 2 {
		return fmt.Errorf("expected 2 announcements, got %d: %s", len(s.items), s.resp.Body)
	}
	if s.items[0].Title != "Second announcement" || s.items[1].Title != "First announcement" {
		return fmt.Errorf("expected newest first, got %q then %q", s.items[0].Title, s.items[1].Title)
	}
	return nil
}

// expectErrorBody checks the answer is a real API error. A bare router 404
// ("404 page not found") means the endpoint does not exist; a genuine refusal
// is a JSON error body carrying its own status.
func (s *announcementScenarioContext) expectErrorBody() error {
	var problem struct {
		Status int `json:"status"`
	}
	if err := s.resp.JSON(&problem); err != nil || problem.Status != s.resp.StatusCode {
		return fmt.Errorf("expected a JSON error body for status %d, got: %s", s.resp.StatusCode, s.resp.Body)
	}
	return nil
}

func (s *announcementScenarioContext) theAnnouncementShouldBeBlocked() error {
	if s.resp.StatusCode != s.expectedBlockStatus {
		return fmt.Errorf("expected the post to be blocked with %d, got %d: %s", s.expectedBlockStatus, s.resp.StatusCode, s.resp.Body)
	}
	return s.expectErrorBody()
}

func (s *announcementScenarioContext) theAnnouncementListShouldStayUnchanged() error {
	count, err := s.announcementCount()
	if err != nil {
		return err
	}
	if count != s.countBefore {
		return fmt.Errorf("expected %d stored announcements, got %d", s.countBefore, count)
	}
	return nil
}

func (s *announcementScenarioContext) theAnnouncementListShouldNotBeAvailable() error {
	if s.resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("expected the list to be unavailable (404), got %d: %s", s.resp.StatusCode, s.resp.Body)
	}
	if err := s.expectErrorBody(); err != nil {
		return err
	}
	if bytes.Contains(s.resp.Body, []byte(s.title)) {
		return fmt.Errorf("the refusal leaks announcement content: %s", s.resp.Body)
	}
	// The announcement really exists, so the refusal is what hides it.
	count, err := s.announcementCount()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("expected the hidden announcement to be stored (1), got %d", count)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	s := &announcementScenarioContext{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*s = announcementScenarioContext{client: apptest.NewClient(app.TournamentBaseURL())}
		// CASCADE also clears the announcements table, which references the
		// tournaments truncated here.
		cleanup := "TRUNCATE tournament_staff, tournament_team_members, tournament_fundings, tournament_status_changes, tournament_teams, tournaments, sponsor_profiles, organizer_profiles, accounts CASCADE"
		if err := app.DB.Exec(cleanup).Error; err != nil {
			return ctx, fmt.Errorf("clean scenario database: %w", err)
		}
		return ctx, nil
	})

	// Background
	sc.Step(`^the Eventory API service is running$`, s.theEventoryAPIServiceIsRunning)

	// Given
	sc.Step(`^an organizer manages a published tournament$`, s.anOrganizerManagesAPublishedTournament)
	sc.Step(`^an organizer manages a tournament that is not published yet$`, s.anOrganizerManagesATournamentThatIsNotPublishedYet)
	sc.Step(`^an admin has been added to that tournament's staff$`, s.anAdminHasBeenAddedToThatTournamentsStaff)
	sc.Step(`^the organizer has posted an announcement$`, s.theOrganizerHasPostedAnAnnouncement)
	sc.Step(`^the organizer has posted two announcements, one after the other$`, s.theOrganizerHasPostedTwoAnnouncementsOneAfterTheOther)
	sc.Step(`^a user prepares an announcement where (.+)$`, s.aUserPreparesAnAnnouncementWhere)

	// When
	sc.Step(`^the organizer posts an announcement with a title and a message$`, s.theOrganizerPostsAnAnnouncementWithATitleAndAMessage)
	sc.Step(`^the admin posts an announcement with a title and a message$`, s.theAdminPostsAnAnnouncementWithATitleAndAMessage)
	sc.Step(`^the user posts the announcement$`, s.theUserPostsTheAnnouncement)
	sc.Step(`^a visitor who is not signed in opens the tournament's announcement list$`, s.aVisitorWhoIsNotSignedInOpensTheAnnouncementList)

	// Then
	sc.Step(`^the announcement should appear first in the tournament's announcement list$`, s.theAnnouncementShouldAppearFirstInTheAnnouncementList)
	sc.Step(`^the announcement should show who posted it and when$`, s.theAnnouncementShouldShowWhoPostedItAndWhen)
	sc.Step(`^the newest announcement should be listed first$`, s.theNewestAnnouncementShouldBeListedFirst)
	sc.Step(`^the announcement should be blocked$`, s.theAnnouncementShouldBeBlocked)
	sc.Step(`^the announcement list should stay unchanged$`, s.theAnnouncementListShouldStayUnchanged)
	sc.Step(`^the announcement list should not be available$`, s.theAnnouncementListShouldNotBeAvailable)
}
