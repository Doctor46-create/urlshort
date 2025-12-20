// Package server provides the entry point for starting the URL shortening HTTP server.
package server

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/Doctor46-create/urlshort/internal/config"
)

func Execute() {
	cfg := config.GetConfig()

	app := NewApplication(cfg)
	app.Init()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		app.Run()
	}()

	sig := <-stop

	app.log.Info("Received signal, shutting down...", zap.String("signal", sig.String()))

	shutdownTimeout := 5 * time.Second
	done := make(chan struct{})

	go func() {
		app.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		app.log.Info("Server shutdown completed")
	case <-time.After(shutdownTimeout):
		app.log.Warn("Server shutdown timed out, exiting forcefully")
	}
}
