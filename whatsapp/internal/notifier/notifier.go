// PROTOTYPE — throwaway. Formats and sends bolão notifications to WhatsApp.
package notifier

import (
	"context"
	"fmt"
	"strconv"
)

type Sender interface {
	SendText(ctx context.Context, jid, text string) error
	LinkedGroup() string
}

type Notifier struct {
	sender Sender
}

func New(sender Sender) *Notifier {
	return &Notifier{sender: sender}
}

// resolveJID returns targetJID if non-empty, otherwise falls back to the global linked group.
func (n *Notifier) resolveJID(targetJID string) (string, error) {
	if targetJID != "" {
		return targetJID, nil
	}
	jid := n.sender.LinkedGroup()
	if jid == "" {
		return "", fmt.Errorf("nenhum grupo vinculado")
	}
	return jid, nil
}

// PartidaAcabou envia resultado da partida e quem pontuou.
// targetJID: JID do grupo destino; se vazio usa o grupo global vinculado.
func (n *Notifier) PartidaAcabou(ctx context.Context, targetJID, homeTeam string, homeScore int, awayTeam string, awayScore int, winners []Winner) error {
	jid, err := n.resolveJID(targetJID)
	if err != nil {
		return err
	}

	linha := homeTeam + " " + strconv.Itoa(homeScore) + " x " + strconv.Itoa(awayScore) + " " + awayTeam
	msg := fmt.Sprintf("⚽ *FIM DE JOGO*\n%s\n\n", linha)

	if len(winners) == 0 {
		msg += "_Ninguém acertou o resultado._"
	} else {
		msg += "🏆 *Pontuaram:*\n"
		for _, w := range winners {
			msg += fmt.Sprintf("• %s — %.1f pts\n", w.Name, w.Pontos)
		}
	}

	return n.sender.SendText(ctx, jid, msg)
}

// FaltamDezMinutos envia lembrete antes da partida.
// targetJID: JID do grupo destino; se vazio usa o grupo global vinculado.
func (n *Notifier) FaltamDezMinutos(ctx context.Context, targetJID, homeTeam, awayTeam string) error {
	jid, err := n.resolveJID(targetJID)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"⏰ *Faltam 10 minutos!*\n%s x %s\nFaçam suas apostas antes que a partida comece! 🎯",
		homeTeam, awayTeam,
	)

	return n.sender.SendText(ctx, jid, msg)
}

// PartidaIniciando envia notificação de encerramento de apostas.
// targetJID: JID do grupo destino; se vazio usa o grupo global vinculado.
func (n *Notifier) PartidaIniciando(ctx context.Context, targetJID, homeTeam, awayTeam string) error {
	jid, err := n.resolveJID(targetJID)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"🚀 *A partida está começando!*\n%s x %s\nApostas encerradas. Boa sorte! 🍀",
		homeTeam, awayTeam,
	)

	return n.sender.SendText(ctx, jid, msg)
}

type Winner struct {
	Name   string
	Pontos float64
}
