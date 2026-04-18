package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"architecture/backend/internal/config"
	identityapp "architecture/backend/internal/modules/identity/application"
	identityinfra "architecture/backend/internal/modules/identity/infrastructure"
	identityhttp "architecture/backend/internal/modules/identity/transport/http"
	todoapp "architecture/backend/internal/modules/todo/application"
	todoinfra "architecture/backend/internal/modules/todo/infrastructure"
	todohttp "architecture/backend/internal/modules/todo/transport/http"
	"architecture/backend/internal/platform/database"
	"architecture/backend/internal/platform/rabbitmq"
	jwtmanager "architecture/backend/internal/shared/auth/jwt"
	"architecture/backend/internal/shared/transport/httpauth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config config.Config
	Router http.Handler
	DB     *sql.DB
	Redis  *redis.Client
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	if err := database.RunMigrations(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	log.Println("bootstrap: postgres connected and migrations applied")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	redisPingCtx, redisPingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisPingCancel()
	if err := redisClient.Ping(redisPingCtx).Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	log.Println("bootstrap: redis connected")

	go rabbitmq.StartEventConsumer(context.Background())

	// Try to set up exchanges/queue/bindings in background until successful.
	go func() {
		for {
			if err := rabbitmq.SetupExchangesAndQueue(context.Background()); err != nil {
				log.Printf("rabbitmq setup failed: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			log.Println("rabbitmq exchanges/queue/bindings set up")
			return
		}
	}()

	jwt := jwtmanager.NewManager(cfg.JWT.Secret)

	identityRepo := identityinfra.NewPostgresRepository(db)
	identityService := identityapp.NewService(identityRepo, jwt, cfg.JWT.ExpiresIn)
	identityHandler := identityhttp.NewHandler(identityService)

	todoRepo := todoinfra.NewPostgresRepository(db)
	todoCache := todoinfra.NewRedisCache(redisClient, cfg.Redis.TodoTTL)
	todoService := todoapp.NewService(todoRepo, todoCache)
	todoHandler := todohttp.NewHandler(todoService)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("db not ready"))
			return
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("redis not ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/identity/register", identityHandler.Register)
		api.Post("/identity/login", identityHandler.Login)

		api.Route("/todos", func(todo chi.Router) {
			todo.Use(httpauth.JWTAuth(jwt))
			todo.Post("/", todoHandler.Create)
			todo.Get("/", todoHandler.List)
			todo.Get("/{id}", todoHandler.GetByID)
			todo.Patch("/{id}/status", todoHandler.UpdateStatus)
			todo.Patch("/{id}/priority", todoHandler.UpdatePriority)
			todo.Delete("/{id}", todoHandler.SoftDelete)
		})

		// Wire session endpoints
		wireSession(redisClient, api)
	})

	return &App{Config: cfg, Router: r, DB: db, Redis: redisClient}, nil
}

func (a *App) Close() {
	if a.Redis != nil {
		_ = a.Redis.Close()
	}

	if a.DB != nil {
		_ = a.DB.Close()
	}
}
