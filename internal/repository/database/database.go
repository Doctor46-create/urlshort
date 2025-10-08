package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type SQLQueries struct {
	GetURL            string
	SaveURL           string
	FindByOriginalURL string
	GetUserURLs       string
}

var queries = SQLQueries{
	GetURL: `SELECT long_url FROM short_urls WHERE short_url = $1`,
	SaveURL: `
        INSERT INTO short_urls (short_url, long_url, created_at)
        VALUES ($1, $2, $3)
    `,
	FindByOriginalURL: `SELECT short_url FROM short_urls WHERE long_url = $1`,
	GetUserURLs:       `SELECT short_url, long_url FROM shortened_urls WHERE user_id = $1 ORDER BY created_at DESC`,
}

type urlRepository struct {
	DB      *sql.DB
	Queries SQLQueries
	mu      sync.RWMutex
}

func NewURLRepository(db *sql.DB) repository.URLRepository {
	return &urlRepository{
		DB:      db,
		Queries: queries,
	}
}

func (r *urlRepository) Get(shortURL string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var longURL string
	err := r.DB.QueryRow(r.Queries.GetURL, shortURL).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("error getting URL: %w", err)
	}

	return longURL, nil
}

func (r *urlRepository) FindByOriginalURL(originalURL string) (string, error) {
	var shortURL string
	err := r.DB.QueryRow(r.Queries.FindByOriginalURL, originalURL).Scan(&shortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("error finding URL: %w", err)
	}

	return shortURL, nil
}

func (r *urlRepository) Save(shortKey, url string, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	_, err := r.DB.Exec(r.Queries.SaveURL, shortKey, url, now)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				existingShortKey, findErr := r.FindByOriginalURL(url)
				if findErr != nil {
					return fmt.Errorf("URL already exists but failed to find short key: %w", findErr)
				}
				return repository.NewURLConflictError(existingShortKey)
			}
		}
		return fmt.Errorf("error saving URL: %w", err)
	}

	return nil
}

func (r *urlRepository) SaveBatch(shortKeys, urls []string, requestID string) error {
	if len(shortKeys) != len(urls) {
		return fmt.Errorf("shortKeys and urls must have the same length")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range urls {
		existingShortKey, err := r.FindByOriginalURL(url)
		if err == nil {
			return repository.NewURLConflictError(existingShortKey)
		}
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	stmt, err := tx.Prepare(r.Queries.SaveURL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, shortKey := range shortKeys {
		_, err := stmt.Exec(shortKey, urls[i], now)
		if err != nil {
			return fmt.Errorf("failed to save URL batch item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *urlRepository) GetUserURLs(userID string) ([]model.UserURL, error) {
	rows, err := r.DB.Query(r.Queries.GetUserURLs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user URLs: %w", err)
	}
	defer rows.Close()

	var urls []model.UserURL
	for rows.Next() {
		var pair model.UserURL
		if err := rows.Scan(&pair.ShortURL, &pair.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL pair: %w", err)
		}
		urls = append(urls, pair)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return urls, nil
}
