CREATE TYPE feed_tipo AS ENUM (
    'palpite_registrado',
    'palpite_alterado',
    'participante_entrou',
    'jogo_iniciado',
    'resultado_apurado'
);

CREATE TABLE feed_eventos (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bolao_id   UUID NOT NULL REFERENCES boloes(id) ON DELETE CASCADE,
    tipo       feed_tipo NOT NULL,
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    jogo_id    UUID REFERENCES jogos(id) ON DELETE SET NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feed_eventos_bolao_created ON feed_eventos(bolao_id, created_at DESC);
