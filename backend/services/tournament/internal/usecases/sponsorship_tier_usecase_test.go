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
	tiers      map[uuid.UUID]*models.SponsorshipTier
	tournament *models.Tournament
	reserved   map[uuid.UUID]int64
	commits    int
	rollbacks  int
}

func newMockTierRepo() *mockTierRepo {
	return &mockTierRepo{tiers: make(map[uuid.UUID]*models.SponsorshipTier), reserved: make(map[uuid.UUID]int64)}
}

func (m *mockTierRepo) List(_ context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	var list []models.SponsorshipTier
	for _, tier := range m.tiers {
		if tier.TournamentID == tournamentID && !tier.Archived {
			copy := *tier
			copy.ReservedSlots = int(m.reserved[tier.ID])
			list = append(list, copy)
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
		if existing.TournamentID == tier.TournamentID && existing.Name == tier.Name && !existing.Archived {
			return repositories.ErrSponsorshipTierNameDuplicate
		}
	}
	tier.ID = uuid.New()
	cpy := *tier
	m.tiers[tier.ID] = &cpy
	return nil
}

type mockTierTransaction struct{ repo *mockTierRepo }

func (m *mockTierRepo) Begin(context.Context) (repositories.SponsorshipTierTransaction, error) {
	return &mockTierTransaction{repo: m}, nil
}

func (tx *mockTierTransaction) LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error) {
	if tx.repo.tournament.ID != tournamentID {
		return nil, nil, repositories.ErrTournamentNotFound
	}
	tier, err := tx.repo.Get(context.Background(), tournamentID, tierID)
	return tx.repo.tournament, tier, err
}

func (tx *mockTierTransaction) Reserved(tierID uuid.UUID) (int64, error) {
	return tx.repo.reserved[tierID], nil
}

func (tx *mockTierTransaction) Commit() error { tx.repo.commits++; return nil }
func (tx *mockTierTransaction) Rollback()     { tx.repo.rollbacks++ }

func (tx *mockTierTransaction) UpdateTier(tier *models.SponsorshipTier) error {
	cpy := *tier
	tx.repo.tiers[tier.ID] = &cpy
	return nil
}

func (tx *mockTierTransaction) ArchiveTier(tierID uuid.UUID) error {
	tx.repo.tiers[tierID].Archived = true
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
	tiers.tournament = tournament
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
		if len(repo.tiers) != 1 || !repo.tiers[tier.ID].Archived {
			t.Fatal("expected the archived tier to remain stored")
		}
		listed, err := uc.ListForOwner(context.Background(), ownerID, tournamentID)
		if err != nil || len(listed) != 0 {
			t.Fatalf("archived tier still listed: %+v, %v", listed, err)
		}
		if _, err := uc.Update(context.Background(), ownerID, tournamentID, tier.ID, input); !errors.Is(err, usecases.ErrSponsorshipTierNotFound) {
			t.Fatalf("archived tier should not be editable: %v", err)
		}
		if err := uc.Delete(context.Background(), ownerID, tournamentID, tier.ID); !errors.Is(err, usecases.ErrSponsorshipTierNotFound) {
			t.Fatalf("archived tier should not be removable again: %v", err)
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

func TestSponsorshipTierUseCase_ProtectedEdits(t *testing.T) {
	for _, operation := range []string{"update", "delete"} {
		for _, reason := range []string{"reserved", "non-owner", "completed", "cancelled"} {
			t.Run(operation+"/"+reason, func(t *testing.T) {
				uc, repo, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
				tier, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput())
				if err != nil {
					t.Fatal(err)
				}
				want := usecases.ErrSponsorshipTierInUse
				switch reason {
				case "reserved":
					repo.reserved[tier.ID] = 1
				case "non-owner":
					ownerID = uuid.New()
					want = usecases.ErrTournamentNotFound
				case "completed":
					repo.tournament.Status = models.TournamentStatusCompleted
					want = usecases.ErrSponsorshipClosed
				case "cancelled":
					repo.tournament.Status = models.TournamentStatusCancelled
					want = usecases.ErrSponsorshipClosed
				}
				if operation == "update" {
					input := validTierInput()
					input.Name = "Changed"
					_, err = uc.Update(context.Background(), ownerID, tournamentID, tier.ID, input)
				} else {
					err = uc.Delete(context.Background(), ownerID, tournamentID, tier.ID)
				}
				if !errors.Is(err, want) {
					t.Fatalf("want %v, got %v", want, err)
				}
				if len(repo.tiers) != 1 || repo.tiers[tier.ID].Name != "Gold" || repo.commits != 0 || repo.rollbacks != 1 {
					t.Fatalf("blocked edit changed data or did not roll back: %+v", repo)
				}
			})
		}
	}
}

func TestSponsorshipTierUseCase_Availability(t *testing.T) {
	for _, reserved := range []int64{0, 1, 2, 3} {
		uc, repo, ownerID, tournamentID := newTierFixture(models.TournamentStatusRegistrationOpen)
		tier, err := uc.Create(context.Background(), ownerID, tournamentID, validTierInput())
		if err != nil {
			t.Fatal(err)
		}
		repo.reserved[tier.ID] = reserved
		tiers, err := uc.ListForOwner(context.Background(), ownerID, tournamentID)
		if err != nil {
			t.Fatal(err)
		}
		if tiers[0].AvailableSlots != max(tier.Slots-int(reserved), 0) || tiers[0].Locked != (reserved > 0) {
			t.Fatalf("incorrect availability for %d reserved slots: %+v", reserved, tiers[0])
		}
	}
}
