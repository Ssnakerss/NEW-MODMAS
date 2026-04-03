package workspace

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	"github.com/Ssnakerss/modmas/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	workspaces, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list workspaces", "error", err, "handler", "workspace.List", "userId", userID)
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "workspace.Create")
		response.BadRequest(w, "invalid request body")
		return
	}
	if input.Name == "" {
		h.logger.Error("validation failed", "error", "name is required", "handler", "workspace.Create")
		response.BadRequest(w, "name is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ws, err := h.service.Create(r.Context(), input.Name, userID)
	if err != nil {
		h.logger.Error("failed to create workspace", "error", err, "handler", "workspace.Create", "userId", userID)
		response.InternalError(w, err.Error())
		return
	}
	response.Created(w, ws)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ws, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get workspace", "error", err, "handler", "workspace.Get", "id", id)
		response.NotFound(w, "workspace not found")
		return
	}
	response.OK(w, ws)
}
