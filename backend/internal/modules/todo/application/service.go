package application

import (
	"context"
	"fmt"
	"log"
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

type Cache interface {
	GetTodo(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, bool, error)
	SetTodo(ctx context.Context, todo tododomain.Todo) error
	DeleteTodo(ctx context.Context, userID, todoID uuid.UUID) error
	GetTodoList(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, bool, error)
	SetTodoList(ctx context.Context, userID uuid.UUID, todos []tododomain.Todo) error
	DeleteTodoList(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	repo  Repository
	cache Cache
}

func NewService(repo Repository, cache Cache) *Service {
	return &Service{repo: repo, cache: cache}
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
		log.Printf("todo.create: failed user_id=%s error=%v", userID, err)
		return tododomain.Todo{}, err
	}
	log.Printf("todo.create: success user_id=%s todo_id=%s", userID, todo.ID)

	if s.cache != nil {
		if err := s.cache.SetTodo(ctx, todo); err != nil {
			log.Printf("todo.create: cache set item failed user_id=%s todo_id=%s error=%v", userID, todo.ID, err)
		}
		if err := s.cache.DeleteTodoList(ctx, userID); err != nil {
			log.Printf("todo.create: cache invalidate list failed user_id=%s error=%v", userID, err)
		}
	}

	return todo, nil
}

func (s *Service) GetByID(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, error) {
	if s.cache != nil {
		todo, found, err := s.cache.GetTodo(ctx, userID, todoID)
		if err == nil && found {
			log.Printf("todo.get_by_id: cache hit user_id=%s todo_id=%s", userID, todoID)
			return todo, nil
		}
		if err != nil {
			log.Printf("todo.get_by_id: cache error user_id=%s todo_id=%s error=%v", userID, todoID, err)
		} else {
			log.Printf("todo.get_by_id: cache miss user_id=%s todo_id=%s", userID, todoID)
		}
	}

	todo, err := s.repo.GetByID(ctx, userID, todoID)
	if err != nil {
		log.Printf("todo.get_by_id: repo failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		return tododomain.Todo{}, err
	}
	log.Printf("todo.get_by_id: repo success user_id=%s todo_id=%s", userID, todoID)

	if s.cache != nil {
		if err := s.cache.SetTodo(ctx, todo); err != nil {
			log.Printf("todo.get_by_id: cache set failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		}
	}

	return todo, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, error) {
	if s.cache != nil {
		todos, found, err := s.cache.GetTodoList(ctx, userID)
		if err == nil && found {
			log.Printf("todo.list: cache hit user_id=%s count=%d", userID, len(todos))
			return todos, nil
		}
		if err != nil {
			log.Printf("todo.list: cache error user_id=%s error=%v", userID, err)
		} else {
			log.Printf("todo.list: cache miss user_id=%s", userID)
		}
	}

	todos, err := s.repo.List(ctx, userID)
	if err != nil {
		log.Printf("todo.list: repo failed user_id=%s error=%v", userID, err)
		return nil, err
	}
	log.Printf("todo.list: repo success user_id=%s count=%d", userID, len(todos))

	if s.cache != nil {
		if err := s.cache.SetTodoList(ctx, userID, todos); err != nil {
			log.Printf("todo.list: cache set failed user_id=%s error=%v", userID, err)
		}
	}

	return todos, nil
}

func (s *Service) UpdateStatus(ctx context.Context, userID, todoID uuid.UUID, status tododomain.Status) error {
	if err := tododomain.ValidateStatus(status); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, userID, todoID, status, time.Now().UTC()); err != nil {
		log.Printf("todo.update_status: failed user_id=%s todo_id=%s status=%s error=%v", userID, todoID, status, err)
		return err
	}
	log.Printf("todo.update_status: success user_id=%s todo_id=%s status=%s", userID, todoID, status)

	if s.cache != nil {
		if err := s.cache.DeleteTodo(ctx, userID, todoID); err != nil {
			log.Printf("todo.update_status: cache delete item failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		}
		if err := s.cache.DeleteTodoList(ctx, userID); err != nil {
			log.Printf("todo.update_status: cache delete list failed user_id=%s error=%v", userID, err)
		}
	}

	return nil
}

func (s *Service) UpdatePriority(ctx context.Context, userID, todoID uuid.UUID, priority int) error {
	if err := tododomain.ValidatePriority(priority); err != nil {
		return err
	}

	if err := s.repo.UpdatePriority(ctx, userID, todoID, priority, time.Now().UTC()); err != nil {
		log.Printf("todo.update_priority: failed user_id=%s todo_id=%s priority=%d error=%v", userID, todoID, priority, err)
		return err
	}
	log.Printf("todo.update_priority: success user_id=%s todo_id=%s priority=%d", userID, todoID, priority)

	if s.cache != nil {
		if err := s.cache.DeleteTodo(ctx, userID, todoID); err != nil {
			log.Printf("todo.update_priority: cache delete item failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		}
		if err := s.cache.DeleteTodoList(ctx, userID); err != nil {
			log.Printf("todo.update_priority: cache delete list failed user_id=%s error=%v", userID, err)
		}
	}

	return nil
}

func (s *Service) SoftDelete(ctx context.Context, userID, todoID uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, userID, todoID, time.Now().UTC()); err != nil {
		log.Printf("todo.soft_delete: failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		return err
	}
	log.Printf("todo.soft_delete: success user_id=%s todo_id=%s", userID, todoID)

	if s.cache != nil {
		if err := s.cache.DeleteTodo(ctx, userID, todoID); err != nil {
			log.Printf("todo.soft_delete: cache delete item failed user_id=%s todo_id=%s error=%v", userID, todoID, err)
		}
		if err := s.cache.DeleteTodoList(ctx, userID); err != nil {
			log.Printf("todo.soft_delete: cache delete list failed user_id=%s error=%v", userID, err)
		}
	}

	return nil
}
