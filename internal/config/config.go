package config

import (
	"errors"
	"os"
	"strings"
)

// Config содержит настройки запуска API и подключения к PostgreSQL.
type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

// Load читает настройки из окружения. DATABASE_URL обязателен, HTTP_ADDR по умолчанию равен :8080.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		HTTPAddr:    strings.TrimSpace(os.Getenv("HTTP_ADDR")),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	return cfg, nil
}
