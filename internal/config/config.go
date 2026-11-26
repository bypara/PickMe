package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config содержит настройки запуска API и подключения к PostgreSQL.
type Config struct {
	DatabaseURL  string
	HTTPAddr     string
	CookieSecure bool
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
	if raw := strings.TrimSpace(os.Getenv("COOKIE_SECURE")); raw != "" {
		secure, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, errors.New("COOKIE_SECURE must be a boolean")
		}
		cfg.CookieSecure = secure
	}
	return cfg, nil
}
