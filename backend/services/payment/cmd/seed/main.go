package main

import (
	"flag"
	"log"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/internal/seeds"
)

func main() {
	allowProd := flag.Bool("allow-prod", false, "Allow seeding mock payment fixtures outside development/local")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if !*allowProd && cfg.Environment != "development" && cfg.Environment != "local" {
		log.Fatal("Mock payment seed is local/development only (pass -allow-prod to override)")
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
