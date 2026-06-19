package handler

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// WAProxyHandler faz proxy das chamadas para o whatsapp service,
// adicionando X-API-Secret server-side. O frontend nunca vê o secret.
type WAProxyHandler struct {
	baseURL   string
	apiSecret string
	client    *http.Client
}

func NewWAProxyHandler(baseURL, apiSecret string) *WAProxyHandler {
	return &WAProxyHandler{
		baseURL:   baseURL,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *WAProxyHandler) proxy(w http.ResponseWriter, r *http.Request, waPath string) {
	target, err := url.Parse(h.baseURL + waPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	req.Header.Set("X-API-Secret", h.apiSecret)

	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, "whatsapp service unavailable", http.StatusBadGateway)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *WAProxyHandler) Status(w http.ResponseWriter, r *http.Request)      { h.proxy(w, r, "/status") }
func (h *WAProxyHandler) QR(w http.ResponseWriter, r *http.Request)          { h.proxy(w, r, "/qr") }
func (h *WAProxyHandler) Connect(w http.ResponseWriter, r *http.Request)     { h.proxy(w, r, "/connect") }
func (h *WAProxyHandler) Disconnect(w http.ResponseWriter, r *http.Request)  { h.proxy(w, r, "/connect") }
func (h *WAProxyHandler) Groups(w http.ResponseWriter, r *http.Request)      { h.proxy(w, r, "/groups") }
func (h *WAProxyHandler) Toggle(w http.ResponseWriter, r *http.Request)      { h.proxy(w, r, "/toggle") }
func (h *WAProxyHandler) Healthcheck(w http.ResponseWriter, r *http.Request) { h.proxy(w, r, "/healthcheck") }
