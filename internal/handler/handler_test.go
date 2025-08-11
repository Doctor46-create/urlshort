// package handler
//
// import (
//
//	"bytes"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//	"github.com/go-chi/chi/v5"
//
//	"github.com/Doctor46-create/urlshort/internal/service"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//
// )
//
//	func TestPostHandler(t *testing.T) {
//		tests := []struct {
//			name       string
//			method     string
//			body       string
//			wantStatus int
//		}{
//			{
//				name:       "SuccessfulPost",
//				method:     http.MethodPost,
//				body:       "www.google.com",
//				wantStatus: http.StatusCreated,
//			},
//			{
//				name:       "Wrong method",
//				method:     http.MethodGet,
//				body:       "www.google.com",
//				wantStatus: http.StatusBadRequest,
//			},
//		}
//
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
//				require.NoError(t, err)
//
//				rr := httptest.NewRecorder()
//				_handler := PostHandler(service.NewURLService())
//				_handler.ServeHTTP(rr, req)
//
//				assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")
//			})
//		}
//	}
//
//	func TestGetHandler(t *testing.T) {
//		srvc := service.NewURLService()
//		originalURL := "www.google.com"
//		shortKey, _ := srvc.Shorten(originalURL)
//
//		tests := []struct {
//			name       string
//			method     string
//			path       string
//			preloadURL string
//			wantStatus int
//			wantBody   string
//		}{
//			{
//				name:       "successful redirect",
//				method:     http.MethodGet,
//				path:       "/" + shortKey,
//				wantStatus: http.StatusTemporaryRedirect,
//			},
//			{
//				name:       "not found",
//				method:     http.MethodGet,
//				path:       "/nonexistent",
//				preloadURL: originalURL,
//				wantStatus: http.StatusBadRequest,
//				wantBody:   "Not found\n",
//			},
//			{
//				name:       "wrong method",
//				method:     http.MethodPost,
//				path:       "/" + shortKey,
//				wantStatus: http.StatusBadRequest,
//				wantBody:   "Method not allowed\n",
//			},
//			{
//				name:       "root path",
//				method:     http.MethodGet,
//				path:       "/",
//				wantStatus: http.StatusBadRequest,
//				wantBody:   "Not found\n",
//			},
//		}
//
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				testSvc := service.NewURLService()
//
//				if tt.preloadURL != "" {
//					testSvc.Shorten(tt.preloadURL)
//				} else if tt.path != "/" && tt.path != "/nonexistent" {
//					testSvc.Shorten(originalURL)
//				}
//
//				req, err := http.NewRequest(tt.method, tt.path, nil)
//				require.NoError(t, err)
//
//				rr := httptest.NewRecorder()
//
//				r := chi.NewRouter()
//				r.Handle("/{id}", GetHandler(testSvc))
//				r.Handle("/", GetHandler(testSvc))
//
//				r.ServeHTTP(rr, req)
//
//				assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")
//
//				if tt.wantBody != "" {
//					assert.Equal(t, tt.wantBody, rr.Body.String(), "response body mismatch")
//				}
//			})
//		}
//	}
package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
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
			repo := repository.NewURLRepository()
			srvc := service.NewURLService(repo)
			h := NewURLHandler(srvc)

			req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			h.ShortenURL(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")
		})
	}
}

func TestGetHandler(t *testing.T) {
	repo := repository.NewURLRepository()
	srvc := service.NewURLService(repo)
	//	h := NewURLHandler(srvc)

	originalURL := "www.google.com"
	shortKey, _ := srvc.Shorten(originalURL)

	tests := []struct {
		name       string
		method     string
		path       string
		preloadURL string
		wantStatus int
		wantBody   string
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
			testRepo := repository.NewURLRepository()
			testSrvc := service.NewURLService(testRepo)
			testH := NewURLHandler(testSrvc)

			if tt.preloadURL != "" {
				testSrvc.Shorten(tt.preloadURL)
			} else if tt.path != "/" && tt.path != "/nonexistent" {
				testSrvc.Shorten(originalURL)
			}

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/{shortKey}", testH.RedirectURL)
			r.Get("/", testH.RedirectURL)
			r.Post("/{shortKey}", testH.RedirectURL)

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code, "status code mismatch")

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String(), "response body mismatch")
			}

			if tt.wantStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, originalURL, rr.Header().Get("Location"))
			}
		})
	}
}
