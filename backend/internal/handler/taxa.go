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

type TaxaHandler struct {
	svc *service.TaxaService
}

func NewTaxaHandler(svc *service.TaxaService) *TaxaHandler {
	return &TaxaHandler{svc: svc}
}

// POST /api/v1/boloes/{id}/taxa/proposta
func (h *TaxaHandler) Propor(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Valor string `json:"valor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Valor == "" {
		apierror.BadRequest(w, "valor must be a non-empty decimal string (e.g. \"50.00\")")
		return
	}
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	proposta, err := h.svc.Propor(r.Context(), bolaoID, userID, in.Valor)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can propose a taxa")
			return
		}
		if errors.Is(err, service.ErrTaxaJaDefinida) {
			apierror.Conflict(w, "taxa de entrada já definida")
			return
		}
		if errors.Is(err, service.ErrPropostaJaExiste) {
			apierror.Conflict(w, "já existe uma proposta ativa")
			return
		}
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, proposta)
}

// POST /api/v1/boloes/{id}/taxa/votar
func (h *TaxaHandler) Votar(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Aprovado *bool `json:"aprovado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Aprovado == nil {
		apierror.BadRequest(w, "aprovado (bool) is required")
		return
	}
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	if err := h.svc.Votar(r.Context(), bolaoID, userID, *in.Aprovado); err != nil {
		if errors.Is(err, service.ErrSemProposta) {
			apierror.NotFound(w, "nenhuma proposta de taxa ativa")
			return
		}
		if errors.Is(err, service.ErrJaVotou) {
			apierror.Conflict(w, "você já votou nesta proposta")
			return
		}
		if errors.Is(err, service.ErrVotoNaoElegivel) {
			apierror.Forbidden(w, "você entrou no bolão após a proposta e não pode votar")
			return
		}
		if errors.Is(err, service.ErrNotParticipante) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/v1/boloes/{id}/taxa
func (h *TaxaHandler) GetEstado(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	estado, err := h.svc.GetEstado(r.Context(), bolaoID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipante) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, estado)
}
