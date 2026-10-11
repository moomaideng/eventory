package main

import (
	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/server"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	app := server.NewApp(db, cfg)
	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: app.HTTP, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("Payment API (mock) listening on %s", httpServer.Addr)
	log.Fatal(httpServer.ListenAndServe())
}
