-- name: UpsertJogo :one
WITH before AS (
    SELECT finished, home_score, away_score FROM jogos WHERE external_id = $1
)
INSERT INTO jogos (external_id, home_team, away_team, home_team_flag, away_team_flag, starts_at, stage, home_score, away_score, finished, winner)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (external_id) DO UPDATE
    SET home_team      = EXCLUDED.home_team,
        away_team      = EXCLUDED.away_team,
        home_team_flag = EXCLUDED.home_team_flag,
        away_team_flag = EXCLUDED.away_team_flag,
        starts_at      = EXCLUDED.starts_at,
        stage          = EXCLUDED.stage,
        home_score     = EXCLUDED.home_score,
        away_score     = EXCLUDED.away_score,
        finished       = EXCLUDED.finished,
        winner         = COALESCE(EXCLUDED.winner, jogos.winner),
        updated_at     = NOW()
RETURNING *,
    (SELECT COALESCE(finished, FALSE) FROM before) AS was_finished,
    (SELECT home_score FROM before) AS was_home_score,
    (SELECT away_score FROM before) AS was_away_score;

-- name: ListJogos :many
SELECT * FROM jogos ORDER BY starts_at ASC;

-- name: GetJogoByID :one
SELECT * FROM jogos WHERE id = $1;

-- name: ListFinishedJogosWithScores :many
SELECT * FROM jogos WHERE finished = TRUE AND home_score IS NOT NULL;

-- name: InsertJogoNotificationIfAbsent :execrows
INSERT INTO jogo_notifications (jogo_id, notification_type)
VALUES ($1, $2)
ON CONFLICT (jogo_id, notification_type) DO NOTHING;
