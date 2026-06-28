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
