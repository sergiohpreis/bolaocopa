package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergiohpreis/bolaocopa/backend/internal/config"
	"github.com/sergiohpreis/bolaocopa/backend/internal/handler"
	apimiddleware "github.com/sergiohpreis/bolaocopa/backend/internal/middleware"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	pool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := runMigrations(cfg.DatabaseURL); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations applied successfully")
	}

	queries := repository.New(pool)

	authSvc := service.NewAuthService(queries, cfg.JWTSecret, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	bolaoSvc := service.NewBolaoService(queries)
	jogoSvc := service.NewJogoService(queries, cfg.FootballDataAPIKey)
	palpiteSvc := service.NewPalpiteService(queries, pool)
	rankingSvc := service.NewRankingService(queries)
	feedSvc := service.NewFeedService(queries)
	taxaSvc := service.NewTaxaService(queries, pool)
	bolaoSvc.SetFeed(feedSvc)
	palpiteSvc.SetFeed(feedSvc)
	rankingSvc.SetFeed(feedSvc)
	taxaSvc.SetFeed(feedSvc)

	// Wire WhatsApp notifier — no-op when WHATSAPP_SERVICE_URL is not set
	var waN service.WANotifier
	if cfg.WhatsAppServiceURL != "" && cfg.WhatsAppAPISecret != "" {
		waN = service.NewHTTPWANotifier(cfg.WhatsAppServiceURL, cfg.WhatsAppAPISecret)
		slog.Info("whatsapp notifier enabled", "url", cfg.WhatsAppServiceURL)
	} else {
		waN = service.NewNoopWANotifier()
		slog.Info("whatsapp notifier disabled (WHATSAPP_SERVICE_URL not set)")
	}
	rankingSvc.SetWANotifier(waN)
	jogoSvc.SetWANotifier(waN)

	allowedOrigins := splitOrigins(cfg.AllowedOrigins)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.CORS(allowedOrigins))

	authMw := apimiddleware.Auth(authSvc)

	handler.RegisterRoutes(
		r,
		handler.NewAuthHandler(authSvc),
		handler.NewBolaoHandler(bolaoSvc),
		handler.NewJogoHandler(jogoSvc),
		handler.NewPalpiteHandler(palpiteSvc),
		handler.NewRankingHandler(rankingSvc, bolaoSvc),
		handler.NewFeedHandler(feedSvc),
		handler.NewTaxaHandler(taxaSvc),
		authMw,
	)

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Background sync: roda imediatamente ao iniciar e depois a cada 5 minutos.
	go func() {
		doSync := func() {
			syncCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := jogoSvc.SyncFromAPI(syncCtx); err != nil {
				slog.Warn("background sync failed", "error", err)
			} else if err := rankingSvc.ComputeScoresForFinishedJogos(syncCtx); err != nil {
				slog.Warn("background scoring failed", "error", err)
			}
		}
		doSync()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				doSync()
			}
		}
	}()

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		slog.Error("server error", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func connectDB(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	slog.Info("database connection established")
	return pool, nil
}

func splitOrigins(s string) []string {
	parts := strings.Split(s, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file:///migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			slog.Warn("migration source close error", "error", srcErr)
		}
		if dbErr != nil {
			slog.Warn("migration db close error", "error", dbErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("m.Up: %w", err)
	}
	return nil
}
