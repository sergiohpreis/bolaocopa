package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

type FeedService struct {
	q repository.Querier
}

func NewFeedService(q repository.Querier) *FeedService {
	return &FeedService{q: q}
}

type FeedEventoResponse struct {
	ID        string    `json:"id"`
	BolaoID   string    `json:"bolao_id"`
	Tipo      string    `json:"tipo"`
	UserID    *string   `json:"user_id,omitempty"`
	UserName  *string   `json:"user_name,omitempty"`
	JogoID    *string   `json:"jogo_id,omitempty"`
	JogoDesc  *string   `json:"jogo_desc,omitempty"`
	Payload   any       `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *FeedService) ListByBolao(ctx context.Context, bolaoID, userID string) ([]FeedEventoResponse, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return nil, ErrBolaoNotFound
	}

	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return nil, ErrNotParticipante
	}

	rows, err := s.q.ListFeedByBolao(ctx, bid)
	if err != nil {
		return nil, fmt.Errorf("listing feed: %w", err)
	}

	now := time.Now()
	result := make([]FeedEventoResponse, 0, len(rows))
	for _, r := range rows {
		ev := FeedEventoResponse{
			ID:        uuidToString(r.ID),
			BolaoID:   uuidToString(r.BolaoID),
			Tipo:      string(r.Tipo),
			CreatedAt: r.CreatedAt.Time,
		}

		if r.UserID.Valid {
			s := uuidToString(r.UserID)
			ev.UserID = &s
		}
		if r.UserName.Valid {
			ev.UserName = &r.UserName.String
		}
		if r.JogoID.Valid {
			s := uuidToString(r.JogoID)
			ev.JogoID = &s
			if r.JogoHomeTeam.Valid && r.JogoAwayTeam.Valid {
				desc := r.JogoHomeTeam.String + " x " + r.JogoAwayTeam.String
				ev.JogoDesc = &desc
			}
		}

		var rawPayload map[string]any
		if err := json.Unmarshal(r.Payload, &rawPayload); err == nil {
			// Hide palpite scores if game hasn't started yet
			if (r.Tipo == repository.FeedTipoPalpiteRegistrado || r.Tipo == repository.FeedTipoPalpiteAlterado) &&
				r.JogoStartsAt.Valid && now.Before(r.JogoStartsAt.Time) {
				delete(rawPayload, "home_score")
				delete(rawPayload, "away_score")
			}
			ev.Payload = rawPayload
		} else {
			ev.Payload = map[string]any{}
		}

		result = append(result, ev)
	}

	if result == nil {
		result = []FeedEventoResponse{}
	}
	return result, nil
}

// InsertEvento is called by other services to publish feed events.
func (s *FeedService) InsertEvento(ctx context.Context, bolaoID string, tipo repository.FeedTipo, userID, jogoID *string, payload map[string]any) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return
	}

	var uid pgtype.UUID
	if userID != nil {
		uid, _ = parseUUID(*userID)
	}

	var jid pgtype.UUID
	if jogoID != nil {
		jid, _ = parseUUID(*jogoID)
	}

	b, _ := json.Marshal(payload)
	if b == nil {
		b = []byte("{}")
	}

	_, _ = s.q.InsertFeedEvento(ctx, repository.InsertFeedEventoParams{
		BolaoID: bid,
		Tipo:    tipo,
		UserID:  uid,
		JogoID:  jid,
		Payload: b,
	})
}

