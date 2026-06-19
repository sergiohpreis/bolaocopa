package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// WANotifier sends bolão notifications to a WhatsApp group.
// groupJID is the WhatsApp group JID to send to; empty string falls back to
// the whatsapp service's globally linked group (useful for healthchecks/tests).
type WANotifier interface {
	NotifyFimDeJogo(ctx context.Context, groupJID, homeTeam string, homeScore int, awayTeam string, awayScore int, winners []WAWinner)
	NotifyFaltamDezMinutos(ctx context.Context, groupJID, homeTeam, awayTeam string)
	NotifyPartidaIniciando(ctx context.Context, groupJID, homeTeam, awayTeam string)
}

type WAWinner struct {
	Name   string
	Pontos int
}

// httpWANotifier is the real implementation — calls the whatsapp service HTTP API.
type httpWANotifier struct {
	baseURL   string
	apiSecret string
	client    *http.Client
}

func NewHTTPWANotifier(baseURL, apiSecret string) WANotifier {
	return &httpWANotifier{
		baseURL:   baseURL,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// noopWANotifier is used when the whatsapp service is not configured.
type noopWANotifier struct{}

func NewNoopWANotifier() WANotifier { return &noopWANotifier{} }

func (n *noopWANotifier) NotifyFimDeJogo(_ context.Context, _, _ string, _ int, _ string, _ int, _ []WAWinner) {
}
func (n *noopWANotifier) NotifyFaltamDezMinutos(_ context.Context, _, _, _ string) {}
func (n *noopWANotifier) NotifyPartidaIniciando(_ context.Context, _, _, _ string) {}

func (n *httpWANotifier) NotifyFimDeJogo(ctx context.Context, groupJID, homeTeam string, homeScore int, awayTeam string, awayScore int, winners []WAWinner) {
	ws := make([]map[string]any, len(winners))
	for i, w := range winners {
		ws[i] = map[string]any{"name": w.Name, "pontos": w.Pontos}
	}
	n.post(ctx, groupJID, map[string]any{
		"type":       "fim_de_jogo",
		"home_team":  homeTeam,
		"home_score": homeScore,
		"away_team":  awayTeam,
		"away_score": awayScore,
		"winners":    ws,
	})
}

func (n *httpWANotifier) NotifyFaltamDezMinutos(ctx context.Context, groupJID, homeTeam, awayTeam string) {
	n.post(ctx, groupJID, map[string]any{
		"type":      "faltam_dez_minutos",
		"home_team": homeTeam,
		"away_team": awayTeam,
	})
}

func (n *httpWANotifier) NotifyPartidaIniciando(ctx context.Context, groupJID, homeTeam, awayTeam string) {
	n.post(ctx, groupJID, map[string]any{
		"type":      "partida_iniciando",
		"home_team": homeTeam,
		"away_team": awayTeam,
	})
}

// post sends the notification with the given groupJID as target_jid.
// The whatsapp service falls back to its globally linked group when target_jid is empty.
func (n *httpWANotifier) post(ctx context.Context, groupJID string, payload map[string]any) {
	if groupJID != "" {
		payload["target_jid"] = groupJID
	}

	b, err := json.Marshal(payload)
	if err != nil {
		slog.Error("wa notifier: marshal payload", "err", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+"/notify", bytes.NewReader(b))
	if err != nil {
		slog.Error("wa notifier: build request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Secret", n.apiSecret)

	resp, err := n.client.Do(req)
	if err != nil {
		slog.Warn("wa notifier: send failed", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		slog.Warn("wa notifier: non-2xx response", "status", resp.StatusCode, "type", fmt.Sprintf("%v", payload["type"]))
	}
}
