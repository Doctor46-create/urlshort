package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/service"
	"go.uber.org/zap"
)

func getUserIDFromContext(
	ctx context.Context,
	w http.ResponseWriter,
	logger *zap.Logger,
) (string, bool) {
	userIDValue := ctx.Value(userIDKey)
	if userIDValue == nil {
		logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", false
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		logger.Error(
			"Invalid user ID type in context",
			zap.Any("type", reflect.TypeOf(userIDValue)),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return "", false
	}

	return userID, true
}

func (h *urlHandler) handleConflictError(
	w http.ResponseWriter,
	err error,
	originalURL string,
	requestID string,
	responseType string,
) bool {
	if !errors.Is(err, service.ErrURLAlreadyShortened) {
		return false
	}

	conflictErr := &service.URLAlreadyShortenedError{}
	if !errors.As(err, &conflictErr) {
		return false
	}

	shortURL := h.cfg.GetBaseURL() + "/" + conflictErr.ShortKey

	switch responseType {
	case "json":
		response := model.JSONResponse{Result: shortURL}
		data, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			h.logger.Error("Failed to marshal conflict JSON response", zap.Error(marshalErr))
			http.Error(w, "Server error", http.StatusInternalServerError)
			return true
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(data)

	default:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURL))
	}

	h.logger.Info("URL already exists",
		zap.String("original_url", originalURL),
		zap.String("short_key", conflictErr.ShortKey),
		zap.String("request_id", requestID),
	)

	return true
}
