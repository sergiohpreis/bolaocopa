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

      <!-- Taxa de entrada card (visível a todos) — só exibe se há algo a mostrar -->
      <div v-if="taxaEstado.taxa_definida || taxaEstado.proposta_ativa" class="taxa-card animate-fade-up stagger-2">

        <!-- Label fixo sempre no topo -->
        <span class="taxa-label">TAXA DE ENTRADA</span>

        <!-- Taxa já definida -->
        <div v-if="taxaEstado.taxa_definida" class="taxa-definida-body">
          <span class="taxa-valor">R$ {{ taxaEstado.taxa_definida }}</span>
          <span class="taxa-badge-definida">✓ APROVADA</span>
        </div>

        <!-- Proposta ativa -->
        <div v-else-if="taxaEstado.proposta_ativa" class="taxa-proposta-body">
          <div class="taxa-proposta-top">
            <span class="taxa-valor">R$ {{ taxaEstado.proposta_ativa.valor }}</span>
            <span class="taxa-pendente-badge">
              <span class="taxa-pendente-dot" />
              {{ taxaEstado.votos_pendentes }} {{ taxaEstado.votos_pendentes === 1 ? 'voto pendente' : 'votos pendentes' }}
            </span>
          </div>
          <div v-if="!jaVotei" class="taxa-vote-btns">
            <button class="taxa-btn-sim" :disabled="taxaLoading" @click="votar(true)">
              <span>✓</span> SIM
            </button>
            <button class="taxa-btn-nao" :disabled="taxaLoading" @click="votar(false)">
              <span>✕</span> NÃO
            </button>
          </div>
          <div v-else class="taxa-votado">
            <span class="taxa-votado-check">✓</span>
            <span>Seu voto foi registrado</span>
          </div>
        </div>

      </div>

      <!-- Drawer de exclusão -->
      <ExcluirBolaoDrawer
        :open="excluirDrawerOpen"
        :participantes="totalParticipantes"
        :palpites="totalPalpites"
        :loading="excluirLoading"
        @close="excluirDrawerOpen = false"
        @confirm="confirmarExclusao"
      />

      <!-- Tab: Admin -->
      <div v-if="isAdmin" v-show="tab === 'admin'" class="admin-tab">

        <!-- Propor taxa de entrada -->
        <div class="settings-card" v-if="!taxaEstado.taxa_definida && !taxaEstado.proposta_ativa">
          <div class="settings-info" style="margin-bottom: 12px;">
            <span class="font-display settings-title">TAXA DE ENTRADA</span>
            <span class="settings-desc">Proponha um valor para votação de todos os participantes. A taxa só é definida se aprovada por unanimidade.</span>
          </div>
          <div style="display: flex; gap: 8px;">
            <input
              v-model="taxaValorInput"
              type="text"
              placeholder="Ex: 50.00"
              class="taxa-input"
              :disabled="taxaLoading"
              @keydown.enter="submeterProposta"
            />
            <button class="taxa-propor-btn" :disabled="taxaLoading || !taxaValorInput.trim()" @click="submeterProposta">
              PROPOR
            </button>
          </div>
        </div>

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

        <!-- Retroativos aprovados (só exibe se retroativo está habilitado) -->

        <div v-if="bolao?.retroativo_enabled && palpitesAprovados.length > 0" class="admin-panel" style="margin-top: 12px;">
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

        <!-- WhatsApp — PROTOTYPE -->
        <WhatsAppAdminPanel
          :bolao-id="bolaoId"
          :linked-group="bolao?.wa_group_jid"
          :notifications-enabled="bolao?.wa_notifications_enabled ?? true"
          @group-changed="reloadBolao"
          @notifications-changed="reloadBolao"
        />

        <!-- Excluir bolão -->
        <div class="danger-zone">
          <div class="danger-zone-label font-display">ZONA DE PERIGO</div>
          <div class="danger-zone-row">
            <div class="danger-zone-info">
              <span class="danger-zone-title font-display">EXCLUIR BOLÃO</span>
              <span class="danger-zone-desc">Remove permanentemente o bolão e todos os dados associados.</span>
            </div>
            <button class="btn-excluir" @click="excluirDrawerOpen = true">Excluir</button>
          </div>
        </div>

      </div>

      <!-- Tab: Jogos -->
      <div v-show="tab === 'jogos'">
        <div v-if="loadingJogos" class="flex justify-center py-16">
          <div class="loader-ring" />
        </div>

        <div v-else>
          <!-- Banner: Ao Vivo -->
          <div v-if="jogosAoVivo.length > 0" class="live-banner animate-fade-up">
            <div class="live-banner-header">
              <span class="live-pulse-dot" />
              <span class="font-display live-banner-label">AO VIVO</span>
            </div>
            <div class="flex flex-col gap-2">
              <JogoCard
                v-for="jogo in jogosAoVivo"
                :key="jogo.id"
                :jogo="jogo"
                :palpite="palpiteMap[jogo.id]"
                :bolao-id="bolaoId"
                :retroativo-enabled="bolao?.retroativo_enabled ?? false"
                @save="(h, a) => savePalpite(jogo.id, h, a)"
                @save-retroativo="(h, a) => savePalpiteRetroativo(jogo.id, h, a)"
              />
            </div>
          </div>

          <!-- Toggle Lista / Chave — só quando há jogos de mata-mata -->
          <div v-if="temMataMata" class="jogo-view-bar">
            <button
              class="jogo-view-btn"
              :class="{ active: jogoView === 'lista' }"
              @click="jogoView = 'lista'"
            >
              LISTA
            </button>
            <button
              class="jogo-view-btn"
              :class="{ active: jogoView === 'chave' }"
              @click="jogoView = 'chave'"
            >
              CHAVE
            </button>
          </div>

          <!-- Modo LISTA -->
          <template v-if="jogoView === 'lista'">
            <!-- Filtros: Próximos / Encerrados -->
            <div class="jogo-filter-bar">
              <button
                class="jogo-filter-btn"
                :class="{ active: jogoFilter === 'proximos' }"
                @click="jogoFilter = 'proximos'"
              >
                PRÓXIMOS
              </button>
              <button
                class="jogo-filter-btn"
                :class="{ active: jogoFilter === 'encerrados' }"
                @click="jogoFilter = 'encerrados'"
              >
                ENCERRADOS
              </button>
            </div>

            <!-- Lista filtrada por fase -->
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
                  :retroativo-enabled="bolao?.retroativo_enabled ?? false"
                  @save="(h, a) => savePalpite(jogo.id, h, a)"
                  @save-retroativo="(h, a) => savePalpiteRetroativo(jogo.id, h, a)"
                />
              </div>
            </div>

            <div v-if="Object.keys(jogosByStage).length === 0 && jogosAoVivo.length === 0" class="empty-state">
              <span style="font-size: 2.5rem;">⚽</span>
              <p class="font-display" style="color: var(--text-muted); font-size: 1.3rem; letter-spacing: 0.06em; margin-top: 12px;">JOGOS NÃO CARREGADOS</p>
            </div>
          </template>

          <!-- Modo CHAVE — acordeão vertical por fase do mata-mata -->
          <BracketView
            v-else
            :colunas="colunasChave"
            :palpite-map="palpiteMap"
            :bolao-id="bolaoId"
            :retroativo-enabled="bolao?.retroativo_enabled ?? false"
            @save="savePalpite"
            @save-retroativo="savePalpiteRetroativo"
          />
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
import { getBolao, listPalpites, upsertPalpite, upsertPalpiteRetroativo, listPalpitesPendentes, aprovarPalpite, rejeitarPalpite, listPalpitesRetroativosAprovados, desaprovarPalpite, setRetroativoEnabled, getTaxaEstado, proporTaxa, votarTaxa, deleteBolao } from '@/api/bolao'
import { listJogos } from '@/api/jogo'
import { useAuthStore } from '@/stores/auth'
import { traduzTime } from '@/utils/teams'
import { formatStage, isMataMata, jogoDefinido, labelColuna, keyColuna, COLUNAS_MATA_MATA } from '@/utils/fases'
import JogoCard from '@/components/bolao/JogoCard.vue'
import FeedPanel from '@/components/bolao/FeedPanel.vue'
import BracketView from '@/components/bolao/BracketView.vue'
import ExcluirBolaoDrawer from '@/components/bolao/ExcluirBolaoDrawer.vue'
import WhatsAppAdminPanel from '@/components/WhatsAppAdminPanel.vue' // PROTOTYPE
import type { Bolao, Jogo, Palpite, PalpitePendente, TaxaEstado } from '@/types'

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

const excluirDrawerOpen = ref(false)
const excluirLoading = ref(false)

const totalParticipantes = computed(() => bolao.value ? 1 + palpitesPendentes.value.length : 0)
const totalPalpites = computed(() => palpites.value.length)

const taxaEstado = ref<TaxaEstado>({ votos_pendentes: 0 })
const taxaValorInput = ref('')
const taxaLoading = ref(false)
const jaVotei = ref(false)

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

const jogoFilter = ref<'proximos' | 'encerrados'>('proximos')
const jogoView = ref<'lista' | 'chave'>('lista')

// Jogos com os dois times conhecidos. A API cria registros de mata-mata com
// times vazios antes do confronto ser decidido — esses não entram na lista nem
// aceitam Palpite (aparecem como "a definir" na chave).
const jogosDefinidos = computed(() => jogos.value.filter(jogoDefinido))

const jogosAoVivo = computed(() =>
  jogosDefinidos.value.filter(j => !j.finished && new Date(j.starts_at) <= new Date())
)
const idsAoVivo = computed(() => new Set(jogosAoVivo.value.map(j => j.id)))

function groupByStage(list: Jogo[]): Record<string, Jogo[]> {
  const groups: Record<string, Jogo[]> = {}
  for (const j of list) {
    if (!groups[j.stage]) groups[j.stage] = []
    groups[j.stage].push(j)
  }
  return groups
}

const jogosByStage = computed(() => {
  const base = jogosDefinidos.value.filter(j => !idsAoVivo.value.has(j.id))
  if (jogoFilter.value === 'proximos') {
    return groupByStage(base.filter(j => !j.finished))
  }
  return groupByStage(base.filter(j => j.finished))
})

// Visão de Chaveamento: só disponível quando há ao menos um Jogo de Mata-mata
// (mesmo que ainda sem times definidos — a fase já existe na chave).
const temMataMata = computed(() => jogos.value.some(j => isMataMata(j.stage)))

// Seções do Mata-mata na ordem canônica. Cada coluna pode agrupar mais de uma
// Fase (Final + Disputa de 3º) e expõe os confrontos já decididos mais a
// contagem de vagas ainda "a definir" (Jogos da fase sem times, que a API cria
// antes do confronto sair). Uma fase sem nenhum Jogo conta como 1 vaga.
const colunasChave = computed(() => {
  const definidosPorStage = groupByStage(jogosDefinidos.value.filter(j => isMataMata(j.stage)))
  const todosPorStage = groupByStage(jogos.value.filter(j => isMataMata(j.stage)))
  return COLUNAS_MATA_MATA.map(stages => {
    const total = stages.reduce((n, s) => n + (todosPorStage[s]?.length ?? 0), 0)
    const jogos = stages
      .flatMap(s => definidosPorStage[s] ?? [])
      .sort((a, b) => new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime())
    // Vagas a definir: confrontos da fase ainda sem times. Se a fase nem existe
    // (total 0), mostra ao menos 1 vaga para sinalizar que a chave continua.
    const vagas = Math.max(total - jogos.length, total === 0 ? 1 : 0)
    return { key: keyColuna(stages), label: labelColuna(stages), jogos, vagas }
  })
})

onMounted(async () => {
  try {
    const [b, j, p, taxa] = await Promise.all([
      getBolao(bolaoId),
      listJogos(),
      listPalpites(bolaoId),
      getTaxaEstado(bolaoId),
    ])
    bolao.value = b
    jogos.value = j
    palpites.value = p
    taxaEstado.value = taxa
    jaVotei.value = taxa.meu_voto != null
    const hasProximos = j.some(jg => !jg.finished && new Date(jg.starts_at) > new Date())
    if (!hasProximos) jogoFilter.value = 'encerrados'
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

async function reloadBolao() {
  try {
    bolao.value = await getBolao(bolaoId)
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro ao recarregar bolão', detail: e.message, life: 3000 })
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
  if (!window.confirm('Remover este palpite retroativo? Esta ação não pode ser desfeita.')) return
  try {
    await desaprovarPalpite(bolaoId, palpiteId)
    palpitesAprovados.value = palpitesAprovados.value.filter((p) => p.id !== palpiteId)
    palpites.value = await listPalpites(bolaoId)
    toast.add({ severity: 'warn', summary: 'Palpite removido.', life: 2000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  }
}

async function carregarTaxa() {
  try {
    taxaEstado.value = await getTaxaEstado(bolaoId)
    jaVotei.value = taxaEstado.value.meu_voto != null
  } catch {
    // silently ignore — taxa card will just not update
  }
}

async function submeterProposta() {
  if (!taxaValorInput.value.trim()) return
  taxaLoading.value = true
  try {
    await proporTaxa(bolaoId, taxaValorInput.value.trim())
    taxaValorInput.value = ''
    await carregarTaxa()
    toast.add({ severity: 'success', summary: 'Proposta enviada!', life: 3000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
  } finally {
    taxaLoading.value = false
  }
}

async function votar(aprovado: boolean) {
  taxaLoading.value = true
  try {
    await votarTaxa(bolaoId, aprovado)
    jaVotei.value = true
    await carregarTaxa()
    toast.add({ severity: 'success', summary: aprovado ? 'Voto registrado!' : 'Proposta cancelada.', life: 3000 })
  } catch (e: any) {
    if (e.response?.status === 409) {
      jaVotei.value = true
      await carregarTaxa()
    } else {
      toast.add({ severity: 'error', summary: 'Erro', detail: e.message, life: 3000 })
    }
  } finally {
    taxaLoading.value = false
  }
}

async function confirmarExclusao() {
  excluirLoading.value = true
  try {
    await deleteBolao(bolaoId)
    toast.add({ severity: 'warn', summary: 'Bolão excluído.', life: 3000 })
    router.push('/boloes')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro ao excluir', detail: e.message, life: 4000 })
  } finally {
    excluirLoading.value = false
    excluirDrawerOpen.value = false
  }
}

function copyInvite() {
  navigator.clipboard.writeText(inviteLink.value)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
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

/* Taxa de entrada */
.taxa-card {
  background: rgba(57,255,106,0.04);
  border: 1px solid rgba(57,255,106,0.18);
  border-radius: 12px;
  padding: 14px 16px;
  margin-bottom: 16px;
}
.taxa-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.7rem;
  letter-spacing: 0.16em;
  color: var(--text-muted);
  display: block;
  margin-bottom: 8px;
}
.taxa-valor {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.75rem;
  color: var(--neon);
  letter-spacing: 0.04em;
  line-height: 1;
  text-shadow: 0 0 16px rgba(57,255,106,0.3);
}

/* Estado: taxa definida */
.taxa-definida-body {
  display: flex;
  align-items: center;
  gap: 12px;
}
.taxa-badge-definida {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.7rem;
  letter-spacing: 0.14em;
  background: rgba(57,255,106,0.1);
  border: 1px solid rgba(57,255,106,0.28);
  border-radius: 4px;
  color: var(--neon);
  padding: 3px 10px;
}

/* Estado: proposta ativa */
.taxa-proposta-body { display: flex; flex-direction: column; gap: 12px; }
.taxa-proposta-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.taxa-pendente-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  background: rgba(255,200,60,0.1);
  border: 1px solid rgba(255,200,60,0.25);
  border-radius: 4px;
  color: rgba(255,200,60,0.85);
  padding: 3px 10px;
  white-space: nowrap;
}
.taxa-pendente-dot {
  display: inline-block;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: rgba(255,200,60,0.85);
  animation: pulse-dot 2s infinite;
}
.taxa-vote-btns {
  display: flex;
  gap: 8px;
}
.taxa-btn-sim,
.taxa-btn-nao {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.9rem;
  letter-spacing: 0.12em;
  border-radius: 8px;
  padding: 10px 0;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
}
.taxa-btn-sim {
  background: rgba(57,255,106,0.1);
  color: var(--neon);
  border: 1px solid rgba(57,255,106,0.3);
}
.taxa-btn-sim:hover:not(:disabled) {
  background: rgba(57,255,106,0.18);
  transform: translateY(-1px);
}
.taxa-btn-nao {
  background: rgba(255,70,70,0.08);
  color: rgba(255,100,100,0.85);
  border: 1px solid rgba(255,70,70,0.25);
}
.taxa-btn-nao:hover:not(:disabled) {
  background: rgba(255,70,70,0.15);
  transform: translateY(-1px);
}
.taxa-btn-sim:disabled,
.taxa-btn-nao:disabled { opacity: 0.45; cursor: not-allowed; transform: none; }

/* Estado: já votei */
.taxa-votado {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(57,255,106,0.55);
  font-size: 0.82rem;
}
.taxa-votado-check {
  font-size: 0.9rem;
  color: var(--neon);
  opacity: 0.7;
}

/* Admin: propor taxa */
.taxa-input {
  flex: 1;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 8px;
  padding: 9px 12px;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-family: 'DM Mono', monospace;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.taxa-input:focus {
  border-color: rgba(57,255,106,0.4);
  box-shadow: 0 0 0 3px rgba(57,255,106,0.08);
}
.taxa-input::placeholder { color: var(--text-muted); }
.taxa-propor-btn {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.85rem;
  letter-spacing: 0.12em;
  background: rgba(57,255,106,0.1);
  border: 1px solid rgba(57,255,106,0.28);
  border-radius: 8px;
  padding: 9px 18px;
  color: var(--neon);
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
  white-space: nowrap;
}
.taxa-propor-btn:hover:not(:disabled) {
  background: rgba(57,255,106,0.18);
  transform: translateY(-1px);
}
.taxa-propor-btn:disabled { opacity: 0.35; cursor: not-allowed; transform: none; }

/* Zona de perigo */
.danger-zone {
  margin-top: 24px;
  background: rgba(255,50,50,0.04);
  border: 1px solid rgba(255,60,60,0.18);
  border-radius: 12px;
  padding: 14px 16px;
}
.danger-zone-label {
  font-size: 0.68rem;
  letter-spacing: 0.18em;
  color: rgba(255,100,100,0.5);
  margin-bottom: 10px;
}
.danger-zone-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.danger-zone-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.danger-zone-title {
  font-size: 0.82rem;
  letter-spacing: 0.1em;
  color: rgba(255,100,100,0.8);
}
.danger-zone-desc {
  font-size: 0.74rem;
  color: var(--text-muted);
  line-height: 1.4;
}
.btn-excluir {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.82rem;
  letter-spacing: 0.1em;
  background: transparent;
  border: 1px solid rgba(255,60,60,0.4);
  border-radius: 8px;
  color: rgba(255,100,100,0.8);
  padding: 7px 14px;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: background 0.15s, border-color 0.15s;
}
.btn-excluir:hover {
  background: rgba(255,50,50,0.08);
  border-color: rgba(255,60,60,0.65);
}

/* Banner AO VIVO */
.live-banner {
  border-left: 3px solid rgba(255, 80, 80, 0.7);
  background: rgba(255, 50, 50, 0.04);
  border-radius: 0 12px 12px 0;
  padding: 12px 0 12px 14px;
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
}
.live-banner::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255,80,80,0.04) 0%, transparent 60%);
  pointer-events: none;
  animation: scan-live 3s linear infinite;
}
@keyframes scan-live {
  0% { transform: translateY(-100%); }
  100% { transform: translateY(200%); }
}
.live-banner-header {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-bottom: 10px;
}
.live-pulse-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #ff8080;
  box-shadow: 0 0 6px rgba(255,80,80,0.6);
  animation: pulse-live-dot 1.4s ease-in-out infinite;
}
@keyframes pulse-live-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.45; transform: scale(0.75); }
}
.live-banner-label {
  font-size: 0.78rem;
  letter-spacing: 0.18em;
  color: #ff8080;
}

/* Filtros de jogos */
.jogo-filter-bar {
  display: flex;
  gap: 0;
  margin-bottom: 20px;
  border-bottom: 1px solid rgba(57,255,106,0.1);
}
.jogo-filter-btn {
  padding: 8px 18px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.82rem;
  letter-spacing: 0.14em;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  position: relative;
  bottom: -1px;
  transition: color 0.2s, border-color 0.2s;
}
.jogo-filter-btn.active {
  color: var(--neon);
  border-bottom-color: var(--neon);
}
.jogo-filter-btn:not(.active):hover {
  color: rgba(255,255,255,0.55);
}

/* Toggle Lista / Chave — pílula segmentada, distinta da barra de filtros */
.jogo-view-bar {
  display: inline-flex;
  gap: 4px;
  margin-bottom: 16px;
  padding: 3px;
  border: 1px solid rgba(57,255,106,0.18);
  border-radius: 999px;
  background: rgba(57,255,106,0.04);
}
.jogo-view-btn {
  padding: 5px 16px;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--text-muted);
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.78rem;
  letter-spacing: 0.14em;
  cursor: pointer;
  transition: color 0.2s, background 0.2s;
}
.jogo-view-btn.active {
  color: var(--neon);
  background: rgba(57,255,106,0.14);
}
.jogo-view-btn:not(.active):hover {
  color: rgba(255,255,255,0.55);
}
</style>
