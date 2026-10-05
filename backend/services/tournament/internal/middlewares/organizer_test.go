package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
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

func TestGetOrganizerProfile_Missing(t *testing.T) {
	_, err := middlewares.GetOrganizerProfile(context.Background())
	if err == nil {
		t.Fatal("expected error when organizer profile is missing from context")
	}
}

func TestGetOrganizerProfile_Present(t *testing.T) {
	profile := &ports.OrganizerInfo{ID: uuid.New(), OrganizerName: "Alice"}
	ctx := context.WithValue(context.Background(), middlewares.OrganizerProfileContextKey, profile)

	extracted, err := middlewares.GetOrganizerProfile(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extracted.ID != profile.ID || extracted.OrganizerName != "Alice" {
		t.Fatalf("unexpected profile extracted: %+v", extracted)
	}
}

func TestOrganizerMiddleware_HTTPGuard(t *testing.T) {
	accountID := uuid.New()
	accountSvc := &stubAccounts{byAccount: map[uuid.UUID]ports.OrganizerInfo{
		accountID: {ID: uuid.New(), AccountID: accountID, OrganizerName: "Organizer One"},
	}}

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	orgMiddleware := middlewares.NewOrganizerMiddleware(api, accountSvc)
	group := huma.NewGroup(api, "/organizer-only")
	group.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		next(huma.WithContext(ctx, context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, accountID.String())))
	})
	group.UseMiddleware(orgMiddleware.HumaMiddleware())
	huma.Register(group, huma.Operation{OperationID: "ok", Method: http.MethodGet, Path: ""},
		func(ctx context.Context, input *struct{}) (*struct{}, error) { return &struct{}{}, nil })

	req := httptest.NewRequest(http.MethodGet, "/organizer-only", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code < 200 || rec.Code >= 300 {
		t.Fatalf("expected 2xx for organizer, got %d", rec.Code)
	}

	missingRouter := chi.NewMux()
	missingAPI := humachi.New(missingRouter, huma.DefaultConfig("Test API", "1.0.0"))
	missingGroup := huma.NewGroup(missingAPI, "/organizer-only")
	missingGroup.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		next(huma.WithContext(ctx, context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, uuid.New().String())))
	})
	missingGroup.UseMiddleware(middlewares.NewOrganizerMiddleware(missingAPI, accountSvc).HumaMiddleware())
	huma.Register(missingGroup, huma.Operation{OperationID: "missing", Method: http.MethodGet, Path: ""},
		func(ctx context.Context, input *struct{}) (*struct{}, error) { return &struct{}{}, nil })

	req = httptest.NewRequest(http.MethodGet, "/organizer-only", nil)
	rec = httptest.NewRecorder()
	missingRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without organizer profile, got %d", rec.Code)
	}
}
