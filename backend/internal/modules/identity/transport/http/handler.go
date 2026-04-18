package http

import (
	"log"
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

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpjson.Decode(r, &req); err != nil {
		log.Printf("identity.register: invalid json error=%v", err)
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("identity.register: failed email=%s error=%v", req.Email, err)
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("identity.register: success user_id=%s email=%s", user.ID, user.Email)

	httpjson.Write(w, http.StatusCreated, registerResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpjson.Decode(r, &req); err != nil {
		log.Printf("identity.login: invalid json error=%v", err)
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("identity.login: failed email=%s error=%v", req.Email, err)
		httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: err.Error()})
		return
	}
	log.Printf("identity.login: success email=%s", req.Email)

	httpjson.Write(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}
