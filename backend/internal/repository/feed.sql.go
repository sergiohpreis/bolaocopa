package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
)

type FeedTipo string

const (
	FeedTipoPalpiteRegistrado  FeedTipo = "palpite_registrado"
	FeedTipoPalpiteAlterado    FeedTipo = "palpite_alterado"
	FeedTipoParticipanteEntrou FeedTipo = "participante_entrou"
	FeedTipoJogoIniciado       FeedTipo = "jogo_iniciado"
	FeedTipoResultadoApurado   FeedTipo = "resultado_apurado"
)

type InsertFeedEventoParams struct {
	BolaoID pgtype.UUID `json:"bolao_id"`
	Tipo    FeedTipo    `json:"tipo"`
	UserID  pgtype.UUID `json:"user_id"`
	JogoID  pgtype.UUID `json:"jogo_id"`
	Payload []byte      `json:"payload"`
}

type FeedEvento struct {
	ID        pgtype.UUID        `json:"id"`
	BolaoID   pgtype.UUID        `json:"bolao_id"`
	Tipo      FeedTipo           `json:"tipo"`
	UserID    pgtype.UUID        `json:"user_id"`
	JogoID    pgtype.UUID        `json:"jogo_id"`
	Payload   []byte             `json:"payload"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type ListFeedByBolaoRow struct {
	ID            pgtype.UUID        `json:"id"`
	BolaoID       pgtype.UUID        `json:"bolao_id"`
	Tipo          FeedTipo           `json:"tipo"`
	UserID        pgtype.UUID        `json:"user_id"`
	JogoID        pgtype.UUID        `json:"jogo_id"`
	Payload       []byte             `json:"payload"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UserName      pgtype.Text        `json:"user_name"`
	JogoHomeTeam  pgtype.Text        `json:"jogo_home_team"`
	JogoAwayTeam  pgtype.Text        `json:"jogo_away_team"`
	JogoStartsAt  pgtype.Timestamptz `json:"jogo_starts_at"`
}

const insertFeedEvento = `
INSERT INTO feed_eventos (bolao_id, tipo, user_id, jogo_id, payload)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, bolao_id, tipo, user_id, jogo_id, payload, created_at
`

func (q *Queries) InsertFeedEvento(ctx context.Context, arg InsertFeedEventoParams) (FeedEvento, error) {
	payload := arg.Payload
	if payload == nil {
		payload = []byte("{}")
	}
	row := q.db.QueryRow(ctx, insertFeedEvento,
		arg.BolaoID,
		arg.Tipo,
		arg.UserID,
		arg.JogoID,
		payload,
	)
	var i FeedEvento
	err := row.Scan(
		&i.ID,
		&i.BolaoID,
		&i.Tipo,
		&i.UserID,
		&i.JogoID,
		&i.Payload,
		&i.CreatedAt,
	)
	return i, err
}

const listFeedByBolao = `
SELECT
    fe.id,
    fe.bolao_id,
    fe.tipo,
    fe.user_id,
    fe.jogo_id,
    fe.payload,
    fe.created_at,
    u.name AS user_name,
    j.home_team AS jogo_home_team,
    j.away_team AS jogo_away_team,
    j.starts_at AS jogo_starts_at
FROM feed_eventos fe
LEFT JOIN users u ON u.id = fe.user_id
LEFT JOIN jogos j ON j.id = fe.jogo_id
WHERE fe.bolao_id = $1
ORDER BY fe.created_at DESC
LIMIT 50
`

func (q *Queries) ListFeedByBolao(ctx context.Context, bolaoID pgtype.UUID) ([]ListFeedByBolaoRow, error) {
	rows, err := q.db.Query(ctx, listFeedByBolao, bolaoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListFeedByBolaoRow
	for rows.Next() {
		var i ListFeedByBolaoRow
		if err := rows.Scan(
			&i.ID,
			&i.BolaoID,
			&i.Tipo,
			&i.UserID,
			&i.JogoID,
			&i.Payload,
			&i.CreatedAt,
			&i.UserName,
			&i.JogoHomeTeam,
			&i.JogoAwayTeam,
			&i.JogoStartsAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// MarshalPayload serializes a map to JSON bytes for use as payload.
func MarshalPayload(v map[string]any) []byte {
	b, _ := json.Marshal(v)
	return b
}
