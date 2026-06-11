package middleware

import (
	"net/http"

	"invitely-api/internal/common"
)

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				common.Error(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if user.Role != role && user.Role != "platform_admin" {
				common.Error(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
