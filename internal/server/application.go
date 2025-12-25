package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/Doctor46-create/urlshort/internal/audit"
	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/config/db"
	grpcServer "github.com/Doctor46-create/urlshort/internal/grpc/server"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"go.uber.org/zap"
)

type Application struct {
	cfg *config.Config
	log *zap.Logger

	audit *audit.Subject

	repo     repository.URLRepository
	dbConfig *db.DBConfig

	service service.Shortener

	httpServer *http.Server
	grpcServer *grpcServer.GRPCServer
}

func NewApplication(cfg *config.Config) *Application {
	return &Application{cfg: cfg}
}

func (a *Application) Init() {
	a.initLogger()
	a.logStartupInfo()

	a.validateHTTPS()

	a.initAudit()
	a.initRepository()
	a.initService()

	a.initHTTPServer()
	a.initGRPCServer()
}

func (a *Application) Run() {
	a.log.Info("HTTP server starting",
		zap.String("addr", a.cfg.GetAddress()),
	)

	go func() {
		var err error
		if a.cfg.IsHTTPSEnabled() {
			err = a.httpServer.ListenAndServeTLS(
				a.cfg.GetTLSCertFile(),
				a.cfg.GetTLSKeyFile(),
			)
		} else {
			err = a.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			a.log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	go func() {
		a.log.Info("gRPC server starting",
			zap.String("addr", a.cfg.GetGRPCAddress()),
		)

		if err := a.grpcServer.Start(); err != nil {
			a.log.Fatal("gRPC server failed", zap.Error(err))
		}
	}()
}

func (a *Application) Shutdown() {
	a.log.Info("Application shutting down")

	if a.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.httpServer.Shutdown(ctx)
	}

	if a.grpcServer != nil {
		a.grpcServer.Stop()
	}

	if shutdowner, ok := a.repo.(repository.Shutdowner); ok {
		shutdowner.Shutdown()
	}

	if a.audit != nil {
		a.audit.CloseAll()
	}

	_ = a.log.Sync()
}

func (a *Application) validateHTTPS() {
	if !a.cfg.IsHTTPSEnabled() {
		return
	}

	if _, err := os.Stat(a.cfg.GetTLSCertFile()); err != nil {
		a.log.Fatal("TLS certificate file not found", zap.Error(err))
	}

	if _, err := os.Stat(a.cfg.GetTLSKeyFile()); err != nil {
		a.log.Fatal("TLS key file not found", zap.Error(err))
	}
}
