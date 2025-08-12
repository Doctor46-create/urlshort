package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Address: ":8080",
			BaseURL: "http://localhost:8080",
		},
	}
}

func TestPostHandler(t *testing.T) {
	cfg := testConfig()
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
			wantContain: cfg.BaseURL,
		},
		{
			name:       "Wrong method",
			method:     http.MethodGet,
			body:       "www.google.com",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewURLRepository()
			srvc := service.NewURLService(repo)
			h := NewURLHandler(srvc, cfg)

			req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			h.ShortenURL(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantContain != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContain)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	cfg := testConfig()
	repo := repository.NewURLRepository()
	srvc := service.NewURLService(repo)
	originalURL := "www.google.com"
	shortKey, _ := srvc.Shorten(originalURL)

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
				s.Shorten(originalURL)
			},
			wantStatus: http.StatusTemporaryRedirect,
			wantLoc:    originalURL,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			path:       "/nonexistent",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Not found\n",
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/" + shortKey,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Method not allowed\n",
		},
		{
			name:       "root path",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := repository.NewURLRepository()
			testSrvc := service.NewURLService(testRepo)
			testH := NewURLHandler(testSrvc, cfg)

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
