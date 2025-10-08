package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/repository"
)

var ErrURLAlreadyShortened = errors.New("URL already shortened")

type URLAlreadyShortenedError struct {
	ShortKey string
}

func (e *URLAlreadyShortenedError) Error() string {
	return fmt.Sprintf("URL already shortened: %s", e.ShortKey)
}

func (e *URLAlreadyShortenedError) Unwrap() error {
	return ErrURLAlreadyShortened
}

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
		if existingShortKey, isConflict := repository.IsURLConflictError(err); isConflict {
			return existingShortKey, &URLAlreadyShortenedError{ShortKey: existingShortKey}
		}
		return "", fmt.Errorf("failed to shorten URL: %w", err)
	}
	return shortKey, nil
}

func (s *urlService) GetOriginal(shortKey string) (string, error) {
	url, err := s.repo.Get(shortKey)
	if err != nil {
		return "", fmt.Errorf("URL not found: %w", err)
	}
	return url, nil
}

func (s *urlService) ShortenBatch(items []model.BatchRequestItem, requestID string) ([]model.BatchResponseItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	shortKeys := make([]string, len(items))
	urls := make([]string, len(items))

	for i, item := range items {
		shortKeys[i] = s.generateShortKey(item.OriginalURL)
		urls[i] = item.OriginalURL
	}

	err := s.repo.SaveBatch(shortKeys, urls, requestID)
	if err != nil {
		if existingShortKey, isConflict := repository.IsURLConflictError(err); isConflict {
			return nil, fmt.Errorf("batch contains duplicate URL with short key %s: %w", existingShortKey, err)
		}
		return nil, fmt.Errorf("failed to shorten URL batch: %w", err)
	}

	results := make([]model.BatchResponseItem, len(items))
	for i, shortKey := range shortKeys {
		results[i] = model.BatchResponseItem{
			CorrelationID: items[i].CorrelationID,
			ShortURL:      shortKey,
		}
	}

	return results, nil
}

func (s *urlService) generateShortKey(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

func (s *urlService) GetUserURLs(userID string) ([]model.UserURL, error) {
	return s.repo.GetUserURLs(userID)
}
