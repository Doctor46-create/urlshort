package handler

import (
	"net/http"
	"net/http/pprof"

	"github.com/Doctor46-create/urlshort/internal/audit"
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	mylogger "github.com/Doctor46-create/urlshort/internal/logger"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	urlHandler   URLHandler
	logger       *zap.Logger
	auditSubject *audit.Subject
}

func NewHandler(srvc service.Shortener, cfg config.ServiceConfig, logger *zap.Logger, dbConfig *db.DBConfig, auditSubject *audit.Subject) *Handler {
	return &Handler{
		urlHandler: NewURLHandler(srvc, cfg, logger, dbConfig, auditSubject),
		logger:     logger,
	}
}

func (h *Handler) InitRouter(logger *zap.Logger, cfg config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(mylogger.NewLoggerMiddleware(logger))
	r.Use(CompressionMiddleware(logger))
	r.Use(AuthMiddleware(logger, cfg))
	r.Use(AuditMiddleware(h.auditSubject, logger))

	r.With(PostOnly(logger)).Post("/", h.urlHandler.ShortenURL)
	r.With(PostOnly(logger)).Post("/api/shorten", h.urlHandler.ShortenURLJSON)
	r.With(PostOnly(logger)).Post("/api/shorten/batch", h.urlHandler.ShortenBatchURL)

	r.With(GetOnly(logger)).Get("/{shortKey}", h.urlHandler.RedirectURL)
	r.With(GetOnly(logger)).Get("/ping", h.urlHandler.PingDB)
	r.With(GetOnly(logger)).Get("/api/user/urls", h.urlHandler.GetUserURLs)
	r.Delete("/api/user/urls", h.urlHandler.DeleteURLs)

	if h.enablePProf() {
		r.Mount("/debug", h.pprofRouter())
		h.logger.Info("pprof profiling enabled at /debug/pprof/")
	}
	return r
}

func (h *Handler) pprofRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(h.pprofAuthMiddleware)

	r.Get("/pprof", pprof.Index)

	r.Get("/pprof/cmdline", pprof.Cmdline)
	r.Get("/pprof/profile", pprof.Profile)
	r.Get("/pprof/symbol", pprof.Symbol)
	r.Get("/pprof/trace", pprof.Trace)

	r.Get("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	r.Get("/pprof/block", pprof.Handler("block").ServeHTTP)
	r.Get("/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/pprof/allocs", pprof.Handler("allocs").ServeHTTP)

	return r
}

func (h *Handler) enablePProf() bool {
	return true
}

func (h *Handler) pprofAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) GetPProfRouter() chi.Router {
	return h.pprofRouter()
}
