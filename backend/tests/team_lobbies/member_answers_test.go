package team_lobbies

import (
	"context"
	handlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"net/http"
	"testing"
	"time"
)

// US4-8: exercise Mongo form answers, PostgreSQL team state, and the
// HTTP create -> join -> edit -> submit -> whole-team rejection workflow.
func TestMemberRegistrationAnswers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{Image: "mongo:7.0", ExposedPorts: []string{"27017/tcp"}, WaitingFor: wait.ForListeningPort("27017/tcp")}, Started: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	mongoClient, err := mongo.Connect(options.Client().ApplyURI("mongodb://" + endpoint))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mongoClient.Disconnect(context.Background()) })
	app = apptest.NewApp(t, pg.DSN, mongoClient.Database("member_answers_test"))
	client := apptest.NewClient(app.TournamentBaseURL())
	s := &teamLobbyScenarioContext{client: client}
	if err := s.createPublishedTournament(); err != nil {
		t.Fatal(err)
	}
	captain, member, outsider := apptest.NewTestUser(), apptest.NewTestUser(), apptest.NewTestUser()
	for _, user := range []*apptest.TestUser{captain, member, outsider} {
		if err := s.persistActiveAccount(user); err != nil {
			t.Fatal(err)
		}
	}
	owner := apptest.NewTestUser()
	owner.ID = s.tournament.Staff[0].AccountID
	owner.Token = apptest.GenerateTestJWT(owner.ID, owner.Email)
	request := func(method, path string, body any, user *apptest.TestUser, status int) *apptest.Response {
		t.Helper()
		resp, err := client.Do(method, path, body, user.AuthHeaders())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != status {
			t.Fatalf("%s %s: %d want %d: %s", method, path, resp.StatusCode, status, resp.Body)
		}
		return resp
	}
	tourPath := "/tournaments/" + s.tournament.ID.String()
	fields := []models.FormField{{ID: "ign", Type: models.FieldTypeShortText, Label: "In-game name", Required: true}}
	saveForm := func() {
		request(http.MethodPut, tourPath+"/form", map[string]any{"title": "Team questions", "fields": fields}, owner, 200)
	}
	saveForm()
	answers := func(value any) map[string]any {
		return map[string]any{"answers": []models.FieldAnswer{{FieldID: "ign", Value: value}}}
	}
	request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Incomplete"}, captain, 400)
	request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Invalid file", "answers": []models.FieldAnswer{{FieldID: "ign", Value: map[string]any{"filename": "bad.exe", "type": "application/octet-stream"}}}}, captain, 400)
	var count int64
	app.DB.Model(&models.TournamentTeam{}).Where("tournament_id = ?", s.tournament.ID).Count(&count)
	if count != 0 {
		t.Fatal("invalid answers created a team")
	}
	created := request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Ready team", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "Captain"}}}, captain, 201)
	var lobby handlers.TeamLobbyResponse
	if err := created.JSON(&lobby); err != nil {
		t.Fatal(err)
	}
	if len(lobby.Members) != 1 || !lobby.Members[0].RegistrationComplete {
		t.Fatal("captain readiness missing")
	}
	joinPath := "/lobbies/" + *lobby.InviteCode + "/join"
	request(http.MethodPost, joinPath, answers(nil), member, 400)
	request(http.MethodPost, joinPath, answers(map[string]any{"filename": "bad.exe"}), member, 400)
	request(http.MethodPost, joinPath, answers("Member"), member, 200)
	registrationPath := "/lobbies/" + lobby.ID.String() + "/registration"
	request(http.MethodPut, registrationPath, answers("Intruder"), outsider, 403)
	request(http.MethodPut, registrationPath, answers(nil), member, 400)
	request(http.MethodPut, registrationPath, answers("Member revised"), member, 200)
	request(http.MethodPost, tourPath+"/form/responses", answers("standalone"), member, 405)
	// A new required question must block both the UI readiness response and API.
	fields = append(fields, models.FormField{ID: "age", Type: models.FieldTypeNumber, Label: "Age", Required: true})
	saveForm()
	var latest handlers.TeamLobbyResponse
	response := request(http.MethodGet, "/lobbies/"+*lobby.InviteCode, nil, captain, 200)
	if err := response.JSON(&latest); err != nil {
		t.Fatal(err)
	}
	for _, entry := range latest.Members {
		if entry.RegistrationComplete {
			t.Fatal("changed form left member ready")
		}
	}
	request(http.MethodPost, "/lobbies/"+lobby.ID.String()+"/lock", nil, captain, 409)
	var persisted models.TournamentTeam
	if err := app.DB.First(&persisted, "id = ?", lobby.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Status != models.TournamentTeamStatusForming || persisted.PaymentID != nil {
		t.Fatal("incomplete team was submitted")
	}
	for _, user := range []*apptest.TestUser{captain, member} {
		request(http.MethodPut, registrationPath, map[string]any{"answers": []models.FieldAnswer{{FieldID: "ign", Value: user.Email}, {FieldID: "age", Value: 21}}}, user, 200)
	}
	request(http.MethodPost, "/lobbies/"+lobby.ID.String()+"/lock", nil, captain, 200)
	request(http.MethodPut, registrationPath, answers("Changed after locking"), member, 409)
	list := request(http.MethodGet, tourPath+"/form/responses", nil, owner, 200)
	var responses struct {
		Items []handlers.FormResponseItem `json:"items"`
	}
	if err := list.JSON(&responses); err != nil {
		t.Fatal(err)
	}
	if len(responses.Items) != 2 {
		t.Fatalf("organizer received %d responses", len(responses.Items))
	}
	for _, entry := range responses.Items {
		if len(entry.Answers) != 2 {
			t.Fatal("organizer did not receive each member's answers")
		}
	}
	request(http.MethodGet, tourPath+"/form/responses", nil, member, 403)
	request(http.MethodPost, tourPath+"/entries/"+lobby.ID.String()+"/review", map[string]any{"decision": "REJECT", "reason": "One member does not meet the requirements"}, owner, 200)
	if err := app.DB.Preload("Members").First(&persisted, "id = ?", lobby.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Status != models.TournamentTeamStatusRejected || len(persisted.Members) != 2 {
		t.Fatal("reject must apply to the whole team")
	}
	// Reapplication updates the existing (tournament, account) Mongo response.
	reapplied := request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Second application", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "New captain"}, {FieldID: "age", Value: 22}}}, captain, 201)
	var newLobby handlers.TeamLobbyResponse
	if err := reapplied.JSON(&newLobby); err != nil {
		t.Fatal(err)
	}
	request(http.MethodPost, "/lobbies/"+*newLobby.InviteCode+"/join", map[string]any{"answers": []models.FieldAnswer{{FieldID: "ign", Value: "New member"}, {FieldID: "age", Value: 23}}}, member, 200)
	var current models.FormResponse
	collection := mongoClient.Database("member_answers_test").Collection("tournament_form_responses")
	filter := bson.M{"tournament_id": s.tournament.ID, "account_id": member.ID}
	if err := collection.FindOne(ctx, filter).Decode(&current); err != nil {
		t.Fatal(err)
	}
	if len(current.Answers) != 2 || current.Answers[0].Value != "New member" {
		t.Fatal("reapplication did not update the member's Mongo response")
	}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil || total != 1 {
		t.Fatalf("duplicate member responses: count=%d err=%v", total, err)
	}
}
