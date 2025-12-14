package server

import (
	defaultLogger "log"
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/audit"
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

	log, logErr := logger.NewLogger(cfg)
	if logErr != nil {
		defaultLogger.Fatalf("Failed to create logger: %v", logErr)
	}
	defer log.Sync()

	log.Info("Starting URL shortener service",
		zap.String("address", cfg.GetAddress()),
		zap.String("base_url", cfg.GetBaseURL()),
		zap.String("file_storage_path", cfg.GetFileStoragePath()),
		zap.String("database_dsn", cfg.GetDSN()),
	)

	var auditSubject *audit.Subject
	if cfg.HasAudit() {
		auditSubject = audit.NewSubject()

		if cfg.GetAuditFile() != "" {
			fileObserver, err := audit.NewFileObserver(cfg.GetAuditFile())
			if err != nil {
				log.Error("Failed to initialize file audit", zap.Error(err))
			} else {
				auditSubject.Register(fileObserver)
				log.Info("File audit enabled", zap.String("path", cfg.GetAuditFile()))
			}
		}

		if cfg.GetAuditURL() != "" {
			httpObserver := audit.NewHTTPObserver(cfg.GetAuditURL())
			auditSubject.Register(httpObserver)
			log.Info("HTTP audit enabled", zap.String("url", cfg.GetAuditURL()))
		}
	} else {
		log.Info("Audit disabled (no audit-file or audit-url configured)")
		auditSubject = nil
	}

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

	if auditSubject != nil {
		go func() {
			<-make(chan struct{})
			log.Info("Shutting down audit...")
			auditSubject.CloseAll()
		}()
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start", zap.Error(err))
	}
}
