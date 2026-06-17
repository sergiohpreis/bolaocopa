// PROTOTYPE — throwaway. Calls the whatsapp service directly from the browser.
// TODO(M1): Before production, remove this file and route all calls through the
// Go backend (JWT-authenticated). The browser must never hold the API secret.
// The VITE_WHATSAPP_API_SECRET env var is inlined into the JS bundle at build
// time and is visible to any user via devtools. Remove before promoting to prod.
import axios from 'axios'
import type { WAStatus, WAGroup, WANotifyPayload } from '@/types'

const SECRET = import.meta.env.VITE_WHATSAPP_API_SECRET ?? 'prototype-secret'

const wa = axios.create({
  baseURL: '/whatsapp',
  headers: { 'X-API-Secret': SECRET },
})

export async function getStatus(): Promise<WAStatus> {
  const { data } = await wa.get<WAStatus>('/status')
  return data
}

export async function getQR(): Promise<string> {
  const { data } = await wa.get<{ qr_base64: string }>('/qr')
  return data.qr_base64
}

export async function connect(): Promise<void> {
  await wa.post('/connect')
}

export async function disconnect(): Promise<void> {
  await wa.delete('/connect')
}

export async function listGroups(): Promise<WAGroup[]> {
  const { data } = await wa.get<WAGroup[]>('/groups')
  return data
}

export async function linkGroup(jid: string): Promise<void> {
  await wa.post('/link', { jid })
}

export async function sendNotification(payload: WANotifyPayload): Promise<void> {
  await wa.post('/notify', payload)
}
