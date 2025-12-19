package handler

import (
	"context"
	"net/http"
	"reflect"

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
