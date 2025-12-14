package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Doctor46-create/urlshort/internal/audit"
	"go.uber.org/zap"
)

const (
	auditOriginalURLKey contextKey = "audit_original_url"
)

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (arw *auditResponseWriter) WriteHeader(code int) {
	arw.statusCode = code
	arw.ResponseWriter.WriteHeader(code)
}

func (arw *auditResponseWriter) Write(b []byte) (int, error) {
	arw.body.Write(b)
	return arw.ResponseWriter.Write(b)
}

func AuditMiddleware(subject *audit.Subject, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if subject == nil {
			logger.Debug("Audit disabled, skipping audit middleware")
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var originalURL string
			if r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "/api/shorten") {
				originalURL = extractOriginalURL(r, logger)
				if originalURL != "" {
					r = restoreRequestBody(r, originalURL, r.URL.Path == "/api/shorten", logger)
				}
			}

			arw := &auditResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.Buffer{},
			}

			ctx := r.Context()
			if originalURL != "" {
				ctx = context.WithValue(ctx, auditOriginalURLKey, originalURL)
			}

			next.ServeHTTP(arw, r.WithContext(ctx))

			if shouldAudit(r, arw.statusCode) {
				go sendAuditEvent(r, arw.statusCode, originalURL, subject, logger)
			}
		})
	}
}

func shouldAudit(r *http.Request, statusCode int) bool {
	if statusCode < 200 || statusCode >= 300 {
		return false
	}

	switch {
	case r.URL.Path == "/" && r.Method == http.MethodPost:
		return true
	case r.URL.Path == "/api/shorten" && r.Method == http.MethodPost:
		return true
	case len(r.URL.Path) > 1 && r.Method == http.MethodGet && r.URL.Path != "/ping" && 
	     r.URL.Path != "/api/user/urls" && !strings.HasPrefix(r.URL.Path, "/api/"):
		return true
	default:
		return false
	}
}

func extractOriginalURL(r *http.Request, logger *zap.Logger) string {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Debug("Failed to read request body for audit", zap.Error(err))
		return ""
	}
	defer r.Body.Close()

	if r.URL.Path == "/api/shorten" {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			logger.Debug("Failed to unmarshal JSON body for audit", zap.Error(err))
			return string(bodyBytes)
		}
		return req.URL
	}

	return string(bodyBytes)
}

func restoreRequestBody(r *http.Request, originalURL string, isJSON bool, logger *zap.Logger) *http.Request {
	var body []byte
	if isJSON {
		jsonBody := map[string]string{"url": originalURL}
		var err error
		body, err = json.Marshal(jsonBody)
		if err != nil {
			logger.Error("Failed to marshal JSON for request restore", zap.Error(err))
			body = []byte(originalURL)
		}
	} else {
		body = []byte(originalURL)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	r.ContentLength = int64(len(body))
	return r
}

func sendAuditEvent(r *http.Request, statusCode int, originalURL string, subject *audit.Subject, logger *zap.Logger) {
	var action audit.Action
	var url string

	switch {
	case r.URL.Path == "/" && r.Method == http.MethodPost:
		action = audit.ActionShorten
		url = originalURL
	case r.URL.Path == "/api/shorten" && r.Method == http.MethodPost:
		action = audit.ActionShorten
		url = originalURL
	case len(r.URL.Path) > 1 && r.Method == http.MethodGet:
		action = audit.ActionFollow
		url = r.URL.String()
	default:
		return
	}

	userID := ""
	if userIDCtx := r.Context().Value(userIDKey); userIDCtx != nil {
		if id, ok := userIDCtx.(string); ok {
			userID = id
		}
	}

	if url != "" {
		event := audit.NewEvent(action, userID, url)
		subject.NotifyAll(event)
		
		logger.Debug("Audit event sent",
			zap.String("action", string(action)),
			zap.String("user_id", userID),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
		)
	}
}

func GetOriginalURLFromContext(ctx context.Context) string {
	if url, ok := ctx.Value(auditOriginalURLKey).(string); ok {
		return url
	}
	return ""
}
