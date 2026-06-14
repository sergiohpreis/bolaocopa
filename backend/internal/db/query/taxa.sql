-- name: ProporTaxa :one
INSERT INTO taxa_entrada_propostas (bolao_id, valor, proposta_por)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPropostaAtiva :one
SELECT * FROM taxa_entrada_propostas WHERE bolao_id = $1;

-- name: RegistrarVoto :one
INSERT INTO taxa_entrada_votos (proposta_id, user_id, aprovado)
VALUES ($1, $2, $3)
ON CONFLICT (proposta_id, user_id) DO NOTHING
RETURNING *;

-- name: CountVotosFavoraveis :one
SELECT COUNT(*) FROM taxa_entrada_votos
WHERE proposta_id = $1 AND aprovado = true;

-- name: CountParticipantesNaMomento :one
SELECT COUNT(*) FROM participantes
WHERE bolao_id = $1 AND joined_at <= $2;

-- name: CancelarProposta :exec
DELETE FROM taxa_entrada_propostas WHERE bolao_id = $1;

-- name: DefinirTaxa :one
UPDATE boloes
SET taxa_entrada = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetTaxaEntrada :one
SELECT taxa_entrada FROM boloes WHERE id = $1;
