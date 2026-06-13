-- name: UpsertUserByGoogleID :one
INSERT INTO users (google_id, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_id) DO UPDATE
    SET email      = EXCLUDED.email,
        name       = EXCLUDED.name,
        avatar_url = EXCLUDED.avatar_url,
        updated_at = NOW()
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUserByEmail :one
INSERT INTO users (email, name, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
