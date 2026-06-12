<template>
  <div class="min-h-screen flex items-center justify-center px-4" style="background: var(--pitch);">
    <div class="w-full max-w-sm animate-fade-up">

      <div class="mb-8">
        <button class="back-btn" @click="router.back()">
          <span>←</span> <span class="font-display" style="letter-spacing: 0.08em; font-size: 0.85rem;">VOLTAR</span>
        </button>
        <h1 class="font-display neon-text mt-6" style="font-size: 2.8rem; line-height: 1;">NOVO BOLÃO</h1>
        <p class="text-sm mt-1" style="color: var(--text-muted);">Crie seu bolão e convide amigos</p>
      </div>

      <div class="card rounded-2xl p-6">
        <div class="field-group">
          <label class="field-label">NOME DO BOLÃO</label>
          <input
            v-model="name"
            type="text"
            class="field-input"
            placeholder="Ex: Bolão do Trabalho"
            required
            autofocus
            @keydown.enter="create"
          />
        </div>

        <button
          class="submit-btn mt-5"
          :disabled="!name.trim() || loading"
          @click="create"
        >
          <span class="font-display" style="font-size: 1rem; letter-spacing: 0.08em;">
            {{ loading ? 'CRIANDO...' : 'CRIAR BOLÃO' }}
          </span>
        </button>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { createBolao } from '@/api/bolao'

const router = useRouter()
const name = ref('')
const loading = ref(false)

async function create() {
  if (!name.value.trim() || loading.value) return
  loading.value = true
  try {
    const bolao = await createBolao(name.value.trim())
    router.push(`/boloes/${bolao.id}`)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
  font-family: 'DM Sans', sans-serif;
  font-size: 0.85rem;
  border: none;
  background: none;
  cursor: pointer;
  padding: 0;
  transition: color 0.2s;
}
.back-btn:hover { color: var(--text-primary); }
.field-group { display: flex; flex-direction: column; gap: 5px; }
.field-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.75rem;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}
.field-input {
  width: 100%;
  padding: 12px 14px;
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(57,255,106,0.18);
  border-radius: 8px;
  color: var(--text-primary);
  font-family: 'DM Sans', sans-serif;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  outline: none;
}
.field-input::placeholder { color: var(--text-muted); }
.field-input:focus {
  border-color: var(--neon);
  box-shadow: 0 0 0 3px rgba(57,255,106,0.12);
}
.submit-btn {
  width: 100%;
  padding: 13px;
  background: var(--neon);
  color: var(--pitch);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.1s, opacity 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.submit-btn:hover:not(:disabled) {
  box-shadow: 0 0 24px rgba(57,255,106,0.4);
  transform: translateY(-1px);
}
.submit-btn:disabled { opacity: 0.35; cursor: not-allowed; }
</style>
