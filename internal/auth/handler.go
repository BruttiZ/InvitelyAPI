package auth

import (
	"net/http"

	"invitely-api/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request LoginRequest
	if err := common.Decode(r, &request); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Login(r.Context(), request)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: response})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request RegisterRequest
	if err := common.Decode(r, &request); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Register(r.Context(), request)
	if err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	common.JSON(w, http.StatusCreated, common.Response{Data: response})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		common.Error(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	user, err := h.service.EnsureUserFromToken(r.Context(), token)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	common.JSON(w, http.StatusOK, common.Response{Data: user})
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}
