package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)


func (h *urlHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	if originalURL == "" {
		h.logger.Error("Empty URL in request")
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}

	requestID := r.Header.Get("X-Request-ID")

	shortKey, err := h.srvc.Shorten(originalURL, requestID)
	if err != nil {
		if errors.Is(err, service.ErrURLAlreadyShortened) {
			conflictErr := &service.URLAlreadyShortenedError{}
			if errors.As(err, &conflictErr) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(h.cfg.GetBaseURL() + "/" + conflictErr.ShortKey))

				h.logger.Info("URL already exists",
					zap.String("original_url", originalURL),
					zap.String("short_key", conflictErr.ShortKey),
					zap.String("request_id", requestID))
				return
			}
		}

		h.logger.Error("Failed to shorten URL", zap.Error(err), zap.String("url", originalURL))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.GetBaseURL() + "/" + shortKey))

	h.logger.Info("URL shortened successfully",
		zap.String("original_url", originalURL),
		zap.String("short_key", shortKey),
		zap.String("request_id", requestID))
}

func (h *urlHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
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

	requestID := r.Header.Get("X-Request-ID")

	shortKey, err := h.srvc.Shorten(newRequest.URL, requestID)
	if err != nil {
		if errors.Is(err, service.ErrURLAlreadyShortened) {
			conflictErr := &service.URLAlreadyShortenedError{}
			if errors.As(err, &conflictErr) {
				response := model.JSONResponse{
					Result: h.cfg.GetBaseURL() + "/" + conflictErr.ShortKey,
				}

				jsonData, err := json.Marshal(response)
				if err != nil {
					h.logger.Error("Failed to marshal JSON response", zap.Error(err))
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(jsonData)

				h.logger.Info("URL already exists (JSON)",
					zap.String("original_url", newRequest.URL),
					zap.String("short_key", conflictErr.ShortKey),
					zap.String("request_id", requestID))
				return
			}
		}

		h.logger.Error("Failed to shorten URL from JSON", zap.Error(err), zap.String("url", newRequest.URL))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	response := model.JSONResponse{
		Result: h.cfg.GetBaseURL() + "/" + shortKey,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal JSON response", zap.Error(err))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

	h.logger.Info("JSON URL shortened successfully",
		zap.String("original_url", newRequest.URL),
		zap.String("short_key", shortKey),
		zap.String("request_id", requestID))
}

func (h *urlHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		h.logger.Info("PingDB called but no database configured")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("No database configured"))
		return
	}

	err := h.db.PingDB()
	if err != nil {
		h.logger.Error("Database ping failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database is available"))
	h.logger.Info("Database ping successful")
}

func (h *urlHandler) ShortenBatchURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	var requestItems []model.BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&requestItems); err != nil {
		h.logger.Error("Failed to decode batch request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(requestItems) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	for _, item := range requestItems {
		if item.OriginalURL == "" {
			http.Error(w, "Empty URL in batch", http.StatusBadRequest)
			return
		}
	}

	requestID := r.Header.Get("X-Request-ID")

	results, err := h.srvc.ShortenBatch(requestItems, requestID)
	if err != nil {
		if errors.Is(err, service.ErrURLAlreadyShortened) {
			conflictErr := &service.URLAlreadyShortenedError{}
			if errors.As(err, &conflictErr) {
				h.logger.Error("Batch contains duplicate URL",
					zap.Error(err),
					zap.String("existing_short_key", conflictErr.ShortKey),
					zap.String("request_id", requestID))
				http.Error(w, fmt.Sprintf("Duplicate URL found with short key: %s", conflictErr.ShortKey), http.StatusConflict)
				return
			}
		}

		h.logger.Error("Failed to shorten URL batch",
			zap.Error(err),
			zap.String("request_id", requestID))
		http.Error(w, "Failed to shorten URLs", http.StatusInternalServerError)
		return
	}

	responseWithFullURL := make([]model.BatchResponseItem, len(results))
	for i, result := range results {
		responseWithFullURL[i] = model.BatchResponseItem{
			CorrelationID: result.CorrelationID,
			ShortURL:      h.cfg.GetBaseURL() + "/" + result.ShortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(responseWithFullURL); err != nil {
		h.logger.Error("Failed to encode batch response",
			zap.Error(err),
			zap.String("request_id", requestID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("URL batch shortened successfully",
		zap.Int("count", len(responseWithFullURL)),
		zap.String("request_id", requestID))
}

func (h *urlHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hadCookie, _ := ctx.Value("had_cookie").(bool)
	cookieWasValid, _ := ctx.Value("cookie_was_valid").(bool)

	if hadCookie && !cookieWasValid {
		h.logger.Warn("Unauthorized access: invalid cookie")
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, nil)
		return
	}

	userIDInterface := ctx.Value("user_id")
	userID, ok := userIDInterface.(string)
	if !ok || userID == "" {
		h.logger.Error("Failed to get user_id from context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	urls, err := h.srvc.GetUserURLs(userID)
	if err != nil {
		h.logger.Error("Failed to retrieve user URLs", zap.Error(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	if len(urls) == 0 {
		h.logger.Info("No URLs found for user", zap.String("user_id", userID))
		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
		return
	}

	for i := range urls {
		fullShortURL := h.cfg.GetBaseURL() + "/" + urls[i].ShortURL
		urls[i].ShortURL = fullShortURL
	}

	h.logger.Info("Successfully retrieved user URLs", zap.String("user_id", userID), zap.Int("url_count", len(urls)))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, urls)
}

