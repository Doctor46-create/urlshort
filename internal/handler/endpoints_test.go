package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/repository/memory"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testConfig() config.ServiceConfig {
	return &config.Config{
		App: config.App{
			Address: ":8080",
			BaseURL: "http://localhost:8080",
		},
	}
}

func testLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	logger, _ := config.Build()
	return logger
}

func testDBConfig() *db.DBConfig {
	return &db.DBConfig{
		DSN: "",
	}
}

func testRepo() repository.URLRepository {
	return memory.NewURLRepository()
}

func TestPostHandler(t *testing.T) {
	cfg := testConfig()
	logger := testLogger()
	defer logger.Sync()
	dbConfig := testDBConfig()

	tests := []struct {
		name        string
		method      string
		body        string
		wantStatus  int
		wantContain string
	}{
		{
			name:        "SuccessfulPost",
			method:      http.MethodPost,
			body:        "www.google.com",
			wantStatus:  http.StatusCreated,
			wantContain: cfg.GetBaseURL(),
		},
		{
			name:       "Wrong method",
			method:     http.MethodGet,
			body:       "www.google.com",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testRepo()
			srvc := service.NewURLService(repo)
			h := NewHandler(srvc, cfg, logger, dbConfig, nil)
			router := h.InitRouter(logger)

			req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantContain != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContain)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	cfg := testConfig()
	logger := testLogger()
	defer logger.Sync()
	dbConfig := testDBConfig()

	repo := testRepo()
	srvc := service.NewURLService(repo)
	originalURL := "www.google.com"
	shortKey, _ := srvc.Shorten(originalURL, "", "46")

	tests := []struct {
		name       string
		method     string
		path       string
		setup      func(service.Shortener)
		wantStatus int
		wantBody   string
		wantLoc    string
	}{
		{
			name:   "successful redirect",
			method: http.MethodGet,
			path:   "/" + shortKey,
			setup: func(s service.Shortener) {
				s.Shorten(originalURL, "", "46")
			},
			wantStatus: http.StatusTemporaryRedirect,
			wantLoc:    originalURL,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			path:       "/nonexistent",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not found\n",
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/" + shortKey,
			wantStatus: http.StatusNotFound,
			wantBody:   "Not found\n",
		},
		{
			name:       "root path",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusNotFound,
			wantBody:   "Not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := testRepo()
			testSrvc := service.NewURLService(testRepo)
			testH := NewURLHandler(testSrvc, cfg, logger, dbConfig, nil)

			if tt.setup != nil {
				tt.setup(testSrvc)
			}

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/{shortKey}", testH.RedirectURL)
			r.Get("/", testH.RedirectURL)
			r.Post("/{shortKey}", testH.RedirectURL)

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
			if tt.wantLoc != "" {
				assert.Equal(t, tt.wantLoc, rr.Header().Get("Location"))
			}
		})
	}
}

func TestShortenURLJSONHandler(t *testing.T) {
	cfg := testConfig()
	logger := testLogger()
	defer logger.Sync()
	dbConfig := testDBConfig()

	tests := []struct {
		name        string
		method      string
		body        string
		wantStatus  int
		wantContain string
	}{
		{
			name:        "Successful JSON shorten",
			method:      http.MethodPost,
			body:        `{"url":"https://example.com"}`,
			wantStatus:  http.StatusCreated,
			wantContain: cfg.GetBaseURL(),
		},
		{
			name:       "Invalid JSON",
			method:     http.MethodPost,
			body:       `invalid json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty URL in JSON",
			method:     http.MethodPost,
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Wrong method for JSON",
			method:     http.MethodGet,
			body:       `{"url":"https://example.com"}`,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testRepo()
			srvc := service.NewURLService(repo)
			h := NewHandler(srvc, cfg, logger, dbConfig, nil)
			router := h.InitRouter(logger)

			req, err := http.NewRequest(tt.method, "/api/shorten", bytes.NewBufferString(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantContain != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContain)
			}
		})
	}
}

func TestPingDBHandler(t *testing.T) {
	cfg := testConfig()
	logger := testLogger()
	defer logger.Sync()

	tests := []struct {
		name       string
		method     string
		dbConfig   *db.DBConfig
		wantStatus int
	}{
		{
			name:       "Successful ping with empty DSN",
			method:     http.MethodGet,
			dbConfig:   &db.DBConfig{DSN: ""},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Wrong method",
			method:     http.MethodPost,
			dbConfig:   &db.DBConfig{DSN: ""},
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testRepo()
			srvc := service.NewURLService(repo)
			h := NewHandler(srvc, cfg, logger, tt.dbConfig, nil)
			router := h.InitRouter(logger)

			req, err := http.NewRequest(tt.method, "/ping", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
