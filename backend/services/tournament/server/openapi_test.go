package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	accountconfig "github.com/moomaideng/eventory/services/account/config"
	accountserver "github.com/moomaideng/eventory/services/account/server"
	paymentconfig "github.com/moomaideng/eventory/services/payment/config"
	paymentserver "github.com/moomaideng/eventory/services/payment/server"
	notificationconfig "github.com/moomaideng/eventory/services/notification/config"
	notificationserver "github.com/moomaideng/eventory/services/notification/server"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/gorm"
)

type dummyAccountSvc struct{}

func (d *dummyAccountSvc) GetAccount(_ context.Context, _ uuid.UUID) (*ports.AccountInfo, error) {
	return nil, nil
}
func (d *dummyAccountSvc) BatchGetAccounts(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]ports.AccountInfo, error) {
	return nil, nil
}
func (d *dummyAccountSvc) BatchGetOrganizerProfiles(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]ports.OrganizerInfo, error) {
	return nil, nil
}
func (d *dummyAccountSvc) GetOrganizerProfileByAccountID(_ context.Context, _ uuid.UUID) (*ports.OrganizerInfo, error) {
	return nil, nil
}
func (d *dummyAccountSvc) GetAccountByHandle(_ context.Context, _ string) (*ports.AccountInfo, error) {
	return nil, nil
}

func fetchOpenAPI(handler http.Handler) (map[string]any, error) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func TestTournamentOpenAPISpec(t *testing.T) {
	db := &gorm.DB{}

	tournamentCfg := config.Config{
		Environment: "development",
		HTTPPort:    "8082",
	}
	accSvc := &dummyAccountSvc{}
	teamLobbyRepo := repositories.NewTeamLobbyRepository(db)
	teamLobbyUseCase := usecases.NewTeamLobbyUseCase(teamLobbyRepo, accSvc, nil, nil, nil)
	sponsorships := usecases.NewSponsorshipUseCase(repositories.NewSponsorshipRepository(db), repositories.NewTournamentRepository(db), accSvc, nil)
	teamReviews := usecases.NewTeamReviewUseCase(repositories.NewTeamReviewRepository(db), accSvc, nil)
	opts := options.Client().ApplyURI("mongodb://localhost:27017").SetServerSelectionTimeout(50 * time.Millisecond)
	client, _ := mongo.Connect(opts)
	mongoDB := client.Database("tournament_db")
	tournamentRouter := newHTTPRouter(db, mongoDB, tournamentCfg, accSvc, teamLobbyUseCase, sponsorships, teamReviews)

	tournamentDoc, err := fetchOpenAPI(tournamentRouter)
	require.NoError(t, err)

	paths, ok := tournamentDoc["paths"].(map[string]any)
	require.True(t, ok)

	// Verify our new bracket endpoints are in the OpenAPI spec!
	assert.Contains(t, paths, "/api/v1/tournaments/{id}/bracket")
	assert.Contains(t, paths, "/api/v1/tournaments/{id}/seed")
	assert.Contains(t, paths, "/api/v1/tournaments/{id}/matches/{matchId}/results")
	assert.Contains(t, paths, "/api/v1/tournaments/sponsorships/mine")

	// If UPDATE_OPENAPI=1, merge account, tournament, and payment into frontend/.openapi.merged.json
	if os.Getenv("UPDATE_OPENAPI") == "1" {
		accountApp := accountserver.NewApp(db, accountconfig.Config{
			Environment: "development",
			HTTPPort:    "8081",
		})
		accountDoc, err := fetchOpenAPI(accountApp.HTTP)
		require.NoError(t, err)

		paymentApp := paymentserver.NewApp(db, paymentconfig.Config{
			Environment: "development",
			HTTPPort:    "8083",
			ServiceKey:  "01234567890123456789012345678901",
		})
		paymentDoc, err := fetchOpenAPI(paymentApp.HTTP)
		require.NoError(t, err)

		notificationApp := notificationserver.NewApp(db, nil, nil, notificationconfig.Config{
			Environment: "development",
			HTTPPort:    "8084",
		})
		notificationDoc, err := fetchOpenAPI(notificationApp.HTTP)
		require.NoError(t, err)

		merged := map[string]any{
			"openapi": "3.1.0",
			"info": map[string]any{
				"title":   "Eventory API",
				"version": "1.0.0",
			},
			"paths": make(map[string]any),
			"components": map[string]any{
				"schemas":         make(map[string]any),
				"securitySchemes": make(map[string]any),
			},
		}

		mergedPaths := merged["paths"].(map[string]any)
		mergedComponents := merged["components"].(map[string]any)
		mergedSchemas := mergedComponents["schemas"].(map[string]any)
		mergedSecSchemes := mergedComponents["securitySchemes"].(map[string]any)

		sources := []struct {
			name string
			doc  map[string]any
		}{
			{"account", accountDoc},
			{"tournament", tournamentDoc},
			{"payment", paymentDoc},
			{"notification", notificationDoc},
		}

		for _, src := range sources {
			srcPaths, _ := src.doc["paths"].(map[string]any)
			for p, op := range srcPaths {
				if len(p) >= 10 && p[:10] == "/internal/" {
					continue
				}
				if src.name == "payment" && (len(p) < 16 || p[:16] != "/api/v1/payments") {
					continue
				}
				if src.name == "notification" && (len(p) < 21 || p[:21] != "/api/v1/notifications") {
					continue
				}
				mergedPaths[p] = op
			}

			srcComponents, _ := src.doc["components"].(map[string]any)
			if srcComponents != nil {
				srcSchemas, _ := srcComponents["schemas"].(map[string]any)
				for k, v := range srcSchemas {
					mergedSchemas[k] = v
				}
				srcSec, _ := srcComponents["securitySchemes"].(map[string]any)
				for k, v := range srcSec {
					mergedSecSchemes[k] = v
				}
			}
		}

		mergedPath := filepath.Join("..", "..", "..", "..", "frontend", ".openapi.merged.json")
		mergedBytes, err := json.MarshalIndent(merged, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(mergedPath, mergedBytes, 0644))
	}
}
