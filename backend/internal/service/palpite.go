package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrJogoNotFound   = errors.New("jogo not found")
	ErrPalpiteFechado = errors.New("palpite fechado: jogo já começou")
)

type PalpiteService struct {
	q repository.Querier
}

func NewPalpiteService(q repository.Querier) *PalpiteService {
	return &PalpiteService{q: q}
}

func (s *PalpiteService) Upsert(ctx context.Context, bolaoID, userID, jogoID string, homeScore, awayScore int) (repository.Palpite, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.Palpite{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Palpite{}, fmt.Errorf("invalid user id: %w", err)
	}
	jid, err := parseUUID(jogoID)
	if err != nil {
		return repository.Palpite{}, ErrJogoNotFound
	}

	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return repository.Palpite{}, fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return repository.Palpite{}, ErrNotParticipante
	}

	jogo, err := s.q.GetJogoByID(ctx, jid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Palpite{}, ErrJogoNotFound
		}
		return repository.Palpite{}, fmt.Errorf("getting jogo: %w", err)
	}

	if time.Now().After(jogo.StartsAt.Time) {
		return repository.Palpite{}, ErrPalpiteFechado
	}

	return s.q.UpsertPalpite(ctx, repository.UpsertPalpiteParams{
		BolaoID:   bid,
		UserID:    uid,
		JogoID:    jid,
		HomeScore: int32(homeScore),
		AwayScore: int32(awayScore),
	})
}

func (s *PalpiteService) ListByUser(ctx context.Context, bolaoID, userID string) ([]repository.Palpite, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return nil, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	items, err := s.q.ListPalpitesByBolaoAndUser(ctx, repository.ListPalpitesByBolaoAndUserParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.Palpite{}
	}
	return items, nil
}
