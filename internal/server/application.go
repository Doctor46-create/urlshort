package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Doctor46-create/urlshort/internal/audit"
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"go.uber.org/zap"
)

type Application struct {
	cfg *config.Config
	log *zap.Logger

	audit *audit.Subject

	repo     repository.URLRepository
	dbConfig *db.DBConfig

	httpServer *http.Server
}

func NewApplication(cfg *config.Config) *Application {
	return &Application{
		cfg: cfg,
	}
}

func (a *Application) Init() {
	a.initLogger()
	a.logStartupInfo()

	a.initAudit()
	a.initRepository()
	a.initHTTPServer()
}

func (a *Application) Run() {
	a.log.Info("Server started",
		zap.String("addr", a.cfg.GetAddress()),
	)

	if err := a.httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		a.log.Fatal("Server failed", zap.Error(err))
	}
}

func (a *Application) Shutdown() {
	a.log.Info("Application shutting down")

	if a.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.log.Error("HTTP server shutdown failed", zap.Error(err))
		}
	}

	if shutdowner, ok := a.repo.(repository.Shutdowner); ok {
		a.log.Info("Shutting down repository")
		shutdowner.Shutdown()
	}

	if a.audit != nil {
		a.log.Info("Shutting down audit")
		a.audit.CloseAll()
	}

	if a.log != nil {
		_ = a.log.Sync()
	}
}
