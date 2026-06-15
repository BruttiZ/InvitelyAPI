package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		common.GinError(c, http.StatusUnauthorized, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: response})
}

func (h *Handler) Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Register(c.Request.Context(), request)
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusCreated, common.Response{Data: response})
}

func (h *Handler) Me(c *gin.Context) {
	token := bearerToken(c)
	if token == "" {
		common.GinError(c, http.StatusUnauthorized, "missing bearer token")
		return
	}

	user, err := h.service.EnsureUserFromToken(c.Request.Context(), token)
	if err != nil {
		common.GinError(c, http.StatusUnauthorized, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: user})
}

func bearerToken(c *gin.Context) string {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token == c.GetHeader("Authorization") {
		return ""
	}
	return token
}
