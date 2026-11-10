package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrSponsorRequired     = errors.New("complete your sponsor profile before sponsoring")
	ErrTierSoldOut         = errors.New("this sponsorship package has no available slots")
	ErrAlreadySponsored    = errors.New("you have already purchased this sponsorship package")
	ErrInvalidSponsorship  = errors.New("invalid sponsorship request")
	ErrSponsorshipNotFound = errors.New("sponsorship not found")
)

type StartSponsorship struct {
	RequestID    uuid.UUID
	TournamentID uuid.UUID
	TierID       uuid.UUID
	PayerID      uuid.UUID
	WebsiteURL   string
	Logo         []byte
}

type SponsorshipUseCase struct {
	repo        repositories.SponsorshipRepository
	tournaments repositories.TournamentRepository
	accounts    ports.AccountService
	payments    ports.PaymentService
}

func NewSponsorshipUseCase(repo repositories.SponsorshipRepository, tournaments repositories.TournamentRepository, accounts ports.AccountService, payments ports.PaymentService) *SponsorshipUseCase {
	return &SponsorshipUseCase{repo: repo, tournaments: tournaments, accounts: accounts, payments: payments}
}

func (u *SponsorshipUseCase) Start(ctx context.Context, input StartSponsorship) (*models.SponsorshipPledge, error) {
	if input.RequestID == uuid.Nil || input.PayerID == uuid.Nil {
		return nil, ErrInvalidSponsorship
	}
	account, err := requireActiveAccount(ctx, u.accounts, input.PayerID)
	if err != nil {
		return nil, err
	}
	if account.Status != ports.AccountStatusActive || account.SponsorName == "" {
		return nil, ErrSponsorRequired
	}
	website, contentType, err := validateSponsorshipBrand(input.WebsiteURL, input.Logo)
	if err != nil {
		return nil, err
	}
	// A retry returns the original attempt, including when its invoice is already closed.
	existing, err := u.repo.Get(ctx, input.RequestID)
	if err == nil {
		if existing.PayerAccountID != input.PayerID || existing.TierID != input.TierID || existing.TournamentID != input.TournamentID {
			return nil, ErrInvalidSponsorship
		}
		return u.GetOwned(ctx, input.PayerID, existing.ID)
	}
	if !errors.Is(err, repositories.ErrPledgeNotFound) {
		return nil, mapSponsorshipError(err)
	}
	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	defer tx.Rollback()
	tournament, tier, err := tx.LockTier(input.TournamentID, input.TierID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if tier.Archived {
		return nil, ErrSponsorshipTierNotFound
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !sponsorshipOpen(tournament, now) {
		return nil, ErrSponsorshipClosed
	}
	existing, err = tx.FindActive(tier.ID, input.PayerID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if existing != nil {
		if existing.Status == models.PledgeConfirmed {
			return nil, ErrAlreadySponsored
		}
		if err := tx.Commit(); err != nil {
			return nil, mapSponsorshipError(err)
		}
		return u.GetOwned(ctx, input.PayerID, existing.ID)
	}
	reserved, err := tx.Reserved(tier.ID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if reserved >= int64(tier.Slots) {
		return nil, ErrTierSoldOut
	}
	if tier.Price <= 0 || tier.Price > 9007199254740991/100 || tournament.Currency != "THB" {
		return nil, ErrInvalidSponsorship
	}
	expires := now.Add(15 * time.Minute)
	if tournament.EndsAt.Before(expires) {
		expires = tournament.EndsAt.UTC().Truncate(time.Microsecond)
	}
	pledge := &models.SponsorshipPledge{
		ID: input.RequestID, TournamentID: tournament.ID, TierID: tier.ID, PayerAccountID: input.PayerID,
		SponsorName: account.SponsorName, WebsiteURL: website, TierName: tier.Name,
		AmountMinor: tier.Price * 100, Currency: tournament.Currency, Status: models.PledgePending, ExpiresAt: expires,
	}
	if err := tx.Create(pledge, &models.SponsorshipLogo{PledgeID: pledge.ID, ContentType: contentType, Data: input.Logo}); err != nil {
		return nil, mapSponsorshipError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, mapSponsorshipError(err)
	}
	return u.GetOwned(ctx, input.PayerID, pledge.ID)
}

func sponsorshipOpen(t *models.Tournament, now time.Time) bool {
	if t.Status == models.TournamentStatusDraft || !now.Before(t.EndsAt) {
		return false
	}
	switch t.Status {
	case models.TournamentStatusCrowdfunding, models.TournamentStatusRegistration, models.TournamentStatusOngoing:
		return true
	}
	return false
}

func (u *SponsorshipUseCase) GetOwned(ctx context.Context, payerID, pledgeID uuid.UUID) (*models.SponsorshipPledge, error) {
	pledge, err := u.repo.Get(ctx, pledgeID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if pledge.PayerAccountID != payerID {
		return nil, ErrSponsorshipNotFound
	}
	if err := u.Reconcile(ctx, pledge); err != nil {
		return nil, err
	}
	pledge, err = u.repo.Get(ctx, pledgeID)
	return pledge, mapSponsorshipError(err)
}

func (u *SponsorshipUseCase) ListPublic(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error) {
	if _, err := u.tournaments.GetPublishedByID(ctx, tournamentID); err != nil {
		return nil, mapSponsorshipError(err)
	}
	pledges, err := u.repo.ListConfirmed(ctx, tournamentID)
	return pledges, mapSponsorshipError(err)
}

func (u *SponsorshipUseCase) ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]repositories.OwnedSponsorship, int64, error) {
	items, total, err := u.repo.ListOwned(ctx, payerID, page, limit)
	return items, total, mapSponsorshipError(err)
}

func (u *SponsorshipUseCase) Logo(ctx context.Context, pledgeID uuid.UUID) (*models.SponsorshipLogo, error) {
	logo, err := u.repo.Logo(ctx, pledgeID)
	return logo, mapSponsorshipError(err)
}

func (u *SponsorshipUseCase) Reconcile(ctx context.Context, pledge *models.SponsorshipPledge) error {
	if pledge.Status != models.PledgePending {
		return nil
	}
	request := ports.RegistrationPaymentRequest{
		PayerAccountID: pledge.PayerAccountID, Source: models.PaymentSourceSponsorship,
		Reference: pledge.ID.String(), IdempotencyKey: pledge.ID.String(),
		Description: "Sponsorship: " + pledge.TierName, AmountMinor: pledge.AmountMinor, Currency: pledge.Currency, ExpiresAt: pledge.ExpiresAt,
	}
	payment, err := u.payments.EnsureInvoice(ctx, request)
	if err != nil || payment == nil {
		return ErrPaymentUnavailable
	}
	if payment.Source != request.Source || payment.Reference != request.Reference || !pledgeMatches(pledge, outcomeOf(payment)) || !settleable(payment.Status) {
		return ErrPaymentUnavailable
	}
	return u.settle(ctx, pledge, payment.ID, payment.Status)
}

// ApplyPaymentOutcome records the result of a payment event on the pledge it
// refers to. A pledge that is gone or already settled is left alone, so
// redelivered events are harmless. An event whose terms differ from the
// pledge returns ErrPaymentMismatch.
func (u *SponsorshipUseCase) ApplyPaymentOutcome(ctx context.Context, outcome PaymentOutcome) error {
	id, err := uuid.Parse(outcome.Reference)
	if err != nil {
		return fmt.Errorf("%w: sponsorship reference %q", ErrPaymentMismatch, outcome.Reference)
	}
	pledge, err := u.repo.Get(ctx, id)
	if errors.Is(err, repositories.ErrPledgeNotFound) {
		return nil
	}
	if err != nil {
		return mapSponsorshipError(err)
	}
	if pledge.Status != models.PledgePending {
		return nil
	}
	if !pledgeMatches(pledge, outcome) || !settleable(outcome.Status) {
		return fmt.Errorf("%w: pledge %s", ErrPaymentMismatch, pledge.ID)
	}
	return u.settle(ctx, pledge, outcome.PaymentID, outcome.Status)
}

// pledgeMatches reports whether the invoice was issued for the pledge.
func pledgeMatches(pledge *models.SponsorshipPledge, p PaymentOutcome) bool {
	return p.PaymentID != uuid.Nil &&
		p.PayerAccountID == pledge.PayerAccountID &&
		p.AmountMinor == pledge.AmountMinor &&
		p.Currency == pledge.Currency &&
		p.ExpiresAt.Equal(pledge.ExpiresAt) &&
		(pledge.PaymentID == nil || *pledge.PaymentID == p.PaymentID)
}

// settle records an invoice status on a pending pledge. An unpaid invoice
// only links the payment.
func (u *SponsorshipUseCase) settle(ctx context.Context, pledge *models.SponsorshipPledge, paymentID uuid.UUID, status models.PaymentStatus) error {
	pledge.PaymentID = &paymentID
	switch status {
	case models.PaymentPaid:
		pledge.Status = models.PledgeConfirmed
		now := time.Now().UTC()
		pledge.ConfirmedAt = &now
	case models.PaymentExpired:
		pledge.Status = models.PledgeExpired
	case models.PaymentCancelled:
		pledge.Status = models.PledgeCancelled
	}
	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return mapSponsorshipError(err)
	}
	defer tx.Rollback()
	if _, _, err := tx.LockTier(pledge.TournamentID, pledge.TierID); err != nil {
		return mapSponsorshipError(err)
	}
	changed, err := tx.SaveResult(pledge)
	if err != nil {
		return mapSponsorshipError(err)
	}
	if changed && pledge.Status == models.PledgeConfirmed {
		if err := tx.RefreshFunding(pledge.TournamentID); err != nil {
			return mapSponsorshipError(err)
		}
	}
	return mapSponsorshipError(tx.Commit())
}

func mapSponsorshipError(err error) error {
	if errors.Is(err, repositories.ErrPledgeNotFound) {
		return ErrSponsorshipNotFound
	}
	return mapTierError(err)
}
