import http from './http'
import type { Bolao, Palpite, PalpitePendente, PalpiteDeJogo, RankingEntry, FeedEvento, TaxaEstado } from '@/types'

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

export async function upsertPalpiteRetroativo(
  bolaoId: string,
  jogoId: string,
  homeScore: number,
  awayScore: number,
): Promise<Palpite> {
  const { data } = await http.put<Palpite>(`/boloes/${bolaoId}/palpites/${jogoId}/retroativo`, {
    home_score: homeScore,
    away_score: awayScore,
  })
  return data
}

export async function listPalpitesPendentes(bolaoId: string): Promise<PalpitePendente[]> {
  const { data } = await http.get<PalpitePendente[]>(`/boloes/${bolaoId}/palpites/pendentes`)
  return data
}

export async function aprovarPalpite(bolaoId: string, palpiteId: string): Promise<Palpite> {
  const { data } = await http.post<Palpite>(`/boloes/${bolaoId}/palpites/${palpiteId}/aprovar`)
  return data
}

export async function rejeitarPalpite(bolaoId: string, palpiteId: string): Promise<Palpite> {
  const { data } = await http.post<Palpite>(`/boloes/${bolaoId}/palpites/${palpiteId}/rejeitar`)
  return data
}

export async function listPalpitesRetroativosAprovados(bolaoId: string): Promise<PalpitePendente[]> {
  const { data } = await http.get<PalpitePendente[]>(`/boloes/${bolaoId}/palpites/retroativos`)
  return data
}

export async function desaprovarPalpite(bolaoId: string, palpiteId: string): Promise<void> {
  await http.delete(`/boloes/${bolaoId}/palpites/${palpiteId}`)
}

export async function setRetroativoEnabled(bolaoId: string, enabled: boolean): Promise<Bolao> {
  const { data } = await http.patch<Bolao>(`/boloes/${bolaoId}/settings`, { retroativo_enabled: enabled })
  return data
}

export async function getTaxaEstado(bolaoId: string): Promise<TaxaEstado> {
  const { data } = await http.get<TaxaEstado>(`/boloes/${bolaoId}/taxa`)
  return data
}

export async function proporTaxa(bolaoId: string, valor: string): Promise<void> {
  await http.post(`/boloes/${bolaoId}/taxa/proposta`, { valor })
}

export async function votarTaxa(bolaoId: string, aprovado: boolean): Promise<void> {
  await http.post(`/boloes/${bolaoId}/taxa/votar`, { aprovado })
}

export async function getRanking(bolaoId: string): Promise<RankingEntry[]> {
  const { data } = await http.get<RankingEntry[]>(`/boloes/${bolaoId}/ranking`)
  return data
}

export async function getPalpitesByJogo(bolaoId: string, jogoId: string): Promise<PalpiteDeJogo[]> {
  const { data } = await http.get<PalpiteDeJogo[]>(`/boloes/${bolaoId}/palpites/${jogoId}`)
  return data
}

export async function getFeed(bolaoId: string): Promise<FeedEvento[]> {
  const { data } = await http.get<FeedEvento[]>(`/boloes/${bolaoId}/feed`)
  return data
}

export async function deleteBolao(id: string): Promise<void> {
  await http.delete(`/boloes/${id}`)
}

export async function setWAGroup(bolaoId: string, jid: string): Promise<Bolao> {
  const { data } = await http.put<Bolao>(`/boloes/${bolaoId}/whatsapp-group`, { jid })
  return data
}
