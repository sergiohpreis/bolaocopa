CREATE TABLE jogo_notifications (
    jogo_id           UUID        NOT NULL REFERENCES jogos(id) ON DELETE CASCADE,
    -- notification_type values: 'faltam_dez_minutos', 'partida_iniciando',
    -- 'fim_de_jogo:<bolao_id>' (per-bolao dedup for fim_de_jogo crash recovery).
    -- sent_at records dispatch attempt, not confirmed delivery (best-effort at-most-once).
    notification_type VARCHAR(100) NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (jogo_id, notification_type)
);
