package middlewares_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	sharedmiddlewares "github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
)

type stubAccounts struct {
	byAccount map[uuid.UUID]ports.OrganizerInfo
}

func (s *stubAccounts) GetAccount(context.Context, uuid.UUID) (*ports.AccountInfo, error) {
	return nil, ports.ErrAccountNotFound
}
func (s *stubAccounts) BatchGetAccounts(context.Context, []uuid.UUID) (map[uuid.UUID]ports.AccountInfo, error) {
	return map[uuid.UUID]ports.AccountInfo{}, nil
}
func (s *stubAccounts) BatchGetOrganizerProfiles(context.Context, []uuid.UUID) (map[uuid.UUID]ports.OrganizerInfo, error) {
	return map[uuid.UUID]ports.OrganizerInfo{}, nil
}
func (s *stubAccounts) GetOrganizerProfileByAccountID(_ context.Context, accountID uuid.UUID) (*ports.OrganizerInfo, error) {
	profile, ok := s.byAccount[accountID]
	if !ok {
		return nil, ports.ErrOrganizerNotFound
	}
	return &profile, nil
}
func (s *stubAccounts) GetAccountByHandle(context.Context, string) (*ports.AccountInfo, error) {
	return nil, ports.ErrAccountNotFound
}

func TestGetOrganizerProfile_Missing(t *testing.T) {
	_, err := middlewares.GetOrganizerProfile(context.Background())
	if err == nil {
		t.Fatal("expected error when organizer profile is missing from context")
	}
}

func TestGetOrganizerProfile_Present(t *testing.T) {
	profile := &ports.OrganizerInfo{AccountID: uuid.New(), OrganizerName: "Alice"}
	ctx := context.WithValue(context.Background(), middlewares.OrganizerProfileContextKey, profile)

	extracted, err := middlewares.GetOrganizerProfile(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extracted.AccountID != profile.AccountID || extracted.OrganizerName != "Alice" {
		t.Fatalf("unexpected profile extracted: %+v", extracted)
	}
}

func TestOrganizerMiddleware_HTTPGuard(t *testing.T) {
	accountID := uuid.New()
	accountSvc := &stubAccounts{byAccount: map[uuid.UUID]ports.OrganizerInfo{
		accountID: {AccountID: accountID, OrganizerName: "Organizer One"},
	}}

	_, api := humatest.New(t)
	orgMiddleware := middlewares.NewOrganizerMiddleware(api, accountSvc)

	huma.Register(api, huma.Operation{
		OperationID: "organizer-success",
		Method:      http.MethodGet,
		Path:        "/organizer-success",
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				next(huma.WithContext(ctx, context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, accountID.String())))
			},
			orgMiddleware.HumaMiddleware(),
		},
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return &struct{}{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "organizer-forbidden",
		Method:      http.MethodGet,
		Path:        "/organizer-forbidden",
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				next(huma.WithContext(ctx, context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, uuid.New().String())))
			},
			orgMiddleware.HumaMiddleware(),
		},
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return &struct{}{}, nil
	})

	t.Run("authorized organizer profile allows access", func(t *testing.T) {
		resp := api.Get("/organizer-success")
		if resp.Code < 200 || resp.Code >= 300 {
			t.Fatalf("expected 2xx for organizer, got %d", resp.Code)
		}
	})

	t.Run("missing organizer profile returns forbidden", func(t *testing.T) {
		resp := api.Get("/organizer-forbidden")
		if resp.Code != http.StatusForbidden {
			t.Fatalf("expected 403 without organizer profile, got %d", resp.Code)
		}
	})
}
