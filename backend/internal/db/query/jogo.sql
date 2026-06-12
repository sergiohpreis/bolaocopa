-- name: UpsertJogo :one
INSERT INTO jogos (external_id, home_team, away_team, home_team_flag, away_team_flag, starts_at, stage, home_score, away_score, finished)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
        updated_at     = NOW()
RETURNING *;

-- name: ListJogos :many
SELECT * FROM jogos ORDER BY starts_at ASC;

-- name: GetJogoByID :one
SELECT * FROM jogos WHERE id = $1;

-- name: ListFinishedJobsWithoutScores :many
SELECT * FROM jogos WHERE finished = TRUE AND home_score IS NOT NULL;
