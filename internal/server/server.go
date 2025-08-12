package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/transport/router"
)

func Execute() {
	cfg := config.GetConfig()
	repo := repository.NewURLRepository()
	srvc := service.NewURLService(repo)
	handler := router.NewHandler(srvc, cfg)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: handler.InitRouter(),
	}

	server.ListenAndServe()
}
