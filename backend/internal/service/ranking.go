package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

type RankingService struct {
	q        repository.Querier
	feed     *FeedService
	waNotif  WANotifier
	notifier matchNotifier
}

func NewRankingService(q repository.Querier) *RankingService {
	noop := NewNoopWANotifier()
	return &RankingService{q: q, waNotif: noop, notifier: matchNotifier{q: q, waNotif: noop}}
}

func (s *RankingService) SetFeed(feed *FeedService) {
	s.feed = feed
}

func (s *RankingService) SetWANotifier(n WANotifier) {
	s.waNotif = n
	s.notifier.waNotif = n
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
// Dedup é por (jogo_id + bolao_id) codificado no notification_type para garantir que um
// crash parcial não impeça bolões restantes de receberem a notificação.
func (s *RankingService) NotifyRecentlyFinished(ctx context.Context, jogos []repository.Jogo) {
	boloes, err := s.q.ListBoloesByWAGroup(ctx)
	if err != nil {
		slog.Warn("wa notify: listing boloes with wa group", "err", err)
		return
	}

	sem := make(chan struct{}, maxConcurrentSends)
	var wg sync.WaitGroup

	for _, jogo := range jogos {
		homeTeam := translateTeam(jogo.HomeTeam)
		awayTeam := translateTeam(jogo.AwayTeam)

		for _, bolao := range boloes {
			// Dedup key includes bolao_id so a crash mid-loop doesn't permanently
			// suppress the remaining bolões for this jogo, and the final score so a
			// corrected result (e.g. an annulled goal) re-notifies with the right
			// score and winners — each distinct score notifies at most once.
			dedupType := fmt.Sprintf("%s:%d-%d:%s", NotifFimDeJogo,
				jogo.HomeScore.Int32, jogo.AwayScore.Int32, uuidToString(bolao.ID))

			sent, err := s.q.InsertJogoNotificationIfAbsent(ctx, repository.InsertJogoNotificationIfAbsentParams{
				JogoID:           jogo.ID,
				NotificationType: dedupType,
			})
			if err != nil {
				slog.Warn("wa notify: insert dedup record", "jogo", jogo.ExternalID, "bolao", uuidToString(bolao.ID), "err", err)
				continue
			}
			if sent == 0 {
				continue
			}

			palpites, err := s.q.ListPalpitesByBolaoAndJogo(ctx, repository.ListPalpitesByBolaoAndJogoParams{
				BolaoID: bolao.ID,
				JogoID:  jogo.ID,
			})
			if err != nil {
				slog.Warn("wa notify: listing palpites per bolao", "bolao", uuidToString(bolao.ID), "jogo", jogo.ExternalID, "err", err)
				continue
			}

			var winners []WAWinner
			for _, r := range palpites {
				if pts, ok := numericToFloat64(r.Pontos); ok && pts > 0 {
					winners = append(winners, WAWinner{Name: r.UserName, Pontos: pts})
				}
			}

			slog.Info("wa notify: dispatching", "type", NotifFimDeJogo, "home", homeTeam, "away", awayTeam, "bolao", uuidToString(bolao.ID), "winners", len(winners))

			jid := bolao.WaGroupJid.String
			home, away := homeTeam, awayTeam
			homeScore, awayScore := int(jogo.HomeScore.Int32), int(jogo.AwayScore.Int32)
			w := winners

			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				s.waNotif.NotifyFimDeJogo(ctx, jid, home, homeScore, away, awayScore, w)
			}()
		}
	}

	wg.Wait()
}

// ComputeScoresForFinishedJogos fetches all finished jogos with known scores
// and (re)computes pontos for every palpite. Re-running is safe and cheap:
// scoreJogo only writes when a palpite's computed score changed, so a corrected
// final score (e.g. an annulled goal) is reflected on the next sync.
func (s *RankingService) ComputeScoresForFinishedJogos(ctx context.Context) error {
	jogos, err := s.q.ListFinishedJogosWithScores(ctx)
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

	if jogo.Stage != "GROUP_STAGE" {
		if _, known := stageMultiplier[jogo.Stage]; !known {
			slog.Warn("unknown knockout stage, scoring as group stage", "stage", jogo.Stage, "jogo_id", jogo.ID)
		}
	}

	for _, p := range palpites {
		pontos := calcPontos(p.HomeScore, p.AwayScore, jogo.HomeScore.Int32, jogo.AwayScore.Int32, jogo.Stage, jogo.Winner.String, p.PenaltyWinner.String)

		// Skip the write when the stored score already matches — keeps the job
		// idempotent so it can run every sync. When the jogo's final score is
		// corrected (e.g. an annulled goal), the recomputed value differs and we
		// update in place, healing stale pontos.
		if stored, ok := numericToFloat64(p.Pontos); ok && fmt.Sprintf("%.1f", stored) == fmt.Sprintf("%.1f", pontos) {
			continue
		}
		pontosNumeric, err := float64ToNumeric(pontos)
		if err != nil {
			slog.Warn("failed to convert pontos to numeric", "palpite_id", p.ID, "pontos", pontos, "error", err)
			continue
		}
		if err := s.q.UpdatePalpitePontos(ctx, repository.UpdatePalpitePontosParams{
			Pontos:  pontosNumeric,
			BolaoID: p.BolaoID,
			JogoID:  p.JogoID,
			UserID:  p.UserID,
		}); err != nil {
			slog.Warn("failed to update pontos", "palpite_id", p.ID, "error", err)
			continue
		}
		// resultado_apurado só é emitido para palpites que ainda não tinham pontos —
		// correções de placar atualizam o valor mas não re-disparam o evento de feed.
		if !p.Pontos.Valid {
			bolaoIDStr := uuidToString(p.BolaoID)
			boloesComPalpiteNovo[bolaoIDStr] = true
		}
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

// stageMultiplier maps football-data.org stage strings to point multipliers.
// LAST_16 and ROUND_OF_16 are both mapped — the API uses different strings
// depending on the tournament edition (e.g. Copa 2026 vs earlier World Cups).
var stageMultiplier = map[string]float64{
	"LAST_32":        1.5,
	"LAST_16":        2.0,
	"ROUND_OF_16":    2.0,
	"QUARTER_FINALS": 2.5,
	"SEMI_FINALS":    3.0,
	"THIRD_PLACE":    3.5,
	"FINAL":          3.5,
}

func calcPontos(palHome, palAway, resHome, resAway int32, stage, apiWinner, penaltyWinner string) float64 {
	mult, isKnockout := stageMultiplier[stage]

	if isKnockout {
		palSide := palSideWinner(palHome, palAway)
		if apiWinner != "" {
			if palSide != "" && palSide == apiWinner {
				// Apostou vitória e acertou quem avançou.
				if palHome == resHome && palAway == resAway {
					return 10.0 * mult
				}
				return 3.0 * mult
			}
			if palSide == "" && palHome == resHome && palAway == resAway {
				// Apostou empate exato e o jogo foi a pênaltis.
				// penaltyWinner ("home"/"away") indica quem o participante escolheu para avançar.
				// NULL (string vazia) ocorre em palpites antigos — tratado como erro de vencedor.
				if (penaltyWinner == "home" && apiWinner == "HOME_TEAM") ||
					(penaltyWinner == "away" && apiWinner == "AWAY_TEAM") {
					return 10.0*mult + 3.0
				}
				return 3.0 * mult
			}
		}
		return 0
	}

	// Fase de Grupos
	if palHome == resHome && palAway == resAway {
		return 10
	}
	if scoreWinner(palHome, palAway) == scoreWinner(resHome, resAway) {
		return 3
	}
	return 0
}

// palSideWinner retorna "HOME_TEAM" ou "AWAY_TEAM" conforme o placar chutado.
// Retorna "" para empate — no mata-mata empate sem winner não pontua "quem avança".
func palSideWinner(home, away int32) string {
	if home > away {
		return "HOME_TEAM"
	}
	if away > home {
		return "AWAY_TEAM"
	}
	return ""
}

func scoreWinner(home, away int32) int {
	if home > away {
		return 1
	}
	if away > home {
		return -1
	}
	return 0
}

func numericToFloat64(n pgtype.Numeric) (float64, bool) {
	if !n.Valid {
		return 0, false
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0, false
	}
	return f.Float64, true
}

func float64ToNumeric(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(fmt.Sprintf("%.1f", f))
	return n, err
}
