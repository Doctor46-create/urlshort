package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

type SQLQueries struct {
	GetURL            string
	SaveURL           string
	FindByOriginalURL string
	GetUserURLs       string
	DeleteURLs        string
}

var queries = SQLQueries{
	GetURL: `SELECT long_url, COALESCE(is_deleted, false) FROM short_urls WHERE short_url = $1`,
	SaveURL: `
		INSERT INTO short_urls (short_url, long_url, created_at, user_id)
		VALUES ($1, $2, $3, $4)
	`,
	FindByOriginalURL: `SELECT short_url FROM short_urls WHERE long_url = $1`,
	GetUserURLs:       `SELECT short_url, long_url FROM short_urls WHERE user_id = $1 ORDER BY created_at DESC`,
	DeleteURLs: `
		UPDATE short_urls
		SET is_deleted = true
		WHERE user_id = $1
		  AND short_url = ANY($2)
		  AND is_deleted = false
	`,
}

type urlRepository struct {
	DB      *sql.DB
	Queries SQLQueries

	mu sync.RWMutex

	DeleteChannel chan model.DeleteURL
	WG            sync.WaitGroup
	shutdown      chan struct{}

	startOnce sync.Once
}

func NewURLRepository(db *sql.DB) repository.URLRepository {
	return &urlRepository{
		DB:            db,
		Queries:       queries,
		DeleteChannel: make(chan model.DeleteURL, 1000),
		shutdown:      make(chan struct{}),
	}
}

func (r *urlRepository) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		r.initDeletionWorkers(ctx)
	})
}

func (r *urlRepository) Shutdown() {
	close(r.shutdown)
	r.WG.Wait()
	log.Printf("Deletion service shutdown completed")
}

func (r *urlRepository) Get(shortURL string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var longURL string
	var deleted bool

	err := r.DB.QueryRow(r.Queries.GetURL, shortURL).Scan(&longURL, &deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("error getting URL: %w", err)
	}

	if deleted {
		return "", model.ErrURLIsDeleted
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

func (r *urlRepository) Save(shortKey, url string, requestID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	_, err := r.DB.Exec(r.Queries.SaveURL, shortKey, url, now, userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortKey, findErr := r.FindByOriginalURL(url)
			if findErr != nil {
				return fmt.Errorf("URL already exists but failed to find short key: %w", findErr)
			}
			return repository.NewURLConflictError(existingShortKey)
		}
		return fmt.Errorf("error saving URL: %w", err)
	}

	return nil
}

func (r *urlRepository) SaveBatch(shortKeys, urls []string, requestID string, userID string) error {
	if len(shortKeys) != len(urls) {
		return fmt.Errorf("shortKeys and urls must have the same length")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range urls {
		if existing, err := r.FindByOriginalURL(url); err == nil {
			return repository.NewURLConflictError(existing)
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
		if _, err := stmt.Exec(shortKey, urls[i], now, userID); err != nil {
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
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return urls, nil
}

func (r *urlRepository) initDeletionWorkers(ctx context.Context) {
	const numWorkers = 3
	const numInputChannels = 5

	inputs := make([]chan model.DeleteURL, numInputChannels)
	for i := range inputs {
		inputs[i] = make(chan model.DeleteURL, 200)
	}

	fanInOutput := make(chan model.DeleteURL, 1000)
	go r.fanInWorker(ctx, inputs, fanInOutput)

	r.DeleteChannel = fanInOutput

	for id := range make([]struct{}, numWorkers) {
		r.WG.Add(1)
		id := id
		go func() {
			defer r.WG.Done()
			log.Printf("Deletion worker %d started", id)
			r.deleteWorker(ctx)
		}()
	}

	r.WG.Add(1)
	go func() {
		defer r.WG.Done()
		r.taskDistributor(ctx, inputs)
	}()
}

func (r *urlRepository) fanInWorker(
	ctx context.Context,
	inputs []chan model.DeleteURL,
	output chan<- model.DeleteURL,
) {
	defer close(output)

	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan model.DeleteURL) {
			defer wg.Done()
			for task := range ch {
				select {
				case output <- task:
				case <-ctx.Done():
					return
				case <-r.shutdown:
					return
				}
			}
		}(input)
	}

	wg.Wait()
}

func (r *urlRepository) taskDistributor(ctx context.Context, inputs []chan model.DeleteURL) {
	var counter uint64
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			r.closeInputs(inputs)
			return
		case <-r.shutdown:
			r.closeInputs(inputs)
			return
		case task, ok := <-r.DeleteChannel:
			if !ok {
				r.closeInputs(inputs)
				return
			}

			idx := counter % uint64(len(inputs))
			counter++

			timer.Reset(100 * time.Millisecond)

			select {
			case inputs[idx] <- task:
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
				log.Printf("Failed to distribute deletion task: user=%s url=%s",
					task.UserID, task.ShortURL)
			}
		}
	}
}

func (r *urlRepository) closeInputs(inputs []chan model.DeleteURL) {
	for _, ch := range inputs {
		close(ch)
	}
}

func (r *urlRepository) deleteWorker(ctx context.Context) {
	const batchSize = 100
	const batchTimeout = 2 * time.Second

	buffer := make([]model.DeleteURL, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			r.processBatch(buffer)
			return
		case <-r.shutdown:
			r.processBatch(buffer)
			return
		case task, ok := <-r.DeleteChannel:
			if !ok {
				r.processBatch(buffer)
				return
			}

			buffer = append(buffer, task)
			if len(buffer) >= batchSize {
				r.processBatch(buffer)
				buffer = buffer[:0]
				timer.Reset(batchTimeout)
			}

		case <-timer.C:
			if len(buffer) > 0 {
				r.processBatch(buffer)
				buffer = buffer[:0]
			}
			timer.Reset(batchTimeout)
		}
	}
}

func (r *urlRepository) processBatch(tasks []model.DeleteURL) {
	if len(tasks) == 0 {
		return
	}

	grouped := make(map[string][]string)
	for _, t := range tasks {
		grouped[t.UserID] = append(grouped[t.UserID], t.ShortURL)
	}

	for userID, urls := range grouped {
		if err := r.batchDeleteURLs(userID, urls); err != nil {
			log.Printf("Batch delete failed for user %s: %v", userID, err)
		}
	}
}

func (r *urlRepository) batchDeleteURLs(userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	_, err := r.DB.Exec(r.Queries.DeleteURLs, userID, pq.Array(shortURLs))
	return err
}

func (r *urlRepository) DeleteURLs(userID string, urlIDs []string) error {
	for _, shortURL := range urlIDs {
		task := model.DeleteURL{
			UserID:   userID,
			ShortURL: shortURL,
		}

		select {
		case r.DeleteChannel <- task:
		case <-time.After(time.Second):
			return fmt.Errorf("deletion service overloaded")
		}
	}
	return nil
}

func (r *urlRepository) PingDB() error {
	if err := r.DB.Ping(); err != nil {
		log.Printf("DB ping failed: %v", err)
		return err
	}
	return nil
}
