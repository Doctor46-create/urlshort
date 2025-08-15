package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/repository"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/handler"
)

func Execute() {
	cfg := config.GetConfig()
	repo := repository.NewURLRepository()
	srvc := service.NewURLService(repo)
	handler := handler.NewHandler(srvc, cfg)

	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: handler.InitRouter(),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
