<template>
  <div class="page-bg min-h-screen pb-12">

    <!-- Header -->
    <div class="page-header animate-fade-up">
      <div class="header-inner">
        <button class="back-btn" @click="router.back()">←</button>
        <div>
          <h1 class="font-display neon-text" style="font-size: 1.8rem; line-height: 1;">RANKING</h1>
        </div>
      </div>
    </div>

    <div class="max-w-lg mx-auto px-4 pt-6">

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-16">
        <div class="loader-ring" />
      </div>

      <!-- Empty -->
      <div v-else-if="ranking.length === 0" class="empty-state animate-fade-up">
        <span style="font-size: 2.5rem;">🏆</span>
        <p class="font-display" style="color: var(--text-muted); font-size: 1.3rem; letter-spacing: 0.06em; margin-top: 12px;">NENHUM PALPITE AINDA</p>
      </div>

      <!-- Podium for top 3 -->
      <div v-else>
        <div v-if="ranking.length >= 2" class="podium animate-fade-up stagger-1">
          <!-- 2nd place -->
          <div class="podium-spot second" v-if="ranking[1]">
            <div class="podium-avatar">
              <img v-if="ranking[1].avatar_url" :src="ranking[1].avatar_url" class="avatar-img" :alt="ranking[1].user_name" />
              <div v-else class="avatar-fallback">{{ ranking[1].user_name?.[0]?.toUpperCase() ?? '?' }}</div>
              <div class="medal silver">2</div>
            </div>
            <div class="podium-name">{{ ranking[1].user_name }}</div>
            <div class="podium-pts silver-pts">{{ ranking[1].total_pontos }}<span style="font-size: 0.7rem; margin-left: 2px;">pts</span></div>
            <div class="podium-bar" style="height: 60px; background: rgba(176,184,193,0.2); border-color: rgba(176,184,193,0.3);" />
          </div>

          <!-- 1st place -->
          <div class="podium-spot first" v-if="ranking[0]">
            <div class="podium-avatar">
              <img v-if="ranking[0].avatar_url" :src="ranking[0].avatar_url" class="avatar-img gold-ring" :alt="ranking[0].user_name" />
              <div v-else class="avatar-fallback gold-ring">{{ ranking[0].user_name?.[0]?.toUpperCase() ?? '?' }}</div>
              <div class="medal gold">1</div>
              <div class="crown">👑</div>
            </div>
            <div class="podium-name first-name">{{ ranking[0].user_name }}</div>
            <div class="podium-pts gold-pts">{{ ranking[0].total_pontos }}<span style="font-size: 0.7rem; margin-left: 2px;">pts</span></div>
            <div class="podium-bar" style="height: 90px; background: rgba(245,200,66,0.1); border-color: rgba(245,200,66,0.3);" />
          </div>

          <!-- 3rd place -->
          <div class="podium-spot third" v-if="ranking[2]">
            <div class="podium-avatar">
              <img v-if="ranking[2].avatar_url" :src="ranking[2].avatar_url" class="avatar-img" :alt="ranking[2].user_name" />
              <div v-else class="avatar-fallback">{{ ranking[2].user_name?.[0]?.toUpperCase() ?? '?' }}</div>
              <div class="medal bronze">3</div>
            </div>
            <div class="podium-name">{{ ranking[2].user_name }}</div>
            <div class="podium-pts bronze-pts">{{ ranking[2].total_pontos }}<span style="font-size: 0.7rem; margin-left: 2px;">pts</span></div>
            <div class="podium-bar" style="height: 40px; background: rgba(205,127,50,0.15); border-color: rgba(205,127,50,0.3);" />
          </div>
        </div>

        <!-- Rest of ranking -->
        <div class="mt-6 flex flex-col gap-2">
          <div
            v-for="(entry, idx) in ranking.slice(ranking.length >= 2 ? 3 : 0)"
            :key="entry.user_id"
            class="rank-row animate-fade-up"
            :class="`stagger-${Math.min(idx + 3, 6)}`"
          >
            <div class="rank-pos">{{ (ranking.length >= 2 ? idx + 4 : idx + 1) }}</div>
            <img v-if="entry.avatar_url" :src="entry.avatar_url" class="rank-avatar" :alt="entry.user_name" />
            <div v-else class="rank-avatar-fb">{{ entry.user_name?.[0]?.toUpperCase() ?? '?' }}</div>
            <div class="flex-1 min-w-0">
              <div class="rank-name">{{ entry.user_name }}</div>
              <div class="rank-sub">{{ entry.palpites_computados }} palpites · {{ entry.palpites_computados > 0 ? Math.round(Number(entry.total_pontos) / Number(entry.palpites_computados)) : 0 }} pts/jogo</div>
            </div>
            <div class="rank-pts">
              <span class="font-display" style="font-size: 1.2rem; color: var(--text-primary);">{{ entry.total_pontos }}</span>
              <span style="font-size: 0.65rem; color: var(--text-muted); letter-spacing: 0.06em;">PTS</span>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { getRanking } from '@/api/bolao'
import type { RankingEntry } from '@/types'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const ranking = ref<RankingEntry[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    ranking.value = await getRanking(route.params.id as string)
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Erro ao carregar ranking', detail: e.message, life: 4000 })
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.page-bg { background: var(--pitch); }
.page-header {
  background: linear-gradient(180deg, rgba(10,26,14,1) 0%, rgba(10,26,14,0.85) 100%);
  border-bottom: 1px solid rgba(57,255,106,0.1);
  position: sticky;
  top: 0;
  z-index: 10;
  backdrop-filter: blur(12px);
}
.header-inner {
  max-width: 512px;
  margin: 0 auto;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.back-btn {
  font-size: 1.4rem;
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
  transition: color 0.2s;
  padding: 4px;
  line-height: 1;
}
.back-btn:hover { color: var(--neon); }
.loader-ring {
  width: 40px; height: 40px;
  border: 2px solid rgba(57,255,106,0.15);
  border-top-color: var(--neon);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.empty-state {
  display: flex; flex-direction: column; align-items: center; padding: 48px 0;
}

/* Podium */
.podium {
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 8px;
  padding: 12px 0 0;
}
.podium-spot {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  flex: 1;
  max-width: 120px;
}
.podium-spot.first { transform: translateY(-12px); }
.podium-avatar {
  position: relative;
  display: flex;
  justify-content: center;
}
.avatar-img {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(57,255,106,0.2);
}
.avatar-fallback {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: rgba(57,255,106,0.08);
  border: 2px solid rgba(57,255,106,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.2rem;
  color: var(--text-muted);
}
.first .avatar-img, .first .avatar-fallback { width: 62px; height: 62px; }
.gold-ring { border-color: var(--gold) !important; box-shadow: 0 0 16px rgba(245,200,66,0.3); }
.medal {
  position: absolute;
  bottom: -6px;
  right: -4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.7rem;
  border: 1px solid rgba(0,0,0,0.3);
}
.medal.gold { background: var(--gold); color: #1a1000; }
.medal.silver { background: var(--silver); color: #1a1a1a; }
.medal.bronze { background: var(--bronze); color: #fff; }
.crown {
  position: absolute;
  top: -22px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 1.1rem;
  animation: float-crown 3s ease-in-out infinite;
}
@keyframes float-crown {
  0%, 100% { transform: translateX(-50%) translateY(0); }
  50% { transform: translateX(-50%) translateY(-3px); }
}
.podium-name {
  font-size: 0.72rem;
  color: var(--text-muted);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}
.first-name { color: var(--text-primary); font-weight: 500; }
.podium-pts {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.1rem;
  line-height: 1;
  display: flex;
  align-items: baseline;
}
.gold-pts { color: var(--gold); }
.silver-pts { color: var(--silver); }
.bronze-pts { color: var(--bronze); }
.podium-bar {
  width: 100%;
  border: 1px solid;
  border-radius: 6px 6px 0 0;
  border-bottom: none;
}

/* Rest of ranking rows */
.rank-row {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 12px 14px;
}
.rank-pos {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 1.1rem;
  color: var(--text-muted);
  min-width: 20px;
  text-align: center;
}
.rank-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}
.rank-avatar-fb {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(57,255,106,0.06);
  border: 1px solid rgba(57,255,106,0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.9rem;
  color: var(--text-muted);
  flex-shrink: 0;
}
.rank-name {
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.rank-sub {
  font-size: 0.68rem;
  color: var(--text-muted);
  margin-top: 1px;
}
.rank-pts {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0;
}
</style>
