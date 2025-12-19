package server

import (
	defaultLogger "log"

	"github.com/Doctor46-create/urlshort/internal/logger"
	"go.uber.org/zap"
)

func (a *Application) initLogger() {
	log, err := logger.NewLogger(a.cfg)
	if err != nil {
		defaultLogger.Fatalf("Failed to create logger: %v", err)
	}
	a.log = log
}

func (a *Application) logStartupInfo() {
	a.log.Info("Starting URL shortener service",
		zap.String("address", a.cfg.GetAddress()),
		zap.String("base_url", a.cfg.GetBaseURL()),
		zap.String("file_storage_path", a.cfg.GetFileStoragePath()),
		zap.String("database_dsn", a.cfg.GetDSN()),
	)
}
