package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", corsOrigin(c.GetHeader("Origin")))
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
