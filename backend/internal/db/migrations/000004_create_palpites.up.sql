CREATE TABLE palpites (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bolao_id       UUID NOT NULL REFERENCES boloes(id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jogo_id        UUID NOT NULL REFERENCES jogos(id) ON DELETE CASCADE,
    home_score     INT NOT NULL,
    away_score     INT NOT NULL,
    pontos         INT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (bolao_id, user_id, jogo_id)
);

CREATE INDEX idx_palpites_bolao_user ON palpites(bolao_id, user_id);
CREATE INDEX idx_palpites_jogo ON palpites(jogo_id);
