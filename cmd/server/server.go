package main

import (
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/task-planner/server/internal/auth"
	"github.com/task-planner/server/internal/correlation"
	"github.com/task-planner/server/internal/httplog"
	"github.com/task-planner/server/internal/task"
	"github.com/task-planner/server/ui"
)

func newServer(cfg Config, db *sql.DB, log *slog.Logger) *http.Server {
	authStorage := auth.NewPostgresStorage(db)
	authService := auth.NewService(authStorage, cfg.JWTSecret, cfg.JWTRefreshSecret)
	authHandler := auth.NewHandler(authService, log)

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
			r.Mount("/tasks", taskHandler.Routes())
			r.Get("/attachments/{id}/file", taskHandler.ServeFile)
		})
	})

	distFS, _ := fs.Sub(ui.FS, "dist")
	r.Handle("/*", spaHandler(distFS))

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
