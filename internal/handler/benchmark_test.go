package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Doctor46-create/urlshort/internal/audit"
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/model"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type mockService struct {
	service.Shortener
}

func (m *mockService) Shorten(originalURL, requestID, userID string) (string, error) {
	return "abc123", nil
}

func (m *mockService) GetOriginal(shortKey string) (string, error) {
	return "https://example.com/original", nil
}

func (m *mockService) ShortenBatch(items []model.BatchRequestItem, requestID, userID string) ([]model.BatchResponseItem, error) {
	result := make([]model.BatchResponseItem, len(items))
	for i, item := range items {
		result[i] = model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "abc123",
		}
	}
	return result, nil
}

func (m *mockService) GetUserURLs(userID string) ([]model.UserURL, error) {
	return []model.UserURL{
		{ShortURL: "abc123", OriginalURL: "https://example.com/original"},
	}, nil
}

func (m *mockService) DeleteURLs(userID string, shortURLs []string) {
}

func (m *mockService) PingDB() error {
	return nil
}

func createTestHandler() *Handler {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		App: config.App{
			Address:         ":8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "short_urls.json",
		},
		Logger: config.Logger{
			LogLevel: 0,
		},
		DB: config.DB{
			DSN: "",
		},
		Audit: config.Audit{
			AuditFile: "",
			AuditURL:  "",
		},
	}
	mockSrvc := &mockService{}
	dbConfig := &db.DBConfig{}
	auditSubject := &audit.Subject{}

	handler := NewHandler(mockSrvc, cfg, logger, dbConfig, auditSubject)

	return handler
}

func BenchmarkShortenURL(b *testing.B) {
	h := createTestHandler()

	originalURL := "https://www.example.com/very/long/url/that/needs/to/be/shortened"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(originalURL)))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		req = req.WithContext(ctx)

		h.urlHandler.ShortenURL(w, req)
	}
}

func BenchmarkShortenURLJSON(b *testing.B) {
	h := createTestHandler()

	request := model.JSONRequest{
		URL: "https://www.example.com/very/long/url/that/needs/to/be/shortened",
	}
	jsonData, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		req = req.WithContext(ctx)

		h.urlHandler.ShortenURLJSON(w, req)
	}
}

func BenchmarkRedirectURL(b *testing.B) {
	h := createTestHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/abc123", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("shortKey", "abc123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.urlHandler.RedirectURL(w, req)
	}
}

func BenchmarkShortenBatchURL(b *testing.B) {
	h := createTestHandler()

	batchRequests := []model.BatchRequestItem{
		{
			CorrelationID: "corr1",
			OriginalURL:   "https://www.example.com/first",
		},
		{
			CorrelationID: "corr2",
			OriginalURL:   "https://www.example.com/second",
		},
		{
			CorrelationID: "corr3",
			OriginalURL:   "https://www.example.com/third",
		},
	}
	jsonData, _ := json.Marshal(batchRequests)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		req = req.WithContext(ctx)

		h.urlHandler.ShortenBatchURL(w, req)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	h := createTestHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/user/urls", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		ctx = context.WithValue(ctx, hadCookieKey, true)
		ctx = context.WithValue(ctx, cookieWasValidKey, true)
		req = req.WithContext(ctx)

		h.urlHandler.GetUserURLs(w, req)
	}
}

func BenchmarkDeleteURLs(b *testing.B) {
	h := createTestHandler()

	deleteRequests := []string{"abc123", "def456", "ghi789"}
	jsonData, _ := json.Marshal(deleteRequests)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		req = req.WithContext(ctx)

		h.urlHandler.DeleteURLs(w, req)
	}
}

func BenchmarkPingDB(b *testing.B) {
	h := createTestHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()
		h.urlHandler.PingDB(w, req)
	}
}

func BenchmarkAllEndpoints(b *testing.B) {
	h := createTestHandler()

	originalURL := "https://www.example.com/very/long/url/that/needs/to/be/shortened"
	request := model.JSONRequest{URL: originalURL}
	jsonData, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(originalURL)))
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), userIDKey, "test-user-id")
		req = req.WithContext(ctx)
		h.urlHandler.ShortenURL(w, req)

		req = httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		req = req.WithContext(ctx)
		h.urlHandler.ShortenURLJSON(w, req)

		req = httptest.NewRequest("GET", "/abc123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("shortKey", "abc123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w = httptest.NewRecorder()
		h.urlHandler.RedirectURL(w, req)

		req = httptest.NewRequest("GET", "/api/user/urls", nil)
		ctx = context.WithValue(req.Context(), userIDKey, "test-user-id")
		ctx = context.WithValue(ctx, hadCookieKey, true)
		ctx = context.WithValue(ctx, cookieWasValidKey, true)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		h.urlHandler.GetUserURLs(w, req)
	}
}
