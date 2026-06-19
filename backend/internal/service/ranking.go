package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

type RankingService struct {
	q       repository.Querier
	feed    *FeedService
	waNotif WANotifier
}

func NewRankingService(q repository.Querier) *RankingService {
	return &RankingService{q: q, waNotif: NewNoopWANotifier()}
}

func (s *RankingService) SetFeed(feed *FeedService) {
	s.feed = feed
}

func (s *RankingService) SetWANotifier(n WANotifier) {
	s.waNotif = n
}

func (s *RankingService) Get(ctx context.Context, bolaoID string) ([]repository.GetRankingRow, error) {
	bid, err := parseUUID(bolaoID)
	if err != nil {
		return nil, fmt.Errorf("invalid bolao id: %w", err)
	}
	items, err := s.q.GetRanking(ctx, bid)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.GetRankingRow{}
	}
	return items, nil
}

// NotifyRecentlyFinished dispara a notificação de fim de jogo para jogos recém-finalizados.
// Envia uma notificação por bolão que tem wa_group_jid configurado e palpites no jogo,
// com os winners daquele bolão específico.
func (s *RankingService) NotifyRecentlyFinished(ctx context.Context, jogos []repository.Jogo) {
	for _, jogo := range jogos {
		boloes, err := s.q.ListBoloesByJogo(ctx, jogo.ID)
		if err != nil {
			slog.Warn("wa notify: listing boloes for jogo", "jogo", jogo.ExternalID, "err", err)
			continue
		}

		for _, bolao := range boloes {
			groupJID := bolao.WaGroupJid.String
			if groupJID == "" {
				continue
			}

			rows, err := s.q.ListPalpitesByBolaoAndJogo(ctx, repository.ListPalpitesByBolaoAndJogoParams{
				BolaoID: bolao.ID,
				JogoID:  jogo.ID,
			})
			if err != nil {
				slog.Warn("wa notify: listing palpites per bolao", "bolao", uuidToString(bolao.ID), "jogo", jogo.ExternalID, "err", err)
				continue
			}

			var winners []WAWinner
			for _, r := range rows {
				if r.Pontos.Valid && r.Pontos.Int32 > 0 {
					winners = append(winners, WAWinner{Name: r.UserName, Pontos: int(r.Pontos.Int32)})
				}
			}

			slog.Info("wa notify: fim_de_jogo", "home", jogo.HomeTeam, "away", jogo.AwayTeam, "bolao", uuidToString(bolao.ID), "winners", len(winners))
			jid := groupJID
			go s.waNotif.NotifyFimDeJogo(context.Background(), jid,
				jogo.HomeTeam, int(jogo.HomeScore.Int32),
				jogo.AwayTeam, int(jogo.AwayScore.Int32),
				winners,
			)
		}
	}
}

// ComputeScoresForFinishedJogos fetches all finished jogos with known scores
// and computes pontos for every palpite.
func (s *RankingService) ComputeScoresForFinishedJogos(ctx context.Context) error {
	jogos, err := s.q.ListFinishedJobsWithoutScores(ctx)
	if err != nil {
		return fmt.Errorf("listing finished jogos: %w", err)
	}

	for _, jogo := range jogos {
		if !jogo.HomeScore.Valid || !jogo.AwayScore.Valid {
			continue
		}
		if err := s.scoreJogo(ctx, jogo); err != nil {
			slog.Warn("failed to score jogo", "jogo_id", jogo.ID, "error", err)
		}
	}
	return nil
}

func (s *RankingService) scoreJogo(ctx context.Context, jogo repository.Jogo) error {
	palpites, err := s.q.ListPalpitesByJogo(ctx, jogo.ID)
	if err != nil {
		return fmt.Errorf("listing palpites: %w", err)
	}

	// boloesComPalpiteNovo rastreia bolões que tinham palpites sem pontos —
	// só esses recebem o evento de feed (evita duplicar resultado_apurado a cada run do job).
	boloesComPalpiteNovo := map[string]bool{}
	jogoIDStr := uuidToString(jogo.ID)

	for _, p := range palpites {
		if p.Pontos.Valid {
			continue
		}
		pontos := calcPontos(p.HomeScore, p.AwayScore, jogo.HomeScore.Int32, jogo.AwayScore.Int32)
		if err := s.q.UpdatePalpitePontos(ctx, repository.UpdatePalpitePontosParams{
			Pontos:  pgtype.Int4{Int32: pontos, Valid: true},
			BolaoID: p.BolaoID,
			JogoID:  p.JogoID,
			UserID:  p.UserID,
		}); err != nil {
			slog.Warn("failed to update pontos", "palpite_id", p.ID, "error", err)
		}
		bolaoIDStr := uuidToString(p.BolaoID)
		boloesComPalpiteNovo[bolaoIDStr] = true
	}

	for bolaoIDStr := range boloesComPalpiteNovo {
		if s.feed != nil {
			s.feed.InsertEvento(ctx, bolaoIDStr, repository.FeedTipoResultadoApurado, nil, &jogoIDStr, map[string]any{
				"home_score": jogo.HomeScore.Int32,
				"away_score": jogo.AwayScore.Int32,
			})
		}
	}

	return nil
}

func calcPontos(palHome, palAway, resHome, resAway int32) int32 {
	if palHome == resHome && palAway == resAway {
		return 10
	}
	if winner(palHome, palAway) == winner(resHome, resAway) {
		return 3
	}
	return 0
}

func winner(home, away int32) int {
	if home > away {
		return 1
	}
	if away > home {
		return -1
	}
	return 0
}
