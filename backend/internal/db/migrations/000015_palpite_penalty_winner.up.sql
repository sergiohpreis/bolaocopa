ALTER TABLE palpites ADD COLUMN penalty_winner TEXT CHECK (penalty_winner IN ('home', 'away'));
