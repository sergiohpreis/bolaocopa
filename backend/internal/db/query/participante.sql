-- name: JoinBolao :one
INSERT INTO participantes (bolao_id, user_id)
VALUES ($1, $2)
ON CONFLICT (bolao_id, user_id) DO UPDATE SET joined_at = participantes.joined_at
RETURNING *;

-- name: ListParticipantesByBolao :many
SELECT p.*, u.name AS user_name, u.avatar_url AS user_avatar
FROM participantes p
JOIN users u ON u.id = p.user_id
WHERE p.bolao_id = $1
ORDER BY p.joined_at ASC;

-- name: IsParticipante :one
SELECT EXISTS(
    SELECT 1 FROM participantes WHERE bolao_id = $1 AND user_id = $2
) AS is_participante;
