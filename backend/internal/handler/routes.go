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
					r.Post("/regenerate-invite", bolao.RegenerateInvite)
					r.Get("/palpites", palpite.ListMine)
					r.Get("/palpites/{jogoId}", palpite.ListByJogo)
					r.Put("/palpites/{jogoId}", palpite.Upsert)
					r.Get("/ranking", ranking.Get)
					r.Get("/feed", feed.List)
				})
			})
		})
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
