package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// GET /api/v1/auth/google — returns the Google OAuth URL
func (h *AuthHandler) GoogleURL(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	writeJSON(w, http.StatusOK, map[string]string{"url": h.svc.GetGoogleAuthURL(state)})
}

// GET /api/v1/auth/google/callback — handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		apierror.BadRequest(w, "missing code parameter")
		return
	}

	tokens, err := h.svc.ExchangeGoogleCode(r.Context(), code)
	if err != nil {
		apierror.Internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Email == "" || in.Name == "" || in.Password == "" {
		apierror.BadRequest(w, "email, name and password are required")
		return
	}
	tokens, err := h.svc.RegisterByEmail(r.Context(), in.Email, in.Name, in.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			apierror.Conflict(w, "email already registered")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tokens)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Email == "" || in.Password == "" {
		apierror.BadRequest(w, "email and password are required")
		return
	}
	tokens, err := h.svc.LoginByEmail(r.Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			apierror.Unauthorized(w, "invalid email or password")
			return
		}
		apierror.Internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		apierror.BadRequest(w, "refresh_token is required")
		return
	}

	tokens, err := h.svc.RefreshToken(r.Context(), in.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			apierror.Unauthorized(w, "invalid or expired refresh token")
			return
		}
		apierror.Internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}
