package infrastructure

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	tododomain "architecture/backend/internal/modules/todo/domain"

	"github.com/google/uuid"
)

//go:embed sql/create_todo.sql
var createTodoQuery string

//go:embed sql/get_todo_by_id.sql
var getTodoByIDQuery string

//go:embed sql/list_todos.sql
var listTodosQuery string

//go:embed sql/update_todo_status.sql
var updateTodoStatusQuery string

//go:embed sql/update_todo_priority.sql
var updateTodoPriorityQuery string

//go:embed sql/soft_delete_todo.sql
var softDeleteTodoQuery string

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, todo tododomain.Todo) error {
	if _, err := r.db.ExecContext(
		ctx,
		createTodoQuery,
		todo.ID,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.Status,
		todo.Priority,
		todo.CreatedAt,
		todo.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create todo: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, error) {
	var todo tododomain.Todo
	var status string
	if err := r.db.QueryRowContext(ctx, getTodoByIDQuery, todoID, userID).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&status,
		&todo.Priority,
		&todo.DeletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return tododomain.Todo{}, fmt.Errorf("todo not found")
		}
		return tododomain.Todo{}, fmt.Errorf("get todo: %w", err)
	}
	todo.Status = tododomain.Status(status)

	return todo, nil
}

func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, listTodosQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	var todos []tododomain.Todo
	for rows.Next() {
		var todo tododomain.Todo
		var status string
		if err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&status,
			&todo.Priority,
			&todo.DeletedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan todo row: %w", err)
		}
		todo.Status = tododomain.Status(status)
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todo rows: %w", err)
	}

	return todos, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, userID, todoID uuid.UUID, status tododomain.Status, updatedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, updateTodoStatusQuery, status, updatedAt, todoID, userID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

func (r *PostgresRepository) UpdatePriority(ctx context.Context, userID, todoID uuid.UUID, priority int, updatedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, updateTodoPriorityQuery, priority, updatedAt, todoID, userID)
	if err != nil {
		return fmt.Errorf("update priority: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, todoID uuid.UUID, deletedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, softDeleteTodoQuery, deletedAt, todoID, userID)
	if err != nil {
		return fmt.Errorf("soft delete todo: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}
