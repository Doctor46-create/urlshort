package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	
	if strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") {
		return g.zw.Write(b)
	}
	
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) Close() error {
	return g.zw.Close()
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func CompressionMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
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
			if !strings.Contains(acceptEncoding, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			zw := gzipWriterPool.Get().(*gzip.Writer)
			defer func() {
				zw.Reset(nil)
				gzipWriterPool.Put(zw)
			}()

			zw.Reset(w)
			
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				zw:             zw,
			}
			defer gzw.Close()

			next.ServeHTTP(gzw, r)
		})
	}
}
