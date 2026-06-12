<template>
  <div class="page-bg min-h-screen">
    <div class="max-w-lg mx-auto px-4 py-8">

      <!-- Header -->
      <div class="page-header animate-fade-up">
        <div>
          <h1 class="font-display neon-text" style="font-size: 3rem; line-height: 1;">MEUS BOLÕES</h1>
          <p class="text-xs mt-0.5" style="color: var(--text-muted); letter-spacing: 0.08em;">COPA DO MUNDO</p>
        </div>
        <button class="new-btn" @click="router.push('/boloes/novo')">
          <span style="font-size: 1.1rem;">+</span>
          <span class="font-display" style="letter-spacing: 0.08em; font-size: 0.9rem;">NOVO</span>
        </button>
      </div>

      <!-- Divider -->
      <div class="neon-divider animate-fade-up stagger-1" />

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-16 animate-fade-in">
        <div class="loader-ring" />
      </div>

      <!-- Empty -->
      <div v-else-if="boloes.length === 0" class="empty-state animate-fade-up stagger-2">
        <div class="empty-icon">⚽</div>
        <p class="font-display" style="font-size: 1.4rem; color: var(--text-muted); letter-spacing: 0.06em;">NENHUM BOLÃO AINDA</p>
        <p class="text-sm mt-1" style="color: var(--text-muted); opacity: 0.6;">Crie um ou entre via link de convite</p>
      </div>

      <!-- List -->
      <div v-else class="flex flex-col gap-3 mt-2">
        <div
          v-for="(bolao, idx) in boloes"
          :key="bolao.id"
          class="bolao-card animate-fade-up"
          :class="`stagger-${Math.min(idx + 2, 6)}`"
          @click="router.push(`/boloes/${bolao.id}`)"
        >
          <div class="bolao-accent" />
          <div class="bolao-content">
            <div>
              <div class="font-display bolao-name">{{ bolao.name }}</div>
              <div class="bolao-date">Desde {{ formatDate(bolao.created_at) }}</div>
            </div>
            <div class="bolao-arrow">›</div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listBoloes } from '@/api/bolao'
import type { Bolao } from '@/types'

const router = useRouter()
const boloes = ref<Bolao[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    boloes.value = await listBoloes()
  } finally {
    loading.value = false
  }
})

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('pt-BR')
}
</script>

<style scoped>
.page-bg {
  background: var(--pitch);
}
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  padding-bottom: 16px;
}
.neon-divider {
  height: 1px;
  background: linear-gradient(90deg, var(--neon) 0%, rgba(57,255,106,0.2) 50%, transparent 100%);
  margin-bottom: 24px;
}
.new-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--neon);
  color: var(--pitch);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-family: 'Bebas Neue', sans-serif;
  transition: box-shadow 0.2s, transform 0.1s;
  margin-bottom: 4px;
}
.new-btn:hover {
  box-shadow: 0 0 20px rgba(57,255,106,0.4);
  transform: translateY(-1px);
}

.loader-ring {
  width: 40px;
  height: 40px;
  border: 2px solid rgba(57,255,106,0.15);
  border-top-color: var(--neon);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  gap: 8px;
}
.empty-icon {
  font-size: 3rem;
  margin-bottom: 8px;
  filter: grayscale(0.3);
}

.bolao-card {
  position: relative;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.15s;
}
.bolao-card:hover {
  border-color: rgba(57,255,106,0.35);
  box-shadow: 0 0 24px rgba(57,255,106,0.1);
  transform: translateX(4px);
}
.bolao-accent {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, var(--neon), rgba(57,255,106,0.3));
}
.bolao-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
}
.bolao-name {
  font-size: 1.3rem;
  letter-spacing: 0.05em;
  color: var(--text-primary);
}
.bolao-date {
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-top: 2px;
  letter-spacing: 0.06em;
}
.bolao-arrow {
  font-size: 1.6rem;
  color: var(--neon);
  opacity: 0.5;
  transition: opacity 0.2s, transform 0.2s;
}
.bolao-card:hover .bolao-arrow {
  opacity: 1;
  transform: translateX(4px);
}
</style>
