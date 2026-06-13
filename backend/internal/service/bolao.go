package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrBolaoNotFound   = errors.New("bolao not found")
	ErrNotAdmin        = errors.New("only the admin can perform this action")
	ErrNotParticipante = errors.New("user is not a participante of this bolao")
)

type BolaoService struct {
	q    repository.Querier
	feed *FeedService
}

func NewBolaoService(q repository.Querier) *BolaoService {
	return &BolaoService{q: q}
}

func (s *BolaoService) SetFeed(feed *FeedService) {
	s.feed = feed
}

func (s *BolaoService) Create(ctx context.Context, name, userID string) (repository.Bolo, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("invalid user id: %w", err)
	}
	bolao, err := s.q.CreateBolao(ctx, repository.CreateBolaoParams{Name: name, AdminID: uid})
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("creating bolao: %w", err)
	}
	if _, err := s.q.JoinBolao(ctx, repository.JoinBolaoParams{BolaoID: bolao.ID, UserID: uid}); err != nil {
		return repository.Bolo{}, fmt.Errorf("adding admin as participante: %w", err)
	}
	return bolao, nil
}

func (s *BolaoService) GetByID(ctx context.Context, bolaoID, userID string) (repository.Bolo, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.Bolo{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Bolo{}, ErrBolaoNotFound
	}
	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return repository.Bolo{}, ErrNotParticipante
	}
	bolao, err := s.q.GetBolaoByID(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Bolo{}, ErrBolaoNotFound
		}
		return repository.Bolo{}, fmt.Errorf("getting bolao: %w", err)
	}
	return bolao, nil
}

func (s *BolaoService) JoinByToken(ctx context.Context, token, userID string) (repository.Bolo, error) {
	bolao, err := s.q.GetBolaoByInviteToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Bolo{}, ErrBolaoNotFound
		}
		return repository.Bolo{}, fmt.Errorf("finding bolao by token: %w", err)
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("invalid user id: %w", err)
	}
	if _, err = s.q.JoinBolao(ctx, repository.JoinBolaoParams{BolaoID: bolao.ID, UserID: uid}); err != nil {
		return repository.Bolo{}, fmt.Errorf("joining bolao: %w", err)
	}
	if s.feed != nil {
		bolaoIDStr := uuidToString(bolao.ID)
		s.feed.InsertEvento(ctx, bolaoIDStr, repository.FeedTipoParticipanteEntrou, &userID, nil, nil)
	}
	return bolao, nil
}

func (s *BolaoService) ListByUser(ctx context.Context, userID string) ([]repository.Bolo, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	items, err := s.q.ListBoloesByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.Bolo{}
	}
	return items, nil
}

func (s *BolaoService) RegenerateInviteToken(ctx context.Context, bolaoID, userID string) (repository.Bolo, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.Bolo{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("invalid user id: %w", err)
	}
	bolao, err := s.q.RegenerateInviteToken(ctx, repository.RegenerateInviteTokenParams{ID: bid, AdminID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Bolo{}, ErrNotAdmin
		}
		return repository.Bolo{}, fmt.Errorf("regenerating token: %w", err)
	}
	return bolao, nil
}

func (s *BolaoService) SetRetroativoEnabled(ctx context.Context, bolaoID, userID string, enabled bool) (repository.Bolo, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.Bolo{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.Bolo{}, fmt.Errorf("invalid user id: %w", err)
	}
	bolao, err := s.q.SetRetroativoEnabled(ctx, repository.SetRetroativoEnabledParams{ID: bid, AdminID: uid, RetroativoEnabled: enabled})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Bolo{}, ErrNotAdmin
		}
		return repository.Bolo{}, fmt.Errorf("setting retroativo_enabled: %w", err)
	}
	return bolao, nil
}
