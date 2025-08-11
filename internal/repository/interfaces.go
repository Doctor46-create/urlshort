package repository

type URLRepository interface {
	Save(shortKey, url string) error
	Get(shortKey string) (string, error)
}
