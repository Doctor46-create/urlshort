package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/handler"
	"github.com/Doctor46-create/urlshort/internal/logger"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/repository/database"
	"github.com/Doctor46-create/urlshort/internal/repository/filestorage"
	"github.com/Doctor46-create/urlshort/internal/repository/memory"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/migrations"
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
		zap.String("database_dsn", cfg.GetDSN()),
	)

	var repo repository.URLRepository
	var dbConfig *db.DBConfig
	switch {
	case cfg.GetDSN() != "":
		dbConfig := db.NewDBConfig(cfg)
		databaseConn, err := db.NewDatabase(dbConfig)
		if err != nil {
			log.Fatal("Failed to connect to database", zap.Error(err))
		}
		log.Info("Running database migrations...")
		if err := migration.MigrateUp(databaseConn.DB); err != nil {
			log.Fatal("Failed to run migrations", zap.Error(err))
		}
		log.Info("Database migrations completed successfully")
		repo = database.NewURLRepository(databaseConn.DB)
		log.Info("Using database storage")
	case cfg.GetFileStoragePath() != "":
		repo = filestorage.NewURLRepository(cfg.GetFileStoragePath())
		log.Info("Using file storage", zap.String("path", cfg.GetFileStoragePath()))
	default:
		repo = memory.NewURLRepository()
		log.Info("Using in-memory storage")
	}

	srvc := service.NewURLService(repo)
	handler := handler.NewHandler(srvc, cfg, log, dbConfig)

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
