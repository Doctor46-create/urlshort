package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Doctor46-create/urlshort/internal/repository"
)

type SQLQueries struct {
	GetURL  string
	SaveURL string
}

var queries = SQLQueries{
	GetURL: `SELECT long_url FROM short_urls WHERE short_url = $1`,
	SaveURL: `
		INSERT INTO short_urls (short_url, long_url, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (short_url)
		DO UPDATE SET
			long_url = EXCLUDED.long_url
	`,
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

func (r *urlRepository) Save(shortKey, url string, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	_, err := r.DB.Exec(r.Queries.SaveURL, shortKey, url, now)
	if err != nil {
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
