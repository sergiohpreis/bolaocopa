package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrTaxaJaDefinida  = errors.New("taxa de entrada já definida e não pode ser alterada")
	ErrPropostaJaExiste = errors.New("já existe uma proposta de taxa ativa")
	ErrJaVotou         = errors.New("participante já votou nesta proposta")
	ErrSemProposta     = errors.New("não há proposta de taxa ativa")
)

type TaxaEstado struct {
	TaxaDefinida   *string          `json:"taxa_definida,omitempty"`
	PropostaAtiva  *PropostaResumo  `json:"proposta_ativa,omitempty"`
	MeuVoto        *bool            `json:"meu_voto,omitempty"`
	VotosPendentes int64            `json:"votos_pendentes"`
}

type PropostaResumo struct {
	ID    string `json:"id"`
	Valor string `json:"valor"`
}

type TaxaService struct {
	q    repository.Querier
	feed *FeedService
}

func NewTaxaService(q repository.Querier) *TaxaService {
	return &TaxaService{q: q}
}

func (s *TaxaService) SetFeed(feed *FeedService) {
	s.feed = feed
}

func (s *TaxaService) Propor(ctx context.Context, bolaoID, adminID string, valor float64) (repository.TaxaEntradaProposta, error) {
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

	_, err = s.q.GetPropostaAtiva(ctx, bid)
	if err == nil {
		return repository.TaxaEntradaProposta{}, ErrPropostaJaExiste
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return repository.TaxaEntradaProposta{}, fmt.Errorf("checking proposta: %w", err)
	}

	var valorNum pgtype.Numeric
	if err := valorNum.Scan(fmt.Sprintf("%.2f", valor)); err != nil {
		return repository.TaxaEntradaProposta{}, fmt.Errorf("invalid valor: %w", err)
	}

	proposta, err := s.q.ProporTaxa(ctx, repository.ProporTaxaParams{
		BolaoID:     bid,
		Valor:       valorNum,
		PropostaPor: uid,
	})
	if err != nil {
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

	proposta, err := s.q.GetPropostaAtiva(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSemProposta
		}
		return fmt.Errorf("getting proposta: %w", err)
	}

	_, err = s.q.RegistrarVoto(ctx, repository.RegistrarVotoParams{
		PropostaID: proposta.ID,
		UserID:     uid,
		Aprovado:   aprovado,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING returned no rows — já votou
			return ErrJaVotou
		}
		return fmt.Errorf("registering vote: %w", err)
	}

	if !aprovado {
		return s.q.CancelarProposta(ctx, bid)
	}

	// Verifica se todos os participantes elegíveis já votaram a favor
	total, err := s.q.CountParticipantesNaMomento(ctx, repository.CountParticipantesNaMomentoParams{
		BolaoID:  bid,
		JoinedAt: proposta.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("counting participantes: %w", err)
	}

	favoraveis, err := s.q.CountVotosFavoraveis(ctx, proposta.ID)
	if err != nil {
		return fmt.Errorf("counting votes: %w", err)
	}

	if favoraveis >= total && total > 0 {
		if _, err := s.q.DefinirTaxa(ctx, repository.DefinirTaxaParams{
			ID:          bid,
			TaxaEntrada: proposta.Valor,
		}); err != nil {
			return fmt.Errorf("defining taxa: %w", err)
		}

		if err := s.q.CancelarProposta(ctx, bid); err != nil {
			return fmt.Errorf("removing proposta: %w", err)
		}

		if s.feed != nil {
			valor := numericToString(proposta.Valor)
			s.feed.InsertEvento(ctx, bolaoID, repository.FeedTipoTaxaAprovada, nil, nil, map[string]any{"valor": valor})
		}
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
		v := numericToString(taxa)
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

	propostaID := uuidToString(proposta.ID)
	estado.PropostaAtiva = &PropostaResumo{
		ID:    propostaID,
		Valor: numericToString(proposta.Valor),
	}

	votosAprovados, err := s.q.CountVotosFavoraveis(ctx, proposta.ID)
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("counting votes: %w", err)
	}

	total, err := s.q.CountParticipantesNaMomento(ctx, repository.CountParticipantesNaMomentoParams{
		BolaoID:  bid,
		JoinedAt: proposta.CreatedAt,
	})
	if err != nil {
		return TaxaEstado{}, fmt.Errorf("counting participantes: %w", err)
	}

	pendentes := total - votosAprovados
	if pendentes < 0 {
		pendentes = 0
	}
	estado.VotosPendentes = pendentes

	return estado, nil
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return ""
	}
	f, _ := n.Float64Value()
	return fmt.Sprintf("%.2f", f.Float64)
}
