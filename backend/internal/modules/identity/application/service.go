package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	identitydomain "architecture/backend/internal/modules/identity/domain"
	jwtmanager "architecture/backend/internal/shared/auth/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, user identitydomain.User) error
	GetByEmail(ctx context.Context, email string) (identitydomain.User, error)
}

type Service struct {
	repo         Repository
	tokenManager *jwtmanager.Manager
	tokenTTL     time.Duration
}

func NewService(repo Repository, tokenManager *jwtmanager.Manager, tokenTTL time.Duration) *Service {
	return &Service{
		repo:         repo,
		tokenManager: tokenManager,
		tokenTTL:     tokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (identitydomain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return identitydomain.User{}, fmt.Errorf("email and password are required")
	}

	user := identitydomain.User{
		ID:        uuid.New(),
		Email:     email,
		CreatedAt: time.Now().UTC(),
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return identitydomain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = passwordHash

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return identitydomain.User{}, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return "", fmt.Errorf("email and password are required")
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := s.tokenManager.Generate(user.ID, s.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}

	return token, nil
}

func hashPassword(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
