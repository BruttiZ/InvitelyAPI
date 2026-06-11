package routes

import "net/http"

func RegisterAuthRoutes(mux *http.ServeMux, handler http.Handler) {
	mux.Handle("/auth/", handler)
}
