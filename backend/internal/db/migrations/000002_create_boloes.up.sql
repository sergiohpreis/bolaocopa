CREATE TABLE boloes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    admin_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_token VARCHAR(64) NOT NULL UNIQUE DEFAULT encode(gen_random_bytes(32), 'hex'),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE participantes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bolao_id   UUID NOT NULL REFERENCES boloes(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (bolao_id, user_id)
);
