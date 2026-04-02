package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	"github.com/Ssnakerss/modmas/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	res, err := h.service.Register(r.Context(), input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	res, err := h.service.Login(r.Context(), input)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.OK(w, res)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}
	response.OK(w, user)
}
