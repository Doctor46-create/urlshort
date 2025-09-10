// Package server
package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/handler"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/logger"
	"go.uber.org/zap"
)

func Execute() {
	cfg := config.GetConfig()

	log := logger.NewLogger(cfg)
	defer log.Sync()

	log.Info("Starting URL shortener service",
		zap.String("address", cfg.GetAddress()),
		zap.String("base_url", cfg.GetBaseURL()),
		zap.String("file_storage_path", cfg.GetFileStoragePath()),
	)

	repo := repository.NewURLRepository(cfg.GetFileStoragePath())
	srvc := service.NewURLService(repo)
	handler := handler.NewHandler(srvc, cfg, log) 

	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: handler.InitRouter(log), 
	}

	log.Info("Server started successfully")

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start", zap.Error(err))
	}
}
