package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"architecture/backend/internal/modules/session/domain"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisSessionRepository(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	sess := &domain.Session{
		ID:        "test-session-id",
		UserID:    "user-123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Minute),
	}

	err := repo.SetSession(ctx, sess, 1*time.Minute)
	require.NoError(t, err)

	got, err := repo.GetSession(ctx, sess.ID)
	require.NoError(t, err)
	require.Equal(t, sess.UserID, got.UserID)

	err = repo.DeleteSession(ctx, sess.ID)
	require.NoError(t, err)

	_, err = repo.GetSession(ctx, sess.ID)
	require.Error(t, err)
}
