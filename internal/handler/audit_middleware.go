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

func wrapResponseWriter(w http.ResponseWriter) *auditResponseWriter {
	return &auditResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           bytes.Buffer{},
	}
}

func AuditMiddleware(subject *audit.Subject, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if subject == nil {
			logger.Debug("Audit disabled, skipping audit middleware")
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r, originalURL := prepareRequestForAudit(r, logger)
			arw := wrapResponseWriter(w)

			next.ServeHTTP(arw, r)

			processAuditAfterResponse(r, arw.statusCode, originalURL, subject, logger)
		})
	}
}

func shouldExtractOriginalURL(r *http.Request) bool {
	return r.Method == http.MethodPost &&
		(r.URL.Path == "/" || r.URL.Path == "/api/shorten")
}

func prepareRequestForAudit(r *http.Request, logger *zap.Logger) (*http.Request, string) {
	if !shouldExtractOriginalURL(r) {
		return r, ""
	}

	originalURL := extractOriginalURL(r, logger)
	if originalURL == "" {
		return r, ""
	}

	r = restoreRequestBody(
		r,
		originalURL,
		r.URL.Path == "/api/shorten",
		logger,
	)

	ctx := context.WithValue(r.Context(), auditOriginalURLKey, originalURL)
	return r.WithContext(ctx), originalURL
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

func restoreRequestBody(
	r *http.Request,
	originalURL string,
	isJSON bool,
	logger *zap.Logger,
) *http.Request {
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

func shouldAudit(r *http.Request, statusCode int) bool {
	if statusCode < 200 || statusCode >= 300 {
		return false
	}

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		return true
	case r.Method == http.MethodPost && r.URL.Path == "/api/shorten":
		return true
	case r.Method == http.MethodGet &&
		len(r.URL.Path) > 1 &&
		r.URL.Path != "/ping" &&
		r.URL.Path != "/api/user/urls" &&
		!strings.HasPrefix(r.URL.Path, "/api/"):
		return true
	default:
		return false
	}
}

func processAuditAfterResponse(
	r *http.Request,
	statusCode int,
	originalURL string,
	subject *audit.Subject,
	logger *zap.Logger,
) {
	if !shouldAudit(r, statusCode) {
		return
	}

	event := buildAuditEvent(r, originalURL)
	if event == nil {
		return
	}

	sendAuditEventAsync(event, subject, logger, statusCode)
}

func buildAuditEvent(r *http.Request, originalURL string) *audit.AuditEvent {
	var (
		action audit.Action
		url    string
	)

	switch {
	case r.Method == http.MethodPost &&
		(r.URL.Path == "/" || r.URL.Path == "/api/shorten"):
		action = audit.ActionShorten
		url = originalURL

	case r.Method == http.MethodGet && len(r.URL.Path) > 1:
		action = audit.ActionFollow
		url = r.URL.String()

	default:
		return nil
	}

	userID := ""
	if id, ok := r.Context().Value(userIDKey).(string); ok {
		userID = id
	}

	return audit.NewEvent(action, userID, url)
}

func sendAuditEventAsync(
	event *audit.AuditEvent,
	subject *audit.Subject,
	logger *zap.Logger,
	statusCode int,
) {
	go func() {
		subject.NotifyAll(event)

		logger.Debug("Audit event sent",
			zap.String("action", string(event.Action)),
			zap.String("user_id", event.UserID),
			zap.String("url", event.URL),
			zap.Int("status_code", statusCode),
		)
	}()
}

func GetOriginalURLFromContext(ctx context.Context) string {
	if url, ok := ctx.Value(auditOriginalURLKey).(string); ok {
		return url
	}
	return ""
}
