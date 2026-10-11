package main

import (
	"flag"
	shareddb "github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/database"
	"log"
)

func main() {
	reset := flag.Bool("reset", false, "Drop the payment database public schema before migrating")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := shareddb.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	if *reset {
		if err := shareddb.ResetPublicSchema(db); err != nil {
			log.Fatal(err)
		}
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	log.Println("Payment database migration completed")
}
