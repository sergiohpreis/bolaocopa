<!-- PROTOTYPE — throwaway. Answers: does the QR→connect→link-group→test-notify flow
     feel right for the admin UX? -->
<template>
  <div class="wa-panel">
    <div class="wa-header">
      <span class="wa-icon">📱</span>
      <span class="wa-title font-display">WHATSAPP</span>
      <span class="wa-badge" :class="badgeClass">{{ badgeLabel }}</span>
    </div>

    <!-- Disconnected / awaiting QR without active QR (show reconnect button) -->
    <template v-if="status?.state === 'disconnected' || (status?.state === 'awaiting_qr' && !qrImage)">
      <p class="wa-hint">Conecte seu WhatsApp para enviar notificações ao grupo do bolão.</p>
      <button class="wa-btn-primary" :disabled="connecting" @click="startConnect">
        {{ connecting ? 'Conectando…' : 'Conectar WhatsApp' }}
      </button>
    </template>

    <!-- QR code -->
    <template v-else-if="status?.state === 'awaiting_qr'">
      <p class="wa-hint">Escaneie o QR code com o WhatsApp do celular.</p>
      <div v-if="qrImage" class="wa-qr-wrap">
        <img :src="`data:image/png;base64,${qrImage}`" alt="QR WhatsApp" class="wa-qr" />
      </div>
      <p v-else class="wa-hint">Gerando QR…</p>
      <button class="wa-btn-ghost" @click="cancelConnect">Cancelar</button>
    </template>

    <!-- Connected -->
    <template v-else-if="status?.state === 'connected'">

      <!-- Group linking -->
      <div v-if="!status.linked_group" class="wa-section">
        <p class="wa-hint">Selecione o grupo do WhatsApp onde as notificações serão enviadas.</p>
        <p v-if="props.bolaoName" class="wa-hint" style="color: var(--neon); opacity: 0.8;">
          Exibindo apenas grupos cujo nome contém <strong>"{{ props.bolaoName }}"</strong>.
        </p>
        <button class="wa-btn-ghost" :disabled="loadingGroups" @click="fetchGroups">
          {{ loadingGroups ? 'Carregando…' : 'Carregar grupos' }}
        </button>
        <div v-if="filteredGroups.length" class="wa-group-list">
          <button
            v-for="g in filteredGroups"
            :key="g.jid"
            class="wa-group-item"
            @click="selectGroup(g.jid)"
          >
            {{ g.name }}
          </button>
        </div>
        <p v-else-if="groups.length && !filteredGroups.length" class="wa-hint" style="color: #ff6b6b;">
          Nenhum grupo encontrado com o nome "{{ props.bolaoName }}". Crie um grupo no WhatsApp com esse nome.
        </p>
      </div>

      <!-- Group linked — test notifications -->
      <div v-else class="wa-section">
        <div class="wa-linked-group">
          <span class="wa-linked-label">GRUPO VINCULADO</span>
          <span class="wa-linked-name">{{ groupName || status.linked_group }}</span>
          <button class="wa-btn-link" @click="unlinkGroup">trocar</button>
        </div>

        <p class="wa-hint" style="margin-top: 1rem;">Testar notificações:</p>
        <div class="wa-test-btns">
          <button class="wa-btn-test" :disabled="sending" @click="testNotify('faltam_dez_minutos')">
            ⏰ Faltam 10 min
          </button>
          <button class="wa-btn-test" :disabled="sending" @click="testNotify('partida_iniciando')">
            🚀 Partida iniciou
          </button>
          <button class="wa-btn-test" :disabled="sending" @click="testNotify('fim_de_jogo')">
            ⚽ Fim de jogo
          </button>
        </div>
        <p v-if="lastResult" class="wa-result" :class="lastResultOk ? 'ok' : 'err'">
          {{ lastResult }}
        </p>

        <button class="wa-btn-ghost wa-disconnect" @click="doDisconnect">
          Desconectar WhatsApp
        </button>
      </div>
    </template>

    <p v-if="error" class="wa-error">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  getStatus, getQR, connect, disconnect, listGroups, linkGroup, sendNotification,
} from '@/api/whatsapp'
import type { WAStatus, WAGroup } from '@/types'
import { useWAPoller } from '@/composables/useWAPoller'

const props = withDefaults(defineProps<{ bolaoName?: string }>(), { bolaoName: '' })

const status = ref<WAStatus | null>(null)
const qrImage = ref<string>('')
const groups = ref<WAGroup[]>([])
const error = ref('')
const connecting = ref(false)
const loadingGroups = ref(false)
const sending = ref(false)
const lastResult = ref('')
const lastResultOk = ref(true)

const badgeClass = computed(() => {
  switch (status.value?.state) {
    case 'connected': return 'badge-connected'
    case 'awaiting_qr': return 'badge-qr'
    default: return 'badge-off'
  }
})

const badgeLabel = computed(() => {
  switch (status.value?.state) {
    case 'connected': return 'conectado'
    case 'awaiting_qr': return 'aguardando QR'
    default: return 'desconectado'
  }
})

const filteredGroups = computed(() => {
  if (!props.bolaoName) return groups.value
  const needle = props.bolaoName.toLowerCase()
  return groups.value.filter(g => g.name.toLowerCase().includes(needle))
})

const groupName = computed(() => {
  if (!status.value?.linked_group) return ''
  return groups.value.find(g => g.jid === status.value!.linked_group)?.name ?? ''
})

async function refresh() {
  try {
    error.value = ''
    status.value = await getStatus()
    if (status.value.state === 'awaiting_qr' && status.value.has_qr) {
      qrImage.value = await getQR()
    }
    // Auto-load groups when connected so we can resolve the linked group name
    if (status.value.state === 'connected' && !groups.value.length && !loadingGroups.value) {
      await fetchGroups()
    }
  } catch (e: any) {
    // service may not be running — show soft error
    error.value = 'Serviço WhatsApp indisponível. Suba o container com --profile whatsapp.'
  }
}

async function startConnect() {
  connecting.value = true
  error.value = ''
  try {
    await connect()
    await refresh()
  } catch (e: any) {
    error.value = e.message
  } finally {
    connecting.value = false
  }
}

async function cancelConnect() {
  try {
    await disconnect()
    await refresh()
  } catch (e: any) {
    error.value = e.message
  }
}

async function doDisconnect() {
  try {
    await disconnect()
    await refresh()
  } catch (e: any) {
    error.value = e.message
  }
}

async function fetchGroups() {
  loadingGroups.value = true
  error.value = ''
  try {
    groups.value = await listGroups()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loadingGroups.value = false
  }
}

async function selectGroup(jid: string) {
  await linkGroup(jid)
  await refresh()
}

async function unlinkGroup() {
  await linkGroup('')
  await refresh()
}

async function testNotify(type: 'fim_de_jogo' | 'faltam_dez_minutos' | 'partida_iniciando') {
  sending.value = true
  lastResult.value = ''
  try {
    await sendNotification({
      type,
      home_team: 'Brasil',
      away_team: 'Argentina',
      home_score: type === 'fim_de_jogo' ? 2 : undefined,
      away_score: type === 'fim_de_jogo' ? 1 : undefined,
      winners: type === 'fim_de_jogo'
        ? [{ name: 'Sergio', pontos: 10 }, { name: 'João', pontos: 3 }]
        : undefined,
    })
    lastResult.value = '✓ Mensagem enviada!'
    lastResultOk.value = true
  } catch (e: any) {
    lastResult.value = '✕ ' + e.message
    lastResultOk.value = false
  } finally {
    sending.value = false
  }
}

refresh()

// Poll while QR is displayed so it refreshes automatically.
// useWAPoller uses onScopeDispose for safe cleanup on unmount.
useWAPoller(() => {
  if (status.value?.state === 'awaiting_qr') refresh()
})
</script>

<style scoped>
.wa-panel {
  background: color-mix(in srgb, var(--bg-card) 100%, transparent);
  border: 1px solid rgba(57, 255, 106, 0.15);
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.wa-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.wa-icon { font-size: 1.2rem; }

.wa-title {
  font-size: 1rem;
  letter-spacing: 0.1em;
  color: var(--text-primary);
}

.wa-badge {
  margin-left: auto;
  font-size: 0.65rem;
  padding: 2px 8px;
  border-radius: 20px;
  font-family: 'DM Sans', sans-serif;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.badge-connected { background: rgba(57, 255, 106, 0.15); color: var(--neon); }
.badge-qr        { background: rgba(255, 200, 0, 0.15);  color: #ffc800; }
.badge-off       { background: rgba(255,255,255, 0.06);  color: var(--text-muted); }

.wa-hint {
  font-size: 0.82rem;
  color: var(--text-muted);
  font-family: 'DM Sans', sans-serif;
  margin: 0;
}

.wa-section { display: flex; flex-direction: column; gap: 0.6rem; }

.wa-qr-wrap {
  display: flex;
  justify-content: center;
  background: white;
  border-radius: 8px;
  padding: 1rem;
}

.wa-qr { width: 220px; height: 220px; }

.wa-group-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  max-height: 200px;
  overflow-y: auto;
}

.wa-group-item {
  text-align: left;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  padding: 0.6rem 0.9rem;
  color: var(--text-primary);
  font-size: 0.85rem;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  transition: border-color 0.15s;
}
.wa-group-item:hover { border-color: var(--neon); }

.wa-linked-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(57, 255, 106, 0.06);
  border: 1px solid rgba(57, 255, 106, 0.2);
  border-radius: 8px;
  padding: 0.6rem 0.9rem;
}

.wa-linked-label {
  font-size: 0.6rem;
  letter-spacing: 0.12em;
  color: var(--neon);
  font-family: 'Bebas Neue', sans-serif;
}

.wa-linked-name {
  flex: 1;
  font-size: 0.85rem;
  color: var(--text-primary);
  font-family: 'DM Sans', sans-serif;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wa-btn-link {
  font-size: 0.75rem;
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
  font-family: 'DM Sans', sans-serif;
  text-decoration: underline;
}

.wa-test-btns {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.wa-btn-test {
  flex: 1;
  min-width: 110px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  color: var(--text-primary);
  font-size: 0.78rem;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.wa-btn-test:hover:not(:disabled) { border-color: var(--neon); background: rgba(57,255,106,0.06); }
.wa-btn-test:disabled { opacity: 0.5; cursor: default; }

.wa-result {
  font-size: 0.8rem;
  font-family: 'DM Sans', sans-serif;
  padding: 0.4rem 0.7rem;
  border-radius: 6px;
}
.wa-result.ok  { color: var(--neon); background: rgba(57,255,106,0.08); }
.wa-result.err { color: #ff6b6b;     background: rgba(255,107,107,0.08); }

.wa-btn-primary {
  background: var(--neon);
  color: var(--bg-dark);
  border: none;
  border-radius: 8px;
  padding: 0.6rem 1.2rem;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.95rem;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: opacity 0.15s;
  align-self: flex-start;
}
.wa-btn-primary:hover:not(:disabled) { opacity: 0.85; }
.wa-btn-primary:disabled { opacity: 0.5; cursor: default; }

.wa-btn-ghost {
  background: none;
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 8px;
  padding: 0.5rem 1rem;
  color: var(--text-muted);
  font-family: 'DM Sans', sans-serif;
  font-size: 0.82rem;
  cursor: pointer;
  align-self: flex-start;
  transition: border-color 0.15s;
}
.wa-btn-ghost:hover { border-color: rgba(255,255,255,0.3); }

.wa-disconnect { margin-top: 0.5rem; color: #ff6b6b; border-color: rgba(255,107,107,0.2); }
.wa-disconnect:hover { border-color: #ff6b6b; }

.wa-error {
  font-size: 0.78rem;
  color: #ff6b6b;
  font-family: 'DM Sans', sans-serif;
  background: rgba(255,107,107,0.08);
  border-radius: 6px;
  padding: 0.4rem 0.7rem;
}
</style>
