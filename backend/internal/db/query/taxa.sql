-- name: ProporTaxa :one
INSERT INTO taxa_entrada_propostas (bolao_id, valor, proposta_por)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPropostaAtiva :one
SELECT * FROM taxa_entrada_propostas WHERE bolao_id = $1;

-- name: GetPropostaAtivaForUpdate :one
SELECT * FROM taxa_entrada_propostas WHERE bolao_id = $1 FOR UPDATE;

-- name: IsParticipanteElegivel :one
-- Returns true if the user was a participant at or before proposta.created_at.
SELECT EXISTS (
    SELECT 1 FROM participantes
    WHERE bolao_id = $1 AND user_id = $2 AND joined_at <= $3
);

-- name: RegistrarVoto :one
-- ON CONFLICT DO NOTHING: when user already voted, RETURNING yields no rows.
-- pgx returns pgx.ErrNoRows, which the service maps to ErrJaVotou.
-- Do NOT change to DO UPDATE without updating the service layer.
INSERT INTO taxa_entrada_votos (proposta_id, user_id, aprovado)
VALUES ($1, $2, $3)
ON CONFLICT (proposta_id, user_id) DO NOTHING
RETURNING *;

-- name: CountVotosFavoraveis :one
-- Counts favorable votes only from participants eligible at proposal time (joined_at <= proposta.created_at).
-- This ensures the same population is used for both total and favorable counts.
SELECT COUNT(*) FROM taxa_entrada_votos tv
JOIN participantes p ON p.user_id = tv.user_id AND p.bolao_id = $2
WHERE tv.proposta_id = $1 AND tv.aprovado = true AND p.joined_at <= $3;

-- name: CountParticipantesNaMomento :one
SELECT COUNT(*) FROM participantes
WHERE bolao_id = $1 AND joined_at <= $2;

-- name: CancelarProposta :exec
DELETE FROM taxa_entrada_propostas WHERE bolao_id = $1;

-- name: DefinirTaxa :one
UPDATE boloes
SET taxa_entrada = $2, updated_at = NOW()
WHERE id = $1 AND taxa_entrada IS NULL
RETURNING *;

-- name: GetTaxaEntrada :one
SELECT taxa_entrada FROM boloes WHERE id = $1;
