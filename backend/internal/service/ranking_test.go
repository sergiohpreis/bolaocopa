package service

import "testing"

func TestCalcPontos(t *testing.T) {
	tests := []struct {
		name          string
		palHome       int32
		palAway       int32
		resHome       int32
		resAway       int32
		stage         string
		apiWinner     string
		penaltyWinner string
		want          float64
	}{
		// Mata-mata: empate exato + acertou penalty winner
		{
			name:          "empate exato + acertou penalty winner (home)",
			palHome:       1, palAway: 1,
			resHome: 1, resAway: 1,
			stage:         "QUARTER_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "home",
			want:          10.0*2.5 + 3.0, // 28
		},
		// Mata-mata: empate exato + errou penalty winner
		{
			name:          "empate exato + errou penalty winner",
			palHome:       1, palAway: 1,
			resHome: 1, resAway: 1,
			stage:         "QUARTER_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "away",
			want:          3.0 * 2.5, // 7.5
		},
		// Mata-mata: empate exato + penalty_winner NULL (palpite antigo, pré-feature)
		{
			name:          "empate exato + penalty_winner null (palpite antigo)",
			palHome:       1, palAway: 1,
			resHome: 1, resAway: 1,
			stage:         "QUARTER_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "",
			want:          3.0 * 2.5, // 7.5
		},
		// Regressão: vitória exata no mata-mata (sem pênaltis)
		{
			name:          "vitória exata no mata-mata",
			palHome:       2, palAway: 0,
			resHome: 2, resAway: 0,
			stage:         "SEMI_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "",
			want:          10.0 * 3.0, // 30
		},
		// Regressão: só acertou o vencedor no mata-mata (placar errado)
		{
			name:          "só acertou o vencedor no mata-mata",
			palHome:       2, palAway: 0,
			resHome: 3, resAway: 1,
			stage:         "SEMI_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "",
			want:          3.0 * 3.0, // 9
		},
		// Fase de grupos: penalty_winner não deve afetar pontuação
		{
			name:          "fase de grupos: placar exato (penalty_winner ignorado)",
			palHome:       2, palAway: 1,
			resHome: 2, resAway: 1,
			stage:         "GROUP_STAGE",
			apiWinner:     "",
			penaltyWinner: "home",
			want:          10,
		},
		{
			name:          "fase de grupos: acertou vencedor",
			palHome:       2, palAway: 0,
			resHome: 3, resAway: 1,
			stage:         "GROUP_STAGE",
			apiWinner:     "",
			penaltyWinner: "",
			want:          3,
		},
		{
			name:          "fase de grupos: errou resultado",
			palHome:       1, palAway: 0,
			resHome: 0, resAway: 1,
			stage:         "GROUP_STAGE",
			apiWinner:     "",
			penaltyWinner: "",
			want:          0,
		},
		// Bug 1: apostou empate, jogo decidido no tempo normal → 0 pts
		{
			name:          "bug1: apostou empate, resultado vitória away no tempo normal",
			palHome:       1, palAway: 1,
			resHome: 0, resAway: 1,
			stage:         "LAST_32",
			apiWinner:     "AWAY_TEAM",
			penaltyWinner: "away",
			want:          0,
		},
		// Bug 2: apostou empate, jogo a pênaltis, errou placar exato, acertou pênaltis → 3×mult + 3
		{
			name:          "bug2: apostou empate (placar errado) + acertou pênaltis",
			palHome:       0, palAway: 0,
			resHome: 1, resAway: 1,
			stage:         "LAST_32",
			apiWinner:     "AWAY_TEAM",
			penaltyWinner: "away",
			want:          3.0*1.5 + 3.0, // 7.5
		},
		// Regressão: apostou empate, jogo a pênaltis, errou placar, errou pênaltis → só 3×mult
		{
			name:          "regressao: apostou empate (placar errado) + errou pênaltis",
			palHome:       0, palAway: 0,
			resHome: 1, resAway: 1,
			stage:         "LAST_32",
			apiWinner:     "AWAY_TEAM",
			penaltyWinner: "home",
			want:          3.0 * 1.5, // 4.5
		},
		// Regressão: bug1 com vitória home
		{
			name:          "bug1: apostou empate, resultado vitória home no tempo normal (QUARTER_FINALS)",
			palHome:       2, palAway: 2,
			resHome: 2, resAway: 0,
			stage:         "QUARTER_FINALS",
			apiWinner:     "HOME_TEAM",
			penaltyWinner: "home",
			want:          0,
		},
		// Apostou vencedor errado no mata-mata → 0 pts
		{
			name:          "apostou vencedor errado no mata-mata",
			palHome:       2, palAway: 0,
			resHome: 0, resAway: 1,
			stage:         "SEMI_FINALS",
			apiWinner:     "AWAY_TEAM",
			penaltyWinner: "",
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcPontos(tt.palHome, tt.palAway, tt.resHome, tt.resAway, tt.stage, tt.apiWinner, tt.penaltyWinner)
			if got != tt.want {
				t.Errorf("calcPontos() = %v, want %v", got, tt.want)
			}
		})
	}
}
