package service

import "github.com/Doctor46-create/urlshort/internal/model"

type Shortener interface {
	Shorten(originalURL string, requestID string, userID string) (shortKey string, err error)
	GetOriginal(shortKey string) (originalURL string, err error)
	ShortenBatch(items []model.BatchRequestItem, requestID string, userID string) ([]model.BatchResponseItem, error)
	GetUserURLs(userID string) ([]model.UserURL, error)
	DeleteURLs(userID string, shortURLs []string)
	PingDB() error
	GetStats() (urls int, users int, err error)
}
