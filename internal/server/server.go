// Package server provides the entry point for starting the URL shortening HTTP server.
package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Doctor46-create/urlshort/internal/config"
)

// Execute is the entry point of the server application.
//
// It initializes the application configuration and all required
// dependencies, starts the HTTP server in a separate goroutine,
// and blocks until an OS termination signal is received.
//
// The function listens for SIGINT and SIGTERM signals and performs
// a graceful shutdown when one of them is caught. During shutdown,
// all application resources (HTTP server, repositories, background
// workers, audit services, etc.) are properly released.
//
// Execute blocks until the application has been fully shut down.
func Execute() {
	cfg := config.GetConfig()

	app := NewApplication(cfg)
	app.Init()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go app.Run()

	<-stop
	app.Shutdown()
}
