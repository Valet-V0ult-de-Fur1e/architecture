package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	identitydomain "architecture/backend/internal/modules/identity/domain"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, user identitydomain.User) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, email, password string) (identitydomain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return identitydomain.User{}, fmt.Errorf("email and password are required")
	}

	user := identitydomain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashPassword(password),
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return identitydomain.User{}, err
	}

	return user, nil
}

func hashPassword(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
