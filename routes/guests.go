package routes

import "net/http"

func RegisterGuestRoutes(mux *http.ServeMux, handler http.Handler) {
	mux.Handle("/guests/", handler)
}
