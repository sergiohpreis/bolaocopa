<template>
  <div class="min-h-screen flex items-center justify-center">
    <ProgressSpinner />
  </div>
</template>

<script setup lang="ts">
import ProgressSpinner from 'primevue/progressspinner'
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import http from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

onMounted(async () => {
  const code = route.query.code as string
  if (!code) {
    toast.add({ severity: 'error', summary: 'Erro', detail: 'Código OAuth ausente', life: 3000 })
    router.push({ name: 'login' })
    return
  }
  try {
    const { data } = await http.get('/auth/google/callback', { params: { code } })
    auth.setTokens(data)

    const pendingInvite = sessionStorage.getItem('pending_invite')
    if (pendingInvite) {
      sessionStorage.removeItem('pending_invite')
      router.push(`/join/${pendingInvite}`)
    } else {
      router.push('/')
    }
  } catch {
    toast.add({ severity: 'error', summary: 'Erro', detail: 'Falha no login com Google', life: 3000 })
    router.push({ name: 'login' })
  }
})
</script>
