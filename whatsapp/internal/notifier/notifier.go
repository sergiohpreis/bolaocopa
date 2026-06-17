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

// PartidaAcabou envia resultado da partida e quem pontuou.
// winners: lista de nomes com pontos.
func (n *Notifier) PartidaAcabou(ctx context.Context, homeTeam string, homeScore int, awayTeam string, awayScore int, winners []Winner) error {
	jid := n.sender.LinkedGroup()
	if jid == "" {
		return fmt.Errorf("nenhum grupo vinculado")
	}

	linha := homeTeam + " " + strconv.Itoa(homeScore) + " x " + strconv.Itoa(awayScore) + " " + awayTeam
	msg := fmt.Sprintf("⚽ *FIM DE JOGO*\n%s\n\n", linha)

	if len(winners) == 0 {
		msg += "_Ninguém acertou o resultado._"
	} else {
		msg += "🏆 *Pontuaram:*\n"
		for _, w := range winners {
			msg += fmt.Sprintf("• %s — %d pts\n", w.Name, w.Pontos)
		}
	}

	return n.sender.SendText(ctx, jid, msg)
}

// FaltamDezMinutos envia lembrete antes da partida.
func (n *Notifier) FaltamDezMinutos(ctx context.Context, homeTeam, awayTeam string) error {
	jid := n.sender.LinkedGroup()
	if jid == "" {
		return fmt.Errorf("nenhum grupo vinculado")
	}

	msg := fmt.Sprintf(
		"⏰ *Faltam 10 minutos!*\n%s x %s\nFaçam suas apostas antes que a partida comece! 🎯",
		homeTeam, awayTeam,
	)

	return n.sender.SendText(ctx, jid, msg)
}

// PartidaIniciando envia notificação de encerramento de apostas.
func (n *Notifier) PartidaIniciando(ctx context.Context, homeTeam, awayTeam string) error {
	jid := n.sender.LinkedGroup()
	if jid == "" {
		return fmt.Errorf("nenhum grupo vinculado")
	}

	msg := fmt.Sprintf(
		"🚀 *A partida começou!*\n%s x %s\nApostas encerradas. Boa sorte! 🍀",
		homeTeam, awayTeam,
	)

	return n.sender.SendText(ctx, jid, msg)
}

type Winner struct {
	Name   string
	Pontos int
}
