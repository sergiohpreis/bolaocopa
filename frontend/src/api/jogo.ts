import http from './http'
import type { Jogo } from '@/types'

export async function listJogos(): Promise<Jogo[]> {
  const { data } = await http.get<Jogo[]>('/jogos')
  return data
}

export async function syncJogos(): Promise<void> {
  await http.post('/jogos/sync')
}
