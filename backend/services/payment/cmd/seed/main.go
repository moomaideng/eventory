package main

import (
	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/internal/seeds"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Environment != "development" && cfg.Environment != "local" {
		log.Fatal("Mock payment seed is local/development only")
	}
	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := seeds.SeedPayments(db); err != nil {
		log.Fatal(err)
	}
	log.Println("Mock payment fixtures created")
}
