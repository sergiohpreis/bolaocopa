-- name: CreateBolao :one
INSERT INTO boloes (name, admin_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetBolaoByID :one
SELECT * FROM boloes WHERE id = $1;

-- name: GetBolaoByInviteToken :one
SELECT * FROM boloes WHERE invite_token = $1;

-- name: ListBoloesByUser :many
SELECT b.* FROM boloes b
JOIN participantes p ON p.bolao_id = b.id
WHERE p.user_id = $1
ORDER BY b.created_at DESC;

-- name: RegenerateInviteToken :one
UPDATE boloes
SET invite_token = encode(gen_random_bytes(32), 'hex'), updated_at = NOW()
WHERE id = $1 AND admin_id = $2
RETURNING *;

-- name: SetRetroativoEnabled :one
UPDATE boloes
SET retroativo_enabled = $3, updated_at = NOW()
WHERE id = $1 AND admin_id = $2
RETURNING *;

-- name: DeleteBolao :exec
DELETE FROM boloes WHERE id = $1 AND admin_id = $2;
