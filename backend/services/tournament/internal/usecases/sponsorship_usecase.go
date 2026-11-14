package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/money"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrSponsorshipNotFound = errors.New("sponsorship not found")
	ErrSponsorRequired     = errors.New("complete your sponsor profile (brand name, websiteUrl, and logoUrl) before sponsoring")
	ErrTierSoldOut         = errors.New("no slots remaining for this sponsorship package")
	ErrAlreadySponsored    = errors.New("you already have an active sponsorship for this tournament")
	ErrInvalidSponsorship  = errors.New("invalid sponsorship request")
)

type StartSponsorship struct {
	RequestID, TournamentID, TierID, PayerID uuid.UUID
}

type PublicSponsorView struct {
	ID          uuid.UUID
	SponsorName string
	TierName    string
	WebsiteURL  string
	LogoURL     string
}

type SponsorshipUseCase struct {
	repo        repositories.SponsorshipRepository
	tournaments repositories.TournamentRepository
	accounts    ports.AccountService
	payments    ports.PaymentService
}

func NewSponsorshipUseCase(
	repo repositories.SponsorshipRepository,
	tournaments repositories.TournamentRepository,
	accounts ports.AccountService,
	payments ports.PaymentService,
) *SponsorshipUseCase {
	return &SponsorshipUseCase{repo: repo, tournaments: tournaments, accounts: accounts, payments: payments}
}

func (u *SponsorshipUseCase) Start(ctx context.Context, input StartSponsorship) (*models.SponsorshipPledge, error) {
	if input.RequestID == uuid.Nil || input.TournamentID == uuid.Nil || input.TierID == uuid.Nil || input.PayerID == uuid.Nil {
		return nil, ErrInvalidSponsorship
	}
	account, err := u.accounts.GetAccount(ctx, input.PayerID)
	if err != nil {
		return nil, err
	}
	if account.Status != ports.AccountStatusActive {
		return nil, ErrSponsorRequired
	}
	profile, err := u.accounts.GetSponsorProfileByAccountID(ctx, input.PayerID)
	if err != nil {
		if errors.Is(err, ports.ErrSponsorNotFound) {
			return nil, ErrSponsorRequired
		}
		return nil, err
	}
	if profile.SponsorName == "" || profile.WebsiteURL == "" || profile.LogoURL == "" {
		return nil, ErrSponsorRequired
	}

	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	defer tx.Rollback()

	if existing, err := tx.FindByID(input.RequestID); err != nil {
		return nil, mapSponsorshipError(err)
	} else if existing != nil {
		if existing.PayerAccountID != input.PayerID || existing.TierID != input.TierID || existing.TournamentID != input.TournamentID {
			return nil, ErrInvalidSponsorship
		}
		if err := tx.Commit(); err != nil {
			return nil, mapSponsorshipError(err)
		}
		return u.refreshIfPending(ctx, existing)
	}

	tournament, tier, err := tx.LockTier(input.TournamentID, input.TierID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !sponsorshipOpen(tournament, now) {
		return nil, ErrSponsorshipClosed
	}

	if active, err := tx.FindActive(tournament.ID, input.PayerID); err != nil {
		return nil, mapSponsorshipError(err)
	} else if active != nil {
		if active.Status == models.PledgeConfirmed {
			return nil, ErrAlreadySponsored
		}
		// An active pending checkout may have expired in payment service; let it finish first.
		_ = tx.Commit()
		settled, err := u.refreshIfPending(ctx, active)
		if err != nil {
			return nil, err
		}
		if settled.Status == models.PledgePending {
			return nil, ErrAlreadySponsored
		}
		if settled.Status == models.PledgeConfirmed {
			return nil, ErrAlreadySponsored
		}
		return u.Start(ctx, input)
	}

	reserved, err := tx.Reserved(tier.ID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if reserved >= int64(tier.Slots) {
		return nil, ErrTierSoldOut
	}

	amt, err := money.FromMinor(tier.PriceMinor, tournament.Currency)
	if err != nil || amt.AmountMinor == 0 {
		return nil, ErrInvalidSponsorship
	}
	expires := now.Add(15 * time.Minute)
	if !tournament.EndsAt.IsZero() && tournament.EndsAt.Before(expires) {
		expires = tournament.EndsAt.UTC().Truncate(time.Microsecond)
	}

	payment, err := u.payments.EnsureInvoice(ctx, ports.RegistrationPaymentRequest{
		PayerAccountID: input.PayerID,
		Source:         models.PaymentSourceSponsorship,
		Reference:      input.RequestID.String(),
		IdempotencyKey: "sponsorship:" + input.RequestID.String(),
		Description:    fmt.Sprintf("Sponsorship: %s (%s)", tournament.Name, tier.Name),
		AmountMinor:    amt.AmountMinor,
		Currency:       amt.Currency,
		ExpiresAt:      expires,
	})
	if err != nil {
		return nil, ErrPaymentUnavailable
	}
	pledge := &models.SponsorshipPledge{
		ID:             input.RequestID,
		TournamentID:   tournament.ID,
		TierID:         tier.ID,
		PayerAccountID: input.PayerID,
		AmountMinor:    amt.AmountMinor,
		Currency:       amt.Currency,
		Status:         models.PledgePending,
		PaymentID:      &payment.ID,
	}
	if !pledgeMatches(pledge, outcomeOf(payment)) || payment.Source != models.PaymentSourceSponsorship {
		return nil, ErrPaymentUnavailable
	}
	if err := tx.Create(pledge); err != nil {
		return nil, mapSponsorshipError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, mapSponsorshipError(err)
	}
	// An invoice can already be settled when an idempotent request is resumed or
	// when Payment emits its event before this transaction commits. Applying the
	// response after the pledge exists closes that otherwise missed-event window.
	if err := u.settle(ctx, pledge, payment.ID, payment.Status); err != nil {
		return nil, err
	}
	return pledge, nil
}

func sponsorshipOpen(t *models.Tournament, now time.Time) bool {
	open := t.Status != models.TournamentStatusDraft && t.Status != models.TournamentStatusCompleted
	return open && (t.EndsAt.IsZero() || now.Before(t.EndsAt))
}

func (u *SponsorshipUseCase) GetOwned(ctx context.Context, payerID, pledgeID uuid.UUID) (*models.SponsorshipPledge, error) {
	pledge, err := u.repo.Get(ctx, pledgeID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if pledge.PayerAccountID != payerID {
		return nil, ErrSponsorshipNotFound
	}
	return u.refreshIfPending(ctx, pledge)
}

func (u *SponsorshipUseCase) refreshIfPending(ctx context.Context, pledge *models.SponsorshipPledge) (*models.SponsorshipPledge, error) {
	if err := u.Reconcile(ctx, pledge); err != nil {
		return nil, err
	}
	return pledge, nil
}

func (u *SponsorshipUseCase) ListPublic(ctx context.Context, tournamentID uuid.UUID) ([]PublicSponsorView, error) {
	if _, err := u.tournaments.GetPublishedByID(ctx, tournamentID); err != nil {
		return nil, mapSponsorshipError(err)
	}
	pledges, err := u.repo.ListConfirmed(ctx, tournamentID)
	if err != nil {
		return nil, mapSponsorshipError(err)
	}
	if len(pledges) == 0 {
		return []PublicSponsorView{}, nil
	}
	ids := make([]uuid.UUID, 0, len(pledges))
	seen := make(map[uuid.UUID]struct{}, len(pledges))
	for _, p := range pledges {
		if _, ok := seen[p.PayerAccountID]; !ok {
			seen[p.PayerAccountID] = struct{}{}
			ids = append(ids, p.PayerAccountID)
		}
	}
	profiles, err := u.accounts.BatchGetSponsorProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	views := make([]PublicSponsorView, 0, len(pledges))
	for _, p := range pledges {
		profile := profiles[p.PayerAccountID]
		tierName := ""
		if p.Tier != nil {
			tierName = p.Tier.Name
		}
		views = append(views, PublicSponsorView{
			ID:          p.ID,
			SponsorName: profile.SponsorName,
			TierName:    tierName,
			WebsiteURL:  profile.WebsiteURL,
			LogoURL:     profile.LogoURL,
		})
	}
	return views, nil
}

func (u *SponsorshipUseCase) ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]repositories.OwnedSponsorship, int64, error) {
	if page < 1 || limit < 1 || limit > 100 {
		return nil, 0, ErrInvalidSponsorship
	}
	return u.repo.ListOwned(ctx, payerID, page, limit)
}

func (u *SponsorshipUseCase) Reconcile(ctx context.Context, pledge *models.SponsorshipPledge) error {
	if pledge.Status != models.PledgePending || pledge.PaymentID == nil {
		return nil
	}
	payment, err := u.payments.GetPayment(ctx, *pledge.PaymentID)
	if err != nil {
		return ErrPaymentUnavailable
	}
	if payment.Source != models.PaymentSourceSponsorship || !settleable(payment.Status) {
		return ErrPaymentUnavailable
	}
	outcome := outcomeOf(payment)
	if !pledgeMatches(pledge, outcome) {
		return ErrPaymentUnavailable
	}
	return u.settle(ctx, pledge, outcome.PaymentID, outcome.Status)
}

func (u *SponsorshipUseCase) ApplyPaymentOutcome(ctx context.Context, outcome PaymentOutcome) error {
	pledgeID, err := uuid.Parse(outcome.Reference)
	if err != nil {
		return fmt.Errorf("%w: invalid pledge reference", ErrPaymentMismatch)
	}
	pledge, err := u.repo.Get(ctx, pledgeID)
	if errors.Is(err, repositories.ErrPledgeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if pledge.Status != models.PledgePending {
		return nil
	}
	if !pledgeMatches(pledge, outcome) {
		return fmt.Errorf("%w: sponsorship details do not match invoice", ErrPaymentMismatch)
	}
	return u.settle(ctx, pledge, outcome.PaymentID, outcome.Status)
}

func pledgeMatches(pledge *models.SponsorshipPledge, outcome PaymentOutcome) bool {
	return outcome.PaymentID != uuid.Nil &&
		outcome.PayerAccountID == pledge.PayerAccountID &&
		outcome.Reference == pledge.ID.String() &&
		outcome.AmountMinor == pledge.AmountMinor &&
		outcome.Currency == pledge.Currency &&
		(pledge.PaymentID == nil || *pledge.PaymentID == outcome.PaymentID)
}

func (u *SponsorshipUseCase) settle(ctx context.Context, pledge *models.SponsorshipPledge, paymentID uuid.UUID, status models.PaymentStatus) error {
	next, ok := map[models.PaymentStatus]models.PledgeStatus{
		models.PaymentPaid:      models.PledgeConfirmed,
		models.PaymentExpired:   models.PledgeExpired,
		models.PaymentCancelled: models.PledgeCancelled,
	}[status]
	if !ok {
		return nil
	}
	pledge.Status = next
	pledge.PaymentID = &paymentID
	tx, err := u.repo.Begin(ctx)
	if err != nil {
		return mapSponsorshipError(err)
	}
	defer tx.Rollback()
	if _, err := tx.SaveResult(pledge); err != nil {
		return mapSponsorshipError(err)
	}
	return mapSponsorshipError(tx.Commit())
}

func mapSponsorshipError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repositories.ErrPledgeNotFound):
		return ErrSponsorshipNotFound
	case errors.Is(err, repositories.ErrTournamentNotFound):
		return ErrTournamentNotFound
	case errors.Is(err, repositories.ErrSponsorshipTierNotFound):
		return ErrSponsorshipTierNotFound
	case errors.Is(err, repositories.ErrSponsorshipTierNameDuplicate):
		return ErrSponsorshipTierNameDuplicate
	default:
		return err
	}
}
