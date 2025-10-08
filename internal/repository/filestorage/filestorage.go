package filestorage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/google/uuid"
)

type fileRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type urlRepository struct {
	storage  map[string]string
	filePath string
	mu       sync.RWMutex
}

func NewURLRepository(filePath string) repository.URLRepository {
	repo := &urlRepository{
		storage:  make(map[string]string),
		filePath: filePath,
	}

	repo.loadFromFile()

	return repo
}

func (r *urlRepository) Save(shortKey, url string, requestID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for existingShortKey, existingURL := range r.storage {
		if existingURL == url {
			return repository.NewURLConflictError(existingShortKey)
		}
	}

	if existingURL, exists := r.storage[shortKey]; exists {
		if existingURL == url {
			return nil 
		}
		return fmt.Errorf("short key already exists for different URL")
	}

	recordUUID := requestID
	if recordUUID == "" {
		recordUUID = uuid.New().String()
	}

	r.storage[shortKey] = url

	if r.filePath != "" {
		record := fileRecord{
			UUID:        recordUUID,
			ShortURL:    shortKey,
			OriginalURL: url,
		}
		return r.appendToFile(record)
	}

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

	for shortKey, url := range r.storage {
		if url == originalURL {
			return shortKey, nil
		}
	}

	return "", fmt.Errorf("URL not found")
}

func (r *urlRepository) SaveBatch(shortKeys, urls []string, requestID string, userID string) error {
	if len(shortKeys) != len(urls) {
		return fmt.Errorf("shortKeys and urls must have the same length")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, url := range urls {
		for existingShortKey, existingURL := range r.storage {
			if existingURL == url {
				return repository.NewURLConflictError(existingShortKey)
			}
		}

		if existingURL, exists := r.storage[shortKeys[i]]; exists {
			if existingURL != urls[i] {
				return fmt.Errorf("short key %s already exists for different URL", shortKeys[i])
			}
		}
	}

	existingRecords, err := r.readAllRecords()
	if err != nil {
		return err
	}

	for i := range shortKeys {
		r.storage[shortKeys[i]] = urls[i]

		recordUUID := requestID
		if recordUUID == "" {
			recordUUID = uuid.New().String()
		}

		record := fileRecord{
			UUID:        recordUUID,
			ShortURL:    shortKeys[i],
			OriginalURL: urls[i],
		}
		existingRecords = append(existingRecords, record)
	}

	if r.filePath != "" {
		return r.writeAllRecords(existingRecords)
	}

	return nil
}

func (r *urlRepository) appendToFile(record fileRecord) error {
	existingRecords, err := r.readAllRecords()
	if err != nil {
		return err
	}

	existingRecords = append(existingRecords, record)

	return r.writeAllRecords(existingRecords)
}

func (r *urlRepository) fileExists() bool {
	if r.filePath == "" {
		return false
	}
	_, err := os.Stat(r.filePath)
	return !os.IsNotExist(err)
}

func (r *urlRepository) readAndParseFile() ([]fileRecord, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer file.Close()

	var records []fileRecord
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&records); err != nil {
		return nil, fmt.Errorf("failed to decode JSON data: %w", err)
	}

	return records, nil
}

func (r *urlRepository) readAllRecords() ([]fileRecord, error) {
	if !r.fileExists() {
		return []fileRecord{}, nil
	}

	return r.readAndParseFile()
}

func (r *urlRepository) writeAllRecords(records []fileRecord) error {
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

func (r *urlRepository) loadFromFile() {
	if r.filePath == "" {
		return
	}

	records, err := r.readAllRecords()
	if err != nil {
		fmt.Printf("Warning: failed to load storage data: %v\n", err)
		return
	}

	for _, record := range records {
		r.storage[record.ShortURL] = record.OriginalURL
	}
}

func (r *urlRepository) GetUserURLs(userID string) ([]model.UserURL, error) {
	return nil, fmt.Errorf("GetUserURLs not implemented for file storage")
}

func (r *urlRepository) DeleteURLs(userID string, urlIDs []string) {
	fmt.Print("DeteleUrls not implemented for filestorage")
}
