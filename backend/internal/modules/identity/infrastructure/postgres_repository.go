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

//go:embed sql/get_user_by_email.sql
var getUserByEmailQuery string

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

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (identitydomain.User, error) {
	var user identitydomain.User
	if err := r.db.QueryRowContext(ctx, getUserByEmailQuery, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return identitydomain.User{}, fmt.Errorf("user not found")
		}

		return identitydomain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}
