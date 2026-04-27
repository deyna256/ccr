package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"github.com/task-planner/server/internal/admin"
	"github.com/task-planner/server/migrations"
	"github.com/tfcp-site/httpx/correlation"
	"github.com/tfcp-site/httpx/httplog"
	"github.com/tfcp-site/httpx/logging"
)

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n") //nolint:errcheck
		os.Exit(1)
	}

	log := logging.New("ccr", parseLogLevel(cfg.LogLevel), correlation.Extractor())

	db, err := openDB(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	apiSrv, err := newAPIServer(cfg, db, log)
	if err != nil {
		log.Error("failed to create API server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	adminStorage := admin.NewAuthStorageAdapter(db, log)
	adminHandler := admin.NewHandler(adminStorage, cfg.AdminSecret, log)
	adminSrv := newAdminServer(cfg.AdminPort, adminHandler.Routes(), log)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("API server started", slog.String("addr", cfg.Addr))
		if err := apiSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("API server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	go func() {
		log.Info("admin server started", slog.String("addr", cfg.AdminPort))
		if err := adminSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("admin server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-quit

	log.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := apiSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("API server shutdown error", slog.String("error", err.Error()))
	}
	if err := adminSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("admin server shutdown error", slog.String("error", err.Error()))
	}

	log.Info("server exited")
}

func openDB(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(time.Minute)
	return db, nil
}

func runMigrations(db *sql.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newAdminServer(addr string, handler http.Handler, log *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           correlation.Middleware(httplog.Middleware(log)(handler)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
