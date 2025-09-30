// Package service
package service

import (
	"fmt"
	"crypto/sha256"
	"encoding/base64"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/model"
)

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) Shortener {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(originalURL string, requestID string) (string, error) {
	shortKey := s.generateShortKey(originalURL)
	err := s.repo.Save(shortKey, originalURL, requestID)
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

func (s *urlService) ShortenBatch(items []model.BatchRequestItem, requestID string) ([]model.BatchResponseItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	results := make([]model.BatchResponseItem, 0, len(items))

	for _, item := range items {
		shortKey, err := s.Shorten(item.OriginalURL, requestID)
		if err != nil {
			return nil, fmt.Errorf("failed to shorten URL for correlation_id %s: %w", item.CorrelationID, err)
		}

		results = append(results, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortKey, 
		})
	}

	return results, nil
}

func (s *urlService) generateShortKey(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
