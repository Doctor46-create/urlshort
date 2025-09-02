package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	zw              *gzip.Writer
	CompressionFlag bool
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")

	if strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") {
		if !g.CompressionFlag {
			g.Header().Set("Content-Encoding", "gzip")
			g.CompressionFlag = true
		}
		return g.zw.Write(b)
	}

	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) Close() error {
	if g.CompressionFlag {
		return g.zw.Close()
	}
	return nil
}

type gzipReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}

func (g gzipReader) Read(p []byte) (int, error) {
	return g.zr.Read(p)
}

func (g gzipReader) Close() error {
	if err := g.zr.Close(); err != nil {
		return err
	}
	return g.ReadCloser.Close()
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func CompressionMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(strings.ToLower(contentEncoding), "gzip")
			if sendsGzip {
				gr, err := newGzipReader(r.Body)
				if err != nil {
					logger.Error("Failed to create gzip reader", zap.Error(err))
					http.Error(w, "Bad Request: Invalid gzip encoding", http.StatusBadRequest)
					return
				}
				defer gr.Close()
				r.Body = gr
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(strings.ToLower(acceptEncoding), "gzip")
			if supportsGzip {
				zw := gzipWriterPool.Get().(*gzip.Writer)
				zw.Reset(w)

				gzw := &gzipWriter{
					ResponseWriter: w,
					zw:             zw,
					CompressionFlag: false,
				}
				ow = gzw

				defer func() {
					if gzw, ok := ow.(*gzipWriter); ok {
						gzw.Close()
						zw.Reset(nil)
						gzipWriterPool.Put(zw)
					}
				}()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
