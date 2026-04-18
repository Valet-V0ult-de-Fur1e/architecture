package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tododomain "architecture/backend/internal/modules/todo/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{client: client, ttl: ttl}
}

func (c *RedisCache) GetTodo(ctx context.Context, userID, todoID uuid.UUID) (tododomain.Todo, bool, error) {
	value, err := c.client.Get(ctx, c.todoKey(userID, todoID)).Result()
	if err != nil {
		if err == redis.Nil {
			return tododomain.Todo{}, false, nil
		}
		return tododomain.Todo{}, false, fmt.Errorf("redis get todo: %w", err)
	}

	var todo tododomain.Todo
	if err := json.Unmarshal([]byte(value), &todo); err != nil {
		return tododomain.Todo{}, false, fmt.Errorf("unmarshal todo cache: %w", err)
	}

	return todo, true, nil
}

func (c *RedisCache) SetTodo(ctx context.Context, todo tododomain.Todo) error {
	payload, err := json.Marshal(todo)
	if err != nil {
		return fmt.Errorf("marshal todo cache: %w", err)
	}

	if err := c.client.Set(ctx, c.todoKey(todo.UserID, todo.ID), payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set todo: %w", err)
	}

	return nil
}

func (c *RedisCache) DeleteTodo(ctx context.Context, userID, todoID uuid.UUID) error {
	if err := c.client.Del(ctx, c.todoKey(userID, todoID)).Err(); err != nil {
		return fmt.Errorf("redis delete todo: %w", err)
	}

	return nil
}

func (c *RedisCache) GetTodoList(ctx context.Context, userID uuid.UUID) ([]tododomain.Todo, bool, error) {
	value, err := c.client.Get(ctx, c.todoListKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("redis get todo list: %w", err)
	}

	var todos []tododomain.Todo
	if err := json.Unmarshal([]byte(value), &todos); err != nil {
		return nil, false, fmt.Errorf("unmarshal todo list cache: %w", err)
	}

	return todos, true, nil
}

func (c *RedisCache) SetTodoList(ctx context.Context, userID uuid.UUID, todos []tododomain.Todo) error {
	payload, err := json.Marshal(todos)
	if err != nil {
		return fmt.Errorf("marshal todo list cache: %w", err)
	}

	if err := c.client.Set(ctx, c.todoListKey(userID), payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set todo list: %w", err)
	}

	return nil
}

func (c *RedisCache) DeleteTodoList(ctx context.Context, userID uuid.UUID) error {
	if err := c.client.Del(ctx, c.todoListKey(userID)).Err(); err != nil {
		return fmt.Errorf("redis delete todo list: %w", err)
	}

	return nil
}

func (c *RedisCache) todoKey(userID, todoID uuid.UUID) string {
	return fmt.Sprintf("todos:%s:item:%s", userID.String(), todoID.String())
}

func (c *RedisCache) todoListKey(userID uuid.UUID) string {
	return fmt.Sprintf("todos:%s:list", userID.String())
}
