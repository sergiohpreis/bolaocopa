package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var footballAPIClient = &http.Client{Timeout: 15 * time.Second}

type JogoService struct {
	q                repository.Querier
	apiKey           string
	waNotif          WANotifier
	recentlyFinished []repository.Jogo // jogos que ficaram FINISHED neste sync run
}

func NewJogoService(q repository.Querier, apiKey string) *JogoService {
	return &JogoService{q: q, apiKey: apiKey, waNotif: NewNoopWANotifier()}
}

func (s *JogoService) SetWANotifier(n WANotifier) {
	s.waNotif = n
}

// DrainRecentlyFinished retorna e limpa os jogos que ficaram FINISHED no último sync.
// Chamado pelo scoring após calcular pontos, para disparar a notificação com winners.
func (s *JogoService) DrainRecentlyFinished() []repository.Jogo {
	jogos := s.recentlyFinished
	s.recentlyFinished = nil
	return jogos
}

func (s *JogoService) ListAll(ctx context.Context) ([]repository.Jogo, error) {
	items, err := s.q.ListJogos(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.Jogo{}
	}
	return items, nil
}

// SyncFromAPI fetches matches from football-data.org and upserts them.
func (s *JogoService) SyncFromAPI(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.football-data.org/v4/competitions/WC/matches", nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Auth-Token", s.apiKey)

	resp, err := footballAPIClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling football-data api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	var apiResp footballDataResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	now := time.Now().UTC()

	for _, m := range apiResp.Matches {
		t, err := time.Parse(time.RFC3339, m.UtcDate)
		if err != nil {
			slog.Warn("failed to parse match date, skipping", "external_id", m.ID, "utc_date", m.UtcDate)
			continue
		}
		startsAt := pgtype.Timestamptz{Time: t, Valid: true}

		finished := m.Status == "FINISHED"
		homeScore := pgtype.Int4{}
		awayScore := pgtype.Int4{}
		if finished && m.Score.FullTime.Home != nil && m.Score.FullTime.Away != nil {
			homeScore = pgtype.Int4{Int32: int32(*m.Score.FullTime.Home), Valid: true}
			awayScore = pgtype.Int4{Int32: int32(*m.Score.FullTime.Away), Valid: true}
		}

		params := repository.UpsertJogoParams{
			ExternalID:   fmt.Sprintf("%d", m.ID),
			HomeTeam:     m.HomeTeam.Name,
			AwayTeam:     m.AwayTeam.Name,
			HomeTeamFlag: pgtype.Text{String: m.HomeTeam.Crest, Valid: m.HomeTeam.Crest != ""},
			AwayTeamFlag: pgtype.Text{String: m.AwayTeam.Crest, Valid: m.AwayTeam.Crest != ""},
			StartsAt:     startsAt,
			Stage:        m.Stage,
			HomeScore:    homeScore,
			AwayScore:    awayScore,
			Finished:     finished,
		}

		upserted, err := s.q.UpsertJogo(ctx, params)
		if err != nil {
			slog.Warn("failed to upsert jogo", "external_id", params.ExternalID, "error", err)
			continue
		}

		// Rastreia jogos recém-finalizados (updated_at < 8 min) para notificar após o scoring.
		if finished && upserted.HomeScore.Valid && upserted.AwayScore.Valid {
			if now.Sub(upserted.UpdatedAt.Time) < 8*time.Minute {
				s.recentlyFinished = append(s.recentlyFinished, upserted)
			}
		}

		// Dispatch WhatsApp notifications based on match timing.
		// The sync runs every 5 min, so each jogo falls in each window exactly once.
		//   faltam_dez_minutos: starts_at in [now+8min, now+13min)
		//   partida_iniciando:  starts_at in [now-3min, now+3min)
		if !finished {
			untilStart := t.Sub(now)
			s.dispatchMatchNotifications(ctx, untilStart, m.HomeTeam.Name, m.AwayTeam.Name)
		}
	}

	slog.Info("synced jogos from football-data", "count", len(apiResp.Matches))
	return nil
}

// dispatchMatchNotifications fires WhatsApp pre-match alerts based on how far
// the match start is from now. Called once per match per sync run.
func (s *JogoService) dispatchMatchNotifications(ctx context.Context, untilStart time.Duration, homeTeam, awayTeam string) {
	const bolaoID = "" // single-group prototype: bolaoID unused by httpWANotifier.post

	switch {
	case untilStart >= 8*time.Minute && untilStart < 13*time.Minute:
		go s.waNotif.NotifyFaltamDezMinutos(ctx, bolaoID, homeTeam, awayTeam)
	case untilStart >= -3*time.Minute && untilStart < 3*time.Minute:
		go s.waNotif.NotifyPartidaIniciando(ctx, bolaoID, homeTeam, awayTeam)
	}
}

type footballDataResponse struct {
	Matches []fdMatch `json:"matches"`
}

type fdMatch struct {
	ID       int     `json:"id"`
	UtcDate  string  `json:"utcDate"`
	Stage    string  `json:"stage"`
	Status   string  `json:"status"`
	HomeTeam fdTeam  `json:"homeTeam"`
	AwayTeam fdTeam  `json:"awayTeam"`
	Score    fdScore `json:"score"`
}

type fdTeam struct {
	Name  string `json:"name"`
	Crest string `json:"crest"`
}

type fdScore struct {
	FullTime fdGoals `json:"fullTime"`
}

type fdGoals struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}
