// PROTOTYPE — throwaway service. Connects to WhatsApp via whatsmeow and exposes
// an HTTP API for the bolaocopa backend to call.
package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sergiohpreis/bolaocopa/whatsapp/internal/notifier"
	"github.com/sergiohpreis/bolaocopa/whatsapp/internal/session"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	storePath := env("STORE_PATH", "/data")
	port := env("PORT", "9090")
	apiSecret := os.Getenv("API_SECRET")
	if apiSecret == "" || apiSecret == "prototype-secret" {
		slog.Error("API_SECRET must be set to a non-default value; refusing to start")
		os.Exit(1)
	}

	mgr := session.New(storePath)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Auto-connect on startup using a detached context — session establishment is
	// long-lived and must not be cancelled by SIGTERM arriving during connect.
	go func() {
		if err := mgr.Connect(context.Background()); err != nil {
			slog.Error("initial connect failed", "err", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Constant-time secret comparison prevents timing side-channels.
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Secret")
			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiSecret)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	ntfr := notifier.New(notifierAdapter{mgr})

	r.With(auth).Route("/", func(r chi.Router) {
		// GET /status — connection state + linked group + enabled flag
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			respond(w, map[string]any{
				"state":        string(mgr.State()),
				"linked_group": mgr.LinkedGroup(),
				"has_qr":       mgr.QRBase64() != "",
				"enabled":      mgr.Enabled(),
			})
		})

		// GET /qr — QR code as base64 PNG (only when awaiting_qr)
		r.Get("/qr", func(w http.ResponseWriter, r *http.Request) {
			qr := mgr.QRBase64()
			if qr == "" {
				http.Error(w, "no qr available", http.StatusNotFound)
				return
			}
			respond(w, map[string]string{"qr_base64": qr})
		})

		// POST /connect — start connection; no-op if already connected or connecting
		r.Post("/connect", func(w http.ResponseWriter, r *http.Request) {
			st := mgr.State()
			if st == session.StateConnected || st == session.StateConnecting {
				respond(w, map[string]string{"status": string(st)})
				return
			}
			go func() {
				if err := mgr.Connect(context.Background()); err != nil {
					slog.Error("connect", "err", err)
				}
			}()
			respond(w, map[string]string{"status": "connecting"})
		})

		// DELETE /connect — disconnect
		r.Delete("/connect", func(w http.ResponseWriter, r *http.Request) {
			mgr.Disconnect()
			respond(w, map[string]string{"status": "disconnected"})
		})

		// GET /groups — list joined groups
		r.Get("/groups", func(w http.ResponseWriter, r *http.Request) {
			groups, err := mgr.ListGroups(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			respond(w, groups)
		})

		// POST /toggle — enable or disable automatic notifications
		r.Post("/toggle", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			mgr.SetEnabled(body.Enabled)
			respond(w, map[string]bool{"enabled": body.Enabled})
		})

		// POST /healthcheck — send a test message; target_jid in body overrides global linked group
		r.Post("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				TargetJID string `json:"target_jid"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			jid := body.TargetJID
			if jid == "" {
				jid = mgr.LinkedGroup()
			}
			if jid == "" {
				http.Error(w, "no linked group", http.StatusConflict)
				return
			}
			msg := "✅ *Bolaocopa* — notificações ativas e funcionando!"
			if err := mgr.SendText(r.Context(), jid, msg); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			respond(w, map[string]string{"status": "sent"})
		})

		// POST /link — link a group JID (empty JID = unlink)
		r.Post("/link", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				JID string `json:"jid"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			mgr.LinkGroup(body.JID)
			respond(w, map[string]string{"linked_group": body.JID})
		})

		// POST /notify — send a typed notification (no-op if disabled)
		r.Post("/notify", func(w http.ResponseWriter, r *http.Request) {
			if !mgr.Enabled() {
				respond(w, map[string]string{"status": "disabled"})
				return
			}
			var body struct {
				Type      string `json:"type"`
				TargetJID string `json:"target_jid"`
				HomeTeam  string `json:"home_team"`
				AwayTeam  string `json:"away_team"`
				HomeScore int    `json:"home_score"`
				AwayScore int    `json:"away_score"`
				Winners   []struct {
					Name   string `json:"name"`
					Pontos int    `json:"pontos"`
				} `json:"winners"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			var sendErr error
			switch body.Type {
			case "fim_de_jogo":
				ws := make([]notifier.Winner, len(body.Winners))
				for i, w := range body.Winners {
					ws[i] = notifier.Winner{Name: w.Name, Pontos: w.Pontos}
				}
				sendErr = ntfr.PartidaAcabou(r.Context(), body.TargetJID, body.HomeTeam, body.HomeScore, body.AwayTeam, body.AwayScore, ws)
			case "faltam_dez_minutos":
				sendErr = ntfr.FaltamDezMinutos(r.Context(), body.TargetJID, body.HomeTeam, body.AwayTeam)
			case "partida_iniciando":
				sendErr = ntfr.PartidaIniciando(r.Context(), body.TargetJID, body.HomeTeam, body.AwayTeam)
			default:
				http.Error(w, "unknown notification type", http.StatusBadRequest)
				return
			}

			if sendErr != nil {
				http.Error(w, sendErr.Error(), http.StatusServiceUnavailable)
				return
			}
			respond(w, map[string]string{"status": "sent"})
		})
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("whatsapp service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		slog.Error("server error", "err", err)
		mgr.Disconnect()
		os.Exit(1)
	}

	slog.Info("shutting down")
	mgr.Disconnect()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("graceful shutdown incomplete", "err", err)
	}
}

func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("respond: json encode", "err", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// notifierAdapter wraps *session.Manager to satisfy notifier.Sender
type notifierAdapter struct {
	*session.Manager
}
