ALTER TABLE palpites
  ADD COLUMN status TEXT NOT NULL DEFAULT 'aprovado'
    CHECK (status IN ('aprovado', 'pendente', 'rejeitado'));
