<template>
  <Toast />
  <WhatsNewModal />
  <RouterView />
</template>

<script setup lang="ts">
import Toast from 'primevue/toast'
import WhatsNewModal from '@/components/WhatsNewModal.vue'
</script>
