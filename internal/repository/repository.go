//Package repository
package repository

import (
	"fmt"
	"sync"
)

type urlRepository struct {
	storage map[string]string
	mu      sync.RWMutex
}

func NewURLRepository() URLRepository {
	return &urlRepository{
		storage: make(map[string]string),
	}
}

func (r *urlRepository) Save(shortKey, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storage[shortKey] = url
	return nil
}

func (r *urlRepository) Get(shortKey string) (string, error) {
	url, exists := r.storage[shortKey]
	r.mu.Lock()
	defer r.mu.Unlock()
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}
