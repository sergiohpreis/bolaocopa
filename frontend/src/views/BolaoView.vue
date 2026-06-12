<template>
  <div class="page-bg min-h-screen pb-12">

    <!-- Header -->
    <div class="page-header animate-fade-up">
      <div class="header-inner">
        <button class="back-btn" @click="router.push('/boloes')">←</button>
        <div class="flex-1 min-w-0">
          <h1 class="font-display page-title neon-text">{{ bolao?.name ?? '...' }}</h1>
        </div>
        <button class="icon-btn" title="Ranking" @click="router.push(`/boloes/${route.params.id}/ranking`)">
          <span style="font-size: 1.2rem;">🏆</span>
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

      <!-- Loading -->
      <div v-if="loadingJogos" class="flex justify-center py-16">
        <div class="loader-ring" />
      </div>

      <!-- Jogos by stage -->
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
              @save="(h, a) => savePalpite(jogo.id, h, a)"
            />
          </div>
        </div>

        <div v-if="Object.keys(jogosByStage).length === 0" class="empty-state">
          <span style="font-size: 2.5rem;">⚽</span>
          <p class="font-display" style="color: var(--text-muted); font-size: 1.3rem; letter-spacing: 0.06em; margin-top: 12px;">JOGOS NÃO CARREGADOS</p>
        </div>
      </div>

      <!-- Feed de atividades -->
      <FeedPanel :bolao-id="bolaoId" />

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { getBolao, listPalpites, upsertPalpite } from '@/api/bolao'
import { listJogos } from '@/api/jogo'
import JogoCard from '@/components/bolao/JogoCard.vue'
import FeedPanel from '@/components/bolao/FeedPanel.vue'
import type { Bolao, Jogo, Palpite } from '@/types'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const bolao = ref<Bolao | null>(null)
const jogos = ref<Jogo[]>([])
const palpites = ref<Palpite[]>([])
const loadingJogos = ref(true)
const copied = ref(false)

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
  padding: 16px;
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
</style>
