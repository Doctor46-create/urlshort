package auth

import (
	"crypto/hmac"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"google.golang.org/grpc/metadata"
)

func SignUserID(userID, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(userID))
	return userID + "." + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func ValidateSignedUserID(signed, secretKey string) (string, bool) {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(userID))
	expected := mac.Sum(nil)

	received, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}

	return userID, hmac.Equal(received, expected)
}

var ErrUnauthorized = grpcUnauthorizedError{}

type grpcUnauthorizedError struct{}

func (grpcUnauthorizedError) Error() string { return "unauthorized" }

func GetUserIDFromMetadata(ctx context.Context, secretKey string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", ErrUnauthorized
	}

	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if token == "" || token != secretKey {
		return "", ErrUnauthorized
	}

	return token, nil
}
