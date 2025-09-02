package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
//	"sync"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	zw              *gzip.Writer
	compressionFlag bool
}
func (g *gzipWriter) Write(b []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		if !g.compressionFlag {
			g.Header().Set("Content-Encoding", "gzip")
			g.compressionFlag = true
		}
		return g.zw.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) Close() error {
	if g.compressionFlag {
		return g.zw.Close()
	}
	return nil
}

//var gzipWriterPool = sync.Pool{
//	New: func() any {
//		return gzip.NewWriter(nil)
//	},
//}

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
				r.Body = gr
				defer gr.Close()
			}

//			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := (strings.Contains(strings.ToLower(acceptEncoding), "gzip") && acceptEncoding != "")
			if supportsGzip {
//				zw := gzipWriterPool.Get().(*gzip.Writer)
//				zw.Reset(w)

				gzw := &gzipWriter{
					ResponseWriter:  w,
					zw:              gzip.NewWriter(w),
					compressionFlag: false,
				}
				w = gzw
				defer gzw.Close()
//				defer func() {
//					if gzw, ok := ow.(*gzipWriter); ok {
//						gzw.Close()
//						zw.Reset(nil)
//						gzipWriterPool.Put(zw)
//					}
//				}()
			}

			next.ServeHTTP(w, r)
		})
	}
}
