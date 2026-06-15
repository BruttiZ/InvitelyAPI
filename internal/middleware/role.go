package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/common"
)

func GinRequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := GinUser(c)
		if !ok {
			common.GinError(c, http.StatusUnauthorized, "unauthenticated")
			c.Abort()
			return
		}
		if user.Role != role && user.Role != "platform_admin" {
			common.GinError(c, http.StatusForbidden, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
