package routes

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/analytics"
	"invitely-api/internal/auth"
	"invitely-api/internal/common"
	"invitely-api/internal/config"
	"invitely-api/internal/dashboard"
	"invitely-api/internal/events"
	"invitely-api/internal/guests"
	"invitely-api/internal/middleware"
	"invitely-api/internal/rsvp"
)

func Register(router *gin.Engine, cfg config.Config, db *sql.DB) {
	middleware.SetAllowedOrigin(cfg.CORSAllowedOrigins)
	router.Use(middleware.GinCORS(), gin.Logger(), gin.Recovery())

	authService := auth.NewService(auth.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, cfg.SupabaseServiceKey), db)
	authHandler := auth.NewHandler(authService)

	eventsHandler := events.NewHandler(events.NewService(events.NewPostgresRepository(db)))
	guestsHandler := guests.NewHandler(guests.NewService(guests.NewPostgresRepository(db)))
	rsvpHandler := rsvp.NewHandler(rsvp.NewService(rsvp.NewPostgresRepository(db)))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(db))
	analyticsHandler := analytics.NewHandler(analytics.NewService(db))

	router.GET("/health", func(c *gin.Context) {
		common.GinJSON(c, http.StatusOK, common.Response{Data: map[string]string{"status": "ok"}})
	})

	authRoutes := router.Group("/auth")
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.GET("/me", authHandler.Me)

	protected := router.Group("/")
	protected.Use(middleware.GinAuth(authService))
	protected.GET("/events", eventsHandler.List)
	protected.POST("/events", eventsHandler.Create)
	protected.GET("/events/:id", eventsHandler.Show)
	protected.GET("/guests", guestsHandler.List)
	protected.POST("/guests", guestsHandler.Create)
	protected.GET("/dashboard", dashboardHandler.Overview)
	protected.GET("/analytics/events/:eventID", analyticsHandler.Summary)

	router.POST("/rsvp", rsvpHandler.Submit)
}
