package service

import (
	"testing"

	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/repository"
)

type mockURLRepository struct {
	repository.URLRepository
}

func (m *mockURLRepository) Save(shortKey, originalURL, requestID, userID string) error {
	return nil
}

func (m *mockURLRepository) Get(shortKey string) (string, error) {
	return "https://example.com/original", nil
}

func (m *mockURLRepository) SaveBatch(shortKeys, originalURLs []string, requestID, userID string) error {
	return nil
}

func (m *mockURLRepository) GetUserURLs(userID string) ([]model.UserURL, error) {
	return []model.UserURL{
		{ShortURL: "abc123", OriginalURL: "https://example.com/original"},
		{ShortURL: "def456", OriginalURL: "https://example.com/another"},
		{ShortURL: "ghi789", OriginalURL: "https://example.com/third"},
	}, nil
}

func (m *mockURLRepository) DeleteURLs(userID string, shortURLs []string) error {
	return nil
}

func (m *mockURLRepository) PingDB() error {
	return nil
}

func createTestService() *urlService {
	repo := &mockURLRepository{}
	return &urlService{repo: repo}
}

func BenchmarkShorten(b *testing.B) {
	s := createTestService()
	originalURL := "https://www.example.com/very/long/url/that/needs/to/be/shortened"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Shorten(originalURL, "req-"+string(rune(i)), "user-123")
		if err != nil {
			b.Fatalf("Shorten failed: %v", err)
		}
	}
}

func BenchmarkGetOriginal(b *testing.B) {
	s := createTestService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.GetOriginal("abc123")
		if err != nil {
			b.Fatalf("GetOriginal failed: %v", err)
		}
	}
}

func BenchmarkShortenBatchSmall(b *testing.B) {
	s := createTestService()

	items := []model.BatchRequestItem{
		{CorrelationID: "corr1", OriginalURL: "https://example.com/first"},
		{CorrelationID: "corr2", OriginalURL: "https://example.com/second"},
		{CorrelationID: "corr3", OriginalURL: "https://example.com/third"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.ShortenBatch(items, "req-"+string(rune(i)), "user-123")
		if err != nil {
			b.Fatalf("ShortenBatch failed: %v", err)
		}
	}
}

func BenchmarkShortenBatchMedium(b *testing.B) {
	s := createTestService()

	items := make([]model.BatchRequestItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = model.BatchRequestItem{
			CorrelationID: "corr" + string(rune(i)),
			OriginalURL:   "https://example.com/url-" + string(rune(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.ShortenBatch(items, "req-"+string(rune(i)), "user-123")
		if err != nil {
			b.Fatalf("ShortenBatch failed: %v", err)
		}
	}
}

// BenchmarkShortenBatchLarge benchmarks the ShortenBatch function with large batch
func BenchmarkShortenBatchLarge(b *testing.B) {
	s := createTestService()

	items := make([]model.BatchRequestItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = model.BatchRequestItem{
			CorrelationID: "corr" + string(rune(i)),
			OriginalURL:   "https://example.com/url-" + string(rune(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.ShortenBatch(items, "req-"+string(rune(i)), "user-123")
		if err != nil {
			b.Fatalf("ShortenBatch failed: %v", err)
		}
	}
}

func BenchmarkGenerateShortKey(b *testing.B) {
	s := createTestService()
	originalURL := "https://www.example.com/very/long/url/that/needs/to/be/shortened"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.generateShortKey(originalURL)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	s := createTestService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.GetUserURLs("user-123")
		if err != nil {
			b.Fatalf("GetUserURLs failed: %v", err)
		}
	}
}

func BenchmarkDeleteURLs(b *testing.B) {
	s := createTestService()
	shortURLs := []string{"abc123", "def456", "ghi789"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.DeleteURLs("user-123", shortURLs)
	}
}

func BenchmarkPingDB(b *testing.B) {
	s := createTestService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.PingDB()
	}
}

func BenchmarkAllServiceFunctions(b *testing.B) {
	s := createTestService()
	originalURL := "https://www.example.com/very/long/url/that/needs/to/be/shortened"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortKey, err := s.Shorten(originalURL, "req-"+string(rune(i)), "user-123")
		if err != nil {
			b.Fatalf("Shorten failed: %v", err)
		}

		_, err = s.GetOriginal(shortKey)
		if err != nil {
			b.Fatalf("GetOriginal failed: %v", err)
		}

		_ = s.generateShortKey(originalURL)

		_, err = s.GetUserURLs("user-123")
		if err != nil {
			b.Fatalf("GetUserURLs failed: %v", err)
		}

		_ = s.PingDB()
	}
}
