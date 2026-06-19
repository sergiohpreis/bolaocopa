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

type BolaoHandler struct {
	svc *service.BolaoService
}

func NewBolaoHandler(svc *service.BolaoService) *BolaoHandler {
	return &BolaoHandler{svc: svc}
}

// POST /api/v1/boloes
func (h *BolaoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		apierror.BadRequest(w, "name is required")
		return
	}
	userID := middleware.UserIDFromContext(r.Context())
	bolao, err := h.svc.Create(r.Context(), in.Name, userID)
	if err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, bolao)
}

// GET /api/v1/boloes
func (h *BolaoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	boloes, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, boloes)
}

// GET /api/v1/boloes/{id}
func (h *BolaoHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolao, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		if errors.Is(err, service.ErrNotParticipante) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bolao)
}

// POST /api/v1/boloes/join/{token}
func (h *BolaoHandler) Join(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID := middleware.UserIDFromContext(r.Context())
	bolao, err := h.svc.JoinByToken(r.Context(), token, userID)
	if err != nil {
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "invite not found")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bolao)
}

// POST /api/v1/boloes/{id}/regenerate-invite
func (h *BolaoHandler) RegenerateInvite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolao, err := h.svc.RegenerateInviteToken(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can regenerate the invite token")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bolao)
}

// DELETE /api/v1/boloes/{id}
func (h *BolaoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can delete this bolao")
			return
		}
		apierror.Internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /api/v1/boloes/{id}/whatsapp-group
func (h *BolaoHandler) SetWAGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	var body struct {
		JID string `json:"jid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.BadRequest(w, "invalid json")
		return
	}

	bolao, err := h.svc.SetWAGroup(r.Context(), bolaoID, userID, body.JID)
	if err != nil {
		if errors.Is(err, service.ErrBolaoNotFound) {
			apierror.NotFound(w, "bolao not found")
			return
		}
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can change whatsapp group")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bolao)
}

// PATCH /api/v1/boloes/{id}/settings
func (h *BolaoHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	var body struct {
		RetroativoEnabled *bool `json:"retroativo_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RetroativoEnabled == nil {
		apierror.BadRequest(w, "retroativo_enabled is required")
		return
	}

	bolao, err := h.svc.SetRetroativoEnabled(r.Context(), bolaoID, userID, *body.RetroativoEnabled)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			apierror.Forbidden(w, "only the admin can change settings")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bolao)
}
