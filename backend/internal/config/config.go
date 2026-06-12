package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL        string
	Port               string
	RunMigrations      bool
	JWTSecret          string
	AllowedOrigins     string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FootballDataAPIKey string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		slog.Warn("ALLOWED_ORIGINS is not set, defaulting to localhost only")
		allowedOrigins = "http://localhost:5173"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET is not set — refusing to start")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		slog.Warn("JWT_SECRET is shorter than 32 characters — consider using a longer secret")
	}

	cfg := Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		Port:               port,
		RunMigrations:      strings.EqualFold(os.Getenv("RUN_MIGRATIONS"), "true"),
		JWTSecret:          jwtSecret,
		AllowedOrigins:     allowedOrigins,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		FootballDataAPIKey: os.Getenv("FOOTBALL_DATA_API_KEY"),
	}

	if cfg.DatabaseURL == "" {
		slog.Warn("DATABASE_URL is not set")
	}
	if cfg.GoogleClientID == "" {
		slog.Warn("GOOGLE_CLIENT_ID is not set")
	}

	return cfg
}
