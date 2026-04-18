package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen   Status = "open"
	StatusDone   Status = "done"
	StatusInWork Status = "in_work"
)

type Todo struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Status      Status
	Priority    int
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ValidateStatus(status Status) error {
	switch status {
	case StatusOpen, StatusDone, StatusInWork:
		return nil
	default:
		return fmt.Errorf("invalid status")
	}
}

func ValidatePriority(priority int) error {
	if priority < 1 || priority > 5 {
		return fmt.Errorf("priority must be between 1 and 5")
	}
	return nil
}
