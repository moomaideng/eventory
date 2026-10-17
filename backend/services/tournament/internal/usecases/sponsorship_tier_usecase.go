package usecases

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrSponsorshipTierNotFound      = errors.New("sponsorship tier not found")
	ErrSponsorshipTierNameDuplicate = errors.New("a tier with this name already exists in this tournament")
	ErrSponsorshipTierInUse         = errors.New("this package has a pending or confirmed sponsorship and cannot be changed")
	ErrInvalidSponsorshipTierName   = errors.New("tier name is required and must be at most 80 characters")
	ErrInvalidSponsorshipTierPrice  = errors.New("tier price must be greater than zero")
	ErrInvalidSponsorshipTierSlots  = errors.New("tier must have at least one slot")
	ErrInvalidSponsorshipTierPerks  = errors.New("tier needs 1 to 10 perks, each at most 200 characters")
	ErrSponsorshipClosed            = errors.New("tiers cannot be changed on a completed or cancelled tournament")
)

const (
	sponsorshipTierNameMaxLength = 80
	sponsorshipTierMaxPerks      = 10
	sponsorshipTierPerkMaxLength = 200
)

type SponsorshipTierInput struct {
	Name      string
	Price     int64
	Slots     int
	Perks     []string
	SortOrder int
}

type SponsorshipTierUseCase struct {
	tiers       repositories.SponsorshipTierRepository
	tournaments repositories.TournamentRepository
}

func NewSponsorshipTierUseCase(
	tiers repositories.SponsorshipTierRepository,
	tournaments repositories.TournamentRepository,
) *SponsorshipTierUseCase {
	return &SponsorshipTierUseCase{tiers: tiers, tournaments: tournaments}
}

// ListPublic returns the tiers sponsors can see on a published tournament.
func (u *SponsorshipTierUseCase) ListPublic(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	if _, err := u.tournaments.GetPublishedByID(ctx, tournamentID); err != nil {
		return nil, mapTierError(err)
	}
	return u.listTiers(ctx, tournamentID)
}

// ListForOwner returns an owned tournament's tiers, published or not, so the
// organizer can manage them before the tournament goes public.
func (u *SponsorshipTierUseCase) ListForOwner(ctx context.Context, accountID, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	tournament, err := u.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, mapTierError(err)
	}
	if !isTournamentOrganizer(tournament, accountID) {
		return nil, ErrTournamentNotFound
	}
	return u.listTiers(ctx, tournamentID)
}

func (u *SponsorshipTierUseCase) listTiers(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	tiers, err := u.tiers.List(ctx, tournamentID)
	if err != nil {
		return nil, mapTierError(err)
	}
	for i := range tiers {
		tiers[i].AvailableSlots = max(tiers[i].Slots-tiers[i].ReservedSlots, 0)
		tiers[i].Locked = tiers[i].ReservedSlots > 0
	}
	return tiers, nil
}

func (u *SponsorshipTierUseCase) Create(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	input SponsorshipTierInput,
) (*models.SponsorshipTier, error) {
	if err := u.requireEditable(ctx, accountID, tournamentID); err != nil {
		return nil, err
	}
	tier := &models.SponsorshipTier{TournamentID: tournamentID}
	if err := applyTierInput(tier, input); err != nil {
		return nil, err
	}
	if err := u.tiers.Create(ctx, tier); err != nil {
		return nil, mapTierError(err)
	}
	tier.AvailableSlots = tier.Slots
	return tier, nil
}

func (u *SponsorshipTierUseCase) Update(
	ctx context.Context,
	accountID, tournamentID, tierID uuid.UUID,
	input SponsorshipTierInput,
) (*models.SponsorshipTier, error) {
	tx, err := u.tiers.Begin(ctx)
	if err != nil {
		return nil, mapTierError(err)
	}
	defer tx.Rollback()
	tier, err := u.lockEditableTier(tx, accountID, tournamentID, tierID)
	if err != nil {
		return nil, mapTierError(err)
	}
	if err := applyTierInput(tier, input); err != nil {
		return nil, err
	}
	if err := tx.UpdateTier(tier); err != nil {
		return nil, mapTierError(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, mapTierError(err)
	}
	tier.AvailableSlots = tier.Slots
	return tier, nil
}

// Delete removes a tier.
func (u *SponsorshipTierUseCase) Delete(ctx context.Context, accountID, tournamentID, tierID uuid.UUID) error {
	tx, err := u.tiers.Begin(ctx)
	if err != nil {
		return mapTierError(err)
	}
	defer tx.Rollback()
	if _, err := u.lockEditableTier(tx, accountID, tournamentID, tierID); err != nil {
		return mapTierError(err)
	}
	// Only closed attempts remain, so their draft branding can be removed with the package.
	if err := tx.DeleteTier(tierID); err != nil {
		return mapTierError(err)
	}
	return mapTierError(tx.Commit())
}

func (u *SponsorshipTierUseCase) lockEditableTier(tx repositories.SponsorshipTierTransaction, accountID, tournamentID, tierID uuid.UUID) (*models.SponsorshipTier, error) {
	tournament, tier, err := tx.LockTier(tournamentID, tierID)
	if err != nil {
		return nil, err
	}
	if err := checkTierEditable(tournament, accountID); err != nil {
		return nil, err
	}
	reserved, err := tx.Reserved(tierID)
	if err != nil {
		return nil, err
	}
	if reserved > 0 {
		return nil, ErrSponsorshipTierInUse
	}
	return tier, nil
}

// requireEditable checks the caller owns the tournament and that it still
// accepts sponsorship changes. Non-owners get ErrTournamentNotFound so the API
// does not reveal tournaments they cannot manage.
func (u *SponsorshipTierUseCase) requireEditable(ctx context.Context, accountID, tournamentID uuid.UUID) error {
	tournament, err := u.tournaments.GetByID(ctx, tournamentID)
	if err != nil {
		return mapTierError(err)
	}
	return checkTierEditable(tournament, accountID)
}

func checkTierEditable(tournament *models.Tournament, accountID uuid.UUID) error {
	if !isTournamentOrganizer(tournament, accountID) {
		return ErrTournamentNotFound
	}
	if tournament.Status == models.TournamentStatusCompleted || tournament.Status == models.TournamentStatusCancelled {
		return ErrSponsorshipClosed
	}
	return nil
}

func applyTierInput(tier *models.SponsorshipTier, input SponsorshipTierInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" || utf8.RuneCountInString(name) > sponsorshipTierNameMaxLength {
		return ErrInvalidSponsorshipTierName
	}
	if input.Price <= 0 {
		return ErrInvalidSponsorshipTierPrice
	}
	if input.Slots <= 0 {
		return ErrInvalidSponsorshipTierSlots
	}

	perks := make([]string, 0, len(input.Perks))
	for _, perk := range input.Perks {
		perk = strings.TrimSpace(perk)
		if perk == "" {
			continue
		}
		if utf8.RuneCountInString(perk) > sponsorshipTierPerkMaxLength {
			return ErrInvalidSponsorshipTierPerks
		}
		perks = append(perks, perk)
	}
	if len(perks) == 0 || len(perks) > sponsorshipTierMaxPerks {
		return ErrInvalidSponsorshipTierPerks
	}

	tier.Name = name
	tier.Price = input.Price
	tier.Slots = input.Slots
	tier.Perks = perks
	tier.SortOrder = input.SortOrder
	return nil
}

func mapTierError(err error) error {
	switch {
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
