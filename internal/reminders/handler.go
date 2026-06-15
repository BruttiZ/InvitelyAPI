package reminders

import (
	"database/sql"
	"errors"
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

func (h *Handler) Send(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request SendRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, issues, err := h.service.Send(c.Request.Context(), user.TenantID, c.Param("id"), request)
	if len(issues) > 0 {
		common.GinJSON(c, http.StatusUnprocessableEntity, common.Response{
			Error: "validation failed",
			Data:  issues,
		})
		return
	}
	if err == sql.ErrNoRows {
		common.GinError(c, http.StatusNotFound, "event not found")
		return
	}
	if err == common.ErrForbidden {
		common.GinError(c, http.StatusForbidden, "forbidden")
		return
	}
	if err == ErrEmailProviderUnavailable {
		common.GinError(c, http.StatusServiceUnavailable, "email provider unavailable")
		return
	}
	var providerError ProviderError
	if errors.As(err, &providerError) {
		status := http.StatusBadGateway
		if providerError.StatusCode == 0 {
			status = http.StatusServiceUnavailable
		}
		common.GinJSON(c, status, common.Response{
			Error: "email provider failed",
			Data: map[string]any{
				"provider_status": providerError.StatusCode,
				"provider_error":  providerError.SafeBody(),
			},
		})
		return
	}
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.GinJSON(c, http.StatusAccepted, common.Response{Data: response})
}
