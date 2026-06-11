package guests

import (
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
	path := strings.TrimPrefix(r.URL.Path, "/guests")
	if path == "" || path == "/" {
		if r.Method == http.MethodGet {
			h.List(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.Create(w, r)
			return
		}
	}

	common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	eventID := r.URL.Query().Get("event_id")
	guests, err := h.service.ListByEvent(r.Context(), user.TenantID, eventID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: guests})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var request CreateGuestRequest
	if err := common.Decode(r, &request); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	guest, err := h.service.Create(r.Context(), user.TenantID, request)
	if err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	common.JSON(w, http.StatusCreated, common.Response{Data: guest})
}
