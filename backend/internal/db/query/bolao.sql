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

-- name: SetBolaoWAGroup :one
UPDATE boloes SET wa_group_jid = $2, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: GetBolaoWAGroup :one
SELECT wa_group_jid FROM boloes WHERE id = $1;

-- name: ListBoloesByJogo :many
SELECT DISTINCT b.id, b.name, b.admin_id, b.invite_token, b.created_at, b.updated_at, b.retroativo_enabled, b.taxa_entrada, b.wa_group_jid, b.wa_notifications_enabled
FROM boloes b
JOIN palpites p ON p.bolao_id = b.id
WHERE p.jogo_id = $1
  AND b.wa_group_jid IS NOT NULL
  AND b.wa_group_jid != ''
  AND b.wa_notifications_enabled = TRUE
ORDER BY b.created_at;

-- name: ListBoloesByWAGroup :many
SELECT id, name, admin_id, invite_token, created_at, updated_at, retroativo_enabled, taxa_entrada, wa_group_jid, wa_notifications_enabled
FROM boloes
WHERE wa_group_jid IS NOT NULL
  AND wa_group_jid != ''
  AND wa_notifications_enabled = TRUE
ORDER BY created_at;

-- name: SetBolaoWANotificationsEnabled :one
UPDATE boloes SET wa_notifications_enabled = $2, updated_at = NOW() WHERE id = $1 RETURNING *;
