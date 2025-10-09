package repository

import "github.com/Doctor46-create/urlshort/internal/model"

type URLRepository interface {
	Save(shortKey, url string, requestID string, userID string) error
	Get(shortKey string) (string, error)
	SaveBatch(shortKeys, urls []string, requestID string, userID string) error
	GetUserURLs(userID string) ([]model.UserURL, error)
	DeleteURLs(userID string, urlIDs []string) error
	PingDB() error
}
