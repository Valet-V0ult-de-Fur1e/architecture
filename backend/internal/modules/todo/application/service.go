package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	tododomain "architecture/backend/internal/modules/todo/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, todo tododomain.Todo) error
	GetByID(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, error)
	List(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, error)
	UpdateStatus(ctx context.Context, userID, todoID uuid.UUID, status tododomain.Status, updatedAt time.Time) error
	UpdatePriority(ctx context.Context, userID, todoID uuid.UUID, priority int, updatedAt time.Time) error
	SoftDelete(ctx context.Context, userID, todoID uuid.UUID, deletedAt time.Time) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, title, description string, priority int) (tododomain.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return tododomain.Todo{}, fmt.Errorf("title is required")
	}
	if err := tododomain.ValidatePriority(priority); err != nil {
		return tododomain.Todo{}, err
	}

	now := time.Now().UTC()
	todo := tododomain.Todo{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      tododomain.StatusOpen,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return tododomain.Todo{}, err
	}

	return todo, nil
}

func (s *Service) GetByID(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, error) {
	return s.repo.GetByID(ctx, userID, todoID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, todoID uuid.UUID, status tododomain.Status) error {
	if err := tododomain.ValidateStatus(status); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, userID, todoID, status, time.Now().UTC())
}

func (s *Service) UpdatePriority(ctx context.Context, userID, todoID uuid.UUID, priority int) error {
	if err := tododomain.ValidatePriority(priority); err != nil {
		return err
	}
	return s.repo.UpdatePriority(ctx, userID, todoID, priority, time.Now().UTC())
}

func (s *Service) SoftDelete(ctx context.Context, userID, todoID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, todoID, time.Now().UTC())
}
