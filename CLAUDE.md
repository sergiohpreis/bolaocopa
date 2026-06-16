# Bolaocopa — Claude Instructions

## Stack

- **Backend**: Go, Chi router, sqlc + pgx/v5, pgxpool, JWT auth
- **Frontend**: Vue 3 + TypeScript, Pinia, Axios, PrimeVue, Tailwind v4
- **DB**: PostgreSQL 16 via Docker Compose
- **Local**: `docker compose up` — proxy at `http://localhost:8001`

## Git Workflow

- Nova branch por feature: `feat/`, `fix/`, `refactor/` etc.
- Commits: inglês, conventional commits, linha única, sem body
- Antes de merge: rodar `/review-specialists`
- Depois de implementar: `/commit-push-pr`

## Backend Conventions

- Queries geradas via `sqlc generate` — nunca editar `repository/*.sql.go` manualmente
- Transações com `pgxpool.Pool` passado direto ao service (não via Querier)
- Erros de domínio como `var Err... = errors.New(...)` no pacote service
- Handler mapeia erros de domínio para HTTP codes explicitamente

## Frontend Conventions

- Design system: pitch verde escuro `#0a1a0e` + neon `#39ff6a` + Bebas Neue + DM Sans
- CSS variables em `src/assets/main.css`
- Tipos globais em `src/types/index.ts`
- API functions em `src/api/bolao.ts` (ou arquivo específico por domínio)

## Feature Announcements

Novas features são anunciadas via **`src/components/WhatsNewModal.vue`**.

- Versão controlada via `STORAGE_KEY = 'whats_new_seen_vX.Y.Z'`
- Bump a versão no badge, no storage key, e atualizar o conteúdo do modal
- Versão atual: **v1.2.0** (taxa de entrada)
- Toda feature nova deve atualizar o modal para **v1.3.0**, etc.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues (`sergiohpreis/bolaocopa`). See `docs/agents/issue-tracker.md`.

### Triage labels

Default label vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout — one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
