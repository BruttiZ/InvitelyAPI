package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/common"
)

func GinAPIKey(expectedKey string) gin.HandlerFunc {
	expectedKey = strings.TrimSpace(expectedKey)

	return func(c *gin.Context) {
		if expectedKey == "" {
			common.GinError(c, http.StatusInternalServerError, "api key is not configured")
			c.Abort()
			return
		}

		providedKey := strings.TrimSpace(c.GetHeader("x-api-key"))
		if providedKey == "" {
			common.GinError(c, http.StatusUnauthorized, "missing api key")
			c.Abort()
			return
		}

		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
			common.GinError(c, http.StatusUnauthorized, "invalid api key")
			c.Abort()
			return
		}

		c.Next()
	}
}
