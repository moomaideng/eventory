package accountgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	accountv1 "github.com/moomaideng/eventory/services/account/pkg/proto/account/v1"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const defaultTimeout = 2 * time.Second

// AccountClient is a gRPC adapter for ports.AccountService.
type AccountClient struct {
	api accountv1.AccountServiceClient
}

func NewAccountClient(addr string) (*AccountClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial account gRPC: %w", err)
	}
	return &AccountClient{api: accountv1.NewAccountServiceClient(conn)}, nil
}

func NewAccountClientFromConn(conn grpc.ClientConnInterface) *AccountClient {
	return &AccountClient{api: accountv1.NewAccountServiceClient(conn)}
}

func (c *AccountClient) GetAccount(ctx context.Context, id uuid.UUID) (*ports.AccountInfo, error) {
	accounts, err := c.BatchGetAccounts(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	account, ok := accounts[id]
	if !ok {
		return nil, ports.ErrAccountNotFound
	}
	return &account, nil
}

func (c *AccountClient) BatchGetAccounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.AccountInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	resp, err := c.api.BatchGetAccounts(ctx, &accountv1.BatchGetAccountsRequest{
		AccountIds: uuidStrings(ids),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrAccountUnavailable, err)
	}
	out := make(map[uuid.UUID]ports.AccountInfo, len(resp.GetAccounts()))
	for _, account := range resp.GetAccounts() {
		id, err := uuid.Parse(account.GetId())
		if err != nil {
			continue
		}
		out[id] = ports.AccountInfo{
			ID:          id,
			Handle:      account.GetHandle(),
			DisplayName: account.GetDisplayName(),
			Status:      ports.AccountStatus(account.GetStatus()),
		}
	}
	return out, nil
}

func (c *AccountClient) BatchGetOrganizerProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.OrganizerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	resp, err := c.api.BatchGetOrganizerProfiles(ctx, &accountv1.BatchGetOrganizerProfilesRequest{
		OrganizerProfileIds: uuidStrings(ids),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrAccountUnavailable, err)
	}
	out := make(map[uuid.UUID]ports.OrganizerInfo, len(resp.GetProfiles()))
	for _, profile := range resp.GetProfiles() {
		info, err := toOrganizerInfo(profile)
		if err != nil {
			continue
		}
		out[info.ID] = info
	}
	return out, nil
}

func (c *AccountClient) GetOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*ports.OrganizerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	resp, err := c.api.GetOrganizerProfileByAccountID(ctx, &accountv1.GetOrganizerProfileByAccountIDRequest{
		AccountId: accountID.String(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ports.ErrOrganizerNotFound
		}
		return nil, fmt.Errorf("%w: %v", ports.ErrAccountUnavailable, err)
	}
	info, err := toOrganizerInfo(resp.GetProfile())
	if err != nil {
		return nil, ports.ErrOrganizerNotFound
	}
	return &info, nil
}

func toOrganizerInfo(profile *accountv1.OrganizerProfile) (ports.OrganizerInfo, error) {
	if profile == nil {
		return ports.OrganizerInfo{}, ports.ErrOrganizerNotFound
	}
	id, err := uuid.Parse(profile.GetId())
	if err != nil {
		return ports.OrganizerInfo{}, err
	}
	accountID, err := uuid.Parse(profile.GetAccountId())
	if err != nil {
		return ports.OrganizerInfo{}, err
	}
	return ports.OrganizerInfo{
		ID:            id,
		AccountID:     accountID,
		OrganizerName: profile.GetOrganizerName(),
	}, nil
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		out = append(out, id.String())
	}
	return out
}
