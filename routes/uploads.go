package routes

import "net/http"

func RegisterUploadRoutes(mux *http.ServeMux, handler http.Handler) {
	mux.Handle("/uploads/", handler)
}
