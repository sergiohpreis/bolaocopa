package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrJogoNotFound        = errors.New("jogo not found")
	ErrPalpiteFechado      = errors.New("palpite fechado: jogo já começou")
	ErrJogoAindaAberto     = errors.New("jogo ainda não começou: palpite retroativo não permitido")
	ErrPalpiteNaoEncontrado = errors.New("palpite not found")
)

type PalpiteService struct {
	q    repository.Querier
	feed *FeedService
}

func NewPalpiteService(q repository.Querier) *PalpiteService {
	return &PalpiteService{q: q}
}

func (s *PalpiteService) SetFeed(feed *FeedService) {
	s.feed = feed
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

	p, err := s.q.UpsertPalpite(ctx, repository.UpsertPalpiteParams{
		BolaoID:   bid,
		UserID:    uid,
		JogoID:    jid,
		HomeScore: int32(homeScore),
		AwayScore: int32(awayScore),
	})
	if err != nil {
		return p, err
	}
	if s.feed != nil {
		tipo := repository.FeedTipoPalpiteRegistrado
		if p.UpdatedAt.Time.After(p.CreatedAt.Time) {
			tipo = repository.FeedTipoPalpiteAlterado
		}
		jogoIDStr := jogoID
		s.feed.InsertEvento(ctx, bolaoID, tipo, &userID, &jogoIDStr, map[string]any{
			"home_score": homeScore,
			"away_score": awayScore,
		})
	}
	return p, nil
}

func (s *PalpiteService) ListByJogo(ctx context.Context, bolaoID, userID, jogoID string) ([]repository.ListPalpitesByBolaoAndJogoRow, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return nil, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	jid, err := parseUUID(jogoID)
	if err != nil {
		return nil, ErrJogoNotFound
	}

	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return nil, ErrNotParticipante
	}

	items, err := s.q.ListPalpitesByBolaoAndJogo(ctx, repository.ListPalpitesByBolaoAndJogoParams{BolaoID: bid, JogoID: jid})
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.ListPalpitesByBolaoAndJogoRow{}
	}
	return items, nil
}

func (s *PalpiteService) UpsertRetroativo(ctx context.Context, bolaoID, userID, jogoID string, homeScore, awayScore int) (repository.Palpite, error) {
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

	if !time.Now().After(jogo.StartsAt.Time) {
		return repository.Palpite{}, ErrJogoAindaAberto
	}

	return s.q.UpsertPalpiteRetroativo(ctx, repository.UpsertPalpiteRetroativoParams{
		BolaoID:   bid,
		UserID:    uid,
		JogoID:    jid,
		HomeScore: int32(homeScore),
		AwayScore: int32(awayScore),
	})
}

func (s *PalpiteService) ListPendentes(ctx context.Context, bolaoID, adminUserID string) ([]repository.ListPalpitesPendentesRow, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return nil, ErrBolaoNotFound
	}
	uid, err := parseUUID(adminUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	bolao, err := s.q.GetBolaoByID(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBolaoNotFound
		}
		return nil, fmt.Errorf("getting bolao: %w", err)
	}
	if bolao.AdminID != uid {
		return nil, ErrNotAdmin
	}

	items, err := s.q.ListPalpitesPendentes(ctx, bid)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.ListPalpitesPendentesRow{}
	}
	return items, nil
}

func (s *PalpiteService) AprovarOuRejeitar(ctx context.Context, bolaoID, palpiteID, adminUserID string, aprovar bool) (repository.Palpite, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.Palpite{}, ErrBolaoNotFound
	}
	pid, err := parseUUID(palpiteID)
	if err != nil {
		return repository.Palpite{}, ErrPalpiteNaoEncontrado
	}
	uid, err := parseUUID(adminUserID)
	if err != nil {
		return repository.Palpite{}, fmt.Errorf("invalid user id: %w", err)
	}

	bolao, err := s.q.GetBolaoByID(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Palpite{}, ErrBolaoNotFound
		}
		return repository.Palpite{}, fmt.Errorf("getting bolao: %w", err)
	}
	if bolao.AdminID != uid {
		return repository.Palpite{}, ErrNotAdmin
	}

	status := "rejeitado"
	if aprovar {
		status = "aprovado"
	}

	p, err := s.q.AtualizarStatusPalpite(ctx, repository.AtualizarStatusPalpiteParams{
		ID:      pid,
		BolaoID: bid,
		Status:  status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Palpite{}, ErrPalpiteNaoEncontrado
		}
		return repository.Palpite{}, fmt.Errorf("updating palpite status: %w", err)
	}

	if aprovar {
		// Se o jogo já terminou, calcular e persistir pontos imediatamente
		jogo, err := s.q.GetJogoByID(ctx, p.JogoID)
		if err == nil && jogo.Finished && jogo.HomeScore.Valid && jogo.AwayScore.Valid {
			pontos := calcPontos(p.HomeScore, p.AwayScore, jogo.HomeScore.Int32, jogo.AwayScore.Int32)
			_ = s.q.UpdatePalpitePontos(ctx, repository.UpdatePalpitePontosParams{
				Pontos:  pgtype.Int4{Int32: pontos, Valid: true},
				BolaoID: p.BolaoID,
				JogoID:  p.JogoID,
				UserID:  p.UserID,
			})
		}

		if s.feed != nil {
			userIDStr := p.UserID.String()
			jogoIDStr := p.JogoID.String()
			s.feed.InsertEvento(ctx, bolaoID, repository.FeedTipoPalpiteRegistrado, &userIDStr, &jogoIDStr, map[string]any{
				"home_score": int(p.HomeScore),
				"away_score": int(p.AwayScore),
			})
		}
	}

	return p, nil
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
