package audit

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

func (s *Subject) Register(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *Subject) NotifyAll(event *AuditEvent) {
	s.mu.RLock()
	observers := s.observers
	s.mu.RUnlock()

	for _, observer := range observers {
		go observer.Notify(event)
	}
}

func (s *Subject) CloseAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, observer := range s.observers {
		observer.Close()
	}
}

type FileObserver struct {
	file   *os.File
	mu     sync.Mutex
	closed bool
}

func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	return &FileObserver{
		file:   file,
		closed: false,
	}, nil
}

func (f *FileObserver) Notify(event *AuditEvent) error {
	if f.closed {
		return fmt.Errorf("file observer is closed")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := f.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (f *FileObserver) Close() error {
	if f.closed {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.closed = true
	return f.file.Close()
}

type HTTPObserver struct {
	url    string
	client *http.Client
	closed bool
	mu     sync.Mutex
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
		closed: false,
	}
}

func (h *HTTPObserver) Notify(event *AuditEvent) error {
	if h.closed {
		return fmt.Errorf("http observer is closed")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest("POST", h.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit server returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (h *HTTPObserver) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	return nil
}
