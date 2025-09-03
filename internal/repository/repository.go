// Package repository
package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/Doctor46-create/urlshort/internal/model"
)

type urlRepository struct {
	storage    map[string]string      
	records    map[string]*model.URLRecord  
	mu         sync.RWMutex
	filePath   string
}

func NewURLRepository(filePath string) URLRepository {
	repo := &urlRepository{
		storage:  make(map[string]string),
		records:  make(map[string]*model.URLRecord),
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
	record := &model.URLRecord{
		UUID:        recordUUID,
		ShortURL:    shortKey,
		OriginalURL: url,
	}
	r.records[shortKey] = record
	
	return r.appendToFile(record)
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

func (r *urlRepository) appendToFile(record *model.URLRecord) error {
	if r.filePath == "" {
		return nil 
	}
	
	existingRecords, err := r.readAllRecords()
	if err != nil {
		return err
	}
	
	existingRecords = append(existingRecords, *record)
	
	return r.writeAllRecords(existingRecords)
}

func (r *urlRepository) readAllRecords() ([]model.URLRecord, error) {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return []model.URLRecord{}, nil
	}
	
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer file.Close()
	
	var records []model.URLRecord
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&records); err != nil {
		return []model.URLRecord{}, nil
	}
	
	return records, nil
}

func (r *urlRepository) writeAllRecords(records []model.URLRecord) error {
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
		r.records[record.ShortURL] = &model.URLRecord{
			UUID:        record.UUID,
			ShortURL:    record.ShortURL,
			OriginalURL: record.OriginalURL,
		}
	}
}

