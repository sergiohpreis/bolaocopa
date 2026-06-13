-- WARNING: Rolling back this migration permanently drops the status column.
-- Any palpites with status='pendente' or 'rejeitado' will lose their state
-- and become indistinguishable from approved ones, affecting rankings.
-- Ensure no pending retroactive palpites exist before rolling back.
DROP INDEX IF EXISTS idx_palpites_bolao_pendente;
ALTER TABLE palpites DROP COLUMN status;
