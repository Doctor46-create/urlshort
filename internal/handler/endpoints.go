package handler

import (
	"encoding/json"
	"io"
	"net/http"
	//"strconv"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *urlHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Error("Method not allowed", zap.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortKey, err := h.srvc.Shorten(originalURL)
	if err != nil {
		h.logger.Error("Failed to shorten URL", zap.Error(err), zap.String("url", originalURL))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.GetBaseURL() + "/" + shortKey))

	h.logger.Info("URL shortened successfully",
		zap.String("original_url", originalURL),
		zap.String("short_key", shortKey))
}

func (h *urlHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Error("Method not allowed", zap.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	shortKey := chi.URLParam(r, "shortKey")
	originalURL, err := h.srvc.GetOriginal(shortKey)
	if err != nil {
		h.logger.Error("URL not found", zap.Error(err), zap.String("short_key", shortKey))
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

	h.logger.Info("URL redirected successfully",
		zap.String("short_key", shortKey),
		zap.String("original_url", originalURL))
}

func (h *urlHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Error("Method not allowed", zap.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	newRequest := &model.JSONRequest{}
	if err := json.NewDecoder(r.Body).Decode(&newRequest); err != nil {
		h.logger.Error("Failed to decode JSON request", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err := model.InputJSONValidate(newRequest)
	if err != nil {
		h.logger.Error("JSON validation failed", zap.Error(err), zap.Any("request", newRequest))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	shortKey, err := h.srvc.Shorten(newRequest.URL)
	if err != nil {
		h.logger.Error("Failed to shorten URL from JSON", zap.Error(err), zap.String("url", newRequest.URL))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fullShortURL := h.cfg.GetBaseURL() + "/" + shortKey

	response := model.JSONResponse{
		Result: fullShortURL,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal JSON response", zap.Error(err))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

	h.logger.Info("JSON URL shortened successfully",
		zap.String("original_url", newRequest.URL),
		zap.String("short_key", shortKey))
}
