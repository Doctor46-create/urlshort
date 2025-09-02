// Package server
package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/handler"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/pkg/logger"
	"go.uber.org/zap"
)

func Execute() {
	cfg := config.GetConfig()

	log := logger.NewLogger(cfg)
	defer log.Sync()

	log.Info("Starting URL shortener service",
		zap.String("address", cfg.GetAddress()),
		zap.String("base_url", cfg.GetBaseURL()),
	)

	repo := repository.NewURLRepository()
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
