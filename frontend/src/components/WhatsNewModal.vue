<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="overlay" @click.self="dismiss">
        <div class="modal">
          <div class="modal-header">
            <span class="version-badge font-display">v1.3.0</span>
            <button class="close-btn" @click="dismiss">×</button>
          </div>
          <h2 class="modal-title font-display">EXCLUIR BOLÃO</h2>
          <p class="modal-subtitle">O que há de novo</p>

          <p class="section-label font-display">PARA O ADMIN</p>
          <ul class="changes">
            <li class="change-item">
              <span class="change-icon">🗑️</span>
              <div>
                <div class="change-title">Exclusão com confirmação</div>
                <div class="change-desc">Na aba ADMIN, o botão "Excluir bolão" abre um resumo do impacto antes de confirmar — participantes, palpites e histórico.</div>
              </div>
            </li>
            <li class="change-item">
              <span class="change-icon">⚠️</span>
              <div>
                <div class="change-title">Ação permanente</div>
                <div class="change-desc">A exclusão remove tudo em cascata e não pode ser desfeita. Somente o admin pode fazer isso.</div>
              </div>
            </li>
          </ul>
          <button class="cta-btn font-display" @click="dismiss">ENTENDIDO</button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const STORAGE_KEY = 'whats_new_seen_v1.3.0'
const show = ref(!localStorage.getItem(STORAGE_KEY))

function dismiss() {
  localStorage.setItem(STORAGE_KEY, '1')
  show.value = false
}
</script>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.modal {
  background: #0d2010;
  border: 1px solid rgba(57, 255, 106, 0.25);
  border-radius: 16px;
  padding: 24px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 0 40px rgba(57, 255, 106, 0.1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.version-badge {
  font-size: 0.7rem;
  letter-spacing: 0.14em;
  color: var(--neon);
  background: rgba(57, 255, 106, 0.1);
  border: 1px solid rgba(57, 255, 106, 0.25);
  border-radius: 4px;
  padding: 3px 8px;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.4rem;
  cursor: pointer;
  line-height: 1;
  padding: 0 4px;
  transition: color 0.2s;
}
.close-btn:hover { color: var(--neon); }

.modal-title {
  font-size: 2rem;
  color: var(--neon);
  line-height: 1;
  margin-bottom: 4px;
}

.modal-subtitle {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-bottom: 20px;
  letter-spacing: 0.04em;
}

.section-label {
  font-size: 0.65rem;
  letter-spacing: 0.14em;
  color: var(--text-muted);
  opacity: 0.6;
  margin-bottom: 10px;
}

.changes {
  list-style: none;
  padding: 0;
  margin: 0 0 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.change-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.change-icon {
  font-size: 1.2rem;
  flex-shrink: 0;
  margin-top: 1px;
}

.change-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 2px;
}

.change-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.cta-btn {
  width: 100%;
  padding: 12px;
  background: rgba(57, 255, 106, 0.12);
  border: 1px solid rgba(57, 255, 106, 0.35);
  border-radius: 10px;
  color: var(--neon);
  font-size: 0.9rem;
  letter-spacing: 0.12em;
  cursor: pointer;
  transition: background 0.2s;
}
.cta-btn:hover { background: rgba(57, 255, 106, 0.2); }

/* Transition */
.modal-enter-active, .modal-leave-active { transition: opacity 0.25s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-active .modal, .modal-leave-active .modal { transition: transform 0.25s ease; }
.modal-enter-from .modal { transform: scale(0.95) translateY(8px); }
.modal-leave-to .modal { transform: scale(0.95) translateY(8px); }
</style>
