package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ssnakerss/modmas/pkg/hasher"
	"github.com/Ssnakerss/modmas/pkg/jwt"
)

type Service struct {
	repo       *Repository
	jwtManager *jwt.Manager
}

// NewService создает новый экземпляр сервиса аутентификации
// Принимает репозиторий для работы с пользователями и менеджер JWT-токенов
func NewService(repo *Repository, jwtManager *jwt.Manager) *Service {
	return &Service{repo: repo, jwtManager: jwtManager}
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        *User  `json:"user"`
}

// Register реализует бизнес-логику регистрации нового пользователя
// Проверяет обязательные поля, хеширует пароль, создает пользователя в БД и генерирует JWT-токен
// Возвращает данные пользователя и токен, или ошибку валидации/операции
func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, errors.New("email, password and name are required")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	hash, err := hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.Create(ctx, input.Email, hash, input.Name)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{AccessToken: token, User: user}, nil
}

// Login реализует бизнес-логику аутентификации пользователя
// Проверяет существование пользователя, сверяет пароль и генерирует JWT-токен
// Возвращает данные пользователя и токен при успешной аутентификации, или ошибку
func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !hasher.Check(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{AccessToken: token, User: user}, nil
}

// Me возвращает информацию о пользователе по его ID
// Используется для получения данных текущего пользователя из JWT-токена
// Возвращает данные пользователя или ошибку, если пользователь не найден
func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}
