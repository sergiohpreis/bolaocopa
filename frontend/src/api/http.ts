import axios from 'axios'
import type { InternalAxiosRequestConfig } from 'axios'
import router from '@/router'

declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

const http = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

const refreshClient = axios.create({ baseURL: '/api/v1' })

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

http.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/auth/login') &&
      !originalRequest.url?.includes('/auth/register') &&
      !originalRequest.url?.includes('/auth/refresh')
    ) {
      originalRequest._retry = true
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const { data } = await refreshClient.post('/auth/refresh', { refresh_token: refreshToken })
          localStorage.setItem('access_token', data.access_token)
          localStorage.setItem('refresh_token', data.refresh_token)
          originalRequest.headers.Authorization = `Bearer ${data.access_token}`
          return http(originalRequest)
        } catch {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          router.push({ name: 'login' })
          return Promise.reject(new Error('Sessão expirada'))
        }
      }
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      router.push({ name: 'login' })
      return Promise.reject(new Error('Sessão expirada'))
    }

    const message =
      error.response?.data?.message || error.message || 'Erro desconhecido'
    return Promise.reject(new Error(message))
  },
)

export default http
