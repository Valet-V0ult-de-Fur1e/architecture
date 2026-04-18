package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	Config config.Config
	Router http.Handler
	DB     *sql.DB
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

	identityRepo := identityinfra.NewPostgresRepository(db)
	identityService := identityapp.NewService(identityRepo)
	identityHandler := identityhttp.NewHandler(identityService)

	todoRepo := todoinfra.NewPostgresRepository(db)
	todoService := todoapp.NewService(todoRepo)
	todoHandler := todohttp.NewHandler(todoService)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
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

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/identity/register", identityHandler.Register)

		api.Route("/todos", func(todo chi.Router) {
			todo.Post("/", todoHandler.Create)
			todo.Get("/", todoHandler.List)
			todo.Get("/{id}", todoHandler.GetByID)
			todo.Patch("/{id}/status", todoHandler.UpdateStatus)
			todo.Patch("/{id}/priority", todoHandler.UpdatePriority)
			todo.Delete("/{id}", todoHandler.SoftDelete)
		})
	})

	return &App{Config: cfg, Router: r, DB: db}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		_ = a.DB.Close()
	}
}
