import http from './http'
import type { Bolao, Palpite, RankingEntry } from '@/types'

export async function createBolao(name: string): Promise<Bolao> {
  const { data } = await http.post<Bolao>('/boloes', { name })
  return data
}

export async function listBoloes(): Promise<Bolao[]> {
  const { data } = await http.get<Bolao[]>('/boloes')
  return data
}

export async function getBolao(id: string): Promise<Bolao> {
  const { data } = await http.get<Bolao>(`/boloes/${id}`)
  return data
}

export async function joinBolao(token: string): Promise<Bolao> {
  const { data } = await http.post<Bolao>(`/boloes/join/${token}`)
  return data
}

export async function regenerateInvite(id: string): Promise<Bolao> {
  const { data } = await http.post<Bolao>(`/boloes/${id}/regenerate-invite`)
  return data
}

export async function listPalpites(bolaoId: string): Promise<Palpite[]> {
  const { data } = await http.get<Palpite[]>(`/boloes/${bolaoId}/palpites`)
  return data
}

export async function upsertPalpite(
  bolaoId: string,
  jogoId: string,
  homeScore: number,
  awayScore: number,
): Promise<Palpite> {
  const { data } = await http.put<Palpite>(`/boloes/${bolaoId}/palpites/${jogoId}`, {
    home_score: homeScore,
    away_score: awayScore,
  })
  return data
}

export async function getRanking(bolaoId: string): Promise<RankingEntry[]> {
  const { data } = await http.get<RankingEntry[]>(`/boloes/${bolaoId}/ranking`)
  return data
}
