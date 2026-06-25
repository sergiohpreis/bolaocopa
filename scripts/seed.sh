#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Erro: .env não encontrado em $ENV_FILE"
  exit 1
fi

source "$ENV_FILE"

# ============================================================
# TRAVA DE SEGURANÇA — este script insere dados FICTÍCIOS e apaga
# seeds anteriores. Roda apenas onde a flag SEED_ALLOW=1 estiver
# presente no .env. O .env de PRODUÇÃO nunca deve conter essa flag,
# então prod fica protegido por padrão (opt-in explícito, não adivinha
# o ambiente). O nome do container Postgres é o mesmo em dev e prod,
# por isso o critério é a flag — não o alvo do `docker exec`.
# ============================================================
DB_CONTAINER="bolaocopa-postgres"
FORCE=0
[[ "${1:-}" == "--force" ]] && FORCE=1

if [[ "${SEED_ALLOW:-}" != "1" ]]; then
  echo "ABORTADO: SEED_ALLOW=1 não está no .env." >&2
  echo "Este seed insere dados FICTÍCIOS e nunca deve rodar em produção." >&2
  echo "Para semear um ambiente de desenvolvimento, adicione 'SEED_ALLOW=1' ao .env." >&2
  exit 1
fi

# Confirmação interativa — apaga e reinsere dados de seed.
# Sem TTY ou resposta diferente de 'sim' cancela limpo (exit 0).
if [[ "$FORCE" != "1" ]]; then
  ans=""
  read -r -p "Rodar seed (dados fictícios) no banco '$POSTGRES_DB' (container '$DB_CONTAINER')? Digite 'sim' para continuar: " ans || true
  if [[ "$ans" != "sim" ]]; then
    echo "Cancelado."
    exit 0
  fi
fi

echo "Rodando seed no banco '$POSTGRES_DB'..."

docker exec -i "$DB_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'

-- ============================================================
-- SEED — dados fictícios para screenshots locais
-- Idempotente: usa ON CONFLICT DO NOTHING onde possível
-- ============================================================

-- Limpa dados anteriores de seed (ordem respeitando FK)
DELETE FROM feed_eventos   WHERE bolao_id IN (SELECT id FROM boloes WHERE name LIKE '[SEED]%');
DELETE FROM taxa_entrada_votos   WHERE proposta_id IN (SELECT id FROM taxa_entrada_propostas WHERE bolao_id IN (SELECT id FROM boloes WHERE name LIKE '[SEED]%'));
DELETE FROM taxa_entrada_propostas WHERE bolao_id IN (SELECT id FROM boloes WHERE name LIKE '[SEED]%');
DELETE FROM palpites       WHERE bolao_id IN (SELECT id FROM boloes WHERE name LIKE '[SEED]%');
DELETE FROM participantes  WHERE bolao_id IN (SELECT id FROM boloes WHERE name LIKE '[SEED]%');
DELETE FROM boloes         WHERE name LIKE '[SEED]%';
DELETE FROM jogos          WHERE external_id LIKE 'seed-%';
DELETE FROM users          WHERE email LIKE '%@seed.bolao';

-- ============================================================
-- USUÁRIOS (senha: "senha123" → bcrypt hash pré-computado)
-- ============================================================
INSERT INTO users (id, google_id, email, name, avatar_url, password_hash) VALUES
  ('a0000000-0000-0000-0000-000000000001', NULL, 'sergio@seed.bolao',   'Sergio',   'https://api.dicebear.com/7.x/avataaars/svg?seed=sergio',   '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq'),
  ('a0000000-0000-0000-0000-000000000002', NULL, 'ana@seed.bolao',      'Ana',      'https://api.dicebear.com/7.x/avataaars/svg?seed=ana',      '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq'),
  ('a0000000-0000-0000-0000-000000000003', NULL, 'carlos@seed.bolao',   'Carlos',   'https://api.dicebear.com/7.x/avataaars/svg?seed=carlos',   '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq'),
  ('a0000000-0000-0000-0000-000000000004', NULL, 'julia@seed.bolao',    'Julia',    'https://api.dicebear.com/7.x/avataaars/svg?seed=julia',    '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq'),
  ('a0000000-0000-0000-0000-000000000005', NULL, 'marcos@seed.bolao',   'Marcos',   'https://api.dicebear.com/7.x/avataaars/svg?seed=marcos',   '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq'),
  ('a0000000-0000-0000-0000-000000000006', NULL, 'fernanda@seed.bolao', 'Fernanda', 'https://api.dicebear.com/7.x/avataaars/svg?seed=fernanda', '$2b$10$l1JJiAs75TEA4F87PfKShu/Ys4dJKz7mybWbjAp/ZyK9VXfcOeloq')
ON CONFLICT (email) DO NOTHING;

-- ============================================================
-- BOLÃO
-- ============================================================
INSERT INTO boloes (id, name, admin_id, invite_token, taxa_entrada, retroativo_enabled, wa_notifications_enabled)
VALUES (
  'b0000000-0000-0000-0000-000000000001',
  '[SEED] Copa dos Amigos 2026',
  'a0000000-0000-0000-0000-000000000001',
  'seed-invite-token-abc123',
  25.00,
  true,
  false
);

-- ============================================================
-- PARTICIPANTES (admin já incluso)
-- ============================================================
INSERT INTO participantes (bolao_id, user_id, joined_at) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', NOW() - INTERVAL '10 days'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', NOW() - INTERVAL '9 days'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', NOW() - INTERVAL '8 days'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004', NOW() - INTERVAL '7 days'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000005', NOW() - INTERVAL '6 days'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000006', NOW() - INTERVAL '5 days')
ON CONFLICT (bolao_id, user_id) DO NOTHING;

-- ============================================================
-- JOGOS
-- stage usa os valores crus da API football-data.org (GROUP_STAGE, LAST_32, LAST_16, ...).
-- Inclui jogos de várias fases de mata-mata como fixture para a Visão de Chaveamento:
-- 2 grupos encerrados, 1 grupo ao vivo, 2 grupos próximos, 1 oitava encerrada e 1 próxima.
-- Quartas/Semi/Final ainda não existem (a chave mostra "a definir" nessas fases).
-- ============================================================
INSERT INTO jogos (id, external_id, home_team, away_team, home_team_flag, away_team_flag, starts_at, stage, home_score, away_score, finished) VALUES
  -- Fase de grupos — encerrados
  ('c0000000-0000-0000-0000-000000000001', 'seed-001', 'Brasil',    'Argentina', 'https://flagcdn.com/br.svg', 'https://flagcdn.com/ar.svg', NOW() - INTERVAL '5 days', 'GROUP_STAGE', 2, 1, true),
  ('c0000000-0000-0000-0000-000000000002', 'seed-002', 'França',    'Alemanha',  'https://flagcdn.com/fr.svg', 'https://flagcdn.com/de.svg', NOW() - INTERVAL '3 days', 'GROUP_STAGE', 1, 1, true),
  -- Fase de grupos — ao vivo
  ('c0000000-0000-0000-0000-000000000003', 'seed-003', 'Espanha',   'Portugal',  'https://flagcdn.com/es.svg', 'https://flagcdn.com/pt.svg', NOW() - INTERVAL '1 hour',  'GROUP_STAGE', NULL, NULL, false),
  -- Fase de grupos — próximos
  ('c0000000-0000-0000-0000-000000000004', 'seed-004', 'Inglaterra','Holanda',   'https://flagcdn.com/gb-eng.svg', 'https://flagcdn.com/nl.svg', NOW() + INTERVAL '2 days', 'GROUP_STAGE', NULL, NULL, false),
  ('c0000000-0000-0000-0000-000000000005', 'seed-005', 'Itália',    'Croácia',   'https://flagcdn.com/it.svg', 'https://flagcdn.com/hr.svg', NOW() + INTERVAL '4 days', 'GROUP_STAGE', NULL, NULL, false),
  -- Mata-mata — 16-avos (1 encerrado, 1 próximo). LAST_32 é a primeira fase
  -- eliminatória da Copa 2026 (48 seleções); LAST_16 vem depois. Manter a ordem
  -- causal: não há confronto definido numa fase sem a anterior ter ocorrido.
  ('c0000000-0000-0000-0000-000000000006', 'seed-006', 'Uruguai',   'México',    'https://flagcdn.com/uy.svg', 'https://flagcdn.com/mx.svg', NOW() - INTERVAL '2 days', 'LAST_32', 0, 3, true),
  ('c0000000-0000-0000-0000-000000000007', 'seed-007', 'Bélgica',   'Croácia',   'https://flagcdn.com/be.svg', 'https://flagcdn.com/hr.svg', NOW() + INTERVAL '7 days', 'LAST_32', NULL, NULL, false)
ON CONFLICT (external_id) DO NOTHING;

-- ============================================================
-- PALPITES — jogos encerrados (com pontos calculados)
-- Jogo 1: Brasil 2x1 Argentina (real)
-- ============================================================
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score, pontos, status) VALUES
  -- Sergio: acertou placar exato → 10pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 2, 1, 10, 'aprovado'),
  -- Ana: acertou vencedor → 3pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000001', 3, 0, 3, 'aprovado'),
  -- Carlos: errou → 0pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000001', 0, 2, 0, 'aprovado'),
  -- Julia: acertou vencedor → 3pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004', 'c0000000-0000-0000-0000-000000000001', 1, 0, 3, 'aprovado'),
  -- Marcos: errou → 0pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000001', 2, 2, 0, 'aprovado'),
  -- Fernanda: acertou placar exato → 10pts
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000006', 'c0000000-0000-0000-0000-000000000001', 2, 1, 10, 'aprovado')
ON CONFLICT (bolao_id, user_id, jogo_id) DO NOTHING;

-- Jogo 2: França 1x1 Alemanha (real)
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score, pontos, status) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 1, 1, 10, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 2, 0,  0, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000002', 0, 0,  3, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004', 'c0000000-0000-0000-0000-000000000002', 1, 2,  0, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000002', 2, 1,  0, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000006', 'c0000000-0000-0000-0000-000000000002', 1, 1, 10, 'aprovado')
ON CONFLICT (bolao_id, user_id, jogo_id) DO NOTHING;

-- Jogo ao vivo (Espanha x Portugal) — palpites registrados, sem pontos ainda
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score, pontos, status) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000003', 2, 0, NULL, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000003', 1, 1, NULL, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 0, 1, NULL, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004', 'c0000000-0000-0000-0000-000000000003', 2, 1, NULL, 'aprovado'),
  -- Marcos: palpite retroativo pendente
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000003', 1, 0, NULL, 'pendente')
ON CONFLICT (bolao_id, user_id, jogo_id) DO NOTHING;

-- Próximos — alguns participantes já chutaram
INSERT INTO palpites (bolao_id, user_id, jogo_id, home_score, away_score, pontos, status) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000004', 2, 1, NULL, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000004', 1, 0, NULL, 'aprovado'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000005', 3, 1, NULL, 'aprovado')
ON CONFLICT (bolao_id, user_id, jogo_id) DO NOTHING;

-- ============================================================
-- FEED EVENTOS
-- ============================================================
INSERT INTO feed_eventos (bolao_id, tipo, user_id, jogo_id, payload, created_at) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'participante_entrou', 'a0000000-0000-0000-0000-000000000002', NULL, '{}', NOW() - INTERVAL '9 days'),
  ('b0000000-0000-0000-0000-000000000001', 'participante_entrou', 'a0000000-0000-0000-0000-000000000003', NULL, '{}', NOW() - INTERVAL '8 days'),
  ('b0000000-0000-0000-0000-000000000001', 'participante_entrou', 'a0000000-0000-0000-0000-000000000004', NULL, '{}', NOW() - INTERVAL '7 days'),
  ('b0000000-0000-0000-0000-000000000001', 'participante_entrou', 'a0000000-0000-0000-0000-000000000005', NULL, '{}', NOW() - INTERVAL '6 days'),
  ('b0000000-0000-0000-0000-000000000001', 'participante_entrou', 'a0000000-0000-0000-0000-000000000006', NULL, '{}', NOW() - INTERVAL '5 days'),
  ('b0000000-0000-0000-0000-000000000001', 'taxa_aprovada',       'a0000000-0000-0000-0000-000000000001', NULL, '{"valor": 25.00}', NOW() - INTERVAL '4 days'),
  ('b0000000-0000-0000-0000-000000000001', 'jogo_iniciado',       NULL, 'c0000000-0000-0000-0000-000000000001', '{}', NOW() - INTERVAL '5 days'),
  ('b0000000-0000-0000-0000-000000000001', 'resultado_apurado',   NULL, 'c0000000-0000-0000-0000-000000000001', '{"home_score": 2, "away_score": 1}', NOW() - INTERVAL '4 days 22 hours'),
  ('b0000000-0000-0000-0000-000000000001', 'jogo_iniciado',       NULL, 'c0000000-0000-0000-0000-000000000002', '{}', NOW() - INTERVAL '3 days'),
  ('b0000000-0000-0000-0000-000000000001', 'resultado_apurado',   NULL, 'c0000000-0000-0000-0000-000000000002', '{"home_score": 1, "away_score": 1}', NOW() - INTERVAL '2 days 22 hours'),
  ('b0000000-0000-0000-0000-000000000001', 'jogo_iniciado',       NULL, 'c0000000-0000-0000-0000-000000000003', '{}', NOW() - INTERVAL '1 hour');

-- ============================================================
-- TAXA DE ENTRADA — já aprovada (valor fixado no bolao)
-- ============================================================
-- A taxa já está fixada via boloes.taxa_entrada = 25.00
-- Proposta e votos são histórico; taxa_aprovada foi o evento de consolidação

SELECT 'Seed concluído com sucesso!' AS status;
SELECT
  u.name,
  COALESCE(SUM(p.pontos), 0) AS total_pontos
FROM participantes pt
JOIN users u ON u.id = pt.user_id
LEFT JOIN palpites p ON p.user_id = pt.user_id AND p.bolao_id = pt.bolao_id AND p.pontos IS NOT NULL
WHERE pt.bolao_id = 'b0000000-0000-0000-0000-000000000001'
GROUP BY u.name
ORDER BY total_pontos DESC;

SQL

echo "Seed finalizado."
