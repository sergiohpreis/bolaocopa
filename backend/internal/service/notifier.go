package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

const (
	NotifFaltamDezMinutos = "faltam_dez_minutos"
	NotifPartidaIniciando = "partida_iniciando"
	NotifFimDeJogo        = "fim_de_jogo"

	// maxConcurrentSends caps the goroutine fan-out per notification event.
	maxConcurrentSends = 10
)

// matchNotifier owns the shared logic for dedup + fan-out across all notification types.
// Both JogoService and RankingService embed it to avoid duplicating the pattern.
type matchNotifier struct {
	q       repository.Querier
	waNotif WANotifier
}

// notifyOnce inserts a dedup record for (jogoID, notificationType) and, if this is the
// first insert (rows == 1), calls sendFn once per bolão with a wa_group_jid configured.
//
// Delivery is best-effort at-most-once: the dedup record is committed before sends are
// dispatched. A transient send failure will NOT be retried on the next sync tick.
// sendFn receives the bolão JID; it must be non-blocking or handle its own timeout.
func (n *matchNotifier) notifyOnce(ctx context.Context, jogoID pgtype.UUID, notificationType string, sendFn func(ctx context.Context, jid string)) {
	rows, err := n.q.InsertJogoNotificationIfAbsent(ctx, repository.InsertJogoNotificationIfAbsentParams{
		JogoID:           jogoID,
		NotificationType: notificationType,
	})
	if err != nil {
		slog.Warn("wa notify: insert dedup record", "type", notificationType, "err", err)
		return
	}
	if rows == 0 {
		// Already dispatched in a previous tick — dedup skip.
		return
	}

	boloes, err := n.q.ListBoloesByWAGroup(ctx)
	if err != nil {
		slog.Warn("wa notify: listing boloes with wa group", "type", notificationType, "err", err)
		return
	}

	sem := make(chan struct{}, maxConcurrentSends)
	var wg sync.WaitGroup
	for _, b := range boloes {
		jid := b.WaGroupJid.String
		wg.Add(1)
		sem <- struct{}{}
		go func(jid string) {
			defer wg.Done()
			defer func() { <-sem }()
			sendFn(ctx, jid)
		}(jid)
	}
	wg.Wait()
}
