// Calls the WhatsApp service via the Go backend proxy (JWT-authenticated).
// The API secret is never exposed to the browser.
import http from './http'
import type { WAStatus, WAGroup } from '@/types'

export async function getStatus(): Promise<WAStatus> {
  const { data } = await http.get<WAStatus>('/whatsapp/status')
  return data
}

export async function getQR(): Promise<string> {
  const { data } = await http.get<{ qr_base64: string }>('/whatsapp/qr')
  return data.qr_base64
}

export async function connect(): Promise<void> {
  await http.post('/whatsapp/connect')
}

export async function disconnect(): Promise<void> {
  await http.delete('/whatsapp/connect')
}

export async function listGroups(): Promise<WAGroup[]> {
  const { data } = await http.get<WAGroup[]>('/whatsapp/groups')
  return data
}

export async function toggleNotifications(enabled: boolean): Promise<void> {
  await http.post('/whatsapp/toggle', { enabled })
}

export async function sendHealthcheck(targetJid?: string): Promise<void> {
  await http.post('/whatsapp/healthcheck', targetJid ? { target_jid: targetJid } : {})
}
