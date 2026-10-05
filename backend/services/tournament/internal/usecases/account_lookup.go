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
		if tournament.OrganizerID == uuid.Nil {
			continue
		}
		if _, ok := seen[tournament.OrganizerID]; ok {
			continue
		}
		seen[tournament.OrganizerID] = struct{}{}
		ids = append(ids, tournament.OrganizerID)
	}
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	profiles, err := accountSvc.BatchGetOrganizerProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[uuid.UUID]string, len(profiles))
	for id, profile := range profiles {
		names[id] = profile.OrganizerName
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
