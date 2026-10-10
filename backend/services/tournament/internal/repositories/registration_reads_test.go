package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRosterQueriesExcludePrivateRegistration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := postgres.Run(ctx, "postgres:18-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	tournament := models.Tournament{ID: uuid.New(), OrganizerID: uuid.New(), Name: "Private upload test", Game: "Test", StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour), RegistrationDeadline: time.Now(), Capacity: 16, Published: true}
	member := models.TournamentTeamMember{
		ID: uuid.New(), AccountID: uuid.New(), Role: models.TournamentTeamMemberRoleCaptain, FormVersion: 1,
		ConsentNotice: "Private consent", RegistrationQuestions: []models.RegistrationQuestion{{ID: "proof", Label: "Proof", Type: "FILE"}},
		RegistrationAnswers: []models.RegistrationAnswer{{QuestionID: "proof", File: &models.RegistrationFile{Name: "proof.pdf", Data: []byte("%PDF-1.4\nPrivate proof")}}},
	}
	tournament.Teams = []models.TournamentTeam{{ID: uuid.New(), Name: "Team", InviteCode: "ABCDEF", Status: models.TournamentTeamStatusAccepted, Members: []models.TournamentTeamMember{member}}}
	if err := db.Create(&tournament).Error; err != nil {
		t.Fatal(err)
	}
	checkRoster := func(tournament *models.Tournament) {
		t.Helper()
		if len(tournament.Teams) != 1 || len(tournament.Teams[0].Members) != 1 {
			t.Fatal("query did not return the roster")
		}
		got := tournament.Teams[0].Members[0]
		if got.ID != member.ID || got.FormVersion != 1 || len(got.RegistrationAnswers) != 0 || len(got.RegistrationQuestions) != 0 || got.ConsentNotice != "" {
			t.Fatal("roster query loaded private submission fields or lost member identity")
		}
	}
	dashboard := repositories.NewOrganizerDashboardRepository(db)
	owned, err := dashboard.GetOwned(ctx, tournament.OrganizerID, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	checkRoster(owned)
	items, _, err := dashboard.ListOwned(ctx, tournament.OrganizerID, 1, 12)
	if err != nil || len(items) != 1 {
		t.Fatalf("owner list: %v, %d items", err, len(items))
	}
	checkRoster(&items[0])
	published, err := repositories.NewTournamentRepository(db).GetPublishedByID(ctx, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	checkRoster(published)
	private, err := repositories.NewTeamLobbyRepository(db).FindMemberRegistration(ctx, member.ID, member.AccountID, uuid.Nil)
	if err != nil || len(private.RegistrationAnswers) != 1 || private.ConsentNotice != member.ConsentNotice {
		t.Fatalf("private submission endpoint lost its data: %v", err)
	}
}
