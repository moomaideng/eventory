package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type refundReviewRepo struct {
	ownerID uuid.UUID
	team    models.TournamentTeam
	marked  bool
}

type refundReviewTx struct {
	repo *refundReviewRepo
}

func (r *refundReviewRepo) Begin(_ context.Context) (repositories.TeamReviewTransaction, error) {
	return &refundReviewTx{repo: r}, nil
}

func (tx *refundReviewTx) LockTournament(tournamentID uuid.UUID) (*models.Tournament, error) {
	return &models.Tournament{
		ID:       tournamentID,
		Status:   models.TournamentStatusRegistration,
		Capacity: 1,
		Staff: []models.TournamentStaff{
			{AccountID: tx.repo.ownerID, Role: models.TournamentStaffRoleOwner},
		},
	}, nil
}

func (tx *refundReviewTx) LockTeamAndAcceptedCount(_, _ uuid.UUID) (*models.TournamentTeam, int, error) {
	return &tx.repo.team, 0, nil
}

func (tx *refundReviewTx) SaveReview(team *models.TournamentTeam, toStatus models.TournamentTeamStatus, review *models.TournamentTeamReview) error {
	team.Status = toStatus
	review.ID = uuid.New()
	return nil
}

func (tx *refundReviewTx) Commit() error { return nil }
func (tx *refundReviewTx) Rollback()     {}

func (r *refundReviewRepo) PendingRefunds(context.Context) ([]models.TournamentTeam, error) {
	if r.team.Status == models.TournamentTeamStatusRejected && r.team.PaymentID != nil && !r.marked {
		return []models.TournamentTeam{r.team}, nil
	}
	return nil, nil
}

func (r *refundReviewRepo) MarkRefunded(_ context.Context, teamID, paymentID uuid.UUID) error {
	if teamID != r.team.ID || paymentID != *r.team.PaymentID {
		return errors.New("wrong refund target")
	}
	r.marked = true
	return nil
}

type refundPaymentStub struct {
	fail  bool
	calls int
}

func (p *refundPaymentStub) RefundRegistration(_ context.Context, paymentID, entryID uuid.UUID) (*ports.RegistrationPayment, error) {
	p.calls++
	if p.fail {
		return nil, errors.New("payment service unavailable")
	}
	return &ports.RegistrationPayment{ID: paymentID, Reference: entryID.String() + ":attempt", Status: "REFUNDED"}, nil
}

func TestRejectedRegistrationRefundRetries(t *testing.T) {
	entryID, paymentID, tournamentID, reviewerID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &refundReviewRepo{ownerID: reviewerID, team: models.TournamentTeam{ID: entryID, Status: models.TournamentTeamStatusLocked, PaymentID: &paymentID}}
	payments := &refundPaymentStub{fail: true}
	uc := usecases.NewTeamReviewUseCase(repo, newStubAccountService(), payments)
	result, err := uc.Review(context.Background(), reviewerID, tournamentID, entryID, "REJECT", "Not eligible")
	if err != nil || result.ToStatus != string(models.TournamentTeamStatusRejected) || payments.calls != 1 || repo.marked {
		t.Fatalf("rejection must commit while refund waits for retry: result=%+v err=%v calls=%d marked=%v", result, err, payments.calls, repo.marked)
	}
	payments.fail = false
	pending, err := uc.PendingRefunds(context.Background())
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending refunds = %d, err=%v; want one", len(pending), err)
	}
	if err := uc.RetryRefund(context.Background(), &pending[0]); err != nil || !repo.marked || payments.calls != 2 {
		t.Fatalf("refund retry failed: err=%v calls=%d marked=%v", err, payments.calls, repo.marked)
	}
	pending, err = uc.PendingRefunds(context.Background())
	if err != nil || len(pending) != 0 || payments.calls != 2 {
		t.Fatal("completed refund was retried")
	}
}

func TestAcceptedRegistrationDoesNotRefund(t *testing.T) {
	entryID, paymentID, reviewerID := uuid.New(), uuid.New(), uuid.New()
	repo := &refundReviewRepo{ownerID: reviewerID, team: models.TournamentTeam{ID: entryID, Status: models.TournamentTeamStatusLocked, PaymentID: &paymentID}}
	payments := &refundPaymentStub{}
	uc := usecases.NewTeamReviewUseCase(repo, newStubAccountService(), payments)
	if _, err := uc.Review(context.Background(), reviewerID, uuid.New(), entryID, "ACCEPT", ""); err != nil || payments.calls != 0 {
		t.Fatalf("accepted entry requested a refund: err=%v calls=%d", err, payments.calls)
	}
}
