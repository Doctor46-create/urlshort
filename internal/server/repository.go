package server

import (
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/repository/database"
	"github.com/Doctor46-create/urlshort/internal/repository/filestorage"
	"github.com/Doctor46-create/urlshort/internal/repository/memory"
	"github.com/Doctor46-create/urlshort/migrations"
	"go.uber.org/zap"
)

func (a *Application) initRepository() {
	switch {
	case a.cfg.GetDSN() != "":
		a.initDatabaseRepository()

	case a.cfg.GetFileStoragePath() != "":
		a.log.Info("Using file storage",
			zap.String("path", a.cfg.GetFileStoragePath()),
		)
		a.repo = filestorage.NewURLRepository(a.cfg.GetFileStoragePath())

	default:
		a.log.Info("Using in-memory storage")
		a.repo = memory.NewURLRepository()
	}
}

func (a *Application) initDatabaseRepository() {
	dbConfig := db.NewDBConfig(a.cfg)

	conn, err := db.NewDatabase(dbConfig)
	if err != nil {
		a.log.Fatal("Failed to connect to database", zap.Error(err))
	}

	a.log.Info("Running database migrations...")
	if err := migration.MigrateUp(conn.DB); err != nil {
		a.log.Fatal("Failed to run migrations", zap.Error(err))
	}

	a.repo = database.NewURLRepository(conn.DB)
	a.dbConfig = dbConfig

	a.log.Info("Using database storage")
}
