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

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}
