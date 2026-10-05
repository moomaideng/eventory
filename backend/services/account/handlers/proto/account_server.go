package proto

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/account/internal/usecases"
	"github.com/moomaideng/eventory/services/account/models"
	accountv1 "github.com/moomaideng/eventory/services/account/pkg/proto/account/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountServer struct {
	accountv1.UnimplementedAccountServiceServer
	accounts *usecases.AccountUseCase
}

func NewAccountServer(accounts *usecases.AccountUseCase) *AccountServer {
	return &AccountServer{accounts: accounts}
}

func (s *AccountServer) BatchGetAccounts(ctx context.Context, req *accountv1.BatchGetAccountsRequest) (*accountv1.BatchGetAccountsResponse, error) {
	ids, err := parseIDs(req.GetAccountIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_ids: %v", err)
	}
	found, missing, err := s.accounts.BatchGetAccounts(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "batch get accounts: %v", err)
	}
	resp := &accountv1.BatchGetAccountsResponse{
		Accounts:    make([]*accountv1.Account, 0, len(found)),
		NotFoundIds: uuidStrings(missing),
	}
	for _, account := range found {
		resp.Accounts = append(resp.Accounts, toProtoAccount(&account))
	}
	return resp, nil
}

func (s *AccountServer) BatchGetOrganizerProfiles(ctx context.Context, req *accountv1.BatchGetOrganizerProfilesRequest) (*accountv1.BatchGetOrganizerProfilesResponse, error) {
	ids, err := parseIDs(req.GetOrganizerProfileIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organizer_profile_ids: %v", err)
	}
	found, missing, err := s.accounts.BatchGetOrganizerProfiles(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "batch get organizer profiles: %v", err)
	}
	resp := &accountv1.BatchGetOrganizerProfilesResponse{
		Profiles:    make([]*accountv1.OrganizerProfile, 0, len(found)),
		NotFoundIds: uuidStrings(missing),
	}
	for _, profile := range found {
		resp.Profiles = append(resp.Profiles, toProtoOrganizer(&profile))
	}
	return resp, nil
}

func (s *AccountServer) GetOrganizerProfileByAccountID(ctx context.Context, req *accountv1.GetOrganizerProfileByAccountIDRequest) (*accountv1.GetOrganizerProfileByAccountIDResponse, error) {
	accountID, err := uuid.Parse(req.GetAccountId())
	if err != nil || accountID == uuid.Nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id")
	}
	profile, err := s.accounts.GetOrganizerProfile(ctx, accountID)
	if err != nil {
		if errors.Is(err, usecases.ErrOrganizerProfileNotFound) {
			return nil, status.Error(codes.NotFound, "organizer profile not found")
		}
		return nil, status.Errorf(codes.Internal, "get organizer profile: %v", err)
	}
	return &accountv1.GetOrganizerProfileByAccountIDResponse{Profile: toProtoOrganizer(profile)}, nil
}

func toProtoAccount(account *models.Account) *accountv1.Account {
	return &accountv1.Account{
		Id:          account.ID.String(),
		Handle:      account.Handle,
		DisplayName: account.DisplayName,
		Status:      string(account.Status),
	}
}

func toProtoOrganizer(profile *models.OrganizerProfile) *accountv1.OrganizerProfile {
	return &accountv1.OrganizerProfile{
		Id:            profile.ID.String(),
		AccountId:     profile.AccountID.String(),
		OrganizerName: profile.OrganizerName,
	}
}

func parseIDs(raw []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
