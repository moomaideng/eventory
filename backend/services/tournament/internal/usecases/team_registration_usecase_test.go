package usecases_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"testing"
	"time"
)

func teamAnswerFixture() (*usecases.TeamLobbyUseCase, *teamLobbyRepositoryStub, *registrationFormRepositoryStub, uuid.UUID, uuid.UUID) {
	captain, member := uuid.New(), uuid.New()
	tournament := &models.Tournament{ID: uuid.New(), RegistrationMode: models.TournamentRegistrationModeTeam,
		Status: models.TournamentStatusRegistrationOpen, RegistrationDeadline: time.Now().Add(time.Hour), MinTeamSize: 2, MaxTeamSize: 5}
	repo := &teamLobbyRepositoryStub{tournament: tournament, team: &models.TournamentTeam{
		ID: uuid.New(), TournamentID: tournament.ID, Tournament: *tournament, Status: models.TournamentTeamStatusForming,
		Members: []models.TournamentTeamMember{
			{ID: uuid.New(), AccountID: captain, Role: models.TournamentTeamMemberRoleCaptain},
			{ID: uuid.New(), AccountID: member, Role: models.TournamentTeamMemberRoleMember},
		}}}
	form := &registrationFormRepositoryStub{form: &models.RegistrationForm{TournamentID: tournament.ID,
		Fields: []models.FormField{{ID: "ign", Type: models.FieldTypeShortText, Label: "In-game name", Required: true}}}}
	form.responses = map[uuid.UUID]*models.FormResponse{captain: {TournamentID: tournament.ID, AccountID: captain, Answers: validTeamAnswers()}}
	accounts := newStubAccountService().withAccount(captain, "captain", "Captain").withAccount(member, "member", "Member")
	return usecases.NewTeamLobbyUseCase(repo, accounts, nil, form, forms.NewDefaultRegistry()), repo, form, captain, member
}

func validTeamAnswers() []models.FieldAnswer {
	return []models.FieldAnswer{{FieldID: "ign", Value: "Player"}}
}

func TestTeamRegistration_CreateValidatesBeforeCreatingMembership(t *testing.T) {
	tests := []struct {
		name    string
		answers []models.FieldAnswer
		want    error
	}{
		{"missing required", nil, forms.ErrFieldRequired},
		{"file payload rejected", []models.FieldAnswer{{FieldID: "ign", Value: map[string]any{"name": "malware.exe", "type": "application/octet-stream", "content": "abc"}}}, forms.ErrInvalidFieldValue},
		{"unknown field", []models.FieldAnswer{{FieldID: "other", Value: "x"}}, forms.ErrUnrecognizedField},
		{"duplicate answer", append(validTeamAnswers(), validTeamAnswers()...), forms.ErrDuplicateFieldAnswer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, repo, _, captain, _ := teamAnswerFixture()
			_, err := u.Create(context.Background(), repo.tournament.ID, captain, "Team", tt.answers)
			if !errors.Is(err, tt.want) || repo.created != nil {
				t.Fatalf("Create error=%v, created=%v", err, repo.created)
			}
		})
	}
	u, repo, form, captain, _ := teamAnswerFixture()
	result, err := u.Create(context.Background(), repo.tournament.ID, captain, "Team", validTeamAnswers())
	if err != nil {
		t.Fatal(err)
	}
	if form.upserted == nil || len(form.upserted.Answers) != 1 || !result.Members[0].RegistrationComplete {
		t.Fatal("captain answers were not saved in MongoDB")
	}
}

func TestTeamRegistration_JoinValidatesBeforeAddingMember(t *testing.T) {
	u, repo, _, _, member := teamAnswerFixture()
	_, err := u.Join(context.Background(), "ABC123", member, nil)
	if !errors.Is(err, forms.ErrFieldRequired) || repo.joined {
		t.Fatalf("Join error=%v joined=%v", err, repo.joined)
	}
	_, err = u.Join(context.Background(), "ABC123", member, validTeamAnswers())
	if err != nil || !repo.joined {
		t.Fatalf("Join valid error=%v", err)
	}
}

func TestTeamRegistration_IncompleteMemberBlocksFreeAndPaidSubmission(t *testing.T) {
	for _, fee := range []int64{0, 250} {
		t.Run(time.Duration(fee).String(), func(t *testing.T) {
			u, repo, _, captain, member := teamAnswerFixture()
			repo.team.Tournament.EntryFee = fee
			_, err := u.Lock(context.Background(), repo.team.ID, captain)
			if !errors.Is(err, usecases.ErrMemberAnswersIncomplete) || repo.locked {
				t.Fatalf("Lock error=%v, locked=%v", err, repo.locked)
			}
			view, err := u.SaveRegistrationAnswers(context.Background(), repo.team.ID, member, validTeamAnswers())
			if err != nil || !view.Members[1].RegistrationComplete {
				t.Fatalf("Save answers error=%v", err)
			}
			_, err = u.Lock(context.Background(), repo.team.ID, captain)
			if err != nil || !repo.locked {
				t.Fatalf("Lock complete error=%v", err)
			}
		})
	}
}

func TestTeamRegistration_OnlyOwnAnswersCanChangeWhileForming(t *testing.T) {
	for _, status := range []models.TournamentTeamStatus{models.TournamentTeamStatusLocked, models.TournamentTeamStatusAccepted, models.TournamentTeamStatusRejected, models.TournamentTeamStatusPaymentPending} {
		u, repo, form, _, member := teamAnswerFixture()
		repo.team.Status = status
		_, err := u.SaveRegistrationAnswers(context.Background(), repo.team.ID, member, validTeamAnswers())
		if !errors.Is(err, repositories.ErrTeamLobbyNotForming) || form.upserted != nil {
			t.Fatalf("status=%s err=%v", status, err)
		}
	}
	u, repo, _, outsider, _ := teamAnswerFixture()
	// The account remains active, but is no longer a member of this roster.
	repo.team.Members = repo.team.Members[1:]
	_, err := u.SaveRegistrationAnswers(context.Background(), repo.team.ID, outsider, validTeamAnswers())
	if !errors.Is(err, usecases.ErrNotTeamMember) {
		t.Fatalf("outsider error=%v, want ErrNotTeamMember", err)
	}
}

func TestTeamRegistration_FormChangesInvalidateReadiness(t *testing.T) {
	u, repo, form, captain, _ := teamAnswerFixture()
	form.responses[repo.team.Members[1].AccountID] = &models.FormResponse{TournamentID: repo.tournament.ID, AccountID: repo.team.Members[1].AccountID, Answers: validTeamAnswers()}
	form.form.Fields = append(form.form.Fields, models.FormField{ID: "age", Type: models.FieldTypeNumber, Label: "Age", Required: true})
	_, err := u.Lock(context.Background(), repo.team.ID, captain)
	if !errors.Is(err, usecases.ErrMemberAnswersIncomplete) || repo.locked {
		t.Fatalf("Lock after form change error=%v", err)
	}
}

func TestTeamRegistration_ReadinessUsesOneBatchAndFailsClosed(t *testing.T) {
	u, repo, form, captain, _ := teamAnswerFixture()
	view, err := u.GetByInviteCode(context.Background(), "ABC123", captain)
	if err != nil {
		t.Fatal(err)
	}
	if form.batchCalls != 1 || !view.Members[0].RegistrationComplete || view.Members[1].RegistrationComplete {
		t.Fatalf("unexpected readiness: batch calls=%d members=%+v", form.batchCalls, view.Members)
	}
	form.batchError = errors.New("MongoDB unavailable")
	_, err = u.Lock(context.Background(), repo.team.ID, captain)
	if !errors.Is(err, form.batchError) || repo.locked {
		t.Fatalf("Lock error=%v locked=%v", err, repo.locked)
	}
}

func TestTeamRegistration_StorageFailureAndDuplicateDoNotChangeMembership(t *testing.T) {
	u, repo, form, captain, _ := teamAnswerFixture()
	form.upsertError = errors.New("MongoDB unavailable")
	_, err := u.Create(context.Background(), repo.tournament.ID, captain, "Team", validTeamAnswers())
	if !errors.Is(err, form.upsertError) || repo.created != nil {
		t.Fatalf("Create error=%v created=%v", err, repo.created)
	}
	form.upsertError = nil
	repo.activeTeam = repo.team
	_, err = u.Create(context.Background(), repo.tournament.ID, captain, "Duplicate", validTeamAnswers())
	if !errors.Is(err, repositories.ErrAlreadyInTournamentLobby) || form.upserted != nil {
		t.Fatalf("duplicate error=%v overwrote=%v", err, form.upserted)
	}
}
