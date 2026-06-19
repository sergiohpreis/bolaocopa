<!-- PROTOTYPE — throwaway. Admin panel for WhatsApp group notifications. -->
<template>
  <div class="wa-panel">
    <div class="wa-header">
      <span class="wa-icon">📱</span>
      <span class="wa-title font-display">WHATSAPP</span>
      <span class="wa-badge" :class="badgeClass">{{ badgeLabel }}</span>
    </div>

    <!-- Disconnected / awaiting QR without active QR (show connect button) -->
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
      <div v-if="!props.linkedGroup" class="wa-section">
        <p class="wa-hint">Selecione o grupo do WhatsApp onde as notificações serão enviadas.</p>
        <button class="wa-btn-ghost" :disabled="loadingGroups" @click="fetchGroups">
          {{ loadingGroups ? 'Carregando…' : 'Carregar grupos' }}
        </button>
        <template v-if="groups.length">
          <input
            v-model="groupSearch"
            class="wa-search"
            placeholder="Buscar grupo…"
            type="search"
          />
          <div class="wa-group-list">
            <button
              v-for="g in filteredGroups"
              :key="g.jid"
              class="wa-group-item"
              @click="selectGroup(g.jid)"
            >
              {{ g.name }}
            </button>
            <p v-if="!filteredGroups.length" class="wa-hint">Nenhum grupo encontrado.</p>
          </div>
        </template>
      </div>

      <!-- Group linked -->
      <div v-else class="wa-section">
        <div class="wa-linked-group">
          <span class="wa-linked-label">GRUPO VINCULADO</span>
          <span class="wa-linked-name">{{ groupName || props.linkedGroup }}</span>
          <button class="wa-btn-link" @click="unlinkGroup">trocar</button>
        </div>

        <!-- Toggle de notificações automáticas -->
        <div class="wa-toggle-row">
          <span class="wa-toggle-label">Notificações automáticas</span>
          <button
            class="wa-toggle"
            :class="{ active: status.enabled }"
            :disabled="toggling"
            @click="doToggle"
          >
            <span class="wa-toggle-knob" />
          </button>
        </div>
        <p class="wa-hint" style="margin-top: -0.25rem;">
          {{ status.enabled
            ? 'Ativas — fim de jogo, faltam 10 min, partida iniciando.'
            : 'Pausadas — nenhuma mensagem será enviada.' }}
        </p>

        <!-- Botão de teste único -->
        <button class="wa-btn-ghost wa-btn-test" :disabled="sendingTest" @click="doHealthcheck">
          {{ sendingTest ? 'Enviando…' : '🔔 Enviar mensagem de teste' }}
        </button>
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
  getStatus, getQR, connect, disconnect, listGroups,
  toggleNotifications, sendHealthcheck,
} from '@/api/whatsapp'
import { setWAGroup } from '@/api/bolao'
import type { WAStatus, WAGroup } from '@/types'
import { useWAPoller } from '@/composables/useWAPoller'

const props = defineProps<{
  bolaoId: string
  linkedGroup: string | null | undefined
}>()

const emit = defineEmits<{ (e: 'group-changed'): void }>()

const status = ref<WAStatus | null>(null)
const qrImage = ref<string>('')
const groups = ref<WAGroup[]>([])
const error = ref('')
const connecting = ref(false)
const loadingGroups = ref(false)
const toggling = ref(false)
const sendingTest = ref(false)
const lastResult = ref('')
const lastResultOk = ref(true)
const groupSearch = ref('')

const filteredGroups = computed(() => {
  const q = groupSearch.value.trim().toLowerCase()
  if (!q) return groups.value
  return groups.value.filter(g => g.name.toLowerCase().includes(q))
})

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

const groupName = computed(() => {
  if (!props.linkedGroup) return ''
  return groups.value.find(g => g.jid === props.linkedGroup)?.name ?? props.linkedGroup
})

async function refresh() {
  try {
    error.value = ''
    status.value = await getStatus()
    if (status.value.state === 'awaiting_qr' && status.value.has_qr) {
      qrImage.value = await getQR()
    }
    if (status.value.state === 'connected' && !props.linkedGroup && !groups.value.length && !loadingGroups.value) {
      await fetchGroups()
    }
  } catch {
    error.value = 'Serviço WhatsApp indisponível. Suba o container com --profile whatsapp.'
  }
}

async function startConnect() {
  connecting.value = true
  error.value = ''
  try {
    await connect()
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    connecting.value = false
  }
}

async function cancelConnect() {
  try {
    await disconnect()
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function doDisconnect() {
  try {
    await disconnect()
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function fetchGroups() {
  loadingGroups.value = true
  error.value = ''
  try {
    groups.value = await listGroups()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loadingGroups.value = false
  }
}

async function selectGroup(jid: string) {
  try {
    await setWAGroup(props.bolaoId, jid)
    emit('group-changed')
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function unlinkGroup() {
  try {
    await setWAGroup(props.bolaoId, '')
    groups.value = []
    groupSearch.value = ''
    emit('group-changed')
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function doToggle() {
  if (!status.value) return
  toggling.value = true
  try {
    await toggleNotifications(!status.value.enabled)
    await refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    toggling.value = false
  }
}

async function doHealthcheck() {
  sendingTest.value = true
  lastResult.value = ''
  try {
    await sendHealthcheck()
    lastResult.value = '✓ Mensagem de teste enviada!'
    lastResultOk.value = true
  } catch (e: unknown) {
    lastResult.value = '✕ ' + (e instanceof Error ? e.message : String(e))
    lastResultOk.value = false
  } finally {
    sendingTest.value = false
  }
}

refresh()

// Poll to keep status up-to-date (QR refresh, connection state changes).
useWAPoller(refresh)
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

.wa-search {
  width: 100%;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  color: var(--text-primary);
  font-size: 0.85rem;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}
.wa-search:focus { border-color: var(--neon); }
.wa-search::placeholder { color: var(--text-muted); }

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

.wa-toggle-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.25rem;
}

.wa-toggle-label {
  font-size: 0.85rem;
  font-family: 'DM Sans', sans-serif;
  color: var(--text-primary);
  flex: 1;
}

.wa-toggle {
  position: relative;
  width: 40px;
  height: 22px;
  border-radius: 11px;
  border: none;
  background: rgba(255,255,255,0.12);
  cursor: pointer;
  transition: background 0.2s;
  flex-shrink: 0;
  padding: 0;
}
.wa-toggle.active { background: var(--neon); }
.wa-toggle:disabled { opacity: 0.5; cursor: default; }

.wa-toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: white;
  transition: transform 0.2s;
  display: block;
}
.wa-toggle.active .wa-toggle-knob { transform: translateX(18px); }

.wa-btn-test { margin-top: 0.25rem; }

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
.wa-btn-ghost:hover:not(:disabled) { border-color: rgba(255,255,255,0.3); }
.wa-btn-ghost:disabled { opacity: 0.5; cursor: default; }

.wa-disconnect { margin-top: 0.5rem; color: #ff6b6b; border-color: rgba(255,107,107,0.2); }
.wa-disconnect:hover:not(:disabled) { border-color: #ff6b6b; }

.wa-error {
  font-size: 0.78rem;
  color: #ff6b6b;
  font-family: 'DM Sans', sans-serif;
  background: rgba(255,107,107,0.08);
  border-radius: 6px;
  padding: 0.4rem 0.7rem;
}
</style>
