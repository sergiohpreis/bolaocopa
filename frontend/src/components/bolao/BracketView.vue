<template>
  <div>
    <div v-for="coluna in colunas" :key="coluna.key" class="stage-group">
      <div class="stage-header">
        <div class="stage-line" />
        <span class="stage-label">{{ coluna.label }}</span>
        <div class="stage-line" />
      </div>
      <div class="flex flex-col gap-2">
        <JogoCard
          v-for="jogo in coluna.jogos"
          :key="jogo.id"
          :jogo="jogo"
          :palpite="palpiteMap[jogo.id]"
          :bolao-id="bolaoId"
          :retroativo-enabled="retroativoEnabled"
          @save="(h, a) => emit('save', jogo.id, h, a)"
          @save-retroativo="(h, a) => emit('saveRetroativo', jogo.id, h, a)"
        />
        <div v-if="coluna.vagas > 0" class="bracket-tbd">
          {{ coluna.vagas }} {{ coluna.vagas === 1 ? 'CONFRONTO' : 'CONFRONTOS' }} A DEFINIR
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import JogoCard from '@/components/bolao/JogoCard.vue'
import type { Jogo, Palpite } from '@/types'

export interface ColunaChave {
  key: string
  label: string
  jogos: Jogo[]
  vagas: number
}

defineProps<{
  colunas: ColunaChave[]
  palpiteMap: Record<string, Palpite>
  bolaoId: string
  retroativoEnabled: boolean
}>()

const emit = defineEmits<{
  (e: 'save', jogoId: string, home: number, away: number): void
  (e: 'saveRetroativo', jogoId: string, home: number, away: number): void
}>()
</script>

<style scoped>
.stage-group { margin-bottom: 28px; }
.stage-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.stage-line {
  flex: 1;
  height: 1px;
  background: rgba(57,255,106,0.12);
}
.stage-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.78rem;
  letter-spacing: 0.14em;
  color: var(--text-muted);
  white-space: nowrap;
}

/* Slot "a definir" — confronto da fase ainda sem times decididos */
.bracket-tbd {
  padding: 18px;
  border: 1px dashed rgba(57,255,106,0.18);
  border-radius: 12px;
  text-align: center;
  color: var(--text-muted);
  font-family: 'Bebas Neue', sans-serif;
  letter-spacing: 0.16em;
  font-size: 0.9rem;
  opacity: 0.7;
}
</style>
