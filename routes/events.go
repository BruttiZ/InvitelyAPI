package routes

import "net/http"

func RegisterEventRoutes(mux *http.ServeMux, handler http.Handler) {
	mux.Handle("/events/", handler)
}
