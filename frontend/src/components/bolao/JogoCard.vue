<template>
  <div class="jogo-card" :class="{ 'is-finished': jogo.finished, 'is-live': !jogo.finished && isClosed }">

    <!-- Teams + score row -->
    <div class="teams-row" :class="{ clickable: canExpand }" @click="canExpand && toggleExpand()">
      <!-- Home team -->
      <div class="team home-team">
        <img v-if="jogo.home_team_flag" :src="jogo.home_team_flag" class="flag" :alt="jogo.home_team" />
        <span class="team-name">{{ traduzTime(jogo.home_team) }}</span>
      </div>

      <!-- Middle: score or time -->
      <div class="score-center">
        <template v-if="jogo.finished">
          <span class="score-num">{{ jogo.home_score }}</span>
          <span class="score-sep">–</span>
          <span class="score-num">{{ jogo.away_score }}</span>
        </template>
        <template v-else-if="isClosed">
          <span class="live-badge">AO VIVO</span>
        </template>
        <template v-else>
          <span class="match-time">{{ formatTime(jogo.starts_at) }}</span>
        </template>
      </div>

      <!-- Away team -->
      <div class="team away-team">
        <span class="team-name">{{ traduzTime(jogo.away_team) }}</span>
        <img v-if="jogo.away_team_flag" :src="jogo.away_team_flag" class="flag" :alt="jogo.away_team" />
      </div>

      <!-- Expand chevron -->
      <div v-if="canExpand" class="chevron" :class="{ open: expanded }">›</div>
    </div>

    <!-- Palpite row -->
    <div class="palpite-row">
      <!-- Active input -->
      <template v-if="!jogo.finished && !isClosed">
        <div class="palpite-input-group">
          <div class="score-input-wrap">
            <button class="score-adj" @click="homeInput = Math.max(0, (homeInput ?? 0) - 1)">−</button>
            <span class="score-display">{{ homeInput }}</span>
            <button class="score-adj" @click="homeInput = (homeInput ?? 0) + 1">+</button>
          </div>
          <span class="score-vs">×</span>
          <div class="score-input-wrap">
            <button class="score-adj" @click="awayInput = Math.max(0, (awayInput ?? 0) - 1)">−</button>
            <span class="score-display">{{ awayInput }}</span>
            <button class="score-adj" @click="awayInput = (awayInput ?? 0) + 1">+</button>
          </div>
          <button
            class="save-btn ready"
            :class="{ saved: isSaved }"
            @click="emit('save', homeInput, awayInput)"
          >
            <span class="font-display" style="font-size: 0.75rem; letter-spacing: 0.1em;">{{ isSaved ? 'SALVO' : 'SALVAR' }}</span>
          </button>
        </div>
      </template>

      <!-- Palpite pendente (aguardando aprovação) -->
      <template v-else-if="palpite && palpite.status === 'pendente'">
        <div class="locked-palpite">
          <span class="locked-label">SEU PALPITE</span>
          <div class="locked-score">
            <span class="locked-num">{{ palpite.home_score }}</span>
            <span class="locked-sep">–</span>
            <span class="locked-num">{{ palpite.away_score }}</span>
          </div>
          <span class="pending-badge font-display">AGUARDANDO</span>
        </div>
      </template>

      <!-- Locked palpite display (aprovado) -->
      <template v-else-if="palpite && palpite.status === 'aprovado'">
        <div class="locked-palpite">
          <span class="locked-label">SEU PALPITE</span>
          <div class="locked-score">
            <span class="locked-num">{{ palpite.home_score }}</span>
            <span class="locked-sep">–</span>
            <span class="locked-num">{{ palpite.away_score }}</span>
          </div>
          <div v-if="palpite.pontos !== undefined && palpite.pontos !== null" class="pontos-badge">
            <span class="font-display">+{{ palpite.pontos }}</span>
            <span style="font-size: 0.6rem; letter-spacing: 0.08em;">PTS</span>
          </div>
        </div>
      </template>

      <!-- Retroativo: jogo fechado, sem palpite aprovado -->
      <template v-else-if="isClosed || jogo.finished">
        <div class="palpite-input-group">
          <div class="score-input-wrap">
            <button class="score-adj" @click="homeInput = Math.max(0, (homeInput ?? 0) - 1)">−</button>
            <span class="score-display">{{ homeInput }}</span>
            <button class="score-adj" @click="homeInput = (homeInput ?? 0) + 1">+</button>
          </div>
          <span class="score-vs">×</span>
          <div class="score-input-wrap">
            <button class="score-adj" @click="awayInput = Math.max(0, (awayInput ?? 0) - 1)">−</button>
            <span class="score-display">{{ awayInput }}</span>
            <button class="score-adj" @click="awayInput = (awayInput ?? 0) + 1">+</button>
          </div>
          <button class="save-btn retro" @click="emit('saveRetroativo', homeInput, awayInput)">
            <span class="font-display" style="font-size: 0.7rem; letter-spacing: 0.08em;">ENVIAR</span>
          </button>
        </div>
      </template>
    </div>

    <!-- Expanded: palpites dos outros -->
    <div v-if="canExpand && expanded" class="outros-palpites">
      <div v-if="loadingOutros" class="outros-loading">
        <div class="loader-ring-xs" />
      </div>
      <div v-else-if="outrosPalpites.length === 0" class="outros-empty">
        Nenhum palpite registrado ainda.
      </div>
      <div v-else class="outros-list">
        <div
          v-for="p in outrosPalpites"
          :key="p.id"
          class="outro-item"
          :class="pontosClass(p.pontos)"
        >
          <div class="outro-avatar">{{ p.user_name?.[0]?.toUpperCase() ?? '?' }}</div>
          <span class="outro-name">{{ p.user_name }}</span>
          <div class="outro-score">
            <span class="outro-num">{{ p.home_score }}</span>
            <span class="outro-sep">–</span>
            <span class="outro-num">{{ p.away_score }}</span>
          </div>
          <div v-if="p.pontos !== undefined && p.pontos !== null" class="outro-pontos" :class="pontosClass(p.pontos)">
            +{{ p.pontos }}
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { getPalpitesByJogo } from '@/api/bolao'
import { traduzTime } from '@/utils/teams'
import type { Jogo, Palpite, PalpiteDeJogo } from '@/types'

const props = defineProps<{ jogo: Jogo; palpite?: Palpite; bolaoId: string }>()
const emit = defineEmits<{
  (e: 'save', home: number, away: number): void
  (e: 'saveRetroativo', home: number, away: number): void
}>()

const homeInput = ref<number>(props.palpite?.home_score ?? 0)
const awayInput = ref<number>(props.palpite?.away_score ?? 0)

const isSaved = computed(
  () =>
    props.palpite != null &&
    homeInput.value === props.palpite.home_score &&
    awayInput.value === props.palpite.away_score
)

watch(() => props.palpite, (p) => {
  homeInput.value = p?.home_score ?? 0
  awayInput.value = p?.away_score ?? 0
}, { deep: true })

const now = ref(new Date())
let clockInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  const gameTime = new Date(props.jogo.starts_at).getTime()
  const remaining = gameTime - Date.now()
  if (remaining > 0 && remaining < 24 * 60 * 60 * 1000) {
    clockInterval = setInterval(() => { now.value = new Date() }, 30_000)
  }
})

onUnmounted(() => {
  if (clockInterval) clearInterval(clockInterval)
})

const isClosed = computed(() => now.value >= new Date(props.jogo.starts_at))
const canExpand = computed(() => isClosed.value || props.jogo.finished)

// Expansion state
const expanded = ref(false)
const loadingOutros = ref(false)
const outrosPalpites = ref<PalpiteDeJogo[]>([])
let fetched = false

async function toggleExpand() {
  expanded.value = !expanded.value
  if (expanded.value && !fetched) {
    loadingOutros.value = true
    try {
      outrosPalpites.value = await getPalpitesByJogo(props.bolaoId, props.jogo.id)
      fetched = true
    } finally {
      loadingOutros.value = false
    }
  }
}

function pontosClass(pontos?: number | null) {
  if (pontos === undefined || pontos === null) return ''
  if (pontos === 10) return 'pts-exact'
  if (pontos === 3) return 'pts-winner'
  return 'pts-miss'
}

function formatTime(iso: string) {
  const d = new Date(iso)
  const day = d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
  const time = d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
  return `${day} · ${time}`
}
</script>

<style scoped>
.jogo-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  overflow: hidden;
  transition: border-color 0.2s;
}
.jogo-card.is-finished {
  border-color: rgba(57,255,106,0.08);
  opacity: 0.9;
}

.teams-row {
  display: flex;
  align-items: center;
  padding: 12px 14px 8px;
  gap: 8px;
}
.teams-row.clickable {
  cursor: pointer;
  transition: background 0.15s;
}
.teams-row.clickable:hover {
  background: rgba(57,255,106,0.04);
}

.chevron {
  font-size: 1.2rem;
  color: var(--text-muted);
  transition: transform 0.2s, color 0.2s;
  flex-shrink: 0;
  line-height: 1;
  margin-left: 2px;
}
.chevron.open {
  transform: rotate(90deg);
  color: var(--neon);
}

.team {
  display: flex;
  align-items: center;
  gap: 7px;
  flex: 1;
  min-width: 0;
}
.away-team {
  justify-content: flex-end;
}
.flag {
  width: 22px;
  height: 22px;
  object-fit: contain;
  border-radius: 2px;
  flex-shrink: 0;
}
.team-name {
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.score-center {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  min-width: 70px;
  justify-content: center;
}
.score-num {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.6rem;
  color: var(--neon);
  line-height: 1;
  min-width: 20px;
  text-align: center;
}
.score-sep {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.2rem;
  color: var(--text-muted);
}
.match-time {
  font-family: 'DM Sans', sans-serif;
  font-size: 0.72rem;
  color: var(--text-muted);
  text-align: center;
  line-height: 1.3;
  white-space: nowrap;
}
.live-badge {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  background: rgba(255,80,80,0.15);
  border: 1px solid rgba(255,80,80,0.3);
  color: #ff8080;
  padding: 2px 7px;
  border-radius: 4px;
  animation: pulse-live 1.5s ease-in-out infinite;
}
@keyframes pulse-live {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Palpite section */
.palpite-row {
  border-top: 1px solid rgba(57,255,106,0.07);
  padding: 10px 14px;
  min-height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.palpite-input-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}
.score-input-wrap {
  display: flex;
  align-items: center;
  gap: 0;
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(57,255,106,0.2);
  border-radius: 8px;
  overflow: hidden;
}
.score-adj {
  width: 28px;
  height: 36px;
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.score-adj:hover {
  color: var(--neon);
  background: rgba(57,255,106,0.08);
}
.score-display {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.4rem;
  color: var(--text-primary);
  min-width: 32px;
  text-align: center;
  line-height: 1;
}
.score-vs {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1rem;
  color: var(--text-muted);
}

.save-btn {
  padding: 7px 12px;
  background: transparent;
  border: 1px solid rgba(57,255,106,0.25);
  border-radius: 7px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}
.save-btn.ready {
  border-color: var(--neon);
  color: var(--neon);
  background: rgba(57,255,106,0.08);
}
.save-btn.ready:hover {
  background: rgba(57,255,106,0.16);
  box-shadow: 0 0 12px rgba(57,255,106,0.2);
}
.save-btn.saved {
  border-color: rgba(57,255,106,0.4);
  color: rgba(57,255,106,0.5);
  background: transparent;
}
.save-btn.saved:hover {
  background: rgba(57,255,106,0.06);
  box-shadow: none;
}
.save-btn.retro {
  border-color: rgba(255,200,60,0.5);
  color: rgba(255,200,60,0.85);
  background: rgba(255,200,60,0.06);
}
.save-btn.retro:hover {
  background: rgba(255,200,60,0.12);
  box-shadow: 0 0 10px rgba(255,200,60,0.15);
}
.save-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.pending-badge {
  font-size: 0.6rem;
  letter-spacing: 0.1em;
  color: rgba(255,200,60,0.85);
  background: rgba(255,200,60,0.08);
  border: 1px solid rgba(255,200,60,0.3);
  border-radius: 4px;
  padding: 2px 7px;
}

.locked-palpite {
  display: flex;
  align-items: center;
  gap: 10px;
}
.locked-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.65rem;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}
.locked-score {
  display: flex;
  align-items: center;
  gap: 4px;
}
.locked-num {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.3rem;
  color: rgba(57,255,106,0.7);
  line-height: 1;
  min-width: 18px;
  text-align: center;
}
.locked-sep {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1rem;
  color: var(--text-muted);
}
.pontos-badge {
  display: flex;
  align-items: baseline;
  gap: 2px;
  background: rgba(57,255,106,0.12);
  border: 1px solid rgba(57,255,106,0.3);
  border-radius: 6px;
  padding: 3px 8px;
  color: var(--neon);
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.1rem;
  line-height: 1;
}
.no-palpite {
  font-size: 0.72rem;
  color: var(--text-muted);
  letter-spacing: 0.04em;
  opacity: 0.7;
}

/* Outros palpites */
.outros-palpites {
  border-top: 1px solid rgba(57,255,106,0.07);
  background: rgba(0,0,0,0.2);
}

.outros-loading {
  display: flex;
  justify-content: center;
  padding: 16px;
}
.loader-ring-xs {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(57,255,106,0.15);
  border-top-color: var(--neon);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.outros-empty {
  text-align: center;
  padding: 14px;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.outros-list {
  display: flex;
  flex-direction: column;
}

.outro-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 8px 14px;
  border-bottom: 1px solid rgba(255,255,255,0.03);
  transition: background 0.15s;
}
.outro-item:last-child { border-bottom: none; }
.outro-item:hover { background: rgba(255,255,255,0.02); }

.outro-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(57,255,106,0.12);
  border: 1px solid rgba(57,255,106,0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.75rem;
  color: var(--neon);
  flex-shrink: 0;
}

.outro-name {
  flex: 1;
  font-size: 0.8rem;
  color: rgba(255,255,255,0.7);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.outro-score {
  display: flex;
  align-items: center;
  gap: 3px;
  flex-shrink: 0;
}
.outro-num {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.1rem;
  color: rgba(255,255,255,0.6);
  min-width: 14px;
  text-align: center;
  line-height: 1;
}
.outro-sep {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.outro-pontos {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.85rem;
  min-width: 28px;
  text-align: right;
  flex-shrink: 0;
  color: var(--text-muted);
}
.pts-exact .outro-num { color: var(--neon); }
.pts-exact .outro-pontos { color: var(--neon); }
.pts-winner .outro-num { color: rgba(255,210,80,0.85); }
.pts-winner .outro-pontos { color: rgba(255,210,80,0.85); }
.pts-miss .outro-pontos { color: rgba(255,80,80,0.5); }
</style>
