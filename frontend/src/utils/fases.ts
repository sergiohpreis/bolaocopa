// Fonte de verdade única para as Fases do torneio.
//
// `stage` viaja cru da API football-data.org (GROUP_STAGE, LAST_32, ...).
// Aqui definimos a ordem canônica do Mata-mata e os labels em português.
//
// A Copa de 2026 tem 48 seleções: o mata-mata começa nos 16-avos (LAST_32),
// confirmado contra os dados reais da API. ROUND_OF_16 é mantido como alias
// para tolerar competições com 32 seleções e os dados de seed.

// Colunas do Mata-mata, na ordem do torneio. O índice = posição na Visão de
// Chaveamento. THIRD_PLACE e FINAL compartilham a última coluna.
export const COLUNAS_MATA_MATA = [
  ['LAST_32'],
  ['LAST_16', 'ROUND_OF_16'],
  ['QUARTER_FINALS'],
  ['SEMI_FINALS'],
  ['THIRD_PLACE', 'FINAL'],
] as const

const FASES_MATA_MATA: readonly string[] = COLUNAS_MATA_MATA.flat()

const LABELS: Record<string, string> = {
  GROUP_STAGE: 'FASE DE GRUPOS',
  LAST_32: '16-AVOS DE FINAL',
  LAST_16: 'OITAVAS DE FINAL',
  ROUND_OF_16: 'OITAVAS DE FINAL',
  QUARTER_FINALS: 'QUARTAS DE FINAL',
  SEMI_FINALS: 'SEMIFINAL',
  THIRD_PLACE: 'DISPUTA DE 3º LUGAR',
  FINAL: 'FINAL',
}

/** Label em português de uma Fase; fallback troca `_` por espaço. */
export function formatStage(stage: string): string {
  return LABELS[stage] ?? stage.replace(/_/g, ' ')
}

/** Uma Fase é do Mata-mata (eliminatória), em oposição à Fase de Grupos. */
export function isMataMata(stage: string): boolean {
  return FASES_MATA_MATA.includes(stage)
}

/**
 * Label de uma coluna da chave. A última Fase do grupo é a mais descritiva
 * (FINAL sobre THIRD_PLACE; OITAVAS sobre o alias ROUND_OF_16).
 */
export function labelColuna(stages: readonly string[]): string {
  return formatStage(stages[stages.length - 1])
}

/** Key estável para `v-for` — única porque nenhuma Fase se repete entre colunas. */
export function keyColuna(stages: readonly string[]): string {
  return stages.join('-')
}

/**
 * Um Jogo está definido quando ambos os times são conhecidos. A API cria
 * registros de Jogos do Mata-mata antes disso, com os nomes dos times vazios
 * (`home_team === ''`); esses Jogos não aceitam Palpite e aparecem como slots
 * "a definir".
 */
export function jogoDefinido(jogo: { home_team: string; away_team: string }): boolean {
  return jogo.home_team.trim() !== '' && jogo.away_team.trim() !== ''
}
