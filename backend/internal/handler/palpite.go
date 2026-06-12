package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sergiohpreis/bolaocopa/backend/internal/middleware"
	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type PalpiteHandler struct {
	svc *service.PalpiteService
}

func NewPalpiteHandler(svc *service.PalpiteService) *PalpiteHandler {
	return &PalpiteHandler{svc: svc}
}

// PUT /api/v1/boloes/{id}/palpites/{jogoId}
func (h *PalpiteHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var in struct {
		HomeScore int `json:"home_score"`
		AwayScore int `json:"away_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		apierror.BadRequest(w, "invalid body")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	jogoID := chi.URLParam(r, "jogoId")

	palpite, err := h.svc.Upsert(r.Context(), bolaoID, userID, jogoID, in.HomeScore, in.AwayScore)
	if err != nil {
		if errors.Is(err, service.ErrPalpiteFechado) {
			apierror.BadRequest(w, "palpite fechado: jogo já começou")
			return
		}
		if errors.Is(err, service.ErrJogoNotFound) {
			apierror.NotFound(w, "jogo not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, palpite)
}

// GET /api/v1/boloes/{id}/palpites/{jogoId}
func (h *PalpiteHandler) ListByJogo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	jogoID := chi.URLParam(r, "jogoId")

	palpites, err := h.svc.ListByJogo(r.Context(), bolaoID, userID, jogoID)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipante) || errors.Is(err, service.ErrBolaoNotFound) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		if errors.Is(err, service.ErrJogoNotFound) {
			apierror.NotFound(w, "jogo not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, palpites)
}

// GET /api/v1/boloes/{id}/palpites
func (h *PalpiteHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	palpites, err := h.svc.ListByUser(r.Context(), bolaoID, userID)
	if err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, palpites)
}
