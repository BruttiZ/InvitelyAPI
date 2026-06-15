package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

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

	router := gin.New()
	routes.Register(router, cfg, db)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Println("invitely-api listening on :" + cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
