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
			{ID: uuid.New(), AccountID: captain, Role: models.TournamentTeamMemberRoleCaptain, RegistrationSubmitted: true, RegistrationAnswers: validTeamAnswers()},
			{ID: uuid.New(), AccountID: member, Role: models.TournamentTeamMemberRoleMember, RegistrationSubmitted: true},
		}}}
	form := &registrationFormRepositoryStub{form: &models.RegistrationForm{TournamentID: tournament.ID,
		Fields: []models.FormField{{ID: "ign", Type: models.FieldTypeShortText, Label: "In-game name", Required: true}}}}
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
	u, repo, _, captain, _ := teamAnswerFixture()
	result, err := u.Create(context.Background(), repo.tournament.ID, captain, "Team", validTeamAnswers())
	if err != nil {
		t.Fatal(err)
	}
	if !repo.captain.RegistrationSubmitted || len(repo.captain.RegistrationAnswers) != 1 || !result.Members[0].RegistrationComplete {
		t.Fatal("captain answers were not bound to membership")
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
		u, repo, _, _, member := teamAnswerFixture()
		repo.team.Status = status
		_, err := u.SaveRegistrationAnswers(context.Background(), repo.team.ID, member, validTeamAnswers())
		if !errors.Is(err, repositories.ErrTeamLobbyNotForming) || repo.team.Members[1].RegistrationVersion != 0 {
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
	repo.team.Members[1].RegistrationAnswers = validTeamAnswers()
	form.form.Fields = append(form.form.Fields, models.FormField{ID: "age", Type: models.FieldTypeNumber, Label: "Age", Required: true})
	_, err := u.Lock(context.Background(), repo.team.ID, captain)
	if !errors.Is(err, usecases.ErrMemberAnswersIncomplete) || repo.locked {
		t.Fatalf("Lock after form change error=%v", err)
	}
}
