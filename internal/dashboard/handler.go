package dashboard

import (
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

func (h *Handler) Overview(c *gin.Context) {
	user, ok := middleware.GinUser(c)
	if !ok {
		common.GinError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	overview, err := h.service.Overview(c.Request.Context(), user.TenantID)
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: overview})
}
