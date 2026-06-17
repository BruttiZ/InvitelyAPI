package rsvp

import (
	"database/sql"
	"errors"
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

func (h *Handler) SubmitPublic(c *gin.Context) {
	var request PublicSubmitRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.GinError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.SubmitPublic(c.Request.Context(), c.Param("slug"), c.ClientIP(), request)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			common.GinError(c, http.StatusNotFound, "event not found")
		case errors.Is(err, ErrPublicRSVPClosed):
			common.GinError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrTooManyCompanions):
			common.GinError(c, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrRateLimited):
			common.GinError(c, http.StatusTooManyRequests, err.Error())
		default:
			common.GinError(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	message := "Resposta registrada."
	if response.Status == "accepted" {
		message = "Presença confirmada."
	}
	common.GinJSON(c, http.StatusCreated, common.Response{Message: message, Data: response})
}
