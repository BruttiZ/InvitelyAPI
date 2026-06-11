package analytics

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
	if r.Method != http.MethodGet {
		common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/analytics/events/"), "/")
	if path == "" || path == r.URL.Path {
		common.Error(w, http.StatusNotFound, "event not found")
		return
	}

	h.Summary(w, r)
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	eventID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/analytics/events/"), "/")
	summary, err := h.service.EventSummary(r.Context(), eventID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: summary})
}
