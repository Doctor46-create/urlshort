package handler

import (
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/service"
	"go.uber.org/zap"
)

type urlHandler struct {
	srvc   service.Shortener
	cfg    config.ServiceConfig
	logger *zap.Logger
	db     *db.DBConfig
}

func NewURLHandler(srvc service.Shortener, cfg config.ServiceConfig, logger *zap.Logger, dbConfig *db.DBConfig) URLHandler {
	return &urlHandler{
		srvc:   srvc,
		cfg:    cfg,
		logger: logger,
		db:     dbConfig,
	}
}
