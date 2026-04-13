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

// NewHandler создает новый экземпляр Handler для обработки HTTP-запросов аутентификации
// Принимает сервис аутентификации и логгер для записи сообщений
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// Register обрабатывает HTTP-запрос на регистрацию нового пользователя
// Декодирует входные данные из тела запроса, валидирует их и передает в сервис
// Возвращает 201 Created с данными пользователя и токеном, или 400 Bad Request в случае ошибки
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

// Login обрабатывает HTTP-запрос на аутентификацию пользователя
// Проверяет логин и пароль, генерирует JWT-токен при успешной аутентификации
// Возвращает 200 OK с данными пользователя и токеном, или 401 Unauthorized в случае ошибки
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

// Me обрабатывает HTTP-запрос на получение данных текущего пользователя
// Извлекает ID пользователя из JWT-токена и возвращает информацию о пользователе
// Возвращает 200 OK с данными пользователя, или 404 Not Found если пользователь не найден
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
