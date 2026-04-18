package bootstrap

import (
	"architecture/backend/internal/modules/session/infrastructure"
	"architecture/backend/internal/modules/session/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// ...existing code...

func wireSession(redisClient *redis.Client, r chi.Router) {
	sessRepo := infrastructure.NewRedisSessionRepository(redisClient)
	sessHandler := http.NewSessionHandler(sessRepo)

	r.Route("/session", func(sr chi.Router) {
		sr.Post("/login", sessHandler.Login)
		sr.Get("/check", sessHandler.Check)
		sr.Post("/logout", sessHandler.Logout)
	})
}
