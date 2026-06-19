package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/config"
	"invitely-api/internal/database"
	"invitely-api/routes"
)

func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" && cfg.APIKey == "" {
		log.Fatal("INVITELY_API_KEY is required in production")
	}

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
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("shutting down invitely-api")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
