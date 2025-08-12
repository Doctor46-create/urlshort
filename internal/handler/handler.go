package handler

import (
	"io"
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/go-chi/chi/v5"
)

type urlHandler struct {
	srvc service.Shortener
	cfg  config.Config
}

func NewURLHandler(srvc service.Shortener, cfg config.Config) URLHandler {
	return &urlHandler{
		srvc: srvc,
		cfg:  cfg,
	}
}

func (h *urlHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortKey, err := h.srvc.Shorten(originalURL)
	if err != nil {
		http.Error(w, "Server error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.BaseURL + shortKey))
}

func (h *urlHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	shortKey := chi.URLParam(r, "shortKey")
	originalURL, err := h.srvc.GetOriginal(shortKey)
	if err != nil {
		http.Error(w, "Not found", http.StatusBadRequest)
		return
	}
	// http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
