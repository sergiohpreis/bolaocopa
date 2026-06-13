<template>
  <div class="feed-panel">
    <div v-if="loading && eventos.length === 0" class="feed-loading">
      <div class="loader-ring-sm" />
    </div>

    <div v-else-if="eventos.length === 0" class="feed-empty">
      <span>Nenhuma atividade ainda.</span>
    </div>

    <div v-else class="feed-list">
      <transition-group name="feed-item" tag="div">
        <div
          v-for="ev in eventos"
          :key="ev.id"
          class="feed-item"
          :class="`feed-item--${ev.tipo}`"
        >
          <div class="feed-icon">{{ iconFor(ev) }}</div>
          <div class="feed-content">
            <div class="feed-text">{{ descFor(ev) }}</div>
            <div class="feed-time">{{ timeAgo(ev.created_at) }}</div>
          </div>
        </div>
      </transition-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { getFeed } from '@/api/bolao'
import { traduzTime } from '@/utils/teams'
import type { FeedEvento } from '@/types'

const props = defineProps<{ bolaoId: string; active?: boolean }>()

const eventos = ref<FeedEvento[]>([])
const loading = ref(true)
const error = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

async function fetch() {
  try {
    eventos.value = await getFeed(props.bolaoId)
    error.value = false
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

// Fetch immediately when tab becomes active
watch(() => props.active, (active) => {
  if (active) fetch()
})

onMounted(() => {
  fetch()
  timer = setInterval(fetch, 15_000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function iconFor(ev: FeedEvento): string {
  if (ev.tipo === 'palpite_registrado' && ev.payload.retroativo) return '⏮'
  const icons: Record<FeedEvento['tipo'], string> = {
    palpite_registrado: '🎯',
    palpite_alterado: '✏️',
    participante_entrou: '👤',
    jogo_iniciado: '⚽',
    resultado_apurado: '📊',
    palpite_removido: '🚫',
  }
  return icons[ev.tipo] ?? '•'
}

function descFor(ev: FeedEvento): string {
  const name = ev.user_name ?? 'Alguém'
  const jogo = ev.jogo_desc
    ? ev.jogo_desc.split(' x ').map(traduzTime).join(' x ')
    : 'um jogo'

  switch (ev.tipo) {
    case 'palpite_registrado':
      if (ev.payload.home_score !== undefined) {
        const placar = `${ev.payload.home_score}×${ev.payload.away_score}`
        if (ev.payload.retroativo) {
          return `${name} teve palpite retroativo aprovado: ${placar} em ${jogo}`
        }
        return `${name} apostou ${placar} em ${jogo}`
      }
      return `${name} registrou palpite em ${jogo}`
    case 'palpite_alterado':
      if (ev.payload.home_score !== undefined) {
        return `${name} alterou para ${ev.payload.home_score}×${ev.payload.away_score} em ${jogo}`
      }
      return `${name} alterou palpite em ${jogo}`
    case 'participante_entrou':
      return `${name} entrou no bolão`
    case 'jogo_iniciado':
      return `${jogo} começou!`
    case 'resultado_apurado':
      if (ev.payload.home_score !== undefined) {
        return `Resultado de ${jogo}: ${ev.payload.home_score}×${ev.payload.away_score}`
      }
      return `Resultado apurado em ${jogo}`
    case 'palpite_removido':
      if (ev.payload.home_score !== undefined) {
        return `Admin removeu o palpite retroativo de ${name}: ${ev.payload.home_score}×${ev.payload.away_score} em ${jogo}`
      }
      return `Admin removeu palpite retroativo de ${name} em ${jogo}`
    default:
      return ''
  }
}

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (diff < 60) return 'agora'
  if (diff < 3600) return `${Math.floor(diff / 60)}min atrás`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h atrás`
  return `${Math.floor(diff / 86400)}d atrás`
}
</script>

<style scoped>
.feed-panel {
  background: rgba(57, 255, 106, 0.03);
  border: 1px solid rgba(57, 255, 106, 0.1);
  border-radius: 12px;
  overflow: hidden;
  margin-top: 16px;
}

.feed-loading {
  display: flex;
  justify-content: center;
  padding: 28px;
}

.loader-ring-sm {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(57, 255, 106, 0.15);
  border-top-color: var(--neon);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.feed-empty {
  text-align: center;
  padding: 28px 16px;
  font-size: 0.82rem;
  color: var(--text-muted);
}

.feed-list {
  max-height: 400px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(57, 255, 106, 0.15) transparent;
}

.feed-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  transition: background 0.2s;
}
.feed-item:last-child { border-bottom: none; }
.feed-item:hover { background: rgba(57, 255, 106, 0.03); }

.feed-icon {
  font-size: 1rem;
  line-height: 1.4;
  flex-shrink: 0;
}

.feed-content { flex: 1; min-width: 0; }

.feed-text {
  font-size: 0.82rem;
  color: var(--text-secondary, rgba(255,255,255,0.75));
  line-height: 1.4;
}

.feed-time {
  font-size: 0.7rem;
  color: var(--text-muted);
  margin-top: 2px;
}

/* Accent per event type */
.feed-item--resultado_apurado .feed-text { color: var(--neon); }
.feed-item--jogo_iniciado .feed-text { color: rgba(255, 210, 80, 0.9); }
.feed-item--participante_entrou .feed-text { color: rgba(80, 180, 255, 0.9); }

/* Transition */
.feed-item-enter-active { transition: all 0.4s ease; }
.feed-item-enter-from { opacity: 0; transform: translateY(-8px); }
</style>
