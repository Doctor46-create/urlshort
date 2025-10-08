package memory

import (
	"fmt"
	"sync"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/repository"
)

type urlRepository struct {
	storage        map[string]string
	reverseStorage map[string]string
	mu             sync.RWMutex
}

func NewURLRepository() repository.URLRepository {
	return &urlRepository{
		storage:        make(map[string]string),
		reverseStorage: make(map[string]string),
	}
}

func (r *urlRepository) Save(shortKey, url string, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShortKey, exists := r.reverseStorage[url]; exists {
		return repository.NewURLConflictError(existingShortKey)
	}

	r.storage[shortKey] = url
	r.reverseStorage[url] = shortKey
	return nil
}

func (r *urlRepository) Get(shortKey string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.storage[shortKey]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}

func (r *urlRepository) FindByOriginalURL(originalURL string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortKey, exists := r.reverseStorage[originalURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return shortKey, nil
}

func (r *urlRepository) SaveBatch(shortKeys, urls []string, requestID string) error {
	if len(shortKeys) != len(urls) {
		return fmt.Errorf("shortKeys and urls must have the same length")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range urls {
		if existingShortKey, exists := r.reverseStorage[url]; exists {
			return repository.NewURLConflictError(existingShortKey)
		}
	}

	for i := range shortKeys {
		r.storage[shortKeys[i]] = urls[i]
		r.reverseStorage[urls[i]] = shortKeys[i]
	}

	return nil
}

func (r *urlRepository) GetUserURLs(userID string) ([]model.UserURL, error) {
	return nil, fmt.Errorf("GetUserURLs not implemented for file storage")
}
