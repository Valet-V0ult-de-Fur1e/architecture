package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"architecture/backend/internal/modules/session/domain"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepository struct {
	client *redis.Client
}

func NewRedisSessionRepository(client *redis.Client) *RedisSessionRepository {
	return &RedisSessionRepository{client: client}
}

func (r *RedisSessionRepository) SetSession(ctx context.Context, session *domain.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("session:%s", session.ID)
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var session domain.Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *RedisSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}
