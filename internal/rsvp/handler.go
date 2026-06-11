package rsvp

import (
	"net/http"
	"strings"

	"invitely-api/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rsvp")
	if path == "" || path == "/" {
		if r.Method == http.MethodPost {
			h.Submit(w, r)
			return
		}
	}

	common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var request SubmitRequest
	if err := common.Decode(r, &request); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Submit(r.Context(), request)
	if err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: response})
}
