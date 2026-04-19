package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/task-planner/server/internal/auth"
	"github.com/task-planner/server/internal/category"
	"github.com/task-planner/server/internal/correlation"
	"github.com/task-planner/server/internal/httplog"
	"github.com/task-planner/server/internal/task"
)

func newServer(cfg Config, db *sql.DB, log *slog.Logger) *http.Server {
	authStorage := auth.NewPostgresStorage(db)
	authService := auth.NewService(authStorage, cfg.JWTSecret, cfg.JWTRefreshSecret)
	authHandler := auth.NewHandler(authService, log)

	catStorage := category.NewPostgresStorage(db)
	catService := category.NewService(catStorage)
	catHandler := category.NewHandler(catService, log)

	taskStorage := task.NewPostgresStorage(db)
	taskService := task.NewService(taskStorage, cfg.FileStoragePath, log)
	taskHandler := task.NewHandler(taskService, log)

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", authHandler.Routes())
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))
			r.Mount("/categories", catHandler.Routes())
			r.Mount("/tasks", taskHandler.Routes())
			r.Get("/attachments/{id}/file", taskHandler.ServeFile)
		})
	})

	handler := correlation.Middleware(httplog.Middleware(log)(r))

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
