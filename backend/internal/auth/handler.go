package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	"github.com/Ssnakerss/modmas/pkg/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "auth.Register")
		response.BadRequest(w, "invalid request body")
		return
	}

	res, err := h.service.Register(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to register user", "error", err, "handler", "auth.Register")
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "auth.Login")
		response.BadRequest(w, "invalid request body")
		return
	}

	res, err := h.service.Login(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to login user", "error", err, "handler", "auth.Login")
		response.Unauthorized(w, err.Error())
		return
	}

	response.OK(w, res)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "handler", "auth.Me", "userId", userID)
		response.NotFound(w, "user not found")
		return
	}
	response.OK(w, user)
}
