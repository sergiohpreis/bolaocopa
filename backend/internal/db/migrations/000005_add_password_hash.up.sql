ALTER TABLE users
    ALTER COLUMN google_id DROP NOT NULL,
    ADD COLUMN password_hash TEXT;
