package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Doctor46-create/urlshort/internal/service"
)

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		wantStatus     int
	}{
		{
			name:       "SuccessfulPost",
			method:     http.MethodPost,
			body:       "www.google.com",
			wantStatus: http.StatusCreated,
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
			req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			_handler := PostHandler(service.NewURLService())
			_handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")

		})
	}
}

func TestGetHandler(t *testing.T) {
	svc := service.NewURLService()
	originalURL := "www.google.com"
	shortKey, _ := svc.Shorten(originalURL) 

	tests := []struct {
		name        string
		method      string
		path        string
		preloadURL  string 
		wantStatus  int
		wantBody    string
	}{
		{
			name:       "successful redirect",
			method:     http.MethodGet,
			path:       "/" + shortKey,
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			path:       "/nonexistent",
			preloadURL: originalURL,
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
			testSvc := service.NewURLService()
			
			if tt.preloadURL != "" {
				testSvc.Shorten(tt.preloadURL)
			} else if tt.path != "/" && tt.path != "/nonexistent" {
				testSvc.Shorten(originalURL)
			}

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := GetHandler(testSvc)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String(), "response body mismatch")
			}
		})
	}
}
