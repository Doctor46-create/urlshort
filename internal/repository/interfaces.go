package repository

type URLRepository interface {
	Save(shortKey, url string, requestID string) error
	Get(shortKey string) (string, error)
}
