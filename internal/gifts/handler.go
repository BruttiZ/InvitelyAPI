package gifts

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

	gifts, err := h.service.ListByEvent(c.Request.Context(), user.TenantID, c.Param("id"))
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: gifts})
}

func (h *Handler) Create(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request CreateGiftRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	gift, err := h.service.Create(c.Request.Context(), user.TenantID, c.Param("id"), request)
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusCreated, common.Response{Data: gift})
}

func (h *Handler) Update(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request UpdateGiftRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	gift, err := h.service.Update(c.Request.Context(), user.TenantID, c.Param("id"), request)
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "gift not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: gift})
}

func (h *Handler) Delete(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	err := h.service.Delete(c.Request.Context(), user.TenantID, c.Param("id"))
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "gift not found")
		return
	}
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
