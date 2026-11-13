package team_lobbies

import (
	"context"
	"github.com/google/uuid"
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
// HTTP create -> join -> reject edits -> submit -> whole-team rejection workflow.
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
	collection := mongoClient.Database("member_answers_test").Collection("tournament_form_responses")
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
		if err := app.DB.Model(s.tournament).Update("status", models.TournamentStatusDraft).Error; err != nil {
			t.Fatal(err)
		}
		request(http.MethodPut, tourPath+"/form", map[string]any{"title": "Team questions", "fields": fields}, owner, 200)
		if err := app.DB.Model(s.tournament).Update("status", models.TournamentStatusRegistration).Error; err != nil {
			t.Fatal(err)
		}
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
	// The answer store still enforces one document per tournament/account.
	if _, err := collection.InsertOne(ctx, bson.M{"tournament_id": s.tournament.ID, "account_id": captain.ID}); !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("composite answer key must be unique: %v", err)
	}
	joinPath := "/lobbies/" + *lobby.InviteCode + "/join"
	request(http.MethodPost, joinPath, answers(nil), member, 400)
	request(http.MethodPost, joinPath, answers(map[string]any{"filename": "bad.exe"}), member, 400)
	request(http.MethodPost, joinPath, answers("Member"), member, 200)
	registrationPath := "/lobbies/" + lobby.ID.String() + "/registration"
	request(http.MethodPut, registrationPath, answers("Intruder"), outsider, 403)
	request(http.MethodPut, registrationPath, answers(nil), member, 400)
	request(http.MethodPut, registrationPath, answers("Member revised"), member, 409)
	request(http.MethodPut, registrationPath, answers("Member"), member, 409)
	request(http.MethodPost, tourPath+"/form/responses", answers("standalone"), member, 405)
	// Rejected edits must leave the original responses and readiness intact.
	var latest handlers.TeamLobbyResponse
	response := request(http.MethodGet, "/lobbies/"+*lobby.InviteCode, nil, captain, 200)
	if err := response.JSON(&latest); err != nil {
		t.Fatal(err)
	}
	for _, entry := range latest.Members {
		if !entry.RegistrationComplete {
			t.Fatal("rejected edit changed member readiness")
		}
	}
	var persisted models.TournamentTeam
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
		if len(entry.Answers) != 1 {
			t.Fatal("organizer did not receive each member's answers")
		}
		if entry.AccountID == member.ID && entry.Answers[0].Value != "Member" {
			t.Fatal("member's submitted answer was overwritten")
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
	// Rejection is final for this team and its members; they cannot re-apply.
	request(http.MethodPost, tourPath+"/entries/"+lobby.ID.String()+"/review", map[string]any{"decision": "ACCEPT"}, owner, 409)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, captain, 200)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, member, 200)
	request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Missing new answers"}, captain, 400)
	request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Second application", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "New captain"}}}, captain, 409)
	request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Member new team", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "New member"}}}, member, 409)

	// Fresh users exercise removing/rejoining and disbanding while FORMING.
	captain2, member2 := apptest.NewTestUser(), apptest.NewTestUser()
	for _, user := range []*apptest.TestUser{captain2, member2} {
		if err := s.persistActiveAccount(user); err != nil {
			t.Fatal(err)
		}
	}
	second := request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Second application", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "New captain"}}}, captain2, 201)
	var newLobby handlers.TeamLobbyResponse
	if err := second.JSON(&newLobby); err != nil {
		t.Fatal(err)
	}
	joined := request(http.MethodPost, "/lobbies/"+*newLobby.InviteCode+"/join", answers("New member"), member2, 200)
	if err := joined.JSON(&newLobby); err != nil {
		t.Fatal(err)
	}
	assertCurrent := func(user *apptest.TestUser, expected string) {
		t.Helper()
		var current handlers.FormResponseItem
		if err := request(http.MethodGet, tourPath+"/form/responses/me", nil, user, 200).JSON(&current); err != nil {
			t.Fatal(err)
		}
		if current.Answers[0].Value != expected {
			t.Fatalf("wrong current response: %+v", current)
		}
	}
	assertCurrent(captain2, "New captain")
	assertCurrent(member2, "New member")
	request(http.MethodPut, "/lobbies/"+newLobby.ID.String()+"/registration", answers("Edited"), member2, 409)
	var history []models.FormResponse
	filter := bson.M{"tournament_id": s.tournament.ID, "account_id": member2.ID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	if err := cursor.All(ctx, &history); err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Answers[0].Value != "New member" {
		t.Fatalf("expected one overwritten answer set, got %+v", history)
	}

	// Removing and rejoining the SAME team is also a new membership/application.
	var memberID uuid.UUID
	for _, entry := range newLobby.Members {
		if entry.AccountID == member2.ID {
			memberID = entry.ID
		}
	}
	request(http.MethodDelete, "/lobbies/"+newLobby.ID.String()+"/members/"+memberID.String(), nil, captain2, 200)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, member2, 404)
	request(http.MethodPost, "/lobbies/"+*newLobby.InviteCode+"/join", answers(nil), member2, 400)
	request(http.MethodPost, "/lobbies/"+*newLobby.InviteCode+"/join", answers("Rejoined member"), member2, 200)
	assertCurrent(member2, "Rejoined member")
	request(http.MethodPut, "/lobbies/"+newLobby.ID.String()+"/registration", answers("Edited again"), member2, 409)

	// Disbanding releases all members; registering again overwrites their answers.
	request(http.MethodDelete, "/lobbies/"+newLobby.ID.String(), nil, captain2, 204)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, captain2, 404)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, member2, 404)
	third := request(http.MethodPost, tourPath+"/lobbies", map[string]any{"name": "Third application", "answers": []models.FieldAnswer{{FieldID: "ign", Value: "Third captain"}}}, captain2, 201)
	var thirdLobby handlers.TeamLobbyResponse
	if err := third.JSON(&thirdLobby); err != nil {
		t.Fatal(err)
	}
	request(http.MethodPost, "/lobbies/"+*thirdLobby.InviteCode+"/join", answers("Other team member"), member2, 200)
	assertCurrent(captain2, "Third captain")
	assertCurrent(member2, "Other team member")
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil || total != 1 {
		t.Fatalf("expected one response per tournament/account, count=%d err=%v", total, err)
	}
	// Solo rejection is also permanent.
	request(http.MethodDelete, "/lobbies/"+thirdLobby.ID.String(), nil, captain2, 204)
	if err := app.DB.Model(&models.Tournament{}).Where("id = ?", s.tournament.ID).Updates(map[string]any{"registration_mode": "SOLO", "min_team_size": 1, "max_team_size": 1}).Error; err != nil {
		t.Fatal(err)
	}
	solo := request(http.MethodPost, tourPath+"/solo-register", answers("Solo first"), captain2, 201)
	var soloLobby handlers.TeamLobbyResponse
	if err := solo.JSON(&soloLobby); err != nil {
		t.Fatal(err)
	}
	assertCurrent(captain2, "Solo first")
	request(http.MethodPost, tourPath+"/solo-register", answers("Solo edit"), captain2, 409)
	request(http.MethodPost, tourPath+"/entries/"+soloLobby.ID.String()+"/review", map[string]any{"decision": "REJECT", "reason": "Try a new application"}, owner, 200)
	request(http.MethodGet, tourPath+"/form/responses/me", nil, captain2, 200)
	request(http.MethodPost, tourPath+"/solo-register", answers(nil), captain2, 400)
	request(http.MethodPost, tourPath+"/solo-register", answers("Solo second"), captain2, 409)
}
