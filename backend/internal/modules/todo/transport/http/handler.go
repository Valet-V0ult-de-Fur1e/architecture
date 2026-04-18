package http

import (
	"net/http"

	todoapp "architecture/backend/internal/modules/todo/application"
	tododomain "architecture/backend/internal/modules/todo/domain"
	"architecture/backend/internal/shared/transport/httpjson"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *todoapp.Service
}

func NewHandler(service *todoapp.Service) *Handler {
	return &Handler{service: service}
}

type createTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

type updatePriorityRequest struct {
	Priority int `json:"priority"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	var req createTodoRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	todo, err := h.service.Create(r.Context(), userID, req.Title, req.Description, req.Priority)
	if err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusCreated, map[string]any{"todo": todo})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	todoID, ok := parseTodoID(w, r)
	if !ok {
		return
	}

	todo, err := h.service.GetByID(r.Context(), userID, todoID)
	if err != nil {
		httpjson.Write(w, http.StatusNotFound, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{"todo": todo})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	todos, err := h.service.List(r.Context(), userID)
	if err != nil {
		httpjson.Write(w, http.StatusInternalServerError, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{"todos": todos})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	todoID, ok := parseTodoID(w, r)
	if !ok {
		return
	}

	var req updateStatusRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	if err := h.service.UpdateStatus(r.Context(), userID, todoID, tododomain.Status(req.Status)); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	todoID, ok := parseTodoID(w, r)
	if !ok {
		return
	}

	var req updatePriorityRequest
	if err := httpjson.Decode(r, &req); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid json"})
		return
	}

	if err := h.service.UpdatePriority(r.Context(), userID, todoID, req.Priority); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]string{"priority": "updated"})
}

func (h *Handler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	todoID, ok := parseTodoID(w, r)
	if !ok {
		return
	}

	if err := h.service.SoftDelete(r.Context(), userID, todoID); err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userIDRaw := r.Header.Get("X-User-ID")
	if userIDRaw == "" {
		httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: "X-User-ID header is required"})
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: "invalid X-User-ID header"})
		return uuid.Nil, false
	}

	return userID, true
}

func parseTodoID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	todoIDRaw := chi.URLParam(r, "id")
	todoID, err := uuid.Parse(todoIDRaw)
	if err != nil {
		httpjson.Write(w, http.StatusBadRequest, httpjson.ErrorResponse{Error: "invalid todo id"})
		return uuid.Nil, false
	}

	return todoID, true
}
