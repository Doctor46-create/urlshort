package repository

import "github.com/Doctor46-create/urlshort/internal/model"

type URLRepository interface {
	Save(shortKey, url string, requestID string) error
	Get(shortKey string) (string, error)
	SaveBatch(shortKeys, urls []string, requestID string) error
	GetUserURLs(userID string) ([]model.UserURL, error)
}
