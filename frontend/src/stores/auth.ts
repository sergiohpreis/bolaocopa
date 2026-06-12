import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { AuthTokens } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(localStorage.getItem('access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refresh_token'))
  const loading = ref(false)

  const isAuthenticated = computed(() => !!accessToken.value)

  const currentUserId = computed<string | null>(() => {
    if (!accessToken.value) return null
    try {
      const base64url = accessToken.value.split('.')[1]
      const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
      const padded = base64.padEnd(base64.length + (4 - (base64.length % 4)) % 4, '=')
      const payload = JSON.parse(atob(padded))
      return payload.sub ?? null
    } catch {
      return null
    }
  })

  async function loginWithGoogle(): Promise<void> {
    loading.value = true
    try {
      const url = await authApi.getGoogleURL()
      window.location.href = url
    } finally {
      loading.value = false
    }
  }

  async function registerByEmail(email: string, name: string, password: string): Promise<void> {
    loading.value = true
    try {
      const tokens = await authApi.register(email, name, password)
      setTokens(tokens)
    } finally {
      loading.value = false
    }
  }

  async function loginByEmail(email: string, password: string): Promise<void> {
    loading.value = true
    try {
      const tokens = await authApi.login(email, password)
      setTokens(tokens)
    } finally {
      loading.value = false
    }
  }

  function setTokens(tokens: AuthTokens) {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    localStorage.setItem('access_token', tokens.access_token)
    localStorage.setItem('refresh_token', tokens.refresh_token)
  }

  function logout() {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  async function refreshAccessToken(): Promise<string | null> {
    if (!refreshToken.value) return null
    try {
      const tokens = await authApi.refresh(refreshToken.value)
      setTokens(tokens)
      return tokens.access_token
    } catch {
      logout()
      return null
    }
  }

  return { accessToken, refreshToken, isAuthenticated, currentUserId, loading, loginWithGoogle, registerByEmail, loginByEmail, setTokens, logout, refreshAccessToken }
})
