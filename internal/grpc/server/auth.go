package server

import (
	"context"
	"strings"

	"github.com/Doctor46-create/urlshort/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UserIDFromMetadata(ctx context.Context, secretKey string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata missing")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization missing")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")

	userID, valid := auth.ValidateSignedUserID(token, secretKey)
	if !valid {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userID, nil
}
