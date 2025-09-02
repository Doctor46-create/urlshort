// Package handler
package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/config"
	mylogger "github.com/Doctor46-create/urlshort/pkg/logger" 
	"go.uber.org/zap"
)

type Handler struct {
	urlHandler URLHandler
	logger     *zap.Logger
}

func NewHandler(srvc service.Shortener, cfg config.ServiceConfig, logger *zap.Logger) *Handler {
	return &Handler{
		urlHandler: NewURLHandler(srvc, cfg, logger),
		logger:     logger,
	}
}

func (h *Handler) InitRouter(logger *zap.Logger) chi.Router {
	r := chi.NewRouter()
	
	r.Use(mylogger.NewLoggerMiddleware(logger))
	r.Use(CompressionMiddleware(logger))
	
	r.Post("/", h.urlHandler.ShortenURL)
	r.Post("/api/shorten", h.urlHandler.ShortenURLJSON)
	r.Get("/{shortKey}", h.urlHandler.RedirectURL)
	return r
}
