<template>
  <div class="page-bg min-h-screen pb-16">

    <!-- Header -->
    <div class="page-header">
      <div class="header-inner">
        <button class="back-btn" @click="router.back()">←</button>
        <h1 class="font-display page-title neon-text">COMO FUNCIONA</h1>
      </div>
    </div>

    <div class="content max-w-lg mx-auto px-4">

      <!-- Intro -->
      <div class="intro-card animate-fade-up">
        <p class="intro-text">
          Bolão da Copa é simples: cada participante faz seus palpites antes de cada jogo.
          Quanto mais preciso, mais pontos.
        </p>
      </div>

      <!-- Pontuação -->
      <section class="section animate-fade-up stagger-1">
        <h2 class="section-title">PONTUAÇÃO</h2>
        <div class="rules-grid">
          <div class="rule-card rule--exact">
            <div class="rule-points">10</div>
            <div class="rule-label">PONTOS</div>
            <div class="rule-desc">Placar exato</div>
            <div class="rule-example">Ex: apostou 2×1, acabou 2×1</div>
          </div>
          <div class="rule-card rule--winner">
            <div class="rule-points">3</div>
            <div class="rule-label">PONTOS</div>
            <div class="rule-desc">Acertou o vencedor ou empate</div>
            <div class="rule-example">Ex: apostou 2×1, acabou 3×0</div>
          </div>
          <div class="rule-card rule--miss">
            <div class="rule-points">0</div>
            <div class="rule-label">PONTOS</div>
            <div class="rule-desc">Errou o resultado</div>
            <div class="rule-example">Ex: apostou vitória, acabou empate</div>
          </div>
        </div>
      </section>

      <!-- Mata-mata -->
      <section class="section animate-fade-up stagger-2">
        <h2 class="section-title">MATA-MATA — PONTUAÇÃO POR FASE</h2>
        <div class="info-card" style="margin-bottom: 12px;">
          <div class="info-icon">⚡</div>
          <div>
            <p class="info-text">No mata-mata, os pontos base da fase de grupos são multiplicados conforme a fase. Quanto mais longe no torneio, mais vale cada acerto.</p>
          </div>
        </div>
        <div class="knockout-table">
          <div class="kt-row kt-header">
            <span>Fase</span>
            <span class="text-center">Mult.</span>
            <span class="text-center">Vencedor</span>
            <span class="text-center">Placar exato</span>
          </div>
          <div class="kt-row">
            <span>16-avos</span>
            <span class="text-center text-muted">1.5×</span>
            <span class="text-center">4.5</span>
            <span class="text-center">15</span>
          </div>
          <div class="kt-row">
            <span>Oitavas</span>
            <span class="text-center text-muted">2×</span>
            <span class="text-center">6</span>
            <span class="text-center">20</span>
          </div>
          <div class="kt-row">
            <span>Quartas</span>
            <span class="text-center text-muted">2.5×</span>
            <span class="text-center">7.5</span>
            <span class="text-center">25</span>
          </div>
          <div class="kt-row">
            <span>Semis</span>
            <span class="text-center text-muted">3×</span>
            <span class="text-center">9</span>
            <span class="text-center">30</span>
          </div>
          <div class="kt-row kt-final">
            <span>Final / 3º lugar</span>
            <span class="text-center text-muted">3.5×</span>
            <span class="text-center">10.5</span>
            <span class="text-center">35</span>
          </div>
        </div>
        <div class="info-card" style="margin-top: 12px;">
          <div class="info-icon">🥅</div>
          <div>
            <p class="info-text">O palpite é sempre o <strong>placar do tempo normal (90')</strong>. Em jogos que vão para os pênaltis, ao apostar empate você deve indicar qual time avança:</p>
            <ul style="margin-top: 8px; padding-left: 16px; font-size: 0.82rem; color: rgba(255,255,255,0.65); line-height: 1.8;">
              <li>Acertou o placar exato <strong>+ indicou o classificado correto</strong> → placar exato × fase <strong>+ 3 bônus</strong></li>
              <li>Acertou o placar exato <strong>+ indicou o time errado</strong> → pontos de "vencedor" × fase</li>
              <li>Apostou empate (qualquer placar) <strong>+ indicou o classificado correto</strong> → pontos de "vencedor" × fase</li>
              <li>Apostou empate (qualquer placar) <strong>sem acertar o classificado</strong> → pontos de "vencedor" × fase</li>
            </ul>
          </div>
        </div>
      </section>

      <!-- Prazo -->
      <section class="section animate-fade-up stagger-3">
        <h2 class="section-title">PRAZO PARA APOSTAR</h2>
        <div class="info-card">
          <div class="info-icon">⏰</div>
          <div>
            <p class="info-text">Você pode registrar ou alterar seu palpite a qualquer momento <strong>até o apito inicial do jogo</strong>. Depois que o jogo começa, o palpite é bloqueado.</p>
          </div>
        </div>
      </section>

      <!-- Como entrar -->
      <section class="section animate-fade-up stagger-4">
        <h2 class="section-title">COMO ENTRAR NO BOLÃO</h2>
        <div class="steps">
          <div class="step">
            <div class="step-num font-display">1</div>
            <div class="step-text">Acesse o link de convite enviado pelo administrador do bolão</div>
          </div>
          <div class="step">
            <div class="step-num font-display">2</div>
            <div class="step-text">Crie sua conta ou entre com o Google</div>
          </div>
          <div class="step">
            <div class="step-num font-display">3</div>
            <div class="step-text">Faça seus palpites antes de cada jogo</div>
          </div>
          <div class="step">
            <div class="step-num font-display">4</div>
            <div class="step-text">Acompanhe o ranking e torça para acertar!</div>
          </div>
        </div>
      </section>

      <!-- Ver palpites dos outros -->
      <section class="section animate-fade-up stagger-4">
        <h2 class="section-title">PALPITES DOS OUTROS</h2>
        <div class="info-card">
          <div class="info-icon">👀</div>
          <div>
            <p class="info-text">Os palpites dos outros participantes ficam <strong>ocultos até o jogo começar</strong>. Quando o apito soa, você pode ver o que todo mundo apostou — e torcer junto (ou contra).</p>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
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
  gap: 14px;
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
  flex-shrink: 0;
}
.back-btn:hover { color: var(--neon); }
.page-title {
  font-size: 1.6rem;
  line-height: 1;
}

.content { padding-top: 8px; }

.intro-card {
  margin: 20px 0 8px;
  padding: 16px;
  background: rgba(57,255,106,0.04);
  border: 1px solid rgba(57,255,106,0.14);
  border-radius: 12px;
}
.intro-text {
  font-size: 0.9rem;
  color: rgba(255,255,255,0.75);
  line-height: 1.6;
}

.section { margin-top: 28px; }
.section-title {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.75rem;
  letter-spacing: 0.16em;
  color: var(--text-muted);
  margin-bottom: 12px;
  padding-left: 2px;
}

/* Pontuação grid */
.rules-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.rule-card {
  border-radius: 10px;
  padding: 14px 10px;
  text-align: center;
  border: 1px solid;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.rule--exact {
  background: rgba(57,255,106,0.06);
  border-color: rgba(57,255,106,0.25);
}
.rule--winner {
  background: rgba(255,210,80,0.05);
  border-color: rgba(255,210,80,0.2);
}
.rule--miss {
  background: rgba(255,255,255,0.02);
  border-color: rgba(255,255,255,0.08);
}
.rule-points {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 2.4rem;
  line-height: 1;
  color: var(--neon);
}
.rule--winner .rule-points { color: rgba(255,210,80,0.9); }
.rule--miss .rule-points { color: var(--text-muted); }

.rule-label {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.62rem;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.rule-desc {
  font-size: 0.72rem;
  color: rgba(255,255,255,0.7);
  line-height: 1.3;
  font-weight: 500;
}
.rule-example {
  font-size: 0.65rem;
  color: var(--text-muted);
  margin-top: 4px;
  line-height: 1.3;
}

/* Info card */
.info-card {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  padding: 14px 16px;
}
.info-icon { font-size: 1.3rem; flex-shrink: 0; margin-top: 1px; }
.info-text {
  font-size: 0.85rem;
  color: rgba(255,255,255,0.7);
  line-height: 1.6;
}
.info-text strong { color: rgba(255,255,255,0.9); font-weight: 600; }

/* Knockout table */
.knockout-table {
  border: 1px solid rgba(57,255,106,0.15);
  border-radius: 10px;
  overflow: hidden;
}
.kt-row {
  display: grid;
  grid-template-columns: 1fr 0.5fr 0.7fr 0.9fr;
  padding: 10px 14px;
  font-size: 0.82rem;
  color: rgba(255,255,255,0.75);
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.kt-row:last-child { border-bottom: none; }
.kt-header {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 0.68rem;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  background: rgba(255,255,255,0.03);
}
.kt-final { color: var(--neon); }
.text-center { text-align: center; }
.text-muted { color: var(--text-muted); }

/* Steps */
.steps { display: flex; flex-direction: column; gap: 12px; }
.step {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}
.step-num {
  font-size: 1.5rem;
  line-height: 1;
  color: var(--neon);
  flex-shrink: 0;
  width: 28px;
  text-align: center;
}
.step-text {
  font-size: 0.85rem;
  color: rgba(255,255,255,0.7);
  line-height: 1.5;
  padding-top: 2px;
}
</style>
