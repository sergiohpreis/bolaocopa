package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sergiohpreis/bolaocopa/backend/internal/middleware"
	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type RankingHandler struct {
	svc      *service.RankingService
	bolaoSvc *service.BolaoService
}

func NewRankingHandler(svc *service.RankingService, bolaoSvc *service.BolaoService) *RankingHandler {
	return &RankingHandler{svc: svc, bolaoSvc: bolaoSvc}
}

// GET /api/v1/boloes/{id}/ranking
func (h *RankingHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	if _, err := h.bolaoSvc.GetByID(r.Context(), bolaoID, userID); err != nil {
		if errors.Is(err, service.ErrNotParticipante) || errors.Is(err, service.ErrBolaoNotFound) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		apierror.Internal(w, err)
		return
	}

	ranking, err := h.svc.Get(r.Context(), bolaoID)
	if err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ranking)
}
