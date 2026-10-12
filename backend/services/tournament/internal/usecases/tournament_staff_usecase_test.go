package usecases_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type mockStaffRepo struct {
	tournaments map[uuid.UUID]bool
	staff       map[string]*models.TournamentStaff
	deleted     map[string]*models.TournamentStaff
}

func newMockStaffRepo() *mockStaffRepo {
	return &mockStaffRepo{
		tournaments: make(map[uuid.UUID]bool),
		staff:       make(map[string]*models.TournamentStaff),
		deleted:     make(map[string]*models.TournamentStaff),
	}
}

func staffKey(tID, aID uuid.UUID) string {
	return tID.String() + ":" + aID.String()
}

func (m *mockStaffRepo) TournamentExists(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	return m.tournaments[tournamentID], nil
}

func (m *mockStaffRepo) ListStaff(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentStaff, error) {
	var list []models.TournamentStaff
	for _, s := range m.staff {
		if s.TournamentID == tournamentID {
			list = append(list, *s)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CreatedAt.Equal(list[j].CreatedAt) {
			return list[i].AccountID.String() < list[j].AccountID.String()
		}
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
	return list, nil
}

func (m *mockStaffRepo) GetStaff(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentStaff, error) {
	s, ok := m.staff[staffKey(tournamentID, accountID)]
	if !ok {
		return nil, nil
	}
	cpy := *s
	return &cpy, nil
}

func (m *mockStaffRepo) AddStaff(ctx context.Context, staff *models.TournamentStaff) error {
	k := staffKey(staff.TournamentID, staff.AccountID)
	if _, ok := m.staff[k]; ok {
		return repositories.ErrStaffAlreadyExists
	}
	if del, ok := m.deleted[k]; ok {
		delete(m.deleted, k)
		del.Role = staff.Role
		del.UpdatedAt = time.Now()
		m.staff[k] = del
		return nil
	}
	cpy := *staff
	if cpy.CreatedAt.IsZero() {
		cpy.CreatedAt = time.Now()
	}
	m.staff[k] = &cpy
	return nil
}

func (m *mockStaffRepo) UpdateStaffRole(ctx context.Context, tournamentID, accountID uuid.UUID, role models.TournamentStaffRole) error {
	k := staffKey(tournamentID, accountID)
	s, ok := m.staff[k]
	if !ok {
		return repositories.ErrStaffNotFound
	}
	s.Role = role
	s.UpdatedAt = time.Now()
	return nil
}

func (m *mockStaffRepo) RemoveStaff(ctx context.Context, tournamentID, accountID uuid.UUID) error {
	k := staffKey(tournamentID, accountID)
	s, ok := m.staff[k]
	if !ok {
		return repositories.ErrStaffNotFound
	}
	m.deleted[k] = s
	delete(m.staff, k)
	return nil
}

func setupStaffFixture() (*mockStaffRepo, *stubAccountService, *usecases.TournamentStaffUseCase, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	repo := newMockStaffRepo()
	accountSvc := newStubAccountService()
	uc := usecases.NewTournamentStaffUseCase(repo, accountSvc)

	tID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	refereeID := uuid.New()

	repo.tournaments[tID] = true

	accountSvc.withAccount(ownerID, "alice_owner", "Alice Owner")
	accountSvc.withAccount(adminID, "bob_admin", "Bob Admin")
	accountSvc.withAccount(refereeID, "charlie_ref", "Charlie Referee")

	_ = repo.AddStaff(context.Background(), &models.TournamentStaff{
		TournamentID: tID,
		AccountID:    ownerID,
		Role:         models.TournamentStaffRoleOwner,
		CreatedAt:    time.Now().Add(-3 * time.Hour),
	})
	_ = repo.AddStaff(context.Background(), &models.TournamentStaff{
		TournamentID: tID,
		AccountID:    adminID,
		Role:         models.TournamentStaffRoleAdmin,
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	})
	_ = repo.AddStaff(context.Background(), &models.TournamentStaff{
		TournamentID: tID,
		AccountID:    refereeID,
		Role:         models.TournamentStaffRoleReferee,
		CreatedAt:    time.Now().Add(-1 * time.Hour),
	})

	return repo, accountSvc, uc, tID, ownerID, adminID, refereeID
}

func TestTournamentStaffUseCase_InviteStaff(t *testing.T) {
	ctx := context.Background()

	t.Run("success by owner", func(t *testing.T) {
		repo, accountSvc, uc, tID, ownerID, _, _ := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		view, err := uc.InviteStaff(ctx, ownerID, tID, "target_user", models.TournamentStaffRoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.AccountID != targetID || view.Role != models.TournamentStaffRoleAdmin || view.Handle != "target_user" {
			t.Fatalf("unexpected view: %+v", view)
		}

		saved, _ := repo.GetStaff(ctx, tID, targetID)
		if saved == nil || saved.Role != models.TournamentStaffRoleAdmin {
			t.Fatalf("staff not saved correctly")
		}
	})

	t.Run("unauthorized caller - admin", func(t *testing.T) {
		_, accountSvc, uc, tID, _, adminID, _ := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "new_ref", "New Referee")

		_, err := uc.InviteStaff(ctx, adminID, tID, "new_ref", models.TournamentStaffRoleReferee)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})

	t.Run("cannot invite as OWNER", func(t *testing.T) {
		_, accountSvc, uc, tID, ownerID, _, _ := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		_, err := uc.InviteStaff(ctx, ownerID, tID, "target_user", models.TournamentStaffRoleOwner)
		if !errors.Is(err, usecases.ErrInvalidStaffRole) {
			t.Fatalf("expected ErrInvalidStaffRole, got %v", err)
		}
	})

	t.Run("reject invalid role", func(t *testing.T) {
		_, accountSvc, uc, tID, ownerID, _, _ := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		_, err := uc.InviteStaff(ctx, ownerID, tID, "target_user", "SUPERADMIN")
		if !errors.Is(err, usecases.ErrInvalidStaffRole) {
			t.Fatalf("expected ErrInvalidStaffRole, got %v", err)
		}
	})

	t.Run("tournament not found", func(t *testing.T) {
		_, accountSvc, uc, _, ownerID, _, _ := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		_, err := uc.InviteStaff(ctx, ownerID, uuid.New(), "target_user", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrTournamentNotFound) {
			t.Fatalf("expected ErrTournamentNotFound, got %v", err)
		}
	})

	t.Run("unauthorized caller - not staff", func(t *testing.T) {
		_, accountSvc, uc, tID, _, _, _ := setupStaffFixture()
		strangerID := uuid.New()
		accountSvc.withAccount(strangerID, "stranger", "Stranger")
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		_, err := uc.InviteStaff(ctx, strangerID, tID, "target_user", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})

	t.Run("unauthorized caller - referee", func(t *testing.T) {
		_, accountSvc, uc, tID, _, _, refereeID := setupStaffFixture()
		targetID := uuid.New()
		accountSvc.withAccount(targetID, "target_user", "Target User")

		_, err := uc.InviteStaff(ctx, refereeID, tID, "target_user", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})

	t.Run("target handle not found", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, _ := setupStaffFixture()
		_, err := uc.InviteStaff(ctx, ownerID, tID, "nonexistent", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, ports.ErrAccountNotFound) {
			t.Fatalf("expected ports.ErrAccountNotFound, got %v", err)
		}
	})

	t.Run("target account not active", func(t *testing.T) {
		_, accountSvc, uc, tID, ownerID, _, _ := setupStaffFixture()
		onboardID := uuid.New()
		accountSvc.withAccountStatus(onboardID, "onboarding_user", "Onboarding", ports.AccountStatusOnboarding)

		_, err := uc.InviteStaff(ctx, ownerID, tID, "onboarding_user", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrAccountNotActive) {
			t.Fatalf("expected ErrAccountNotActive, got %v", err)
		}
	})

	t.Run("target already staff", func(t *testing.T) {
		_, _, uc, tID, ownerID, adminID, _ := setupStaffFixture()
		_, err := uc.InviteStaff(ctx, ownerID, tID, "bob_admin", models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrStaffAlreadyExists) {
			t.Fatalf("expected ErrStaffAlreadyExists, got %v", err)
		}
		_ = adminID
	})
}

func TestTournamentStaffUseCase_UpdateStaffRole(t *testing.T) {
	ctx := context.Background()

	t.Run("success updating referee to admin", func(t *testing.T) {
		repo, _, uc, tID, ownerID, _, refereeID := setupStaffFixture()
		view, err := uc.UpdateStaffRole(ctx, ownerID, tID, refereeID, models.TournamentStaffRoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.Role != models.TournamentStaffRoleAdmin {
			t.Fatalf("expected ADMIN, got %v", view.Role)
		}
		saved, _ := repo.GetStaff(ctx, tID, refereeID)
		if saved.Role != models.TournamentStaffRoleAdmin {
			t.Fatalf("expected ADMIN in repo, got %v", saved.Role)
		}
	})

	t.Run("admin caller unauthorized", func(t *testing.T) {
		_, _, uc, tID, _, adminID, _ := setupStaffFixture()
		_, err := uc.UpdateStaffRole(ctx, adminID, tID, adminID, models.TournamentStaffRoleReferee)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})

	t.Run("cannot update role to OWNER", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, refereeID := setupStaffFixture()
		_, err := uc.UpdateStaffRole(ctx, ownerID, tID, refereeID, models.TournamentStaffRoleOwner)
		if !errors.Is(err, usecases.ErrInvalidStaffRole) {
			t.Fatalf("expected ErrInvalidStaffRole, got %v", err)
		}
	})

	t.Run("cannot modify OWNER staff member", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, _ := setupStaffFixture()
		_, err := uc.UpdateStaffRole(ctx, ownerID, tID, ownerID, models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}
	})

	t.Run("target staff not found", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, _ := setupStaffFixture()
		_, err := uc.UpdateStaffRole(ctx, ownerID, tID, uuid.New(), models.TournamentStaffRoleAdmin)
		if !errors.Is(err, usecases.ErrStaffNotFound) {
			t.Fatalf("expected ErrStaffNotFound, got %v", err)
		}
	})

	t.Run("referee caller unauthorized", func(t *testing.T) {
		_, _, uc, tID, _, adminID, refereeID := setupStaffFixture()
		_, err := uc.UpdateStaffRole(ctx, refereeID, tID, adminID, models.TournamentStaffRoleReferee)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})
}

func TestTournamentStaffUseCase_ListStaff(t *testing.T) {
	ctx := context.Background()

	t.Run("success lists all staff in order", func(t *testing.T) {
		_, _, uc, tID, ownerID, adminID, refereeID := setupStaffFixture()
		list, err := uc.ListStaff(ctx, ownerID, tID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 3 {
			t.Fatalf("expected 3 staff members, got %d", len(list))
		}
		// Expect order: owner (-3h), admin (-2h), referee (-1h)
		if list[0].AccountID != ownerID || list[0].Role != models.TournamentStaffRoleOwner {
			t.Errorf("expected 1st to be owner, got %+v", list[0])
		}
		if list[1].AccountID != adminID || list[1].Role != models.TournamentStaffRoleAdmin {
			t.Errorf("expected 2nd to be admin, got %+v", list[1])
		}
		if list[2].AccountID != refereeID || list[2].Role != models.TournamentStaffRoleReferee {
			t.Errorf("expected 3rd to be referee, got %+v", list[2])
		}
		if list[0].Handle != "alice_owner" {
			t.Errorf("expected handle alice_owner, got %v", list[0].Handle)
		}
	})

	t.Run("unauthorized caller", func(t *testing.T) {
		_, accountSvc, uc, tID, _, _, refereeID := setupStaffFixture()
		strangerID := uuid.New()
		accountSvc.withAccount(strangerID, "stranger", "Stranger")

		_, err := uc.ListStaff(ctx, strangerID, tID)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized for stranger, got %v", err)
		}

		_, err = uc.ListStaff(ctx, refereeID, tID)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized for referee, got %v", err)
		}
	})

	t.Run("tournament not found", func(t *testing.T) {
		_, _, uc, _, ownerID, _, _ := setupStaffFixture()
		_, err := uc.ListStaff(ctx, ownerID, uuid.New())
		if !errors.Is(err, usecases.ErrTournamentNotFound) {
			t.Fatalf("expected ErrTournamentNotFound, got %v", err)
		}
	})
}

func TestTournamentStaffUseCase_RemoveStaff(t *testing.T) {
	ctx := context.Background()

	t.Run("success removing referee", func(t *testing.T) {
		repo, _, uc, tID, ownerID, _, refereeID := setupStaffFixture()
		err := uc.RemoveStaff(ctx, ownerID, tID, refereeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		saved, _ := repo.GetStaff(ctx, tID, refereeID)
		if saved != nil {
			t.Fatalf("expected referee to be deleted from repo")
		}
	})

	t.Run("admin caller unauthorized", func(t *testing.T) {
		_, _, uc, tID, _, adminID, refereeID := setupStaffFixture()
		err := uc.RemoveStaff(ctx, adminID, tID, refereeID)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})

	t.Run("owner cannot remove themselves (US8-3 Scenario 2)", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, _ := setupStaffFixture()
		err := uc.RemoveStaff(ctx, ownerID, tID, ownerID)
		if !errors.Is(err, usecases.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}
	})

	t.Run("re-inviting previously removed staff succeeds", func(t *testing.T) {
		repo, _, uc, tID, ownerID, _, refereeID := setupStaffFixture()
		err := uc.RemoveStaff(ctx, ownerID, tID, refereeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		saved, _ := repo.GetStaff(ctx, tID, refereeID)
		if saved != nil {
			t.Fatalf("expected referee to be deleted from repo")
		}

		// Re-invite the same account as ADMIN
		view, err := uc.InviteStaff(ctx, ownerID, tID, "charlie_ref", models.TournamentStaffRoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error re-inviting: %v", err)
		}
		if view.Role != models.TournamentStaffRoleAdmin {
			t.Fatalf("expected role ADMIN, got %v", view.Role)
		}
		saved, _ = repo.GetStaff(ctx, tID, refereeID)
		if saved == nil || saved.Role != models.TournamentStaffRoleAdmin {
			t.Fatalf("expected restored staff to be ADMIN in repo, got %+v", saved)
		}
	})

	t.Run("staff not found", func(t *testing.T) {
		_, _, uc, tID, ownerID, _, _ := setupStaffFixture()
		err := uc.RemoveStaff(ctx, ownerID, tID, uuid.New())
		if !errors.Is(err, usecases.ErrStaffNotFound) {
			t.Fatalf("expected ErrStaffNotFound, got %v", err)
		}
	})

	t.Run("referee caller unauthorized", func(t *testing.T) {
		_, _, uc, tID, _, adminID, refereeID := setupStaffFixture()
		err := uc.RemoveStaff(ctx, refereeID, tID, adminID)
		if !errors.Is(err, usecases.ErrCallerNotAuthorized) {
			t.Fatalf("expected ErrCallerNotAuthorized, got %v", err)
		}
	})
}

func TestTournamentStaff_RolesAndCapabilities(t *testing.T) {
	owner := &models.TournamentStaff{Role: models.TournamentStaffRoleOwner}
	admin := &models.TournamentStaff{Role: models.TournamentStaffRoleAdmin}
	referee := &models.TournamentStaff{Role: models.TournamentStaffRoleReferee}
	var nilStaff *models.TournamentStaff

	// Nil safety
	if nilStaff.IsOwner() || nilStaff.IsAdmin() || nilStaff.IsReferee() {
		t.Fatal("nil staff must return false for all role checks")
	}

	// Owner
	if !owner.IsOwner() || owner.IsAdmin() || owner.IsReferee() {
		t.Fatal("owner capabilities mismatch")
	}

	// Admin
	if admin.IsOwner() || !admin.IsAdmin() || admin.IsReferee() {
		t.Fatal("admin capabilities mismatch")
	}

	// Referee
	if referee.IsOwner() || referee.IsAdmin() || !referee.IsReferee() {
		t.Fatal("referee capabilities mismatch")
	}

	// Assignable roles
	if !models.IsAssignableRole(models.TournamentStaffRoleAdmin) {
		t.Fatal("ADMIN must be assignable")
	}
	if !models.IsAssignableRole(models.TournamentStaffRoleReferee) {
		t.Fatal("REFEREE must be assignable")
	}
	if models.IsAssignableRole(models.TournamentStaffRoleOwner) {
		t.Fatal("OWNER must NOT be assignable")
	}
	if models.IsAssignableRole("UNKNOWN") {
		t.Fatal("UNKNOWN must NOT be assignable")
	}
}

func TestTournament_StaffHelpers(t *testing.T) {
	ownerID := uuid.New()
	adminID := uuid.New()
	tourney := models.Tournament{
		Staff: []models.TournamentStaff{
			{AccountID: ownerID, Role: models.TournamentStaffRoleOwner},
			{AccountID: adminID, Role: models.TournamentStaffRoleAdmin},
		},
	}

	if tourney.OwnerID() != ownerID {
		t.Fatalf("expected owner ID %v, got %v", ownerID, tourney.OwnerID())
	}

	ownerStaff, ok := tourney.StaffMember(ownerID)
	if !ok || ownerStaff.Role != models.TournamentStaffRoleOwner {
		t.Fatalf("expected owner staff, got %v, ok=%v", ownerStaff, ok)
	}

	adminStaff, ok := tourney.StaffMember(adminID)
	if !ok || adminStaff.Role != models.TournamentStaffRoleAdmin {
		t.Fatalf("expected admin staff, got %v, ok=%v", adminStaff, ok)
	}

	strangerStaff, ok := tourney.StaffMember(uuid.New())
	if ok || strangerStaff != nil {
		t.Fatalf("expected false for non-staff member")
	}

	emptyTourney := models.Tournament{}
	if emptyTourney.OwnerID() != uuid.Nil {
		t.Fatalf("expected uuid.Nil for empty tournament, got %v", emptyTourney.OwnerID())
	}
}
