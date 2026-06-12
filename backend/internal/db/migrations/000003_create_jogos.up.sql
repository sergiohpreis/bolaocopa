CREATE TABLE jogos (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id      VARCHAR(64) NOT NULL UNIQUE,
    home_team        VARCHAR(100) NOT NULL,
    away_team        VARCHAR(100) NOT NULL,
    home_team_flag   TEXT,
    away_team_flag   TEXT,
    starts_at        TIMESTAMPTZ NOT NULL,
    stage            VARCHAR(50) NOT NULL,
    home_score       INT,
    away_score       INT,
    finished         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
