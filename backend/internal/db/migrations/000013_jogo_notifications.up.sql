CREATE TABLE jogo_notifications (
    jogo_id           UUID NOT NULL REFERENCES jogos(id),
    notification_type VARCHAR(50) NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (jogo_id, notification_type)
);
