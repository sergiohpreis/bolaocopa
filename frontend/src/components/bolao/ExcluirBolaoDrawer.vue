<template>
  <Transition name="drawer">
    <div v-if="open" class="drawer-overlay" @click.self="$emit('close')">
      <div class="drawer">
        <div class="drawer-handle" />
        <div class="font-display drawer-title">EXCLUIR BOLÃO?</div>
        <p class="drawer-subtitle">Esta ação é permanente e não pode ser desfeita.</p>

        <div class="impact-list">
          <div class="impact-item">
            <span class="impact-icon">👥</span>
            <div class="impact-text">
              <span class="impact-count">{{ participantes }}</span>
              {{ participantes === 1 ? 'participante será removido' : 'participantes serão removidos' }}
            </div>
          </div>
          <div class="impact-item">
            <span class="impact-icon">⚽</span>
            <div class="impact-text">
              <span class="impact-count">{{ palpites }}</span>
              {{ palpites === 1 ? 'palpite será perdido' : 'palpites serão perdidos' }}
            </div>
          </div>
          <div class="impact-item">
            <span class="impact-icon">📋</span>
            <div class="impact-text">Todo o histórico de atividades será apagado</div>
          </div>
        </div>

        <div class="drawer-actions">
          <button class="btn-delete" :disabled="loading" @click="$emit('confirm')">
            <span v-if="loading" class="loader-ring" />
            <span v-else>✕ Excluir permanentemente</span>
          </button>
          <button class="btn-cancel" :disabled="loading" @click="$emit('close')">Cancelar</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
defineProps<{
  open: boolean
  participantes: number
  palpites: number
  loading: boolean
}>()
defineEmits<{
  close: []
  confirm: []
}>()
</script>

<style scoped>
.drawer-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  z-index: 50;
  display: flex; align-items: flex-end; justify-content: center;
}
.drawer {
  background: var(--pitch-mid);
  border: 1px solid rgba(255,60,60,0.2);
  border-bottom: none;
  border-radius: 16px 16px 0 0;
  padding: 12px 20px 36px;
  width: 100%;
  max-width: 512px;
}
.drawer-handle {
  width: 36px; height: 4px;
  background: rgba(255,255,255,0.15);
  border-radius: 2px;
  margin: 0 auto 16px;
}
.drawer-title {
  font-size: 1.4rem; letter-spacing: 0.08em;
  color: rgba(255,100,100,0.9);
  margin-bottom: 6px;
}
.drawer-subtitle {
  font-size: 0.82rem; color: var(--text-muted);
  margin-bottom: 20px; line-height: 1.5;
}
.impact-list { display: flex; flex-direction: column; gap: 10px; margin-bottom: 24px; }
.impact-item {
  display: flex; align-items: center; gap: 12px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 8px; padding: 10px 14px;
}
.impact-icon { font-size: 1.1rem; flex-shrink: 0; }
.impact-text { font-size: 0.82rem; color: var(--text-muted); line-height: 1.4; }
.impact-count { color: rgba(255,100,100,0.85); font-weight: 700; }
.drawer-actions { display: flex; flex-direction: column; gap: 8px; }

.btn-delete {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.9rem; letter-spacing: 0.12em;
  background: rgba(255,50,50,0.15);
  border: 1px solid rgba(255,60,60,0.5);
  border-radius: 8px;
  color: rgba(255,100,100,0.9);
  padding: 12px 20px;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
  width: 100%;
  display: flex; align-items: center; justify-content: center;
}
.btn-delete:hover:not(:disabled) {
  background: rgba(255,50,50,0.25);
  transform: translateY(-1px);
}
.btn-delete:disabled { opacity: 0.4; cursor: not-allowed; transform: none; }

.btn-cancel {
  font-family: 'DM Sans', sans-serif;
  font-size: 0.82rem;
  background: transparent;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: var(--text-muted);
  padding: 12px 18px;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
  width: 100%;
}
.btn-cancel:hover:not(:disabled) { border-color: rgba(255,255,255,0.2); color: var(--text-primary); }
.btn-cancel:disabled { opacity: 0.4; cursor: not-allowed; }

.loader-ring {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(255,100,100,0.2);
  border-top-color: rgba(255,100,100,0.8);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.drawer-enter-active, .drawer-leave-active { transition: opacity 0.2s; }
.drawer-enter-from, .drawer-leave-to { opacity: 0; }
.drawer-enter-from .drawer, .drawer-leave-to .drawer { transform: translateY(100%); }
.drawer-enter-active .drawer, .drawer-leave-active .drawer { transition: transform 0.25s ease; }
</style>
