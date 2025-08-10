package handler

import (
	"io"
	"net/http"
	"strings"
	"github.com/Doctor46-create/urlshort/internal/service"
)

func PostHandler(shortener service.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		shortKey, err := shortener.Shorten(string(body))
		if err != nil {
			http.Error(w, "Couldn't generate short url", http.StatusBadRequest)
			return
		}

		shortURL := "http://localhost:8080/" + shortKey
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func GetHandler(shortener service.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusBadRequest)
			return
		}

		shortKey := strings.TrimPrefix(r.URL.Path, "/")
		if originalURL, exists := shortener.GetOriginal(shortKey); exists {
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "Not found", http.StatusBadRequest)
		}
	}
}
