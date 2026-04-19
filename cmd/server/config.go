package main

import (
	"log"
	"log/slog"
	"os"
)

type Config struct {
	Addr             string
	DatabaseURL      string
	LogLevel         slog.Level
	JWTSecret        string
	JWTRefreshSecret string
	FileStoragePath  string
}

func configFromEnv() Config {
	return Config{
		Addr:             env("ADDR", ":8080"),
		DatabaseURL:      mustEnv("DATABASE_URL"),
		LogLevel:         parseLogLevel(env("LOG_LEVEL", "info")),
		JWTSecret:        mustEnv("JWT_SECRET"),
		JWTRefreshSecret: mustEnv("JWT_REFRESH_SECRET"),
		FileStoragePath:  mustEnv("FILE_STORAGE_PATH"),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env variable %q is not set", key)
	}
	return v
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
