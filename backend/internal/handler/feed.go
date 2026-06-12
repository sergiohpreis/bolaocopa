package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sergiohpreis/bolaocopa/backend/internal/middleware"
	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type FeedHandler struct {
	svc *service.FeedService
}

func NewFeedHandler(svc *service.FeedService) *FeedHandler {
	return &FeedHandler{svc: svc}
}

// GET /api/v1/boloes/{id}/feed
func (h *FeedHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	bolaoID := chi.URLParam(r, "id")

	eventos, err := h.svc.ListByBolao(r.Context(), bolaoID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipante) || errors.Is(err, service.ErrBolaoNotFound) {
			apierror.Forbidden(w, "you are not a member of this bolao")
			return
		}
		apierror.Internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, eventos)
}
