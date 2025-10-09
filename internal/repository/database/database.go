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
	DeleteURLs: `UPDATE short_urls SET is_deleted = true 
        			 WHERE user_id = $1 AND short_url = ANY($2) AND is_deleted = false`,
}

type urlRepository struct {
	DB            *sql.DB
	Queries       SQLQueries
	mu            sync.RWMutex
	DeleteChannel chan model.DeleteURL
	WG            sync.WaitGroup
	shutdown      chan struct{}
}

func NewURLRepository(db *sql.DB) repository.URLRepository {
	repo := &urlRepository{
		DB:            db,
		Queries:       queries,
		DeleteChannel: make(chan model.DeleteURL, 1000),
		shutdown:      make(chan struct{}),
	}
	repo.initDeletionWorkers(context.Background())
	return repo
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

func (r *urlRepository) SaveBatch(shortKeys, urls []string, requestID string, userID string) error {
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
		_, err := stmt.Exec(shortKey, urls[i], now, userID)
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

func (r *urlRepository) fanInWorker(ctx context.Context, inputs []chan model.DeleteURL, output chan<- model.DeleteURL) {
	defer close(output)

	var wg sync.WaitGroup

	forward := func(input <-chan model.DeleteURL) {
		defer wg.Done()
		for task := range input {
			select {
			case output <- task:
			case <-ctx.Done():
				return
			case <-r.shutdown:
				return
			}
		}
	}

	wg.Add(len(inputs))
	for _, input := range inputs {
		go forward(input)
	}

	wg.Wait()
}

func (r *urlRepository) initDeletionWorkers(ctx context.Context) {
	const numWorkers = 3
	const numInputChannels = 5

	inputChannels := make([]chan model.DeleteURL, numInputChannels)
	for i := range inputChannels {
		inputChannels[i] = make(chan model.DeleteURL, 200)
	}

	fanInOutput := make(chan model.DeleteURL, 1000)
	go r.fanInWorker(ctx, inputChannels, fanInOutput)

	r.DeleteChannel = fanInOutput

	for id := range numWorkers {
		r.WG.Add(1)
		go func(id int) {
			defer r.WG.Done()
			log.Printf("Deletion worker %d started", id)
			r.deleteWorker(ctx)
		}(id)
	}

	r.WG.Add(1)
	go func() {
		defer r.WG.Done()
		r.taskDistributor(ctx, inputChannels)
	}()
}

func (r *urlRepository) taskDistributor(ctx context.Context, inputs []chan model.DeleteURL) {
	var counter uint64

	for {
		select {
		case <-ctx.Done():
			for _, input := range inputs {
				close(input)
			}
			return
		case <-r.shutdown:
			for _, input := range inputs {
				close(input)
			}
			return
		case task, ok := <-r.DeleteChannel:
			if !ok {
				for _, input := range inputs {
					close(input)
				}
				return
			}

			idx := counter % uint64(len(inputs))
			counter++

			select {
			case inputs[idx] <- task:
			case <-time.After(100 * time.Millisecond):
				log.Printf("Failed to distribute deletion task after timeout: user=%s, url=%s",
					task.UserID, task.ShortURL)
			case <-ctx.Done():
				return
			case <-r.shutdown:
				return
			}
		}
	}
}

func (r *urlRepository) deleteWorker(ctx context.Context) {
	const batchSize = 100
	const batchTimeout = 2 * time.Second

	taskBuffer := make([]model.DeleteURL, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			r.processBatch(taskBuffer)
			return
		case task, ok := <-r.DeleteChannel:
			if !ok {
				r.processBatch(taskBuffer)
				return
			}
			taskBuffer = append(taskBuffer, task)
			if len(taskBuffer) >= batchSize {
				r.processBatch(taskBuffer)
				taskBuffer = taskBuffer[:0]
				timer.Reset(batchTimeout)
			}
		case <-timer.C:
			if len(taskBuffer) > 0 {
				r.processBatch(taskBuffer)
				taskBuffer = taskBuffer[:0]
			}
			timer.Reset(batchTimeout)
		}
	}
}

func (r *urlRepository) processBatch(tasks []model.DeleteURL) {
	if len(tasks) == 0 {
		return
	}

	groups := make(map[string][]string, len(tasks))
	for _, task := range tasks {
		groups[task.UserID] = append(groups[task.UserID], task.ShortURL)
	}

	for userID, shortURLs := range groups {
		if err := r.batchDeleteURLs(userID, shortURLs); err != nil {
			log.Printf("Failed to delete URLs for user %s: %v", userID, err)
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
		task := model.DeleteURL{UserID: userID, ShortURL: shortURL}

		select {
		case r.DeleteChannel <- task:
		case <-time.After(1 * time.Second):
			return fmt.Errorf("deletion service is overloaded, try again later")
		}
	}
	return nil
}

func (r *urlRepository) Shutdown() {
	close(r.shutdown)
	close(r.DeleteChannel)
	r.WG.Wait()
	log.Printf("Deletion service shutdown completed")
}

func (r *urlRepository) PingDB() error {
	if err := r.DB.Ping(); err != nil {
		log.Printf("Error connecting to database: %v", err)
		return err
	}

	fmt.Println("Database connection successful")
	return nil
}
