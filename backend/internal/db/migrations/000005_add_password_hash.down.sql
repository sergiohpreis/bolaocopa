ALTER TABLE users
    ALTER COLUMN google_id SET NOT NULL,
    DROP COLUMN password_hash;
