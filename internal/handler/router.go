// Package handler
package handler

import (
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	mylogger "github.com/Doctor46-create/urlshort/internal/logger"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	urlHandler URLHandler
	logger     *zap.Logger
}

func NewHandler(srvc service.Shortener, cfg config.ServiceConfig, logger *zap.Logger, dbConfig *db.DBConfig) *Handler {
	return &Handler{
		urlHandler: NewURLHandler(srvc, cfg, logger, dbConfig),
		logger:     logger,
	}
}

func (h *Handler) InitRouter(logger *zap.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(mylogger.NewLoggerMiddleware(logger))
	r.Use(CompressionMiddleware(logger))
	r.Use(AuthMiddleware(logger))

	r.With(PostOnly(logger)).Post("/", h.urlHandler.ShortenURL)
	r.With(PostOnly(logger)).Post("/api/shorten", h.urlHandler.ShortenURLJSON)
	r.With(PostOnly(logger)).Post("/api/shorten/batch", h.urlHandler.ShortenBatchURL)

	r.With(GetOnly(logger)).Get("/{shortKey}", h.urlHandler.RedirectURL)
	r.With(GetOnly(logger)).Get("/ping", h.urlHandler.PingDB)
	r.With(GetOnly(logger)).Get("/api/user/urls", h.urlHandler.GetUserURLs)
	return r
}
