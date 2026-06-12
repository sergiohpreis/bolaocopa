<template>
  <Toast />
  <WhatsNewModal v-if="auth.isAuthenticated" />
  <RouterView />
</template>

<script setup lang="ts">
import Toast from 'primevue/toast'
import WhatsNewModal from '@/components/WhatsNewModal.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>
