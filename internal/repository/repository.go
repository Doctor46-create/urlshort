package repository

import "fmt"

type urlRepository struct {
	storage map[string]string
}

func NewURLRepository() *urlRepository {
	return &urlRepository{
		storage: make(map[string]string),
	}
}

func (r *urlRepository) Save(shortKey, url string) error {
	r.storage[shortKey] = url
	return nil
}

func (r *urlRepository) Get(shortKey string) (string, error) {
	url, exists := r.storage[shortKey]
	if !exists {
		return "", fmt.Errorf("URL not found") 
	}
	return url, nil
}
