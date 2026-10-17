package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type mockTierRepo struct {
	tiers map[uuid.UUID]*models.SponsorshipTier
}

func newMockTierRepo() *mockTierRepo {
	return &mockTierRepo{tiers: make(map[uuid.UUID]*models.SponsorshipTier)}
}

func (m *mockTierRepo) List(_ context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	var list []models.SponsorshipTier
	for _, tier := range m.tiers {
		if tier.TournamentID == tournamentID {
			list = append(list, *tier)
		}
	}
	return list, nil
}

func (m *mockTierRepo) Get(_ context.Context, tournamentID, tierID uuid.UUID) (*models.SponsorshipTier, error) {
	tier, ok := m.tiers[tierID]
	if !ok || tier.TournamentID != tournamentID {
		return nil, repositories.ErrSponsorshipTierNotFound
	}
	cpy := *tier
	return &cpy, nil
}

func (m *mockTierRepo) Create(_ context.Context, tier *models.SponsorshipTier) error {
	for _, existing := range m.tiers {
		if existing.TournamentID == tier.TournamentID && existing.Name == tier.Name {
			return repositories.ErrSponsorshipTierNameDuplicate
		}
	}
	tier.ID = uuid.New()
	cpy := *tier
	m.tiers[tier.ID] = &cpy
	return nil
}

func (m *mockTierRepo) Update(_ context.Context, tier *models.SponsorshipTier) error {
	cpy := *tier
	m.tiers[tier.ID] = &cpy
	return nil
}

func (m *mockTierRepo) Delete(_ context.Context, tournamentID, tierID uuid.UUID) error {
	tier, ok := m.tiers[tierID]
	if !ok || tier.TournamentID != tournamentID {
		return repositories.ErrSponsorshipTierNotFound
	}
	delete(m.tiers, tierID)
	return nil
}

func newTierFixture(status models.TournamentStatus) (*usecases.SponsorshipTierUseCase, *mockTierRepo, uuid.UUID, uuid.UUID) {
	ownerID := uuid.New()
	tournament := &models.Tournament{
		ID:     uuid.New(),
		Status: status,
		Staff:  []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
	}
	tiers := newMockTierRepo()
	uc := usecases.NewSponsorshipTierUseCase(tiers, &mockTournamentRepository{details: tournament})
	return uc, tiers, ownerID, tournament.ID
}

func validTierInput() usecases.SponsorshipTierInput {
	return usecases.SponsorshipTierInput{Name: "Gold", Price: 10000, Slots: 2, Perks: []string{"Logo on stream"}}
}

func TestSponsorshipTierUseCase_Create(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*usecases.SponsorshipTierInput)
		wantErr error
	}{
		{name: "valid tier", mutate: func(*usecases.SponsorshipTierInput) {}},
		{name: "blank name", mutate: func(in *usecases.SponsorshipTierInput) { in.Name = "   " }, wantErr: usecases.ErrInvalidSponsorshipTierName},
		{name: "zero price", mutate: func(in *usecases.SponsorshipTierInput) { in.Price = 0 }, wantErr: usecases.ErrInvalidSponsorshipTierPrice},
		{name: "no slots", mutate: func(in *usecases.SponsorshipTierInput) { in.Slots = 0 }, wantErr: usecases.ErrInvalidSponsorshipTierSlots},
		{name: "only blank perks", mutate: func(in *usecases.SponsorshipTierInput) { in.Perks = []string{" ", ""} }, wantErr: usecases.ErrInvalidSponsorshipTierPerks},
		{name: "too many perks", mutate: func(in *usecases.SponsorshipTierInput) {
			in.Perks = make([]string, 11)
			for i := range in.Perks {
				in.Perks[i] = "perk"
			}
		}, wantErr: usecases.ErrInvalidSponsorshipTierPerks},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
			input := validTierInput()
			tt.mutate(&input)

			tier, err := uc.Create(context.Background(), ownerID, tournamentID, input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && (tier == nil || tier.Name != "Gold") {
				t.Fatalf("expected created Gold tier, got %+v", tier)
			}
		})
	}

	t.Run("trims name and drops blank perks", func(t *testing.T) {
		uc, _, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		input := validTierInput()
		input.Name = "  Gold  "
		input.Perks = []string{"Logo", " ", " Banner "}

		tier, err := uc.Create(context.Background(), ownerID, tournamentID, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tier.Name != "Gold" || len(tier.Perks) != 2 || tier.Perks[1] != "Banner" {
			t.Fatalf("unexpected normalized tier: %+v", tier)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		uc, _, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		if _, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput())
		if !errors.Is(err, usecases.ErrSponsorshipTierNameDuplicate) {
			t.Fatalf("expected ErrSponsorshipTierNameDuplicate, got %v", err)
		}
	})

	t.Run("non-owner sees not found", func(t *testing.T) {
		uc, _, _, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		_, err := uc.Create(context.Background(), uuid.New(), tournamentID, validTierInput())
		if !errors.Is(err, usecases.ErrTournamentNotFound) {
			t.Fatalf("expected ErrTournamentNotFound, got %v", err)
		}
	})

	for _, status := range []models.TournamentStatus{models.TournamentStatusCompleted, models.TournamentStatusCancelled} {
		t.Run("locked when "+string(status), func(t *testing.T) {
			uc, _, ownerID, tournamentID := newTierFixture(status)
			_, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput())
			if !errors.Is(err, usecases.ErrSponsorshipClosed) {
				t.Fatalf("expected ErrSponsorshipClosed, got %v", err)
			}
		})
	}
}

func TestSponsorshipTierUseCase_UpdateAndDelete(t *testing.T) {
	t.Run("owner renames then deletes", func(t *testing.T) {
		uc, repo, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		tier, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		input := validTierInput()
		input.Name = "Platinum"
		updated, err := uc.Update(context.Background(), ownerID, tournamentID, tier.ID, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Platinum" {
			t.Fatalf("expected Platinum, got %q", updated.Name)
		}

		if err := uc.Delete(context.Background(), ownerID, tournamentID, tier.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.tiers) != 0 {
			t.Fatalf("expected tier deleted, %d remain", len(repo.tiers))
		}
	})

	t.Run("unknown tier", func(t *testing.T) {
		uc, _, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		err := uc.Delete(context.Background(), ownerID, tournamentID, uuid.New())
		if !errors.Is(err, usecases.ErrSponsorshipTierNotFound) {
			t.Fatalf("expected ErrSponsorshipTierNotFound, got %v", err)
		}
	})
}

func TestSponsorshipTierUseCase_ListForOwner(t *testing.T) {
	uc, _, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
	if _, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tiers, err := uc.ListForOwner(context.Background(), ownerID, tournamentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tiers) != 1 {
		t.Fatalf("expected 1 tier, got %d", len(tiers))
	}

	if _, err := uc.ListForOwner(context.Background(), uuid.New(), tournamentID); !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected ErrTournamentNotFound for non-owner, got %v", err)
	}
}
