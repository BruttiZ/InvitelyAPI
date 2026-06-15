package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/auth"
	"invitely-api/internal/common"
)

type userContextKey struct{}

func UserFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(auth.User)
	return user, ok
}

func GinAuth(service *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || token == c.GetHeader("Authorization") {
			common.GinError(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		user, err := service.EnsureUserFromToken(c.Request.Context(), token)
		if err != nil {
			common.GinError(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set("user", user)
		ctx := context.WithValue(c.Request.Context(), userContextKey{}, user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GinUser(c *gin.Context) (auth.User, bool) {
	if value, ok := c.Get("user"); ok {
		user, ok := value.(auth.User)
		return user, ok
	}

	return UserFromContext(c.Request.Context())
}

func GinRequireAuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := GinUser(c); !ok {
			common.GinError(c, http.StatusUnauthorized, "unauthenticated")
			c.Abort()
			return
		}
		c.Next()
	}
}
