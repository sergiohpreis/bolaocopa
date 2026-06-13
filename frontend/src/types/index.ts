export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface User {
  id: string
  name: string
  email: string
  avatar_url?: string
}

export interface Bolao {
  id: string
  name: string
  admin_id: string
  invite_token: string
  retroativo_enabled: boolean
  created_at: string
  updated_at: string
}

export interface Jogo {
  id: string
  external_id: string
  home_team: string
  away_team: string
  home_team_flag?: string
  away_team_flag?: string
  starts_at: string
  stage: string
  home_score?: number
  away_score?: number
  finished: boolean
}

export interface Palpite {
  id: string
  bolao_id: string
  user_id: string
  jogo_id: string
  home_score: number
  away_score: number
  pontos?: number
  status: 'aprovado' | 'pendente' | 'rejeitado'
}

export interface PalpitePendente {
  id: string
  bolao_id: string
  user_id: string
  jogo_id: string
  home_score: number
  away_score: number
  pontos?: number
  status: string
  user_name: string
  home_team: string
  away_team: string
  starts_at: string
  finished: boolean
  jogo_home_score?: number
  jogo_away_score?: number
}

export interface PalpiteDeJogo {
  id: string
  bolao_id: string
  user_id: string
  jogo_id: string
  home_score: number
  away_score: number
  pontos?: number
  user_name: string
  user_avatar?: string
}

export interface FeedEvento {
  id: string
  bolao_id: string
  tipo: 'palpite_registrado' | 'palpite_alterado' | 'participante_entrou' | 'jogo_iniciado' | 'resultado_apurado' | 'palpite_removido'
  user_id?: string
  user_name?: string
  jogo_id?: string
  jogo_desc?: string
  payload: Record<string, any>
  created_at: string
}

export interface RankingEntry {
  user_id: string
  user_name: string
  avatar_url?: string
  total_pontos: number
  palpites_computados: number
}
