-- name: UpsertPalpite :one
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (bolao_id, user_id, jogo_id) DO UPDATE
    SET home_score = EXCLUDED.home_score,
        away_score = EXCLUDED.away_score,
        updated_at = NOW()
RETURNING *;

-- name: ListPalpitesByBolaoAndUser :many
SELECT * FROM palpites WHERE bolao_id = $1 AND user_id = $2 ORDER BY created_at ASC;

-- name: ListPalpitesByBolaoAndJogo :many
SELECT p.*, u.name AS user_name, u.avatar_url AS user_avatar
FROM palpites p
JOIN users u ON u.id = p.user_id
WHERE p.bolao_id = $1 AND p.jogo_id = $2;

-- name: UpdatePalpitePontos :exec
UPDATE palpites SET pontos = $1, updated_at = NOW()
WHERE bolao_id = $2 AND jogo_id = $3 AND user_id = $4;

-- name: GetRanking :many
SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.avatar_url,
    COALESCE(SUM(p.pontos), 0) AS total_pontos,
    COUNT(p.id) FILTER (WHERE p.pontos IS NOT NULL) AS palpites_computados
FROM participantes pt
JOIN users u ON u.id = pt.user_id
LEFT JOIN palpites p ON p.user_id = pt.user_id AND p.bolao_id = pt.bolao_id
WHERE pt.bolao_id = $1
GROUP BY u.id, u.name, u.avatar_url
ORDER BY total_pontos DESC, u.name ASC;
