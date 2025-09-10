package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
	"slices"

	"go.uber.org/zap"
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
	if (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) && g.compressionFlag {
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
