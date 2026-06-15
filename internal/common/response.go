package common

import "github.com/gin-gonic/gin"

type Response struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func GinJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func GinError(c *gin.Context, status int, message string) {
	GinJSON(c, status, Response{Error: message})
}
