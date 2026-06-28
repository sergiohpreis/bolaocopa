-- WARNING: This down migration is LOSSY.
-- Fractional knockout scores (e.g. 4.5, 7.5, 10.5) are rounded to INT by Postgres.
-- Rolling back after any knockout match has been scored will corrupt the ranking.
-- Do NOT run this in production after the first knockout game is finished.
ALTER TABLE palpites ALTER COLUMN pontos TYPE INT USING pontos::INT;
ALTER TABLE jogos DROP COLUMN winner;
