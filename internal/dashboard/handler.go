package dashboard

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
	path := strings.TrimPrefix(r.URL.Path, "/dashboard")
	if (path == "" || path == "/") && r.Method == http.MethodGet {
		h.Overview(w, r)
		return
	}

	common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		common.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	overview, err := h.service.Overview(r.Context(), user.TenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: overview})
}
