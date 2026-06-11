package routes

import (
	"database/sql"
	"net/http"

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

func Register(mux *http.ServeMux, cfg config.Config, db *sql.DB) {
	middleware.SetAllowedOrigin(cfg.CORSAllowedOrigins)

	authService := auth.NewService(auth.NewSupabaseClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, cfg.SupabaseServiceKey), db)
	authHandler := auth.NewHandler(authService)

	eventsHandler := events.NewHandler(events.NewService(events.NewPostgresRepository(db)))
	guestsHandler := guests.NewHandler(guests.NewService(guests.NewPostgresRepository(db)))
	rsvpHandler := rsvp.NewHandler(rsvp.NewService(rsvp.NewPostgresRepository(db)))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(db))
	analyticsHandler := analytics.NewHandler(analytics.NewService(db))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		common.JSON(w, http.StatusOK, common.Response{Data: map[string]string{"status": "ok"}})
	})

	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/me", authHandler.Me)

	protected := middleware.Auth(authService)
	mux.Handle("/events", protected(eventsHandler))
	mux.Handle("/events/", protected(eventsHandler))
	mux.Handle("/guests", protected(guestsHandler))
	mux.Handle("/guests/", protected(guestsHandler))
	mux.Handle("/dashboard", protected(dashboardHandler))
	mux.Handle("/dashboard/", protected(dashboardHandler))
	mux.Handle("/analytics/", protected(analyticsHandler))

	mux.Handle("/rsvp", rsvpHandler)
	mux.Handle("/rsvp/", rsvpHandler)
}

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return middleware.Recovery(middleware.Logger(middleware.CORS(handler)))
}
