ALTER TABLE boloes ADD COLUMN taxa_entrada NUMERIC(10,2);

CREATE TABLE taxa_entrada_propostas (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bolao_id     UUID NOT NULL REFERENCES boloes(id) ON DELETE CASCADE,
    valor        NUMERIC(10,2) NOT NULL,
    proposta_por UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (bolao_id)
);

CREATE TABLE taxa_entrada_votos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposta_id UUID NOT NULL REFERENCES taxa_entrada_propostas(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    aprovado    BOOLEAN NOT NULL,
    voted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (proposta_id, user_id)
);

ALTER TYPE feed_tipo ADD VALUE 'taxa_aprovada';
