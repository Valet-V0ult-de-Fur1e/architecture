package infrastructure

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	identitydomain "architecture/backend/internal/modules/identity/domain"
)

//go:embed sql/create_user.sql
var createUserQuery string

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user identitydomain.User) error {
	if _, err := r.db.ExecContext(ctx, createUserQuery, user.ID, user.Email, user.PasswordHash, user.CreatedAt); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
