package handler

import (
	"net/http"

	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type JogoHandler struct {
	svc *service.JogoService
}

func NewJogoHandler(svc *service.JogoService) *JogoHandler {
	return &JogoHandler{svc: svc}
}

// GET /api/v1/jogos
func (h *JogoHandler) List(w http.ResponseWriter, r *http.Request) {
	jogos, err := h.svc.ListAll(r.Context())
	if err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jogos)
}

// POST /api/v1/jogos/sync — admin-only in production (protected by auth middleware)
func (h *JogoHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.SyncFromAPI(r.Context()); err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}
