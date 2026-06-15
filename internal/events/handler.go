package events

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/common"
	"invitely-api/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	events, err := h.service.List(c.Request.Context(), user.TenantID)
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: events})
}

func (h *Handler) Create(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request CreateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.service.Create(c.Request.Context(), user.TenantID, request)
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusCreated, common.Response{Data: event})
}

func (h *Handler) Show(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	event, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if event.TenantID != user.TenantID && user.Role != "platform_admin" {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: event})
}
