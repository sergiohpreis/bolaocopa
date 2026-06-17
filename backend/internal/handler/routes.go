package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	auth *AuthHandler,
	bolao *BolaoHandler,
	jogo *JogoHandler,
	palpite *PalpiteHandler,
	ranking *RankingHandler,
	feed *FeedHandler,
	taxa *TaxaHandler,
	authMw func(http.Handler) http.Handler,
) {
	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		// Public
		r.Route("/auth", func(r chi.Router) {
			r.Get("/google", auth.GoogleURL)
			r.Get("/google/callback", auth.GoogleCallback)
			r.Post("/refresh", auth.Refresh)
			r.Post("/register", auth.Register)
			r.Post("/login", auth.Login)
		})

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(authMw)

			r.Get("/jogos", jogo.List)
			r.Post("/jogos/sync", jogo.Sync)

			r.Route("/boloes", func(r chi.Router) {
				r.Post("/", bolao.Create)
				r.Get("/", bolao.List)
				r.Post("/join/{token}", bolao.Join)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", bolao.Get)
					r.Delete("/", bolao.Delete)
					r.Post("/regenerate-invite", bolao.RegenerateInvite)
					r.Patch("/settings", bolao.UpdateSettings)
					r.Get("/palpites", palpite.ListMine)
					r.Get("/palpites/pendentes", palpite.ListPendentes)
					r.Get("/palpites/retroativos", palpite.ListRetroativosAprovados)
					r.Route("/palpites/{jogoId}", func(r chi.Router) {
						r.Get("/", palpite.ListByJogo)
						r.Put("/", palpite.Upsert)
						r.Put("/retroativo", palpite.UpsertRetroativo)
					})
					r.Post("/palpites/{palpiteId}/aprovar", palpite.Aprovar)
					r.Post("/palpites/{palpiteId}/rejeitar", palpite.Rejeitar)
					r.Delete("/palpites/{palpiteId}", palpite.Desaprovar)
					r.Get("/ranking", ranking.Get)
					r.Get("/feed", feed.List)
					r.Get("/taxa", taxa.GetEstado)
					r.Post("/taxa/proposta", taxa.Propor)
					r.Post("/taxa/votar", taxa.Votar)
				})
			})
		})
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
