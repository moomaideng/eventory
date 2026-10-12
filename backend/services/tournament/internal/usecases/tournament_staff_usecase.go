package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrCallerNotAuthorized = errors.New("caller is not authorized to manage tournament staff")
	ErrInvalidStaffRole    = errors.New("invalid staff role: must be ADMIN or REFEREE")
	ErrAccountNotActive    = errors.New("invited account is not active")
	ErrStaffAlreadyExists  = errors.New("account is already a staff member")
	ErrStaffNotFound       = errors.New("staff member not found")
	ErrCannotModifyOwner   = errors.New("tournament owner role cannot be modified or removed")
)

type TournamentStaffView struct {
	TournamentID uuid.UUID
	AccountID    uuid.UUID
	Handle       string
	DisplayName  string
	Role         models.TournamentStaffRole
	CreatedAt    time.Time
}

type TournamentStaffUseCase struct {
	repo       repositories.TournamentStaffRepository
	accountSvc ports.AccountService
}

func NewTournamentStaffUseCase(repo repositories.TournamentStaffRepository, accountSvc ports.AccountService) *TournamentStaffUseCase {
	return &TournamentStaffUseCase{
		repo:       repo,
		accountSvc: accountSvc,
	}
}

func (u *TournamentStaffUseCase) checkCallerAuth(ctx context.Context, tournamentID, callerID uuid.UUID) error {
	exists, err := u.repo.TournamentExists(ctx, tournamentID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTournamentNotFound
	}

	callerStaff, err := u.repo.GetStaff(ctx, tournamentID, callerID)
	if err != nil {
		return err
	}
	if !callerStaff.IsOwner() {
		return ErrCallerNotAuthorized
	}
	return nil
}

func (u *TournamentStaffUseCase) InviteStaff(
	ctx context.Context,
	callerID, tournamentID uuid.UUID,
	handle string,
	role models.TournamentStaffRole,
) (*TournamentStaffView, error) {
	if !models.IsAssignableRole(role) {
		return nil, ErrInvalidStaffRole
	}

	if err := u.checkCallerAuth(ctx, tournamentID, callerID); err != nil {
		return nil, err
	}

	account, err := u.accountSvc.GetAccountByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	if account.Status != ports.AccountStatusActive {
		return nil, ErrAccountNotActive
	}

	existing, err := u.repo.GetStaff(ctx, tournamentID, account.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrStaffAlreadyExists
	}

	staff := &models.TournamentStaff{
		TournamentID: tournamentID,
		AccountID:    account.ID,
		Role:         role,
	}
	if err := u.repo.AddStaff(ctx, staff); err != nil {
		if errors.Is(err, repositories.ErrStaffAlreadyExists) {
			return nil, ErrStaffAlreadyExists
		}
		return nil, err
	}

	return &TournamentStaffView{
		TournamentID: tournamentID,
		AccountID:    account.ID,
		Handle:       account.Handle,
		DisplayName:  account.DisplayName,
		Role:         role,
		CreatedAt:    staff.CreatedAt,
	}, nil
}

func (u *TournamentStaffUseCase) UpdateStaffRole(
	ctx context.Context,
	callerID, tournamentID, targetAccountID uuid.UUID,
	newRole models.TournamentStaffRole,
) (*TournamentStaffView, error) {
	if !models.IsAssignableRole(newRole) {
		return nil, ErrInvalidStaffRole
	}

	if err := u.checkCallerAuth(ctx, tournamentID, callerID); err != nil {
		return nil, err
	}

	targetStaff, err := u.repo.GetStaff(ctx, tournamentID, targetAccountID)
	if err != nil {
		return nil, err
	}
	if targetStaff == nil {
		return nil, ErrStaffNotFound
	}
	if targetStaff.IsOwner() {
		return nil, ErrCannotModifyOwner
	}

	if err := u.repo.UpdateStaffRole(ctx, tournamentID, targetAccountID, newRole); err != nil {
		if errors.Is(err, repositories.ErrStaffNotFound) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}

	handle := ""
	displayName := ""
	if acc, err := u.accountSvc.GetAccount(ctx, targetAccountID); err == nil && acc != nil {
		handle = acc.Handle
		displayName = acc.DisplayName
	}

	return &TournamentStaffView{
		TournamentID: tournamentID,
		AccountID:    targetAccountID,
		Handle:       handle,
		DisplayName:  displayName,
		Role:         newRole,
		CreatedAt:    targetStaff.CreatedAt,
	}, nil
}

func (u *TournamentStaffUseCase) ListStaff(
	ctx context.Context,
	callerID, tournamentID uuid.UUID,
) ([]TournamentStaffView, error) {
	if err := u.checkCallerAuth(ctx, tournamentID, callerID); err != nil {
		return nil, err
	}

	staffList, err := u.repo.ListStaff(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	accountIDs := make([]uuid.UUID, 0, len(staffList))
	for _, s := range staffList {
		accountIDs = append(accountIDs, s.AccountID)
	}

	accountsMap, err := u.accountSvc.BatchGetAccounts(ctx, accountIDs)
	if err != nil {
		accountsMap = make(map[uuid.UUID]ports.AccountInfo)
	}

	views := make([]TournamentStaffView, 0, len(staffList))
	for _, s := range staffList {
		acc := accountsMap[s.AccountID]
		views = append(views, TournamentStaffView{
			TournamentID: s.TournamentID,
			AccountID:    s.AccountID,
			Handle:       acc.Handle,
			DisplayName:  acc.DisplayName,
			Role:         s.Role,
			CreatedAt:    s.CreatedAt,
		})
	}

	return views, nil
}

func (u *TournamentStaffUseCase) RemoveStaff(
	ctx context.Context,
	callerID, tournamentID, targetAccountID uuid.UUID,
) error {
	if err := u.checkCallerAuth(ctx, tournamentID, callerID); err != nil {
		return err
	}

	targetStaff, err := u.repo.GetStaff(ctx, tournamentID, targetAccountID)
	if err != nil {
		return err
	}
	if targetStaff == nil {
		return ErrStaffNotFound
	}
	if targetStaff.IsOwner() {
		return ErrCannotModifyOwner
	}

	if err := u.repo.RemoveStaff(ctx, tournamentID, targetAccountID); err != nil {
		if errors.Is(err, repositories.ErrStaffNotFound) {
			return ErrStaffNotFound
		}
		return err
	}

	return nil
}
