package events

import (
	"database/sql"
	"net/http"
	"strings"

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
		if isInvalidIDError(err) {
			common.GinError(c, http.StatusBadRequest, err.Error())
			return
		}
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
		if isInvalidIDError(err) {
			common.GinError(c, http.StatusBadRequest, err.Error())
			return
		}
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if event.TenantID != user.TenantID && user.Role != "platform_admin" {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: event})
}

func (h *Handler) PublicShow(c *gin.Context) {
	event, err := h.service.FindPublicBySlug(c.Request.Context(), c.Param("slug"))
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	status := strings.ToLower(strings.TrimSpace(event.Status))
	if status != "published" && status != "active" {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	common.GinJSON(c, http.StatusOK, common.Response{Data: event})
}

func (h *Handler) Update(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request UpdateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.service.Update(c.Request.Context(), user.TenantID, c.Param("id"), request)
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: event})
}

func isInvalidIDError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "invalid tenant id") || strings.Contains(message, "invalid event id")
}

func (h *Handler) Delete(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	err := h.service.Delete(c.Request.Context(), user.TenantID, c.Param("id"))
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
