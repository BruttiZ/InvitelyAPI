package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"invitely-api/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *gin.Context) {
	summary, err := h.service.EventSummary(c.Request.Context(), c.Param("eventID"))
	if err != nil {
		common.GinError(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: summary})
}
