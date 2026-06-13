<template>
  <div class="page-bg min-h-screen pb-12">

    <!-- Header -->
    <div class="page-header animate-fade-up">
      <div class="header-inner">
        <button class="back-btn" @click="router.push('/boloes')">←</button>
        <div class="flex-1 min-w-0">
          <h1 class="font-display page-title neon-text">{{ bolao?.name ?? '...' }}</h1>
        </div>
        <button class="icon-btn" title="Como funciona" @click="router.push('/como-funciona')" style="font-size: 1rem; color: var(--text-muted); font-weight: 600; font-family: 'DM Sans', sans-serif;">
          ?
        </button>
        <button class="icon-btn" title="Ranking" @click="router.push(`/boloes/${route.params.id}/ranking`)">
          <span style="font-size: 1.2rem;">🏆</span>
        </button>
      </div>

      <!-- Tabs -->
      <div class="tabs-bar">
        <button
          class="tab-btn"
          :class="{ active: tab === 'jogos' }"
          @click="tab = 'jogos'"
        >
          <span>⚽</span> JOGOS
        </button>
        <button
          class="tab-btn"
          :class="{ active: tab === 'feed' }"
          @click="tab = 'feed'"
        >
          <span class="feed-badge-dot" :class="{ pulse: tab !== 'feed' }" />
          ATIVIDADE
        </button>
        <button
          v-if="isAdmin"
          class="tab-btn"
          :class="{ active: tab === 'admin' }"
          @click="tab = 'admin'"
        >
          ⚙ ADMIN
        </button>
      </div>
    </div>

    <div class="max-w-lg mx-auto px-4">

      <!-- Invite card -->
      <div v-if="bolao" class="invite-card animate-fade-up stagger-1">
        <div class="invite-label">
          <span style="font-size: 0.7rem; letter-spacing: 0.12em; color: var(--text-muted); font-family: 'Bebas Neue', sans-serif;">LINK DE CONVITE</span>
        </div>
        <div class="invite-body">
          <span class="invite-link">{{ inviteLink }}</span>
          <button class="copy-btn" @click="copyInvite">
            <span class="font-display" style="font-size: 0.78rem; letter-spacing: 0.1em;">{{ copied ? 'COPIADO ✓' : 'COPIAR' }}</span>
          </button>
        </div>
      </div>

      <!-- Tab: Admin -->
      <div v-if="isAdmin" v-show="tab === 'admin'" class="admin-tab">

        <!-- Toggle retroativo -->
        <div class="settings-card">
          <div class="settings-row">
            <div class="settings-info">
              <span class="font-display settings-title">PALPITES RETROATIVOS</span>
              <span class="settings-desc">Permite que participantes apostem em jogos que já começaram (sujeito à sua aprovação)</span>
            </div>
            <button
              class="toggle-btn"
              :class="{ active: bolao?.retroativo_enabled }"
              @click="toggleRetroativo"
            >
              <span class="toggle-knob" />
            </button>
          </div>
        </div>

        <!-- Palpites pendentes -->
        <div v-if="palpitesPendentes.length > 0" class="admin-panel">
          <div class="admin-panel-header">
            <span class="font-display" style="font-size: 0.85rem; letter-spacing: 0.1em; color: rgba(255,200,60,0.9);">AGUARDANDO APROVAÇÃO</span>
            <span class="pending-count">{{ palpitesPendentes.length }}</span>
          </div>
          <div class="pending-list">
            <div v-for="p in palpitesPendentes" :key="p.id" class="pending-item">
              <div class="pending-info">
                <span class="pending-user">{{ p.user_name }}</span>
                <span class="pending-jogo">{{ traduzTime(p.home_team) }} × {{ traduzTime(p.away_team) }}</span>
                <span v-if="p.finished && p.jogo_home_score != null" class="pending-resultado">
                  Resultado: {{ p.jogo_home_score }} – {{ p.jogo_away_score }}
                </span>
              </div>
              <div class="pending-score">
                <span class="font-display" style="font-size: 1.2rem; color: rgba(255,200,60,0.85);">{{ p.home_score }} – {{ p.away_score }}</span>
              </div>
              <div class="pending-actions">
                <button class="action-btn approve" @click="aprovar(p.id)">✓</button>
                <button class="action-btn reject" @click="rejeitar(p.id)">✕</button>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="admin-empty">Nenhum palpite aguardando aprovação.</div>

        <!-- Retroativos aprovados -->
        <div v-if="palpitesAprovados.length > 0" class="admin-panel" style="margin-top: 12px;">
          <div class="admin-panel-header">
            <span class="font-display" style="font-size: 0.85rem; letter-spacing: 0.1em; color: rgba(100,220,130,0.9);">RETROATIVOS APROVADOS</span>
            <span class="pending-count" style="background: rgba(100,220,130,0.15); color: rgba(100,220,130,0.9);">{{ palpitesAprovados.length }}</span>
          </div>
          <div class="pending-list">
            <div v-for="p in palpitesAprovados" :key="p.id" class="pending-item">
              <div class="pending-info">
                <span class="pending-user">{{ p.user_name }}</span>
                <span class="pending-jogo">{{ traduzTime(p.home_team) }} × {{ traduzTime(p.away_team) }}</span>
                <span v-if="p.finished && p.jogo_home_score != null" class="pending-resultado">
                  Resultado: {{ p.jogo_home_score }} – {{ p.jogo_away_score }}
                </span>
              </div>
              <div class="pending-score">
                <span class="font-display" style="font-size: 1.2rem; color: rgba(100,220,130,0.85);">{{ p.home_score }} – {{ p.away_score }}</span>
              </div>
              <div class="pending-actions">
                <button class="action-btn reject" title="Remover palpite" @click="desaprovar(p.id)">✕</button>
              </div>
            </div>
          </div>
        </div>

      </div>

      <!-- Tab: Jogos -->
      <div v-show="tab === 'jogos'">
        <div v-if="loadingJogos" class="flex justify-center py-16">
          <div class="loader-ring" />
        </div>

        <div v-else>
          <div v-for="(group, stage) in jogosByStage" :key="stage" class="stage-group">
            <div class="stage-header">
              <div class="stage-line" />
              <span class="stage-label">{{ formatStage(String(stage)) }}</span>
              <div class="stage-line" />
            </div>
            <div class="flex flex-col gap-2">
              <JogoCard
                v-for="jogo in group"
                :key="jogo.id"
                :jogo="jogo"
                :palpite="palpiteMap[jogo.id]"
                :bolao-id="bolaoId"
                @save="(h, a) => savePalpite(jogo.id, h, a)"
                @save-retroativo="(h, a) => savePalpiteRetroativo(jogo.id, h, a)"
              />
            </div>
          </div>

          <div v-if="Object.keys(jogosByStage).length === 0" class="empty-state">
            <span style="font-size: 2.5rem;">⚽</span>
            <p class="font-display" style="color: var(--text-muted); font-size: 1.3rem; letter-spacing: 0.06em; margin-top: 12px;">JOGOS NÃO CARREGADOS</p>
          </div>
        </div>
      </div>

      <!-- Tab: Feed -->
      <div v-show="tab === 'feed'">
        <FeedPanel :bolao-id="bolaoId" :active="tab === 'feed'" />
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { getBolao, listPalpites, upsertPalpite, upsertPalpiteRetroativo, listPalpitesPendentes, aprovarPalpite, rejeitarPalpite, listPalpitesRetroativosAprovados, desaprovarPalpite, setRetroativoEnabled } from '@/api/bolao'
import { listJogos } from '@/api/jogo'
import { useAuthStore } from '@/stores/auth'
import { traduzTime } from '@/utils/teams'
import JogoCard from '@/components/bolao/JogoCard.vue'
import FeedPanel from '@/components/bolao/FeedPanel.vue'
import type { Bolao, Jogo, Palpite, PalpitePendente } from '@/types'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()

const bolao = ref<Bolao | null>(null)
const jogos = ref<Jogo[]>([])
const palpites = ref<Palpite[]>([])
const palpitesPendentes = ref<PalpitePendente[]>([])
const palpitesAprovados = ref<PalpitePendente[]>([])
const loadingJogos = ref(true)
const copied = ref(false)
const tab = ref<'jogos' | 'feed' | 'admin'>('jogos')

const isAdmin = computed(() => bolao.value?.admin_id === authStore.currentUserId)

const bolaoId = route.params.id as string

const inviteLink = computed(() =>
  bolao.value ? `${window.location.origin}/join/${bolao.value.invite_token}` : '',
)

const palpiteMap = computed(() => {
  const map: Record<string, Palpite> = {}
  for (const p of palpites.value) map[p.jogo_id] = p
  return map
})

const jogosByStage = computed(() => {
  const groups: Record<string, Jogo[]> = {}
  for (const j of jogos.value) {
    if (!groups[j.stage]) groups[j.stage] = []
    groups[j.stage].push(j)
  }
  return groups
})

onMounted(async () => {
  try {
    const [b, j, p] = await Promise.all([
      getBolao(bolaoId),
      listJogos(),
      listPalpites(bolaoId),
    ])
    bolao.value = b
    jogos.value = j
    palpites.value = p
    if (b.admin_id === authStore.currentUserId) {
      ;[palpitesPendentes.value, palpitesAprovados.value] = await Promise.all([
        listPalpitesPendentes(bolaoId),
        listPalpitesRetroativosAprovados(bolaoId),
      ])
    }
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro ao carregar bolão', detail: e.message, life: 4000 })
  } finally {
    loadingJogos.value = false
  }
})

async function savePalpite(jogoId: string, home: number, away: number) {
  try {
    const p = await upsertPalpite(bolaoId, jogoId, home, away)
    const idx = palpites.value.findIndex((x) => x.jogo_id === jogoId)
    if (idx >= 0) palpites.value[idx] = p
    else palpites.value.push(p)
    toast.add({ severity: 'success', summary: 'Palpite salvo!', life: 2000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function savePalpiteRetroativo(jogoId: string, home: number, away: number) {
  try {
    const p = await upsertPalpiteRetroativo(bolaoId, jogoId, home, away)
    const idx = palpites.value.findIndex((x) => x.jogo_id === jogoId)
    if (idx >= 0) palpites.value[idx] = p
    else palpites.value.push(p)
    toast.add({ severity: 'success', summary: 'Palpite enviado para aprovação!', life: 3000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function aprovar(palpiteId: string) {
  try {
    await aprovarPalpite(bolaoId, palpiteId)
    palpitesPendentes.value = palpitesPendentes.value.filter((p) => p.id !== palpiteId)
    palpites.value = await listPalpites(bolaoId)
    toast.add({ severity: 'success', summary: 'Palpite aprovado!', life: 2000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function rejeitar(palpiteId: string) {
  try {
    await rejeitarPalpite(bolaoId, palpiteId)
    palpitesPendentes.value = palpitesPendentes.value.filter((p) => p.id !== palpiteId)
    palpites.value = await listPalpites(bolaoId)
    toast.add({ severity: 'success', summary: 'Palpite rejeitado.', life: 2000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function toggleRetroativo() {
  if (!bolao.value) return
  try {
    const updated = await setRetroativoEnabled(bolaoId, !bolao.value.retroativo_enabled)
    bolao.value = updated
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function desaprovar(palpiteId: string) {
  try {
    await desaprovarPalpite(bolaoId, palpiteId)
    palpitesAprovados.value = palpitesAprovados.value.filter((p) => p.id !== palpiteId)
    palpites.value = await listPalpites(bolaoId)
    toast.add({ severity: 'warn', summary: 'Palpite removido.', life: 2000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

function copyInvite() {
  navigator.clipboard.writeText(inviteLink.value)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

function formatStage(stage: string) {
  const map: Record<string, string> = {
    GROUP_STAGE: 'FASE DE GRUPOS',
    ROUND_OF_16: 'OITAVAS DE FINAL',
    QUARTER_FINALS: 'QUARTAS DE FINAL',
    SEMI_FINALS: 'SEMIFINAL',
    THIRD_PLACE: 'TERCEIRO LUGAR',
    FINAL: 'FINAL',
  }
  return map[stage] ?? stage.replace(/_/g, ' ')
}
</script>

<style scoped>
.page-bg {
  background: var(--pitch);
}
.page-header {
  background: linear-gradient(180deg, rgba(10,26,14,1) 0%, rgba(10,26,14,0.85) 100%);
  border-bottom: 1px solid rgba(57,255,106,0.1);
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(12px);
}
.header-inner {
  max-width: 512px;
  margin: 0 auto;
  padding: 16px 16px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}
.page-title {
  font-size: 1.8rem;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.back-btn {
  font-size: 1.4rem;
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
  transition: color 0.2s;
  padding: 4px;
  line-height: 1;
  flex-shrink: 0;
}
.back-btn:hover { color: var(--neon); }
.icon-btn {
  background: rgba(57,255,106,0.06);
  border: 1px solid rgba(57,255,106,0.15);
  border-radius: 8px;
  padding: 8px;
  cursor: pointer;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.icon-btn:hover { background: rgba(57,255,106,0.12); }

/* Tabs */
.tabs-bar {
  max-width: 512px;
  margin: 0 auto;
  display: flex;
  padding: 10px 16px 0;
  gap: 4px;
}
.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.82rem;
  letter-spacing: 0.12em;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: color 0.2s, border-color 0.2s;
  position: relative;
  bottom: -1px;
}
.tab-btn.active {
  color: var(--neon);
  border-bottom-color: var(--neon);
}
.tab-btn:not(.active):hover {
  color: rgba(255,255,255,0.6);
}

.feed-badge-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: background 0.3s;
}
.tab-btn.active .feed-badge-dot {
  background: var(--neon);
  box-shadow: 0 0 5px var(--neon);
}
.feed-badge-dot.pulse {
  background: var(--neon);
  box-shadow: 0 0 5px var(--neon);
  animation: pulse-dot 2s infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.invite-card {
  background: rgba(57,255,106,0.04);
  border: 1px solid rgba(57,255,106,0.18);
  border-radius: 12px;
  padding: 12px 16px;
  margin: 16px 0 20px;
}
.invite-label { margin-bottom: 6px; }
.invite-body {
  display: flex;
  align-items: center;
  gap: 10px;
}
.invite-link {
  flex: 1;
  font-size: 0.78rem;
  color: rgba(57,255,106,0.7);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'DM Mono', monospace;
}
.copy-btn {
  background: rgba(57,255,106,0.12);
  border: 1px solid rgba(57,255,106,0.3);
  border-radius: 6px;
  padding: 5px 10px;
  color: var(--neon);
  cursor: pointer;
  transition: background 0.2s;
  flex-shrink: 0;
  font-family: 'Bebas Neue', sans-serif;
}
.copy-btn:hover { background: rgba(57,255,106,0.2); }

.loader-ring {
  width: 40px;
  height: 40px;
  border: 2px solid rgba(57,255,106,0.15);
  border-top-color: var(--neon);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.stage-group { margin-bottom: 28px; }
.stage-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.stage-line {
  flex: 1;
  height: 1px;
  background: rgba(57,255,106,0.12);
}
.stage-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.78rem;
  letter-spacing: 0.14em;
  color: var(--text-muted);
  white-space: nowrap;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
}

/* Admin panel */
.admin-panel {
  margin-top: 24px;
  margin-bottom: 24px;
  background: rgba(255,200,60,0.04);
  border: 1px solid rgba(255,200,60,0.2);
  border-radius: 12px;
  overflow: hidden;
}
.admin-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255,200,60,0.12);
}
.pending-count {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.9rem;
  background: rgba(255,200,60,0.15);
  border: 1px solid rgba(255,200,60,0.3);
  border-radius: 4px;
  color: rgba(255,200,60,0.9);
  padding: 1px 8px;
}
.pending-list { display: flex; flex-direction: column; }
.pending-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.03);
}
.pending-item:last-child { border-bottom: none; }
.pending-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.pending-user {
  font-size: 0.82rem;
  font-weight: 600;
  color: rgba(255,255,255,0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pending-jogo {
  font-size: 0.72rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pending-resultado {
  font-size: 0.68rem;
  color: rgba(57,255,106,0.6);
  font-family: 'Bebas Neue', sans-serif;
  letter-spacing: 0.06em;
}
.pending-score { flex-shrink: 0; }
.pending-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.action-btn {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.action-btn.approve {
  background: rgba(57,255,106,0.12);
  color: var(--neon);
}
.action-btn.approve:hover { background: rgba(57,255,106,0.22); }
.action-btn.reject {
  background: rgba(255,80,80,0.1);
  color: rgba(255,80,80,0.7);
}
.action-btn.reject:hover { background: rgba(255,80,80,0.2); }

.admin-tab {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding-top: 8px;
}

.settings-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}

.settings-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.settings-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings-title {
  font-size: 0.85rem;
  letter-spacing: 0.1em;
  color: var(--text-primary);
}

.settings-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.toggle-btn {
  width: 48px;
  height: 26px;
  border-radius: 13px;
  background: rgba(255,255,255,0.1);
  border: 1px solid rgba(255,255,255,0.15);
  position: relative;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  flex-shrink: 0;
}
.toggle-btn.active {
  background: rgba(57,255,106,0.3);
  border-color: var(--neon);
}
.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: transform 0.2s, background 0.2s;
}
.toggle-btn.active .toggle-knob {
  transform: translateX(22px);
  background: var(--neon);
}

.admin-empty {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
  padding: 20px 0;
}
</style>
