package rsvp

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

func (h *Handler) Submit(c *gin.Context) {
	var request SubmitRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Submit(c.Request.Context(), request)
	if err != nil {
		common.GinError(c, http.StatusBadRequest, err.Error())
		return
	}

	common.GinJSON(c, http.StatusOK, common.Response{Data: response})
}
