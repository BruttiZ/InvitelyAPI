package events

import (
	"database/sql"
	"net/http"
	"strings"

	"invitely-api/internal/common"
	"invitely-api/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/events")
	switch {
	case path == "" || path == "/":
		if r.Method == http.MethodGet {
			h.List(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.Create(w, r)
			return
		}
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		h.Show(w, r)
		return
	}

	common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	events, err := h.service.List(r.Context(), user.TenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: events})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request CreateEventRequest
	if err := common.Decode(r, &request); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.service.Create(r.Context(), user.TenantID, request)
	if err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	common.JSON(w, http.StatusCreated, common.Response{Data: event})
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/events/"), "/")
	event, err := h.service.FindByID(r.Context(), id)
	if err == sql.ErrNoRows {
		common.Error(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if event.TenantID != user.TenantID && user.Role != "platform_admin" {
		common.Error(w, http.StatusNotFound, "event not found")
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: event})
}
