package usecases

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrSponsorRequired    = errors.New("complete your sponsor profile before sponsoring")
	ErrTierSoldOut        = errors.New("this sponsorship package has no available slots")
	ErrAlreadySponsored   = errors.New("you have already purchased this sponsorship package")
	ErrInvalidSponsorship = errors.New("invalid sponsorship request")
)

const SponsorshipPaymentSource = "SPONSORSHIP"

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
		return nil, err
	}
	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	tournament, tier, err := tx.LockTier(input.TournamentID, input.TierID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !sponsorshipOpen(tournament, now) {
		return nil, ErrSponsorshipClosed
	}
	existing, err = tx.FindActive(tier.ID, input.PayerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == models.PledgeConfirmed {
			return nil, ErrAlreadySponsored
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return u.GetOwned(ctx, input.PayerID, existing.ID)
	}
	reserved, err := tx.Reserved(tier.ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return u.GetOwned(ctx, input.PayerID, pledge.ID)
}

func sponsorshipOpen(t *models.Tournament, now time.Time) bool {
	if !t.Published || !now.Before(t.EndsAt) {
		return false
	}
	switch t.Status {
	case models.TournamentStatusCrowdfunding, models.TournamentStatusRegistrationOpen, models.TournamentStatusRegistrationClosed, models.TournamentStatusOngoing:
		return true
	}
	return false
}

func (u *SponsorshipUseCase) GetOwned(ctx context.Context, payerID, pledgeID uuid.UUID) (*models.SponsorshipPledge, error) {
	pledge, err := u.repo.Get(ctx, pledgeID)
	if err != nil {
		return nil, err
	}
	if pledge.PayerAccountID != payerID {
		return nil, repositories.ErrPledgeNotFound
	}
	if err := u.Reconcile(ctx, pledge); err != nil {
		return nil, err
	}
	return u.repo.Get(ctx, pledgeID)
}

func (u *SponsorshipUseCase) ListPublic(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error) {
	if _, err := u.tournaments.GetPublishedByID(ctx, tournamentID); err != nil {
		return nil, err
	}
	return u.repo.ListConfirmed(ctx, tournamentID)
}

func (u *SponsorshipUseCase) ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]repositories.OwnedSponsorship, int64, error) {
	return u.repo.ListOwned(ctx, payerID, page, limit)
}

func (u *SponsorshipUseCase) Logo(ctx context.Context, pledgeID uuid.UUID) (*models.SponsorshipLogo, error) {
	return u.repo.Logo(ctx, pledgeID)
}

func (u *SponsorshipUseCase) Reconcile(ctx context.Context, pledge *models.SponsorshipPledge) error {
	if pledge.Status != models.PledgePending {
		return nil
	}
	request := ports.RegistrationPaymentRequest{
		PayerAccountID: pledge.PayerAccountID, Source: SponsorshipPaymentSource,
		Reference: pledge.ID.String(), IdempotencyKey: pledge.ID.String(),
		Description: "Sponsorship: " + pledge.TierName, AmountMinor: pledge.AmountMinor, Currency: pledge.Currency, ExpiresAt: pledge.ExpiresAt,
	}
	payment, err := u.payments.EnsureInvoice(ctx, request)
	if err != nil || payment == nil {
		return ErrPaymentUnavailable
	}
	if payment.ID == uuid.Nil ||
		payment.PayerAccountID != request.PayerAccountID ||
		payment.Source != request.Source ||
		payment.Reference != request.Reference ||
		payment.AmountMinor != request.AmountMinor ||
		payment.Currency != request.Currency ||
		!payment.ExpiresAt.Equal(request.ExpiresAt) ||
		(pledge.PaymentID != nil && *pledge.PaymentID != payment.ID) {
		return ErrPaymentUnavailable
	}
	pledge.PaymentID = &payment.ID
	switch payment.Status {
	case "PAID":
		pledge.Status = models.PledgeConfirmed
		now := time.Now().UTC()
		pledge.ConfirmedAt = &now
	case "EXPIRED":
		pledge.Status = models.PledgeExpired
	case "CANCELLED":
		pledge.Status = models.PledgeCancelled
	case "UNPAID":
	default:
		return ErrPaymentUnavailable
	}
	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, _, err := tx.LockTier(pledge.TournamentID, pledge.TierID); err != nil {
		return err
	}
	changed, err := tx.SaveResult(pledge)
	if err != nil {
		return err
	}
	if changed && pledge.Status == models.PledgeConfirmed {
		if err := tx.RefreshFunding(pledge.TournamentID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (u *SponsorshipUseCase) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pledges, err := u.repo.Pending(ctx)
			if err != nil {
				slog.Error("list pending sponsorships", "error", err)
				continue
			}
			for i := range pledges {
				if ctx.Err() != nil {
					return
				}
				if err := u.Reconcile(ctx, &pledges[i]); err != nil {
					slog.Warn("reconcile sponsorship", "pledge", pledges[i].ID, "error", err)
				}
			}
		}
	}
}
