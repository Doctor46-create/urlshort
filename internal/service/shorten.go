// Package service
package service

import (
	"fmt"
	"crypto/sha256"
	"encoding/base64"
	"github.com/Doctor46-create/urlshort/internal/repository"
)

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) Shortener {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(originalURL string) (string, error) {
	shortKey := generateShortKey(originalURL)
	err := s.repo.Save(shortKey, originalURL)
	if err != nil {
		return "", err
	}
	return shortKey, nil
}

func (s *urlService) GetOriginal(shortKey string) (string, error) {
	url, err := s.repo.Get(shortKey)
	if err != nil {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}

func generateShortKey(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
