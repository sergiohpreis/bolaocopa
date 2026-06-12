import http from './http'
import type { AuthTokens } from '@/types'

export async function getGoogleURL(): Promise<string> {
  const { data } = await http.get<{ url: string }>('/auth/google')
  return data.url
}

export async function refresh(refreshToken: string): Promise<AuthTokens> {
  const { data } = await http.post<AuthTokens>('/auth/refresh', { refresh_token: refreshToken })
  return data
}

export async function register(email: string, name: string, password: string): Promise<AuthTokens> {
  const { data } = await http.post<AuthTokens>('/auth/register', { email, name, password })
  return data
}

export async function login(email: string, password: string): Promise<AuthTokens> {
  const { data } = await http.post<AuthTokens>('/auth/login', { email, password })
  return data
}
