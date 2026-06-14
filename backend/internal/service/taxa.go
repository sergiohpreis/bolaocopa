package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrTaxaJaDefinida   = errors.New("taxa de entrada já definida e não pode ser alterada")
	ErrPropostaJaExiste = errors.New("já existe uma proposta de taxa ativa")
	ErrJaVotou          = errors.New("participante já votou nesta proposta")
	ErrSemProposta      = errors.New("não há proposta de taxa ativa")
	ErrVotoNaoElegivel  = errors.New("participante entrou após a proposta e não pode votar")
)

type TaxaEstado struct {
	TaxaDefinida   *string         `json:"taxa_definida,omitempty"`
	PropostaAtiva  *PropostaResumo `json:"proposta_ativa,omitempty"`
	VotosPendentes int64           `json:"votos_pendentes"`
	MeuVoto        *bool           `json:"meu_voto,omitempty"`
}

type PropostaResumo struct {
	ID    string `json:"id"`
	Valor string `json:"valor"`
}

type TaxaService struct {
	q    repository.Querier
	pool *pgxpool.Pool
	feed *FeedService
}

func NewTaxaService(q repository.Querier, pool *pgxpool.Pool) *TaxaService {
	return &TaxaService{q: q, pool: pool}
}

func (s *TaxaService) SetFeed(feed *FeedService) {
	s.feed = feed
}

func (s *TaxaService) Propor(ctx context.Context, bolaoID, adminID string, valor string) (repository.TaxaEntradaProposta, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return repository.TaxaEntradaProposta{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(adminID)
	if err != nil {
		return repository.TaxaEntradaProposta{}, fmt.Errorf("invalid user id: %w", err)
	}

	bolao, err := s.q.GetBolaoByID(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.TaxaEntradaProposta{}, ErrBolaoNotFound
		}
		return repository.TaxaEntradaProposta{}, fmt.Errorf("getting bolao: %w", err)
	}
	if bolao.AdminID != uid {
		return repository.TaxaEntradaProposta{}, ErrNotAdmin
	}
	if bolao.TaxaEntrada.Valid {
		return repository.TaxaEntradaProposta{}, ErrTaxaJaDefinida
	}

	var params repository.ProporTaxaParams
	params.BolaoID = bid
	params.PropostaPor = uid
	if err := params.Valor.Scan(valor); err != nil {
		return repository.TaxaEntradaProposta{}, fmt.Errorf("invalid valor: %w", err)
	}

	proposta, err := s.q.ProporTaxa(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.TaxaEntradaProposta{}, ErrPropostaJaExiste
		}
		return repository.TaxaEntradaProposta{}, fmt.Errorf("creating proposta: %w", err)
	}
	return proposta, nil
}

func (s *TaxaService) Votar(ctx context.Context, bolaoID, userID string, aprovado bool) error {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return ErrNotParticipante
	}

	// Check proposta exists before starting tx (fast-fail without locking)
	propostaCheck, err := s.q.GetPropostaAtiva(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSemProposta
		}
		return fmt.Errorf("getting proposta: %w", err)
	}

	// Verify voter is eligible (joined at or before proposal creation)
	elegivel, err := s.q.IsParticipanteElegivel(ctx, repository.IsParticipanteElegivelParams{
		BolaoID:  bid,
		UserID:   uid,
		JoinedAt: propostaCheck.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("checking eligibility: %w", err)
	}
	if !elegivel {
		return ErrVotoNaoElegivel
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := repository.New(tx)

	// Lock the proposta row to serialize concurrent votes
	proposta, err := qtx.GetPropostaAtivaForUpdate(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSemProposta
		}
		return fmt.Errorf("locking proposta: %w", err)
	}

	_, err = qtx.RegistrarVoto(ctx, repository.RegistrarVotoParams{
		PropostaID: proposta.ID,
		UserID:     uid,
		Aprovado:   aprovado,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrJaVotou
		}
		return fmt.Errorf("registering vote: %w", err)
	}

	if !aprovado {
		if err := qtx.CancelarProposta(ctx, bid); err != nil {
			return fmt.Errorf("canceling proposta: %w", err)
		}
		return tx.Commit(ctx)
	}

	// Check unanimity over the eligible population (same set for both counts)
	total, err := qtx.CountParticipantesNaMomento(ctx, repository.CountParticipantesNaMomentoParams{
		BolaoID:  bid,
		JoinedAt: proposta.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("counting participantes: %w", err)
	}

	favoraveis, err := qtx.CountVotosFavoraveis(ctx, repository.CountVotosFavoraveisParams{
		PropostaID: proposta.ID,
		BolaoID:    bid,
		JoinedAt:   proposta.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("counting votes: %w", err)
	}

	taxaDefinida := false
	if total > 0 && favoraveis == total {
		if _, err := qtx.DefinirTaxa(ctx, repository.DefinirTaxaParams{
			ID:          bid,
			TaxaEntrada: proposta.Valor,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// taxa_entrada IS NULL check failed — another tx already set it
				return ErrTaxaJaDefinida
			}
			return fmt.Errorf("defining taxa: %w", err)
		}

		if err := qtx.CancelarProposta(ctx, bid); err != nil {
			return fmt.Errorf("removing proposta: %w", err)
		}
		taxaDefinida = true
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	// Emit feed event after commit — fire-and-forget
	if taxaDefinida && s.feed != nil {
		valor, _ := numericToString(proposta.Valor)
		s.feed.InsertEvento(ctx, bolaoID, repository.FeedTipoTaxaAprovada, nil, nil, map[string]any{"valor": valor})
	}

	return nil
}

func (s *TaxaService) GetEstado(ctx context.Context, bolaoID, userID string) (TaxaEstado, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return TaxaEstado{}, ErrBolaoNotFound
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("invalid user id: %w", err)
	}

	ok, err := s.q.IsParticipante(ctx, repository.IsParticipanteParams{BolaoID: bid, UserID: uid})
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("checking membership: %w", err)
	}
	if !ok {
		return TaxaEstado{}, ErrNotParticipante
	}

	taxa, err := s.q.GetTaxaEntrada(ctx, bid)
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("getting taxa: %w", err)
	}

	estado := TaxaEstado{}

	if taxa.Valid {
		v, err := numericToString(taxa)
		if err != nil {
			return TaxaEstado{}, fmt.Errorf("formatting taxa: %w", err)
		}
		estado.TaxaDefinida = &v
		return estado, nil
	}

	proposta, err := s.q.GetPropostaAtiva(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return estado, nil
		}
		return TaxaEstado{}, fmt.Errorf("getting proposta: %w", err)
	}

	valor, err := numericToString(proposta.Valor)
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("formatting proposta valor: %w", err)
	}
	estado.PropostaAtiva = &PropostaResumo{
		ID:    uuidToString(proposta.ID),
		Valor: valor,
	}

	total, err := s.q.CountParticipantesNaMomento(ctx, repository.CountParticipantesNaMomentoParams{
		BolaoID:  bid,
		JoinedAt: proposta.CreatedAt,
	})
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("counting participantes: %w", err)
	}

	votosAprovados, err := s.q.CountVotosFavoraveis(ctx, repository.CountVotosFavoraveisParams{
		PropostaID: proposta.ID,
		BolaoID:    bid,
		JoinedAt:   proposta.CreatedAt,
	})
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("counting votes: %w", err)
	}

	pendentes := total - votosAprovados
	if pendentes < 0 {
		pendentes = 0
	}
	estado.VotosPendentes = pendentes

	meuVoto, err := s.q.GetMeuVoto(ctx, repository.GetMeuVotoParams{
		PropostaID: proposta.ID,
		UserID:     uid,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return TaxaEstado{}, fmt.Errorf("getting meu_voto: %w", err)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		estado.MeuVoto = &meuVoto
	}

	return estado, nil
}
