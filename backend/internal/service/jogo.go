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
	q        repository.Querier
	apiKey   string
	waNotif  WANotifier
	notifier matchNotifier
}

func NewJogoService(q repository.Querier, apiKey string) *JogoService {
	noop := NewNoopWANotifier()
	return &JogoService{q: q, apiKey: apiKey, waNotif: noop, notifier: matchNotifier{q: q, waNotif: noop}}
}

func (s *JogoService) SetWANotifier(n WANotifier) {
	s.waNotif = n
	s.notifier.waNotif = n
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
		winner := pgtype.Text{}
		if finished {
			// For penalty shootout matches, fullTime includes the penalty goals on top
			// of regulation — use regularTime for the display/scoring score.
			scoreGoals := m.Score.FullTime
			if m.Score.Duration == "PENALTY_SHOOTOUT" && m.Score.RegularTime.Home != nil && m.Score.RegularTime.Away != nil {
				scoreGoals = m.Score.RegularTime
			}
			if scoreGoals.Home != nil && scoreGoals.Away != nil {
				homeScore = pgtype.Int4{Int32: int32(*scoreGoals.Home), Valid: true}
				awayScore = pgtype.Int4{Int32: int32(*scoreGoals.Away), Valid: true}
			}
			if m.Score.Winner != "" {
				winner = pgtype.Text{String: m.Score.Winner, Valid: true}
			}
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
			Winner:       winner,
		}

		upserted, err := s.q.UpsertJogo(ctx, params)
		if err != nil {
			slog.Warn("failed to upsert jogo", "external_id", params.ExternalID, "error", err)
			continue
		}

		// Notify on two transitions, both surfaced via the score-aware dedup key in
		// NotifyRecentlyFinished (which keys on the final score, so each distinct score
		// notifies at most once):
		//   1. !finished → finished (first fim-de-jogo)
		//   2. already finished but the final score changed (e.g. an annulled goal)
		// WasFinished/WasHomeScore/WasAwayScore come from the CTE reading pre-upsert state.
		//
		// For knockout matches we wait until winner is populated — the API sometimes
		// sets FINISHED before filling winner, and the WA message would show wrong points.
		scoreChanged := upserted.HomeScore != upserted.WasHomeScore || upserted.AwayScore != upserted.WasAwayScore
		newlyFinished := upserted.Finished && !upserted.WasFinished
		correctedScore := upserted.Finished && upserted.WasFinished && scoreChanged
		_, isKnockout := stageMultiplier[m.Stage]
		winnerReady := !isKnockout || upserted.Winner.Valid
		if (newlyFinished || correctedScore) && upserted.HomeScore.Valid && upserted.AwayScore.Valid && winnerReady {
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
				Winner:     upserted.Winner,
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
			s.dispatchMatchNotifications(context.Background(), upserted.ID, untilStart, m.HomeTeam.Name, m.AwayTeam.Name)
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
// The gap [2min, 7min) triggers nothing — intentional, avoids double-alerting.
// Notifications are sent to all bolões that have a wa_group_jid configured.
func (s *JogoService) dispatchMatchNotifications(ctx context.Context, jogoID pgtype.UUID, untilStart time.Duration, homeTeam, awayTeam string) {
	var notificationType string

	switch {
	case untilStart >= 7*time.Minute && untilStart < 12*time.Minute:
		notificationType = NotifFaltamDezMinutos
	case untilStart >= -2*time.Minute && untilStart < 2*time.Minute:
		notificationType = NotifPartidaIniciando
	default:
		return
	}

	home := translateTeam(homeTeam)
	away := translateTeam(awayTeam)
	slog.Info("wa notify: dispatching", "type", notificationType, "home", home, "away", away)

	s.notifier.notifyOnce(ctx, jogoID, notificationType, func(ctx context.Context, b repository.Bolo) {
		jid := b.WaGroupJid.String
		switch notificationType {
		case NotifFaltamDezMinutos:
			pendentes, err := s.q.ListParticipantesSemPalpite(ctx, repository.ListParticipantesSemPalpiteParams{
				BolaoID: b.ID,
				JogoID:  jogoID,
			})
			if err != nil {
				slog.Warn("wa notify: listing pending bettors", "bolao_id", b.ID, "err", err)
			}
			nomes := make([]string, len(pendentes))
			for i, p := range pendentes {
				nomes[i] = p.UserName
			}
			s.waNotif.NotifyFaltamDezMinutos(ctx, jid, home, away, nomes)
		case NotifPartidaIniciando:
			s.waNotif.NotifyPartidaIniciando(ctx, jid, home, away)
		}
	})
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
	Winner      string  `json:"winner"`
	Duration    string  `json:"duration"`
	FullTime    fdGoals `json:"fullTime"`
	RegularTime fdGoals `json:"regularTime"`
}

type fdGoals struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}
