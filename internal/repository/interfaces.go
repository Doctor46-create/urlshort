package repository

type URLRepository interface {
	Save(shortKey, url string, requestID string) error
	Get(shortKey string) (string, error)
	SaveBatch(shortKeys, urls []string, requestID string) error
}
