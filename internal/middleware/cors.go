package middleware

import (
	"net/http"
	"strings"
)

var allowedOrigins = []string{"*"}

func SetAllowedOrigin(origin string) {
	origin = strings.TrimSpace(strings.TrimPrefix(origin, "CORS_ALLOWED_ORIGINS="))
	if origin == "" {
		return
	}

	allowedOrigins = []string{}
	for _, item := range strings.Split(origin, ",") {
		item = strings.TrimRight(strings.TrimSpace(item), "/")
		if item != "" {
			allowedOrigins = append(allowedOrigins, item)
		}
	}

	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin(r.Header.Get("Origin")))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func corsOrigin(requestOrigin string) string {
	requestOrigin = strings.TrimRight(strings.TrimSpace(requestOrigin), "/")
	for _, origin := range allowedOrigins {
		if origin == "*" {
			return "*"
		}
		if requestOrigin != "" && origin == requestOrigin {
			return requestOrigin
		}
	}

	return allowedOrigins[0]
}
