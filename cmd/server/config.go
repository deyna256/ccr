package main

import (
	"fmt"
	"os"
)

type Config struct {
	Addr             string
	AdminPort        string
	AdminSecret      string
	DatabaseURL      string
	LogLevel         string
	JWTSecret        string
	JWTRefreshSecret string
	FileStoragePath  string
}

func configFromEnv() (Config, error) {
	var missing []string
	return Config{
		Addr:             env("ADDR", ":8080"),
		AdminPort:        env("ADMIN_PORT", ":8081"),
		AdminSecret:      required("ADMIN_SECRET", &missing),
		DatabaseURL:      required("DATABASE_URL", &missing),
		LogLevel:         env("LOG_LEVEL", "info"),
		JWTSecret:        required("JWT_SECRET", &missing),
		JWTRefreshSecret: required("JWT_REFRESH_SECRET", &missing),
		FileStoragePath:  required("FILE_STORAGE_PATH", &missing),
	}, missingError(missing)
}

func required(key string, missing *[]string) string {
	v := os.Getenv(key)
	if v == "" {
		*missing = append(*missing, key)
	}
	return v
}

func missingError(missing []string) error {
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required env variables: %v", missing)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
