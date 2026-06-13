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
WHERE p.bolao_id = $1 AND p.jogo_id = $2 AND p.status = 'aprovado';

-- name: ListPalpitesByJogo :many
SELECT * FROM palpites WHERE jogo_id = $1 AND status = 'aprovado';

-- name: UpdatePalpitePontos :exec
UPDATE palpites SET pontos = $1, updated_at = NOW()
WHERE bolao_id = $2 AND jogo_id = $3 AND user_id = $4;

-- name: UpsertPalpiteRetroativo :one
-- When the conflict row has status='aprovado', the WHERE clause causes Postgres to skip
-- the DO UPDATE, and RETURNING emits 0 rows. pgx surfaces this as pgx.ErrNoRows,
-- which the service maps to ErrPalpiteJaAprovado.
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score, status)
VALUES ($1, $2, $3, $4, $5, 'pendente')
ON CONFLICT (bolao_id, user_id, jogo_id) DO UPDATE
    SET home_score = EXCLUDED.home_score,
        away_score = EXCLUDED.away_score,
        status = 'pendente',
        updated_at = NOW()
WHERE palpites.status != 'aprovado'
RETURNING *;

-- name: ListPalpitesPendentes :many
SELECT p.*, u.name AS user_name, j.home_team, j.away_team, j.starts_at, j.finished, j.home_score AS jogo_home_score, j.away_score AS jogo_away_score
FROM palpites p
JOIN users u ON u.id = p.user_id
JOIN jogos j ON j.id = p.jogo_id
WHERE p.bolao_id = $1 AND p.status = 'pendente'
ORDER BY p.created_at ASC;

-- name: AtualizarStatusPalpite :one
UPDATE palpites
SET status = $3, updated_at = NOW()
WHERE id = $1 AND bolao_id = $2 AND status = 'pendente'
RETURNING *;

-- name: GetRanking :many
SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.avatar_url,
    COALESCE(SUM(p.pontos), 0) AS total_pontos,
    COUNT(p.id) FILTER (WHERE p.pontos IS NOT NULL) AS palpites_computados
FROM participantes pt
JOIN users u ON u.id = pt.user_id
LEFT JOIN palpites p ON p.user_id = pt.user_id AND p.bolao_id = pt.bolao_id AND p.status = 'aprovado'
WHERE pt.bolao_id = $1
GROUP BY u.id, u.name, u.avatar_url
ORDER BY total_pontos DESC, u.name ASC;

-- name: ListPalpitesRetroativosAprovados :many
SELECT p.*, u.name AS user_name, j.home_team, j.away_team, j.starts_at, j.finished, j.home_score AS jogo_home_score, j.away_score AS jogo_away_score
FROM palpites p
JOIN users u ON u.id = p.user_id
JOIN jogos j ON j.id = p.jogo_id
WHERE p.bolao_id = $1 AND p.status = 'aprovado' AND j.starts_at <= NOW()
ORDER BY p.updated_at DESC;

-- name: GetPalpiteByID :one
SELECT * FROM palpites WHERE id = $1 AND bolao_id = $2;

-- name: DeletePalpite :exec
DELETE FROM palpites WHERE id = $1 AND bolao_id = $2;
