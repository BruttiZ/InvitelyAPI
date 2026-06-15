package routes

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/analytics"
	"invitely-api/internal/apidocs"
	"invitely-api/internal/auth"
	"invitely-api/internal/budget"
	"invitely-api/internal/common"
	"invitely-api/internal/config"
	"invitely-api/internal/dashboard"
	"invitely-api/internal/events"
	"invitely-api/internal/gifts"
	"invitely-api/internal/guests"
	"invitely-api/internal/middleware"
	"invitely-api/internal/reminders"
	"invitely-api/internal/rsvp"
)

func Register(router *gin.Engine, cfg config.Config, db *sql.DB) {
	middleware.SetAllowedOrigin(cfg.CORSAllowedOrigins)
	router.Use(middleware.GinCORS(), gin.Logger(), gin.Recovery())

	authService := auth.NewService(auth.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, cfg.SupabaseServiceKey), db)
	authHandler := auth.NewHandler(authService)

	eventsHandler := events.NewHandler(events.NewService(events.NewPostgresRepository(db)))
	guestsHandler := guests.NewHandler(guests.NewService(guests.NewPostgresRepository(db)))
	budgetHandler := budget.NewHandler(budget.NewService(budget.NewPostgresRepository(db)))
	giftsHandler := gifts.NewHandler(gifts.NewService(gifts.NewPostgresRepository(db)))
	remindersHandler := reminders.NewHandler(reminders.NewService(
		reminders.NewPostgresRepository(db),
		events.NewPostgresRepository(db),
		reminders.NewSender(reminders.SenderConfig{
			APIKey:       cfg.BrevoAPIKey,
			FromName:     cfg.BrevoFromName,
			SMTPHost:     cfg.BrevoSMTPHost,
			SMTPPort:     cfg.BrevoSMTPPort,
			SMTPUsername: cfg.BrevoSMTPUsername,
			SMTPKey:      cfg.BrevoSMTPKey,
		}),
	))
	rsvpHandler := rsvp.NewHandler(rsvp.NewService(rsvp.NewPostgresRepository(db)))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(db))
	analyticsHandler := analytics.NewHandler(analytics.NewService(db))

	router.GET("/health", func(c *gin.Context) {
		common.GinJSON(c, http.StatusOK, common.Response{Data: map[string]string{"status": "ok"}})
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger")
	})
	router.GET("/swagger", swaggerUI)
	router.GET("/swagger/", swaggerUI)
	router.GET("/swagger/index.html", swaggerUI)
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		document, err := apidocs.SwaggerJSON()
		if err != nil {
			common.GinError(c, http.StatusInternalServerError, "failed to render swagger json")
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(document))
	})
	router.GET("/swagger/doc.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", []byte(apidocs.SwaggerYAML))
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
	protected.PUT("/events/:id", eventsHandler.Update)
	protected.DELETE("/events/:id", eventsHandler.Delete)
	protected.POST("/events/:id/reminders", remindersHandler.Send)
	protected.GET("/events/:id/budget", budgetHandler.List)
	protected.POST("/events/:id/budget", budgetHandler.Create)
	protected.PUT("/budget/:id", budgetHandler.Update)
	protected.DELETE("/budget/:id", budgetHandler.Delete)
	protected.GET("/events/:id/gifts", giftsHandler.List)
	protected.POST("/events/:id/gifts", giftsHandler.Create)
	protected.PUT("/gifts/:id", giftsHandler.Update)
	protected.DELETE("/gifts/:id", giftsHandler.Delete)
	protected.GET("/guests", guestsHandler.List)
	protected.POST("/guests", guestsHandler.Create)
	protected.GET("/dashboard", dashboardHandler.Overview)
	protected.GET("/analytics/events/:eventID", analyticsHandler.Summary)

	router.POST("/rsvp", rsvpHandler.Submit)
}

func swaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(apidocs.SwaggerHTML))
}
