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
		HomeScore     int    `json:"home_score"`
		AwayScore     int    `json:"away_score"`
		PenaltyWinner string `json:"penalty_winner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		apierror.BadRequest(w, "invalid body")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	jogoID := chi.URLParam(r, "jogoId")

	palpite, err := h.svc.Upsert(r.Context(), bolaoID, userID, jogoID, in.HomeScore, in.AwayScore, in.PenaltyWinner)
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

// PUT /api/v1/boloes/{id}/palpites/{jogoId}/retroativo
func (h *PalpiteHandler) UpsertRetroativo(w http.ResponseWriter, r *http.Request) {
	var in struct {
		HomeScore     int    `json:"home_score"`
		AwayScore     int    `json:"away_score"`
		PenaltyWinner string `json:"penalty_winner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		apierror.BadRequest(w, "invalid body")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	jogoID := chi.URLParam(r, "jogoId")

	palpite, err := h.svc.UpsertRetroativo(r.Context(), bolaoID, userID, jogoID, in.HomeScore, in.AwayScore, in.PenaltyWinner)
	if err != nil {
		if errors.Is(err, service.ErrRetroativoDesabilitado) {
			apierror.BadRequest(w, "palpites retroativos desabilitados neste bolão")
			return
		}
		if errors.Is(err, service.ErrJogoAindaAberto) {
			apierror.BadRequest(w, "jogo ainda não começou")
			return
		}
		if errors.Is(err, service.ErrPalpiteJaAprovado) {
			apierror.BadRequest(w, "palpite já aprovado: não pode ser alterado")
			return
		}
		if errors.Is(err, service.ErrJogoNotFound) {
			apierror.NotFound(w, "jogo not found")
			return
		}
		if errors.Is(err, service.ErrNotParticipante) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, palpite)
}

// GET /api/v1/boloes/{id}/palpites/pendentes
func (h *PalpiteHandler) ListPendentes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	pendentes, err := h.svc.ListPendentes(r.Context(), bolaoID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can view pending palpites")
			return
		}
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pendentes)
}

// POST /api/v1/boloes/{id}/palpites/{palpiteId}/aprovar
func (h *PalpiteHandler) Aprovar(w http.ResponseWriter, r *http.Request) {
	h.avaliar(w, r, true)
}

// POST /api/v1/boloes/{id}/palpites/{palpiteId}/rejeitar
func (h *PalpiteHandler) Rejeitar(w http.ResponseWriter, r *http.Request) {
	h.avaliar(w, r, false)
}

func (h *PalpiteHandler) avaliar(w http.ResponseWriter, r *http.Request, aprovar bool) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	palpiteID := chi.URLParam(r, "palpiteId")

	palpite, err := h.svc.AprovarOuRejeitar(r.Context(), bolaoID, palpiteID, userID, aprovar)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can approve or reject palpites")
			return
		}
		if errors.Is(err, service.ErrPalpiteNaoEncontrado) || errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "palpite not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, palpite)
}

// GET /api/v1/boloes/{id}/palpites/retroativos
func (h *PalpiteHandler) ListRetroativosAprovados(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	items, err := h.svc.ListRetroativosAprovados(r.Context(), bolaoID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can list approved retroativos")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// DELETE /api/v1/boloes/{id}/palpites/{palpiteId}
func (h *PalpiteHandler) Desaprovar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")
	palpiteID := chi.URLParam(r, "palpiteId")

	if err := h.svc.Desaprovar(r.Context(), bolaoID, palpiteID, userID); err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can remove palpites")
			return
		}
		if errors.Is(err, service.ErrPalpiteNaoEncontrado) || errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "palpite not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
