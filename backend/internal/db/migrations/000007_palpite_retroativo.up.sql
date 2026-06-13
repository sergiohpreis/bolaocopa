ALTER TABLE palpites
  ADD COLUMN status TEXT NOT NULL DEFAULT 'aprovado'
    CHECK (status IN ('aprovado', 'pendente', 'rejeitado'));

-- Filtered index for admin approval queue (pendente is a transient minority state).
CREATE INDEX idx_palpites_bolao_pendente ON palpites(bolao_id, created_at) WHERE status = 'pendente';
