package main

import (
	"log"
	"net/http"
	"time"

	"invitely-api/internal/config"
	"invitely-api/internal/database"
	"invitely-api/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	routes.Register(mux, cfg, db)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes.Chain(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Println("invitely-api listening on :" + cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
