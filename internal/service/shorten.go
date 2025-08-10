package service

import (
	"crypto/sha256"
	"encoding/base64"
)

type Shortener interface {
	Shorten(originalURL string) (shortKey string, err error)
	GetOriginal(shortKey string) (originalURL string, exists bool)
}

type URLService struct {
	storage map[string]string
}

func NewURLService() *URLService {
	return &URLService{
		storage: make(map[string]string),
	}
}

func (s *URLService) Shorten(originalURL string) (string, error) {
	shortKey := generateShortKey(originalURL)
	s.storage[shortKey] = originalURL
	return shortKey, nil
}

func (s *URLService) GetOriginal(shortKey string) (string, bool) {
	originalURL, exists := s.storage[shortKey]
	return originalURL, exists
}

func generateShortKey(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

