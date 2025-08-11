package server

import (
	"net/http"
	"github.com/Doctor46-create/urlshort/internal/transport/router"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/repository"
)

func Execute() {

	repo := repository.NewURLRepository()
	srvc := service.NewURLService(repo)
	handler := router.NewHandler(srvc)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler.InitRouter(),
	}

	server.ListenAndServe()
}
