package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
)

func organizerNames(ctx context.Context, accountSvc ports.AccountService, tournaments ...models.Tournament) (map[uuid.UUID]string, error) {
	ids := make([]uuid.UUID, 0, len(tournaments))
	seen := make(map[uuid.UUID]struct{}, len(tournaments))
	for _, tournament := range tournaments {
		ownerID := tournament.OwnerID()
		if ownerID == uuid.Nil {
			continue
		}
		if _, ok := seen[ownerID]; ok {
			continue
		}
		seen[ownerID] = struct{}{}
		ids = append(ids, ownerID)
	}
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	profiles, err := accountSvc.BatchGetOrganizerProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[uuid.UUID]string, len(tournaments))
	for _, tournament := range tournaments {
		ownerID := tournament.OwnerID()
		if profile, ok := profiles[ownerID]; ok {
			names[tournament.ID] = profile.OrganizerName
		}
	}
	return names, nil
}

func accountInfos(ctx context.Context, accountSvc ports.AccountService, members ...models.TournamentTeamMember) (map[uuid.UUID]ports.AccountInfo, error) {
	ids := make([]uuid.UUID, 0, len(members))
	seen := make(map[uuid.UUID]struct{}, len(members))
	for _, member := range members {
		if member.AccountID == uuid.Nil {
			continue
		}
		if _, ok := seen[member.AccountID]; ok {
			continue
		}
		seen[member.AccountID] = struct{}{}
		ids = append(ids, member.AccountID)
	}
	if len(ids) == 0 {
		return map[uuid.UUID]ports.AccountInfo{}, nil
	}
	return accountSvc.BatchGetAccounts(ctx, ids)
}

func requireActiveAccount(ctx context.Context, accountSvc ports.AccountService, accountID uuid.UUID) (*ports.AccountInfo, error) {
	account, err := accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Status == ports.AccountStatusOnboarding {
		return nil, ports.ErrAccountOnboarding
	}
	return account, nil
}
