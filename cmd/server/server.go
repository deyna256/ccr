package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/task-planner/server/internal/auth"
	"github.com/task-planner/server/internal/sync"
	"github.com/task-planner/server/internal/task"
	"github.com/task-planner/server/ui"
	"github.com/tfcp-site/httpx/correlation"
	"github.com/tfcp-site/httpx/httplog"
)

func newAPIServer(cfg Config, db *sql.DB, log *slog.Logger) (*http.Server, error) {
	authStorage := auth.NewPostgresStorage(db)
	authService := auth.NewService(authStorage, cfg.JWTSecret, cfg.JWTRefreshSecret, log)
	authHandler := auth.NewHandler(authService, log)

	taskStorage := task.NewPostgresStorage(db, log)
	taskFileStorage := task.NewOsFileStorage(cfg.FileStoragePath)
	taskService := task.NewService(taskStorage, taskFileStorage, log)
	taskHandler := task.NewHandler(taskService, log)

	syncStorage := sync.NewPostgresStorage(db)
	syncService := sync.NewService(syncStorage, log)
	syncHandler := sync.NewHandler(syncService, log)

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", authHandler.Routes())
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret, log))
			r.Mount("/tasks", taskHandler.Routes())
			r.Get("/attachments/{id}/file", taskHandler.ServeFile)
			r.Mount("/sync", syncHandler.Routes())
		})
	})

	distFS, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return nil, fmt.Errorf("ui dist not found: %w", err)
	}
	r.Handle("/*", spaHandler(distFS))

	handler := correlation.Middleware(httplog.Middleware(log)(r))

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}, nil
}

func spaHandler(fsys fs.FS) http.Handler {
	srv := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "" {
			srv.ServeHTTP(w, r)
			return
		}
		if _, err := fsys.Open(path[1:]); err != nil {
			r.URL.Path = "/"
		}
		srv.ServeHTTP(w, r)
	})
}
