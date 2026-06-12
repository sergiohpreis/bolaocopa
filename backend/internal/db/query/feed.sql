-- name: InsertFeedEvento :one
INSERT INTO feed_eventos (bolao_id, tipo, user_id, jogo_id, payload)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListFeedByBolao :many
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
LIMIT 50;
