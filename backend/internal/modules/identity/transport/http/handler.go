package http

import (
	"net/http"

	identityapp "architecture/backend/internal/modules/identity/application"
	"architecture/backend/internal/shared/transport/httpjson"
)

type Handler struct {
	service *identityapp.Service
}

func NewHandler(service *identityapp.Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusCreated, registerResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
	})
}
