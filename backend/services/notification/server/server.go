package server

import (
	"context"
	"net/http"

	"github.com/moomaideng/eventory/internal/middlewares"
	"gorm.io/gorm"
)

type BrokerHealth interface{ Ready() bool }

func New(db *gorm.DB, broker BrokerHealth, slowRequestMS int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("GET /health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("GET /ready", func(writer http.ResponseWriter, request *http.Request) {
		if !databaseReady(request.Context(), db) || !broker.Ready() {
			http.Error(writer, "not ready", http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
	return middlewares.RequestID(middlewares.AccessLog(slowRequestMS)(middlewares.Recover(router)))
}

func databaseReady(ctx context.Context, db *gorm.DB) bool {
	sqlDB, err := db.DB()
	return err == nil && sqlDB.PingContext(ctx) == nil
}
