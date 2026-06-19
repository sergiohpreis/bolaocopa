package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var footballAPIClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

type JogoService struct {
	q       repository.Querier
	apiKey  string
	waNotif WANotifier
}

func NewJogoService(q repository.Querier, apiKey string) *JogoService {
	return &JogoService{q: q, apiKey: apiKey, waNotif: NewNoopWANotifier()}
}

func (s *JogoService) SetWANotifier(n WANotifier) {
	s.waNotif = n
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
// Returns jogos that transitioned to FINISHED in this sync run (for fim-de-jogo notifications).
func (s *JogoService) SyncFromAPI(ctx context.Context) ([]repository.Jogo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.football-data.org/v4/competitions/WC/matches", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Auth-Token", s.apiKey)

	resp, err := footballAPIClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling football-data api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var apiResp footballDataResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	now := time.Now().UTC()

	var recentlyFinished []repository.Jogo

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

		// Track jogos that transitioned !finished → finished in this sync.
		// WasFinished comes from the CTE that reads the pre-upsert state.
		if upserted.Finished && !upserted.WasFinished && upserted.HomeScore.Valid && upserted.AwayScore.Valid {
			recentlyFinished = append(recentlyFinished, repository.Jogo{
				ID:         upserted.ID,
				ExternalID: upserted.ExternalID,
				HomeTeam:   upserted.HomeTeam,
				AwayTeam:   upserted.AwayTeam,
				HomeScore:  upserted.HomeScore,
				AwayScore:  upserted.AwayScore,
				Finished:   upserted.Finished,
				StartsAt:   upserted.StartsAt,
				Stage:      upserted.Stage,
				CreatedAt:  upserted.CreatedAt,
				UpdatedAt:  upserted.UpdatedAt,
			})
		}

		// Dispatch WhatsApp pre-match notifications based on time until start.
		// Windows are 5 min wide (matching the sync interval) with 1 min buffer:
		//   faltam_dez_minutos: untilStart in [7min, 12min)
		//   partida_iniciando:  untilStart in [-2min, +2min)
		if !finished {
			untilStart := t.Sub(now)
			s.dispatchMatchNotifications(context.Background(), untilStart, m.HomeTeam.Name, m.AwayTeam.Name)
		}
	}

	slog.Info("synced jogos from football-data", "count", len(apiResp.Matches))
	return recentlyFinished, nil
}

// dispatchMatchNotifications fires WhatsApp pre-match alerts based on how far
// the match start is from now. Windows are 5 min wide with 1 min slack:
//
//	faltam_dez_minutos: [7min, 12min)
//	partida_iniciando:  [-2min, +2min)
//
// Notifications are sent to all bolões that have a wa_group_jid configured,
// regardless of whether they have palpites on this jogo.
func (s *JogoService) dispatchMatchNotifications(ctx context.Context, untilStart time.Duration, homeTeam, awayTeam string) {
	var notifyFn func(ctx context.Context, groupJID, homeTeam, awayTeam string)

	switch {
	case untilStart >= 7*time.Minute && untilStart < 12*time.Minute:
		slog.Info("wa notify: faltam_dez_minutos", "home", homeTeam, "away", awayTeam)
		notifyFn = s.waNotif.NotifyFaltamDezMinutos
	case untilStart >= -2*time.Minute && untilStart < 2*time.Minute:
		slog.Info("wa notify: partida_iniciando", "home", homeTeam, "away", awayTeam)
		notifyFn = s.waNotif.NotifyPartidaIniciando
	default:
		return
	}

	boloes, err := s.q.ListBoloesByWAGroup(ctx)
	if err != nil {
		slog.Warn("wa notify: listing boloes with wa group", "err", err)
		return
	}
	for _, b := range boloes {
		jid := b.WaGroupJid.String
		ht, at := homeTeam, awayTeam
		go notifyFn(ctx, jid, ht, at)
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
