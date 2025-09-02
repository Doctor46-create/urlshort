package handler

import (
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/service"
	"go.uber.org/zap"
)

type urlHandler struct {
	srvc   service.Shortener
	cfg    config.ServiceConfig
	logger *zap.Logger
}

func NewURLHandler(srvc service.Shortener, cfg config.ServiceConfig, logger *zap.Logger) URLHandler {
	return &urlHandler{
		srvc:   srvc,
		cfg:    cfg,
		logger: logger,
	}
}

