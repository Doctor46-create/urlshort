package handler

import (
	"compress/gzip"
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/auth"

)

type gzipWriter struct {
	http.ResponseWriter
	zw              *gzip.Writer
	compressionFlag bool
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	if g.compressionFlag {
		return g.zw.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	contentType := g.Header().Get("Content-Type")

	if g.compressionFlag &&
		(strings.HasPrefix(contentType, "application/json") ||
			strings.HasPrefix(contentType, "text/html")) {

		g.Header().Set("Content-Encoding", "gzip")
	}

	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipWriter) Close() error {
	if g.compressionFlag {
		return g.zw.Close()
	}
	return nil
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

type contextKey string

const (
	userIDKey         contextKey = "user_id"
	cookieWasValidKey contextKey = "cookie_was_valid"
	hadCookieKey      contextKey = "had_cookie"
)

func CompressionMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(strings.ToLower(contentEncoding), "gzip")
			if sendsGzip {
				gr, err := gzip.NewReader(r.Body)
				if err != nil {
					logger.Error("Failed to create gzip reader", zap.Error(err))
					http.Error(w, "Bad Request: Invalid gzip encoding", http.StatusBadRequest)
					return
				}
				defer gr.Close()
				r.Body = gr
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(strings.ToLower(acceptEncoding), "gzip") && acceptEncoding != ""

			if supportsGzip {
				zw := gzipWriterPool.Get().(*gzip.Writer)
				zw.Reset(w)

				gzw := &gzipWriter{
					ResponseWriter:  w,
					zw:              zw,
					compressionFlag: false,
				}
				defer func() {
					gzw.Close()
					zw.Reset(nil)
					gzipWriterPool.Put(zw)
				}()
				w = gzw
			}

			next.ServeHTTP(w, r)
		})
	}
}

func MethodAllowed(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(methods, r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		})
	}
}

func PostOnly(logger *zap.Logger) func(http.Handler) http.Handler {
	return MethodAllowed(http.MethodPost)
}

func GetOnly(logger *zap.Logger) func(http.Handler) http.Handler {
	return MethodAllowed(http.MethodGet)
}

func AuthMiddleware(logger *zap.Logger, cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				userID    string
				isValid   bool
				hadCookie bool
			)

			cookie, err := r.Cookie("user_id")
			hadCookie = err == nil && cookie != nil && cookie.Value != ""

			if !hadCookie {
				userID = uuid.New().String()
				isValid = false
			} else {
				userID, isValid = auth.ValidateSignedUserID(cookie.Value, cfg.SecretKey)
				if !isValid {
					userID = uuid.New().String()
				}
			}

			signedValue := auth.SignUserID(userID, cfg.SecretKey)

			logger.Info("Setting cookie",
				zap.String("cookie_name", "user_id"),
				zap.String("cookie_value", signedValue),
			)

			http.SetCookie(w, &http.Cookie{
				Name:     "user_id",
				Value:    signedValue,
				MaxAge:   3600 * 24 * 30,
				Path:     "/",
				Secure:   false,
				HttpOnly: true,
			})

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, userID)
			ctx = context.WithValue(ctx, cookieWasValidKey, isValid)
			ctx = context.WithValue(ctx, hadCookieKey, hadCookie)

			logger.Info("Context set",
				zap.String("user_id", userID),
				zap.Bool("cookie_was_valid", isValid),
				zap.Bool("had_cookie", hadCookie),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

