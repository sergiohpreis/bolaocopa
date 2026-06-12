<template>
  <div class="login-bg min-h-screen flex items-center justify-center px-4 py-12">
    <div class="pitch-lines" aria-hidden="true" />

    <div class="w-full max-w-sm animate-fade-up">
      <!-- Header -->
      <div class="flex flex-col items-center mb-8">
        <div class="invite-badge">
          <span style="font-size: 2.2rem;">🏆</span>
        </div>
        <h1 class="font-display neon-text mt-4" style="font-size: 2.2rem; line-height:1;">CONVITE ESPECIAL</h1>
        <p class="text-sm mt-1 text-center" style="color: var(--text-muted);">Você foi convidado para um bolão</p>
      </div>

      <div class="card rounded-2xl p-6">
        <div class="tab-row mb-6">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-btn"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key; error = ''"
          >{{ tab.label }}</button>
        </div>

        <!-- Error global -->
        <div v-if="error && activeTab === 'google'" class="error-msg mb-4">{{ error }}</div>

        <!-- Google -->
        <div v-if="activeTab === 'google'" class="flex flex-col gap-4 animate-fade-in">
          <p class="text-sm text-center" style="color: var(--text-muted); line-height: 1.6;">
            Entre com o Google e entre automaticamente no bolão.
          </p>
          <button class="google-btn" :disabled="loading" @click="handleGoogle">
            <span class="google-icon">
              <svg width="18" height="18" viewBox="0 0 18 18"><path fill="#4285F4" d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.874 2.684-6.615z"/><path fill="#34A853" d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z"/><path fill="#FBBC05" d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332z"/><path fill="#EA4335" d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 6.29C4.672 4.163 6.656 3.58 9 3.58z"/></svg>
            </span>
            <span class="font-display" style="font-size: 1rem; letter-spacing: 0.06em;">
              {{ loading ? 'AGUARDE...' : 'ENTRAR COM GOOGLE' }}
            </span>
          </button>
        </div>

        <!-- Login -->
        <form v-else-if="activeTab === 'login'" class="flex flex-col gap-3 animate-fade-in" @submit.prevent="handleLogin">
          <div class="field-group">
            <label class="field-label">EMAIL</label>
            <input v-model="email" type="email" class="field-input" placeholder="seu@email.com" required />
          </div>
          <div class="field-group">
            <label class="field-label">SENHA</label>
            <div class="relative">
              <input v-model="password" :type="showPass ? 'text' : 'password'" class="field-input pr-10" placeholder="••••••••" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">{{ showPass ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <div v-if="error" class="error-msg">{{ error }}</div>
          <button type="submit" class="submit-btn mt-1" :disabled="loading">
            <span class="font-display" style="font-size: 1rem; letter-spacing: 0.08em;">
              {{ loading ? 'ENTRANDO...' : 'ENTRAR E PARTICIPAR' }}
            </span>
          </button>
          <p class="text-center text-xs mt-1" style="color: var(--text-muted);">
            Não tem conta?
            <button type="button" class="link-btn" @click="activeTab = 'register'; error = ''">Cadastrar</button>
          </p>
        </form>

        <!-- Register -->
        <form v-else class="flex flex-col gap-3 animate-fade-in" @submit.prevent="handleRegister">
          <div class="field-group">
            <label class="field-label">NOME</label>
            <input v-model="name" type="text" class="field-input" placeholder="Seu nome" required />
          </div>
          <div class="field-group">
            <label class="field-label">EMAIL</label>
            <input v-model="email" type="email" class="field-input" placeholder="seu@email.com" required />
          </div>
          <div class="field-group">
            <label class="field-label">SENHA</label>
            <div class="relative">
              <input v-model="password" :type="showPass ? 'text' : 'password'" class="field-input pr-10" placeholder="Mín. 8 caracteres" required />
              <button type="button" class="pass-toggle" @click="showPass = !showPass">{{ showPass ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <div v-if="error" class="error-msg">{{ error }}</div>
          <button type="submit" class="submit-btn mt-1" :disabled="loading">
            <span class="font-display" style="font-size: 1rem; letter-spacing: 0.08em;">
              {{ loading ? 'CRIANDO...' : 'CRIAR CONTA E PARTICIPAR' }}
            </span>
          </button>
          <p class="text-center text-xs mt-1" style="color: var(--text-muted);">
            Já tem conta?
            <button type="button" class="link-btn" @click="activeTab = 'login'; error = ''">Entrar</button>
          </p>
        </form>
      </div>
    </div>
    <p class="text-center mt-6" style="font-size: 0.75rem; color: var(--text-muted);">
      <router-link to="/como-funciona" style="color: var(--text-muted); text-decoration: underline; text-underline-offset: 3px; opacity: 0.7; transition: opacity 0.2s;" onmouseover="this.style.opacity='1'" onmouseout="this.style.opacity='0.7'">
        Como funciona o bolão?
      </router-link>
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { joinBolao } from '@/api/bolao'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)

const tabs: { key: 'google' | 'login' | 'register'; label: string }[] = [
  { key: 'google', label: 'Google' },
  { key: 'login', label: 'Entrar' },
  { key: 'register', label: 'Cadastrar' },
]
const activeTab = ref<'google' | 'login' | 'register'>('google')

const email = ref('')
const name = ref('')
const password = ref('')
const error = ref('')
const showPass = ref(false)

onMounted(async () => {
  if (auth.isAuthenticated) {
    await tryJoin()
  }
})

async function tryJoin() {
  loading.value = true
  try {
    const bolao = await joinBolao(route.params.token as string)
    router.push(`/boloes/${bolao.id}`)
  } catch (e: any) {
    error.value = e?.message ?? 'Link de convite inválido ou expirado'
  } finally {
    loading.value = false
  }
}

async function handleGoogle() {
  sessionStorage.setItem('pending_invite', route.params.token as string)
  await auth.loginWithGoogle()
}

async function handleLogin() {
  error.value = ''
  try {
    await auth.loginByEmail(email.value, password.value)
    await tryJoin()
  } catch (e: any) {
    error.value = e?.message ?? 'Erro ao entrar'
  }
}

async function handleRegister() {
  error.value = ''
  if (password.value.length < 8) {
    error.value = 'Senha deve ter no mínimo 8 caracteres'
    return
  }
  try {
    await auth.registerByEmail(email.value, name.value, password.value)
    await tryJoin()
  } catch (e: any) {
    error.value = e?.message ?? 'Erro ao cadastrar'
  }
}
</script>

<style scoped>
.login-bg {
  background: var(--pitch);
  position: relative;
  overflow: hidden;
}
.pitch-lines {
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(ellipse 70% 50% at 50% 110%, rgba(245,200,66,0.05) 0%, transparent 60%),
    radial-gradient(ellipse 40% 40% at 80% -10%, rgba(57,255,106,0.04) 0%, transparent 60%);
}
.invite-badge {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: conic-gradient(from 0deg, var(--gold) 0%, rgba(245,200,66,0.3) 50%, var(--gold) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse-neon 3s ease-in-out infinite;
}
.tab-row {
  display: flex;
  background: rgba(0,0,0,0.3);
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
  border: 1px solid rgba(57,255,106,0.1);
}
.tab-btn {
  flex: 1;
  padding: 8px 4px;
  border-radius: 8px;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.9rem;
  letter-spacing: 0.07em;
  color: var(--text-muted);
  transition: all 0.2s;
  cursor: pointer;
  border: none;
  background: transparent;
}
.tab-btn.active {
  background: var(--neon);
  color: var(--pitch);
}
.tab-btn:not(.active):hover { color: var(--text-primary); }
.field-group { display: flex; flex-direction: column; gap: 5px; }
.field-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.75rem;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}
.field-input {
  width: 100%;
  padding: 10px 14px;
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(57,255,106,0.18);
  border-radius: 8px;
  color: var(--text-primary);
  font-family: 'DM Sans', sans-serif;
  font-size: 0.9rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  outline: none;
}
.field-input::placeholder { color: var(--text-muted); }
.field-input:focus {
  border-color: var(--neon);
  box-shadow: 0 0 0 3px rgba(57,255,106,0.12);
}
.pass-toggle {
  position: absolute; right: 10px; top: 50%; transform: translateY(-50%);
  font-size: 0.9rem; cursor: pointer; opacity: 0.6; transition: opacity 0.15s;
  border: none; background: transparent;
}
.pass-toggle:hover { opacity: 1; }
.submit-btn {
  width: 100%; padding: 12px; background: var(--neon); color: var(--pitch);
  border: none; border-radius: 8px; cursor: pointer;
  transition: box-shadow 0.2s, transform 0.1s, opacity 0.2s;
  display: flex; align-items: center; justify-content: center;
}
.submit-btn:hover:not(:disabled) {
  box-shadow: 0 0 24px rgba(57,255,106,0.4);
  transform: translateY(-1px);
}
.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.google-btn {
  width: 100%; padding: 12px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 8px; color: var(--text-primary); cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 10px;
  transition: background 0.2s, border-color 0.2s;
}
.google-btn:hover:not(:disabled) {
  background: rgba(255,255,255,0.1);
  border-color: rgba(255,255,255,0.25);
}
.google-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.google-icon { display: flex; align-items: center; }
.error-msg {
  background: rgba(255, 80, 80, 0.12);
  border: 1px solid rgba(255, 80, 80, 0.3);
  border-radius: 8px;
  padding: 8px 12px;
  color: #ff8080;
  font-size: 0.82rem;
}
.link-btn {
  color: var(--neon); font-weight: 500; cursor: pointer;
  border: none; background: none; padding: 0; font-size: inherit;
}
.link-btn:hover { text-decoration: underline; }
</style>
