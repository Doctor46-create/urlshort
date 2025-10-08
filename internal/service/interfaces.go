package service

import "github.com/Doctor46-create/urlshort/internal/model"

type Shortener interface {
	Shorten(originalURL string, requestID string) (shortKey string, err error)
	GetOriginal(shortKey string) (originalURL string, err error)
	ShortenBatch(items []model.BatchRequestItem, requestID string) ([]model.BatchResponseItem, error)
	GetUserURLs(userID string) ([]model.UserURL, error)
}
