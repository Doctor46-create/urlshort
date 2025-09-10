// Package repository
package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

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

func NewURLRepository(filePath string) URLRepository {
	repo := &urlRepository{
		storage:  make(map[string]string),
		filePath: filePath,
	}
	
	repo.loadFromFile()
	
	return repo
}

func (r *urlRepository) Save(shortKey, url string, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	recordUUID := requestID
	if recordUUID == "" {
		recordUUID = uuid.New().String()
	}
	
	if existingURL, exists := r.storage[shortKey]; exists {
		if existingURL == url {
			return nil
		}
		return fmt.Errorf("short key already exists")
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
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
