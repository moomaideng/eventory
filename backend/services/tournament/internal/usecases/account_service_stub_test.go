package usecases_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
)

type stubAccountService struct {
	accounts   map[uuid.UUID]ports.AccountInfo
	organizers map[uuid.UUID]ports.OrganizerInfo // by profile ID
	byAccount  map[uuid.UUID]ports.OrganizerInfo // by account ID
	sponsors   map[uuid.UUID]ports.SponsorInfo   // by account ID
}

func newStubAccountService() *stubAccountService {
	return &stubAccountService{
		accounts:   make(map[uuid.UUID]ports.AccountInfo),
		organizers: make(map[uuid.UUID]ports.OrganizerInfo),
		byAccount:  make(map[uuid.UUID]ports.OrganizerInfo),
		sponsors:   make(map[uuid.UUID]ports.SponsorInfo),
	}
}

func (s *stubAccountService) withAccount(id uuid.UUID, handle, displayName string) *stubAccountService {
	s.accounts[id] = ports.AccountInfo{
		ID: id, Handle: handle, DisplayName: displayName, Status: ports.AccountStatusActive,
	}
	return s
}

func (s *stubAccountService) withAccountEmail(id uuid.UUID, handle, displayName, email string) *stubAccountService {
	s.accounts[id] = ports.AccountInfo{
		ID: id, Handle: handle, DisplayName: displayName, Email: email, Status: ports.AccountStatusActive,
	}
	return s
}

func (s *stubAccountService) withAccountStatus(id uuid.UUID, handle, displayName string, status ports.AccountStatus) *stubAccountService {
	s.accounts[id] = ports.AccountInfo{
		ID: id, Handle: handle, DisplayName: displayName, Status: status,
	}
	return s
}

func (s *stubAccountService) withOrganizer(accountID uuid.UUID, organizerName string) *stubAccountService {
	info := ports.OrganizerInfo{
		AccountID:     accountID,
		OrganizerName: organizerName,
	}
	s.organizers[accountID] = info
	s.byAccount[accountID] = info
	return s
}

func (s *stubAccountService) withSponsor(accountID uuid.UUID, sponsorName, websiteURL, logoURL string) *stubAccountService {
	s.sponsors[accountID] = ports.SponsorInfo{
		AccountID: accountID, SponsorName: sponsorName, WebsiteURL: websiteURL, LogoURL: logoURL,
	}
	return s
}

func (s *stubAccountService) GetAccount(ctx context.Context, id uuid.UUID) (*ports.AccountInfo, error) {
	if account, ok := s.accounts[id]; ok {
		return &account, nil
	}
	// Default active account for unit tests that only care about persistence rules.
	return &ports.AccountInfo{
		ID: id, Handle: "stub", DisplayName: "Stub User", Status: ports.AccountStatusActive,
	}, nil
}

func (s *stubAccountService) BatchGetAccounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.AccountInfo, error) {
	out := make(map[uuid.UUID]ports.AccountInfo)
	for _, id := range ids {
		if account, ok := s.accounts[id]; ok {
			out[id] = account
			continue
		}
		out[id] = ports.AccountInfo{
			ID: id, Handle: "stub", DisplayName: "Stub User", Status: ports.AccountStatusActive,
		}
	}
	return out, nil
}

func (s *stubAccountService) BatchGetOrganizerProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.OrganizerInfo, error) {
	out := make(map[uuid.UUID]ports.OrganizerInfo)
	for _, id := range ids {
		if profile, ok := s.organizers[id]; ok {
			out[id] = profile
		}
	}
	return out, nil
}

func (s *stubAccountService) GetOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*ports.OrganizerInfo, error) {
	profile, ok := s.byAccount[accountID]
	if !ok {
		return nil, ports.ErrOrganizerNotFound
	}
	return &profile, nil
}

func (s *stubAccountService) BatchGetSponsorProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.SponsorInfo, error) {
	out := make(map[uuid.UUID]ports.SponsorInfo)
	for _, id := range ids {
		if profile, ok := s.sponsors[id]; ok {
			out[id] = profile
		}
	}
	return out, nil
}

func (s *stubAccountService) GetSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*ports.SponsorInfo, error) {
	profile, ok := s.sponsors[accountID]
	if !ok {
		return nil, ports.ErrSponsorNotFound
	}
	return &profile, nil
}

func (s *stubAccountService) GetAccountByHandle(ctx context.Context, handle string) (*ports.AccountInfo, error) {
	for _, acc := range s.accounts {
		if acc.Handle == handle {
			return &acc, nil
		}
	}
	return nil, ports.ErrAccountNotFound
}
